package internal

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

type eventAttributes map[string]interface{}

func (e eventAttributes) has(key string) bool {
	_, ok := e[key]
	return ok
}

func (e eventAttributes) isString(key string, expected string) error {
	actual, ok := e[key].(string)
	if !ok {
		return fmt.Errorf("key %s is not a string; got type %t with value %v", key, e[key], e[key])
	}

	if actual != expected {
		return fmt.Errorf("key %s has unexpected value: expected=%s; got=%s", key, expected, actual)
	}

	return nil
}

type harvestedTxnEvent struct {
	intrinsics      eventAttributes
	userAttributes  eventAttributes
	agentAttributes eventAttributes
}

func (h *harvestedTxnEvent) UnmarshalJSON(data []byte) error {
	var arr []eventAttributes

	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}

	if len(arr) != 3 {
		return fmt.Errorf("unexpected number of transaction event items: %d", len(arr))
	}

	h.intrinsics = arr[0]
	h.userAttributes = arr[1]
	h.agentAttributes = arr[2]

	return nil
}

func harvestTxnDataEvent(t *TxnData) (*harvestedTxnEvent, error) {
	// Since transaction event JSON is built using string manipulation, we have
	// to do an awkward marshal/unmarshal shuffle to be able to verify the
	// intrinsics.
	js, err := json.Marshal(&t.TxnEvent)
	if err != nil {
		return nil, err
	}

	event := &harvestedTxnEvent{}
	if err := json.Unmarshal(js, event); err != nil {
		return nil, err
	}

	return event, nil
}

// This function implements as close as we can get to the round trip tests in
// the cross agent tests.
func TestCatMap(t *testing.T) {
	var testcases []struct {
		Name                       string            `json:"name"`
		AppName                    string            `json:"appName"`
		TransactionName            string            `json:"transactionName"`
		TransactionGUID            string            `json:"transactionGuid"`
		InboundPayload             []interface{}     `json:"inboundPayload"`
		ExpectedIntrinsicFields    map[string]string `json:"expectedIntrinsicFields"`
		NonExpectedIntrinsicFields []string          `json:"nonExpectedIntrinsicFields"`
		OutboundRequests           []struct {
			OutboundTxnName         string          `json:"outboundTxnName"`
			ExpectedOutboundPayload json.RawMessage `json:"expectedOutboundPayload"`
		} `json:"outboundRequests"`
	}

	err := crossagent.ReadJSON("cat/cat_map.json", &testcases)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		// Fake enough transaction data to run the test.
		tr := &TxnData{
			Name: tc.TransactionName,
		}

		tr.CrossProcess.Init(true, &ConnectReply{
			CrossProcessID:  "1#1",
			EncodingKey:     "foo",
			TrustedAccounts: map[int]struct{}{1: struct{}{}},
		}, CrossProcessMetadata{})

		// Marshal the inbound payload into JSON for easier testing.
		txnData, err := json.Marshal(tc.InboundPayload)
		if err != nil {
			t.Errorf("%s: error marshalling inbound payload: %v", tc.Name, err)
		}

		// Set up the GUID.
		if tc.TransactionGUID != "" {
			tr.CrossProcess.GUID = tc.TransactionGUID
		}

		// Swallow errors, since some of these tests are testing the behaviour when
		// erroneous headers are provided.
		tr.CrossProcess.handleInboundRequestTxnData(txnData)

		// Simulate outbound requests.
		for _, req := range tc.OutboundRequests {
			metadata, err := tr.CrossProcess.CreateCrossProcessMetadata(req.OutboundTxnName, tc.AppName)
			if err != nil {
				t.Errorf("%s: error creating outbound request headers: %v", tc.Name, err)
			}

			// Grab and deobfuscate the txndata that would have been sent to the
			// external service.
			txnData, err := deobfuscate(metadata.TxnData, tr.CrossProcess.EncodingKey)
			if err != nil {
				t.Errorf("%s: error deobfuscating outbound request header: %v", tc.Name, err)
			}

			// Check the JSON against the expected value.
			compacted := CompactJSONString(string(txnData))
			expected := CompactJSONString(string(req.ExpectedOutboundPayload))
			if compacted != expected {
				t.Errorf("%s: outbound metadata does not match expected value: expected=%s; got=%s", tc.Name, expected, compacted)
			}
		}

		// Finalise the transaction, ignoring errors.
		tr.CrossProcess.Finalise(tc.TransactionName, tc.AppName)

		// Harvest the event.
		event, err := harvestTxnDataEvent(tr)
		if err != nil {
			t.Errorf("%s: error harvesting event data: %v", tc.Name, err)
		}

		// Now we have the event, let's look for the expected intrinsics.
		for key, value := range tc.ExpectedIntrinsicFields {
			// First, check if the key exists at all.
			if !event.intrinsics.has(key) {
				t.Fatalf("%s: missing intrinsic %s", tc.Name, key)
			}

			// Everything we're looking for is a string, so we can be a little lazy
			// here.
			if err := event.intrinsics.isString(key, value); err != nil {
				t.Errorf("%s: %v", tc.Name, err)
			}
		}

		// Finally, we verify that the unexpected intrinsics didn't miraculously
		// appear.
		for _, key := range tc.NonExpectedIntrinsicFields {
			if event.intrinsics.has(key) {
				t.Errorf("%s: expected intrinsic %s to be missing; instead, got value %v", tc.Name, key, event.intrinsics[key])
			}
		}
	}
}
