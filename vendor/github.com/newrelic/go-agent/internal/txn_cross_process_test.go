package internal

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/newrelic/go-agent/internal/cat"
)

var (
	replyAccountOne = &ConnectReply{
		CrossProcessID:  "1#1",
		EncodingKey:     "foo",
		TrustedAccounts: map[int]struct{}{1: struct{}{}},
	}

	replyAccountTwo = &ConnectReply{
		CrossProcessID:  "2#2",
		EncodingKey:     "foo",
		TrustedAccounts: map[int]struct{}{2: struct{}{}},
	}

	requestEmpty            = newRequest().Request
	requestCATOne           = newRequest().withCAT(newTxnCrossProcessFromConnectReply(replyAccountOne), "txn", "app").Request
	requestSyntheticsOne    = newRequest().withSynthetics(1, "foo").Request
	requestCATSyntheticsOne = newRequest().withCAT(newTxnCrossProcessFromConnectReply(replyAccountOne), "txn", "app").withSynthetics(1, "foo").Request
)

func mustObfuscate(input, encodingKey string) string {
	output, err := obfuscate([]byte(input), []byte(encodingKey))
	if err != nil {
		panic(err)
	}

	return string(output)
}

func newTxnCrossProcessFromConnectReply(reply *ConnectReply) *TxnCrossProcess {
	txp := &TxnCrossProcess{GUID: "abcdefgh"}
	txp.InitFromHTTPRequest(true, reply, nil)

	return txp
}

type request struct {
	*http.Request
}

func newRequest() *request {
	req, err := http.NewRequest("GET", "http://foo.bar", nil)
	if err != nil {
		panic(err)
	}

	return &request{Request: req}
}

func (req *request) withCAT(txp *TxnCrossProcess, txnName, appName string) *request {
	metadata, err := txp.CreateCrossProcessMetadata(txnName, appName)
	if err != nil {
		panic(err)
	}

	for k, values := range MetadataToHTTPHeader(metadata) {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}

	return req
}

func (req *request) withSynthetics(account int, encodingKey string) *request {
	header := fmt.Sprintf(`[1,%d,"resource","job","monitor"]`, account)
	obfuscated, err := obfuscate([]byte(header), []byte(encodingKey))
	if err != nil {
		panic(err)
	}

	req.Header.Add(cat.NewRelicSyntheticsName, string(obfuscated))
	return req
}

func TestTxnCrossProcessInit(t *testing.T) {
	for _, tc := range []struct {
		name          string
		enabled       bool
		reply         *ConnectReply
		req           *http.Request
		expected      *TxnCrossProcess
		expectedError bool
	}{
		{
			name:    "disabled",
			enabled: false,
			reply:   replyAccountOne,
			req:     nil,
			expected: &TxnCrossProcess{
				CrossProcessID:  []byte("1#1"),
				EncodingKey:     []byte("foo"),
				Enabled:         false,
				TrustedAccounts: map[int]struct{}{1: struct{}{}},
			},
			expectedError: false,
		},
		{
			name:    "normal connect reply without a request",
			enabled: true,
			reply:   replyAccountOne,
			req:     nil,
			expected: &TxnCrossProcess{
				CrossProcessID:  []byte("1#1"),
				EncodingKey:     []byte("foo"),
				Enabled:         true,
				TrustedAccounts: map[int]struct{}{1: struct{}{}},
			},
			expectedError: false,
		},
		{
			name:    "normal connect reply with a request without headers",
			enabled: true,
			reply:   replyAccountOne,
			req:     requestEmpty,
			expected: &TxnCrossProcess{
				CrossProcessID:  []byte("1#1"),
				EncodingKey:     []byte("foo"),
				Enabled:         true,
				TrustedAccounts: map[int]struct{}{1: struct{}{}},
			},
			expectedError: false,
		},
		{
			name:    "normal connect reply with a request with untrusted headers",
			enabled: true,
			reply:   replyAccountTwo,
			req:     requestCATOne,
			expected: &TxnCrossProcess{
				CrossProcessID:  []byte("2#2"),
				EncodingKey:     []byte("foo"),
				Enabled:         true,
				TrustedAccounts: map[int]struct{}{2: struct{}{}},
			},
			expectedError: true,
		},
		{
			name:    "normal connect reply with a request with trusted headers",
			enabled: true,
			reply:   replyAccountOne,
			req:     requestCATOne,
			expected: &TxnCrossProcess{
				CrossProcessID:  []byte("1#1"),
				EncodingKey:     []byte("foo"),
				Enabled:         true,
				TrustedAccounts: map[int]struct{}{1: struct{}{}},
			},
			expectedError: false,
		},
	} {
		actual := &TxnCrossProcess{}

		id := ""
		txnData := ""
		synthetics := ""
		if tc.req != nil {
			id = tc.req.Header.Get(cat.NewRelicIDName)
			txnData = tc.req.Header.Get(cat.NewRelicTxnName)
			synthetics = tc.req.Header.Get(cat.NewRelicSyntheticsName)
		}

		err := actual.Init(tc.enabled, tc.reply, CrossProcessMetadata{id, txnData, synthetics})
		if tc.expectedError == false && err != nil {
			t.Errorf("%s: unexpected error returned from Init: %v", tc.name, err)
		} else if tc.expectedError && err == nil {
			t.Errorf("%s: no error returned from Init when one was expected", tc.name)
		}

		if !reflect.DeepEqual(actual.EncodingKey, tc.expected.EncodingKey) {
			t.Errorf("%s: EncodingKey mismatch: expected=%v; got=%v", tc.name, tc.expected.EncodingKey, actual.EncodingKey)
		}

		if !reflect.DeepEqual(actual.CrossProcessID, tc.expected.CrossProcessID) {
			t.Errorf("%s: CrossProcessID mismatch: expected=%v; got=%v", tc.name, tc.expected.CrossProcessID, actual.CrossProcessID)
		}

		if !reflect.DeepEqual(actual.TrustedAccounts, tc.expected.TrustedAccounts) {
			t.Errorf("%s: TrustedAccounts mismatch: expected=%v; got=%v", tc.name, tc.expected.TrustedAccounts, actual.TrustedAccounts)
		}

		if actual.Enabled != tc.expected.Enabled {
			t.Errorf("%s: Enabled mismatch: expected=%v; got=%v", tc.name, tc.expected.Enabled, actual.Enabled)
		}
	}
}

func TestTxnCrossProcessCreateCrossProcessMetadata(t *testing.T) {
	for _, tc := range []struct {
		name             string
		enabled          bool
		reply            *ConnectReply
		req              *http.Request
		txnName          string
		appName          string
		expectedError    bool
		expectedMetadata CrossProcessMetadata
	}{
		{
			name:             "disabled, no header",
			enabled:          false,
			reply:            replyAccountOne,
			req:              nil,
			txnName:          "txn",
			appName:          "app",
			expectedError:    false,
			expectedMetadata: CrossProcessMetadata{},
		},
		{
			name:             "disabled, header",
			enabled:          false,
			reply:            replyAccountOne,
			req:              requestCATOne,
			txnName:          "txn",
			appName:          "app",
			expectedError:    false,
			expectedMetadata: CrossProcessMetadata{},
		},
		{
			name:          "disabled, synthetics",
			enabled:       false,
			reply:         replyAccountOne,
			req:           requestSyntheticsOne,
			txnName:       "txn",
			appName:       "app",
			expectedError: false,
			expectedMetadata: CrossProcessMetadata{
				Synthetics: mustObfuscate(`[1,1,"resource","job","monitor"]`, "foo"),
			},
		},
		{
			name:          "disabled, header, synthetics",
			enabled:       false,
			reply:         replyAccountOne,
			req:           requestCATSyntheticsOne,
			txnName:       "txn",
			appName:       "app",
			expectedError: false,
			expectedMetadata: CrossProcessMetadata{
				Synthetics: mustObfuscate(`[1,1,"resource","job","monitor"]`, "foo"),
			},
		},
		{
			name:          "enabled, no header, no synthetics",
			enabled:       true,
			reply:         replyAccountOne,
			req:           requestEmpty,
			txnName:       "txn",
			appName:       "app",
			expectedError: false,
			expectedMetadata: CrossProcessMetadata{
				ID:      mustObfuscate(`1#1`, "foo"),
				TxnData: mustObfuscate(`["00000000",false,"00000000","b95be233"]`, "foo"),
			},
		},
		{
			name:          "enabled, no header, synthetics",
			enabled:       true,
			reply:         replyAccountOne,
			req:           requestSyntheticsOne,
			txnName:       "txn",
			appName:       "app",
			expectedError: false,
			expectedMetadata: CrossProcessMetadata{
				ID:         mustObfuscate(`1#1`, "foo"),
				TxnData:    mustObfuscate(`["00000000",false,"00000000","b95be233"]`, "foo"),
				Synthetics: mustObfuscate(`[1,1,"resource","job","monitor"]`, "foo"),
			},
		},
		{
			name:          "enabled, header, no synthetics",
			enabled:       true,
			reply:         replyAccountOne,
			req:           requestCATOne,
			txnName:       "txn",
			appName:       "app",
			expectedError: false,
			expectedMetadata: CrossProcessMetadata{
				ID:      mustObfuscate(`1#1`, "foo"),
				TxnData: mustObfuscate(`["00000000",false,"abcdefgh","cbec2654"]`, "foo"),
			},
		},
		{
			name:          "enabled, header, synthetics",
			enabled:       true,
			reply:         replyAccountOne,
			req:           requestCATSyntheticsOne,
			txnName:       "txn",
			appName:       "app",
			expectedError: false,
			expectedMetadata: CrossProcessMetadata{
				ID:         mustObfuscate(`1#1`, "foo"),
				TxnData:    mustObfuscate(`["00000000",false,"abcdefgh","cbec2654"]`, "foo"),
				Synthetics: mustObfuscate(`[1,1,"resource","job","monitor"]`, "foo"),
			},
		},
	} {
		txp := &TxnCrossProcess{GUID: "00000000"}
		txp.InitFromHTTPRequest(tc.enabled, tc.reply, tc.req)
		metadata, err := txp.CreateCrossProcessMetadata(tc.txnName, tc.appName)

		if tc.expectedError == false && err != nil {
			t.Errorf("%s: unexpected error returned from CreateCrossProcessMetadata: %v", tc.name, err)
		} else if tc.expectedError && err == nil {
			t.Errorf("%s: no error returned from CreateCrossProcessMetadata when one was expected", tc.name)
		}

		if !reflect.DeepEqual(tc.expectedMetadata, metadata) {
			t.Errorf("%s: metadata mismatch: expected=%v; got=%v", tc.name, tc.expectedMetadata, metadata)
		}

		// Ensure that a path hash was generated if TxnData was created.
		if metadata.TxnData != "" && txp.PathHash == "" {
			t.Errorf("%s: no path hash generated", tc.name)
		}
	}
}

func TestTxnCrossProcessCreateCrossProcessMetadataError(t *testing.T) {
	// Ensure errors bubble back up from deeper within our obfuscation code.
	// It's likely impossible to get outboundTxnData() to fail, but we can get
	// outboundID() to fail by having an empty encoding key.
	txp := &TxnCrossProcess{Enabled: true}
	metadata, err := txp.CreateCrossProcessMetadata("txn", "app")
	if metadata.ID != "" || metadata.TxnData != "" || metadata.Synthetics != "" {
		t.Errorf("one or more metadata fields were set unexpectedly; got %v", metadata)
	}
	if err == nil {
		t.Errorf("did not get expected error with an empty encoding key")
	}

	// Test the above with Synthetics support to ensure that the Synthetics
	// payload is still set.
	txp = &TxnCrossProcess{
		Enabled:          true,
		Type:             txnCrossProcessSynthetics,
		SyntheticsHeader: "foo",
		// This won't be actually examined, but can't be nil for the IsSynthetics()
		// check to pass.
		Synthetics: &cat.SyntheticsHeader{},
	}
	metadata, err = txp.CreateCrossProcessMetadata("txn", "app")
	if metadata.ID != "" || metadata.TxnData != "" {
		t.Errorf("one or more metadata fields were set unexpectedly; got %v", metadata)
	}
	if metadata.Synthetics != "foo" {
		t.Errorf("unexpected synthetics metadata: expected %s; got %s", "foo", metadata.Synthetics)
	}
	if err == nil {
		t.Errorf("did not get expected error with an empty encoding key")
	}
}

func TestTxnCrossProcessFinalise(t *testing.T) {
	// No CAT.
	txp := &TxnCrossProcess{}
	txp.InitFromHTTPRequest(true, replyAccountOne, nil)
	if err := txp.Finalise("txn", "app"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if txp.PathHash != "" {
		t.Errorf("unexpected path hash: %s", txp.PathHash)
	}

	// CAT, but no path hash.
	txp = &TxnCrossProcess{}
	txp.InitFromHTTPRequest(true, replyAccountOne, requestCATOne)
	if txp.PathHash != "" {
		t.Errorf("unexpected path hash: %s", txp.PathHash)
	}
	if err := txp.Finalise("txn", "app"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if txp.PathHash == "" {
		t.Error("unexpected lack of path hash")
	}

	// CAT, with a path hash.
	txp = &TxnCrossProcess{}
	txp.InitFromHTTPRequest(true, replyAccountOne, requestCATOne)
	txp.CreateCrossProcessMetadata("txn", "app")
	if txp.PathHash == "" {
		t.Error("unexpected lack of path hash")
	}
	if err := txp.Finalise("txn", "app"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if txp.PathHash == "" {
		t.Error("unexpected lack of path hash")
	}
}

func TestTxnCrossProcessIsInbound(t *testing.T) {
	for _, tc := range []struct {
		txpType  uint8
		expected bool
	}{
		{0, false},
		{txnCrossProcessSynthetics, false},
		{txnCrossProcessInbound, true},
		{txnCrossProcessOutbound, false},
		{txnCrossProcessSynthetics | txnCrossProcessInbound, true},
		{txnCrossProcessSynthetics | txnCrossProcessOutbound, false},
		{txnCrossProcessInbound | txnCrossProcessOutbound, true},
		{txnCrossProcessSynthetics | txnCrossProcessInbound | txnCrossProcessOutbound, true},
	} {
		txp := &TxnCrossProcess{Type: tc.txpType}
		actual := txp.IsInbound()
		if actual != tc.expected {
			t.Errorf("unexpected IsInbound result for input %d: expected=%v; got=%v", tc.txpType, tc.expected, actual)
		}
	}
}

func TestTxnCrossProcessIsOutbound(t *testing.T) {
	for _, tc := range []struct {
		txpType  uint8
		expected bool
	}{
		{0, false},
		{txnCrossProcessSynthetics, false},
		{txnCrossProcessInbound, false},
		{txnCrossProcessOutbound, true},
		{txnCrossProcessSynthetics | txnCrossProcessInbound, false},
		{txnCrossProcessSynthetics | txnCrossProcessOutbound, true},
		{txnCrossProcessInbound | txnCrossProcessOutbound, true},
		{txnCrossProcessSynthetics | txnCrossProcessInbound | txnCrossProcessOutbound, true},
	} {
		txp := &TxnCrossProcess{Type: tc.txpType}
		actual := txp.IsOutbound()
		if actual != tc.expected {
			t.Errorf("unexpected IsOutbound result for input %d: expected=%v; got=%v", tc.txpType, tc.expected, actual)
		}
	}
}

func TestTxnCrossProcessIsSynthetics(t *testing.T) {
	for _, tc := range []struct {
		txpType    uint8
		synthetics *cat.SyntheticsHeader
		expected   bool
	}{
		{0, nil, false},
		{txnCrossProcessSynthetics, nil, false},
		{txnCrossProcessInbound, nil, false},
		{txnCrossProcessOutbound, nil, false},
		{txnCrossProcessSynthetics | txnCrossProcessInbound, nil, false},
		{txnCrossProcessSynthetics | txnCrossProcessOutbound, nil, false},
		{txnCrossProcessInbound | txnCrossProcessOutbound, nil, false},
		{txnCrossProcessSynthetics | txnCrossProcessInbound | txnCrossProcessOutbound, nil, false},
		{0, &cat.SyntheticsHeader{}, false},
		{txnCrossProcessSynthetics, &cat.SyntheticsHeader{}, true},
		{txnCrossProcessInbound, &cat.SyntheticsHeader{}, false},
		{txnCrossProcessOutbound, &cat.SyntheticsHeader{}, false},
		{txnCrossProcessSynthetics | txnCrossProcessInbound, &cat.SyntheticsHeader{}, true},
		{txnCrossProcessSynthetics | txnCrossProcessOutbound, &cat.SyntheticsHeader{}, true},
		{txnCrossProcessInbound | txnCrossProcessOutbound, &cat.SyntheticsHeader{}, false},
		{txnCrossProcessSynthetics | txnCrossProcessInbound | txnCrossProcessOutbound, &cat.SyntheticsHeader{}, true},
	} {
		txp := &TxnCrossProcess{Type: tc.txpType, Synthetics: tc.synthetics}
		actual := txp.IsSynthetics()
		if actual != tc.expected {
			t.Errorf("unexpected IsSynthetics result for input %d and %p: expected=%v; got=%v", tc.txpType, tc.synthetics, tc.expected, actual)
		}
	}
}

func TestTxnCrossProcessUsed(t *testing.T) {
	for _, tc := range []struct {
		txpType  uint8
		expected bool
	}{
		{0, false},
		{txnCrossProcessSynthetics, true},
		{txnCrossProcessInbound, true},
		{txnCrossProcessOutbound, true},
		{txnCrossProcessSynthetics | txnCrossProcessInbound, true},
		{txnCrossProcessSynthetics | txnCrossProcessOutbound, true},
		{txnCrossProcessInbound | txnCrossProcessOutbound, true},
		{txnCrossProcessSynthetics | txnCrossProcessInbound | txnCrossProcessOutbound, true},
	} {
		txp := &TxnCrossProcess{Type: tc.txpType}
		actual := txp.Used()
		if actual != tc.expected {
			t.Errorf("unexpected Used result for input %d: expected=%v; got=%v", tc.txpType, tc.expected, actual)
		}
	}
}

func TestTxnCrossProcessSetInbound(t *testing.T) {
	txp := &TxnCrossProcess{Type: 0}

	txp.SetInbound(false)
	if txp.IsInbound() != false {
		t.Error("Inbound is not false after being set to false from false")
	}

	txp.SetInbound(true)
	if txp.IsInbound() != true {
		t.Error("Inbound is not true after being set to true from false")
	}

	txp.SetInbound(true)
	if txp.IsInbound() != true {
		t.Error("Inbound is not true after being set to true from true")
	}

	txp.SetInbound(false)
	if txp.IsInbound() != false {
		t.Error("Inbound is not false after being set to false from true")
	}
}

func TestTxnCrossProcessSetOutbound(t *testing.T) {
	txp := &TxnCrossProcess{Type: 0}

	txp.SetOutbound(false)
	if txp.IsOutbound() != false {
		t.Error("Outbound is not false after being set to false from false")
	}

	txp.SetOutbound(true)
	if txp.IsOutbound() != true {
		t.Error("Outbound is not true after being set to true from false")
	}

	txp.SetOutbound(true)
	if txp.IsOutbound() != true {
		t.Error("Outbound is not true after being set to true from true")
	}

	txp.SetOutbound(false)
	if txp.IsOutbound() != false {
		t.Error("Outbound is not false after being set to false from true")
	}
}

func TestTxnCrossProcessSetSynthetics(t *testing.T) {
	// We'll always set SyntheticsHeader, since we're not really testing the full
	// behaviour of IsSynthetics() here.
	txp := &TxnCrossProcess{
		Type:       0,
		Synthetics: &cat.SyntheticsHeader{},
	}

	txp.SetSynthetics(false)
	if txp.IsSynthetics() != false {
		t.Error("Synthetics is not false after being set to false from false")
	}

	txp.SetSynthetics(true)
	if txp.IsSynthetics() != true {
		t.Error("Synthetics is not true after being set to true from false")
	}

	txp.SetSynthetics(true)
	if txp.IsSynthetics() != true {
		t.Error("Synthetics is not true after being set to true from true")
	}

	txp.SetSynthetics(false)
	if txp.IsSynthetics() != false {
		t.Error("Synthetics is not false after being set to false from true")
	}
}

func TestTxnCrossProcessParseAppData(t *testing.T) {
	for _, tc := range []struct {
		name            string
		encodingKey     string
		input           string
		expectedAppData *cat.AppDataHeader
		expectedError   bool
	}{
		{
			name:            "empty string",
			encodingKey:     "foo",
			input:           "",
			expectedAppData: nil,
			expectedError:   false,
		},
		{
			name:            "invalidly encoded string",
			encodingKey:     "foo",
			input:           "xxx",
			expectedAppData: nil,
			expectedError:   true,
		},
		{
			name:            "invalid JSON",
			encodingKey:     "foo",
			input:           mustObfuscate("xxx", "foo"),
			expectedAppData: nil,
			expectedError:   true,
		},
		{
			name:            "invalid encoding key",
			encodingKey:     "foo",
			input:           mustObfuscate(`["xp","txn",1,2,3,"guid",false]`, "bar"),
			expectedAppData: nil,
			expectedError:   true,
		},
		{
			name:        "success",
			encodingKey: "foo",
			input:       mustObfuscate(`["xp","txn",1,2,3,"guid",false]`, "foo"),
			expectedAppData: &cat.AppDataHeader{
				CrossProcessID:        "xp",
				TransactionName:       "txn",
				QueueTimeInSeconds:    1,
				ResponseTimeInSeconds: 2,
				ContentLength:         3,
				TransactionGUID:       "guid",
			},
			expectedError: false,
		},
	} {
		txp := &TxnCrossProcess{
			Enabled:     true,
			EncodingKey: []byte(tc.encodingKey),
		}

		actualAppData, actualErr := txp.ParseAppData(tc.input)

		if tc.expectedError && actualErr == nil {
			t.Errorf("%s: expected an error, but didn't get one", tc.name)
		} else if tc.expectedError == false && actualErr != nil {
			t.Errorf("%s: expected no error, but got %v", tc.name, actualErr)
		}

		if !reflect.DeepEqual(actualAppData, tc.expectedAppData) {
			t.Errorf("%s: app data mismatched: expected=%v; got=%v", tc.name, tc.expectedAppData, actualAppData)
		}
	}
}

func TestTxnCrossProcessCreateAppData(t *testing.T) {
	for _, tc := range []struct {
		name            string
		enabled         bool
		crossProcessID  string
		encodingKey     string
		txnName         string
		queueTime       time.Duration
		responseTime    time.Duration
		contentLength   int64
		guid            string
		expectedAppData string
		expectedError   bool
	}{
		{
			name:            "cat disabled",
			enabled:         false,
			crossProcessID:  "1#1",
			encodingKey:     "foo",
			txnName:         "txn",
			queueTime:       1 * time.Second,
			responseTime:    2 * time.Second,
			contentLength:   4096,
			guid:            "",
			expectedAppData: "",
			expectedError:   false,
		},
		{
			name:            "invalid encoding key",
			enabled:         true,
			crossProcessID:  "1#1",
			encodingKey:     "",
			txnName:         "txn",
			queueTime:       1 * time.Second,
			responseTime:    2 * time.Second,
			contentLength:   4096,
			guid:            "",
			expectedAppData: "",
			expectedError:   true,
		},
		{
			name:            "success",
			enabled:         true,
			crossProcessID:  "1#1",
			encodingKey:     "foo",
			txnName:         "txn",
			queueTime:       1 * time.Second,
			responseTime:    2 * time.Second,
			contentLength:   4096,
			guid:            "guid",
			expectedAppData: mustObfuscate(`["1#1","txn",1,2,4096,"guid",false]`, "foo"),
			expectedError:   false,
		},
	} {
		txp := &TxnCrossProcess{
			Enabled:        tc.enabled,
			EncodingKey:    []byte(tc.encodingKey),
			CrossProcessID: []byte(tc.crossProcessID),
			GUID:           tc.guid,
		}

		actualAppData, actualErr := txp.CreateAppData(tc.txnName, tc.queueTime, tc.responseTime, tc.contentLength)

		if tc.expectedError && actualErr == nil {
			t.Errorf("%s: expected an error, but didn't get one", tc.name)
		} else if tc.expectedError == false && actualErr != nil {
			t.Errorf("%s: expected no error, but got %v", tc.name, actualErr)
		}

		if !reflect.DeepEqual(actualAppData, tc.expectedAppData) {
			t.Errorf("%s: app data mismatched: expected=%v; got=%v", tc.name, tc.expectedAppData, actualAppData)
		}
	}
}

func TestTxnCrossProcessHandleInboundRequestHeaders(t *testing.T) {
	for _, tc := range []struct {
		name          string
		enabled       bool
		reply         *ConnectReply
		metadata      CrossProcessMetadata
		expectedError bool
	}{
		{
			name:    "disabled, invalid encoding key, invalid synthetics",
			enabled: false,
			reply: &ConnectReply{
				EncodingKey: "",
			},
			metadata: CrossProcessMetadata{
				Synthetics: "foo",
			},
			expectedError: true,
		},
		{
			name:    "disabled, valid encoding key, invalid synthetics",
			enabled: false,
			reply:   replyAccountOne,
			metadata: CrossProcessMetadata{
				Synthetics: "foo",
			},
			expectedError: true,
		},
		{
			name:    "disabled, valid encoding key, valid synthetics",
			enabled: false,
			reply:   replyAccountOne,
			metadata: CrossProcessMetadata{
				Synthetics: mustObfuscate(`[1,1,"resource","job","monitor"]`, "foo"),
			},
			expectedError: false,
		},
		{
			name:    "enabled, invalid encoding key, valid input",
			enabled: true,
			reply: &ConnectReply{
				EncodingKey: "",
			},
			metadata: CrossProcessMetadata{
				ID:      mustObfuscate(`1#1`, "foo"),
				TxnData: mustObfuscate(`["00000000",false,"00000000","b95be233"]`, "foo"),
			},
			expectedError: true,
		},
		{
			name:    "enabled, valid encoding key, invalid id",
			enabled: true,
			reply:   replyAccountOne,
			metadata: CrossProcessMetadata{
				ID:      mustObfuscate(`1`, "foo"),
				TxnData: mustObfuscate(`["00000000",false,"00000000","b95be233"]`, "foo"),
			},
			expectedError: true,
		},
		{
			name:    "enabled, valid encoding key, invalid txndata",
			enabled: true,
			reply:   replyAccountOne,
			metadata: CrossProcessMetadata{
				ID:      mustObfuscate(`1#1`, "foo"),
				TxnData: mustObfuscate(`["00000000",alse,"00000000","b95be233"]`, "foo"),
			},
			expectedError: true,
		},
		{
			name:    "enabled, valid encoding key, valid input",
			enabled: true,
			reply:   replyAccountOne,
			metadata: CrossProcessMetadata{
				ID:      mustObfuscate(`1#1`, "foo"),
				TxnData: mustObfuscate(`["00000000",false,"00000000","b95be233"]`, "foo"),
			},
			expectedError: false,
		},
	} {
		txp := &TxnCrossProcess{Enabled: tc.enabled}
		txp.Init(tc.enabled, tc.reply, CrossProcessMetadata{})

		err := txp.handleInboundRequestHeaders(tc.metadata)
		if tc.expectedError && err == nil {
			t.Errorf("%s: expected error, but didn't get one", tc.name)
		} else if tc.expectedError == false && err != nil {
			t.Errorf("%s: expected no error, but got %v", tc.name, err)
		}
	}
}
