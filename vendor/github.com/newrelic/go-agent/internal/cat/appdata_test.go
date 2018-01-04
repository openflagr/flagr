package cat

import (
	"encoding/json"
	"testing"
)

func TestAppDataRoundTrip(t *testing.T) {
	for _, test := range []struct {
		json    string
		appData AppDataHeader
	}{
		{
			json: `["xpid","txn",1,2,4096,"guid",false]`,
			appData: AppDataHeader{
				CrossProcessID:        "xpid",
				TransactionName:       "txn",
				QueueTimeInSeconds:    1.0,
				ResponseTimeInSeconds: 2.0,
				ContentLength:         4096,
				TransactionGUID:       "guid",
			},
		},
	} {
		// Test unmarshalling.
		appData := &AppDataHeader{}
		if err := json.Unmarshal([]byte(test.json), appData); err != nil {
			t.Errorf("given %s: error expected to be nil; got %v", test.json, err)
		}

		if test.appData.CrossProcessID != appData.CrossProcessID {
			t.Errorf("given %s: CrossProcessID expected to be %s; got %s", test.json, test.appData.CrossProcessID, appData.CrossProcessID)
		}

		if test.appData.TransactionName != appData.TransactionName {
			t.Errorf("given %s: TransactionName expected to be %s; got %s", test.json, test.appData.TransactionName, appData.TransactionName)
		}

		if test.appData.QueueTimeInSeconds != appData.QueueTimeInSeconds {
			t.Errorf("given %s: QueueTimeInSeconds expected to be %f; got %f", test.json, test.appData.QueueTimeInSeconds, appData.QueueTimeInSeconds)
		}

		if test.appData.ResponseTimeInSeconds != appData.ResponseTimeInSeconds {
			t.Errorf("given %s: ResponseTimeInSeconds expected to be %f; got %f", test.json, test.appData.ResponseTimeInSeconds, appData.ResponseTimeInSeconds)
		}

		if test.appData.ContentLength != appData.ContentLength {
			t.Errorf("given %s: ContentLength expected to be %d; got %d", test.json, test.appData.ContentLength, appData.ContentLength)
		}

		if test.appData.TransactionGUID != appData.TransactionGUID {
			t.Errorf("given %s: TransactionGUID expected to be %s; got %s", test.json, test.appData.TransactionGUID, appData.TransactionGUID)
		}

		// Test marshalling.
		data, err := json.Marshal(&test.appData)
		if err != nil {
			t.Errorf("given %s: error expected to be nil; got %v", test.json, err)
		}

		if string(data) != test.json {
			t.Errorf("given %s: unexpected JSON %s", test.json, string(data))
		}
	}
}

func TestAppDataUnmarshal(t *testing.T) {
	// Test error cases where we get a generic error from the JSON package.
	for _, input := range []string{
		// Basic malformed JSON test: beyond this, we're not going to unit test the
		// Go standard library's JSON package.
		``,
	} {
		appData := &AppDataHeader{}

		if err := json.Unmarshal([]byte(input), appData); err == nil {
			t.Errorf("given %s: error expected to be non-nil; got nil", input)
		}
	}

	// Test error cases where a specific variable is returned.
	for _, tc := range []struct {
		input string
		err   error
	}{
		// Unexpected JSON types.
		{`false`, errInvalidAppDataJSON},
		{`true`, errInvalidAppDataJSON},
		{`1234`, errInvalidAppDataJSON},
		{`{}`, errInvalidAppDataJSON},
		{`""`, errInvalidAppDataJSON},

		// Invalid data types for each field in turn.
		{`[0,"txn",1.0,2.0,4096,"guid",false]`, errInvalidAppDataCrossProcessID},
		{`["xpid",0,1.0,2.0,4096,"guid",false]`, errInvalidAppDataTransactionName},
		{`["xpid","txn","queue",2.0,4096,"guid",false]`, errInvalidAppDataQueueTimeInSeconds},
		{`["xpid","txn",1.0,"response",4096,"guid",false]`, errInvalidAppDataResponseTimeInSeconds},
		{`["xpid","txn",1.0,2.0,"content length","guid",false]`, errInvalidAppDataContentLength},
		{`["xpid","txn",1.0,2.0,4096,0,false]`, errInvalidAppDataTransactionGUID},
	} {
		appData := &AppDataHeader{}

		if err := json.Unmarshal([]byte(tc.input), appData); err != tc.err {
			t.Errorf("given %s: error expected to be %v; got %v", tc.input, tc.err, err)
		}
	}

	// Test error cases where the incorrect number of elements was provided.
	for _, input := range []string{
		`[]`,
		`[1,2,3,4,5,6]`,
	} {
		appData := &AppDataHeader{}

		err := json.Unmarshal([]byte(input), appData)
		if _, ok := err.(errUnexpectedArraySize); !ok {
			t.Errorf("given %s: error expected to be errUnexpectedArraySize; got %v", input, err)
		}
	}
}
