package cat

import (
	"encoding/json"
	"testing"
)

func TestTxnDataRoundTrip(t *testing.T) {
	for _, test := range []struct {
		input   string
		output  string
		txnData TxnDataHeader
	}{
		{
			input:  `["guid",false]`,
			output: `["guid",false,"",""]`,
			txnData: TxnDataHeader{
				GUID:     "guid",
				TripID:   "",
				PathHash: "",
			},
		},
		{
			input:  `["guid",false,"trip"]`,
			output: `["guid",false,"trip",""]`,
			txnData: TxnDataHeader{
				GUID:     "guid",
				TripID:   "trip",
				PathHash: "",
			},
		},
		{
			input:  `["guid",false,null]`,
			output: `["guid",false,"",""]`,
			txnData: TxnDataHeader{
				GUID:     "guid",
				TripID:   "",
				PathHash: "",
			},
		},
		{
			input:  `["guid",false,"trip",null]`,
			output: `["guid",false,"trip",""]`,
			txnData: TxnDataHeader{
				GUID:     "guid",
				TripID:   "trip",
				PathHash: "",
			},
		},
		{
			input:  `["guid",false,"trip","hash"]`,
			output: `["guid",false,"trip","hash"]`,
			txnData: TxnDataHeader{
				GUID:     "guid",
				TripID:   "trip",
				PathHash: "hash",
			},
		},
	} {
		// Test unmarshalling.
		txnData := &TxnDataHeader{}
		if err := json.Unmarshal([]byte(test.input), txnData); err != nil {
			t.Errorf("given %s: error expected to be nil; got %v", test.input, err)
		}

		if test.txnData.GUID != txnData.GUID {
			t.Errorf("given %s: GUID expected to be %s; got %s", test.input, test.txnData.GUID, txnData.GUID)
		}

		if test.txnData.TripID != txnData.TripID {
			t.Errorf("given %s: TripID expected to be %s; got %s", test.input, test.txnData.TripID, txnData.TripID)
		}

		if test.txnData.PathHash != txnData.PathHash {
			t.Errorf("given %s: PathHash expected to be %s; got %s", test.input, test.txnData.PathHash, txnData.PathHash)
		}

		// Test marshalling.
		data, err := json.Marshal(&test.txnData)
		if err != nil {
			t.Errorf("given %s: error expected to be nil; got %v", test.output, err)
		}

		if string(data) != test.output {
			t.Errorf("given %s: unexpected JSON %s", test.output, string(data))
		}
	}
}

func TestTxnDataUnmarshal(t *testing.T) {
	// Test error cases where we get a generic error from the JSON package.
	for _, input := range []string{
		// Basic malformed JSON test: beyond this, we're not going to unit test the
		// Go standard library's JSON package.
		``,
	} {
		txnData := &TxnDataHeader{}

		if err := json.Unmarshal([]byte(input), txnData); err == nil {
			t.Errorf("given %s: error expected to be non-nil; got nil", input)
		}
	}

	// Test error cases where the incorrect number of elements was provided.
	for _, input := range []string{
		`[]`,
		`[1]`,
	} {
		txnData := &TxnDataHeader{}

		err := json.Unmarshal([]byte(input), txnData)
		if _, ok := err.(errUnexpectedArraySize); !ok {
			t.Errorf("given %s: error expected to be errUnexpectedArraySize; got %v", input, err)
		}
	}

	// Test error cases where a specific variable is returned.
	for _, tc := range []struct {
		input string
		err   error
	}{
		// Unexpected JSON types.
		{`false`, errInvalidTxnDataJSON},
		{`true`, errInvalidTxnDataJSON},
		{`1234`, errInvalidTxnDataJSON},
		{`{}`, errInvalidTxnDataJSON},
		{`""`, errInvalidTxnDataJSON},

		// Invalid data types for each field in turn.
		{`[false,false,"trip","hash"]`, errInvalidTxnDataGUID},
		{`["guid",false,0,"hash"]`, errInvalidTxnDataTripID},
		{`["guid",false,"trip",[]]`, errInvalidTxnDataPathHash},
	} {
		txnData := &TxnDataHeader{}

		if err := json.Unmarshal([]byte(tc.input), txnData); err != tc.err {
			t.Errorf("given %s: error expected to be %v; got %v", tc.input, tc.err, err)
		}
	}
}
