package cat

import (
	"encoding/json"
	"testing"
)

func TestSyntheticsUnmarshalInvalid(t *testing.T) {
	// Test error cases where we get a generic error from the JSON package.
	for _, input := range []string{
		// Basic malformed JSON test: beyond this, we're not going to unit test the
		// Go standard library's JSON package.
		``,
	} {
		synthetics := &SyntheticsHeader{}

		if err := json.Unmarshal([]byte(input), synthetics); err == nil {
			t.Errorf("given %s: error expected to be non-nil; got nil", input)
		}
	}

	// Test error cases where the incorrect number of elements was provided.
	for _, input := range []string{
		`[]`,
		`[1,2,3,4]`,
	} {
		synthetics := &SyntheticsHeader{}

		err := json.Unmarshal([]byte(input), synthetics)
		if _, ok := err.(errUnexpectedArraySize); !ok {
			t.Errorf("given %s: error expected to be errUnexpectedArraySize; got %v", input, err)
		}
	}

	// Test error cases with invalid version numbers.
	for _, input := range []string{
		`[0,1234,"resource","job","monitor"]`,
		`[2,1234,"resource","job","monitor"]`,
	} {
		synthetics := &SyntheticsHeader{}

		err := json.Unmarshal([]byte(input), synthetics)
		if _, ok := err.(errUnexpectedSyntheticsVersion); !ok {
			t.Errorf("given %s: error expected to be errUnexpectedSyntheticsVersion; got %v", input, err)
		}
	}

	// Test error cases where a specific variable is returned.
	for _, tc := range []struct {
		input string
		err   error
	}{
		// Unexpected JSON types.
		{`false`, errInvalidSyntheticsJSON},
		{`true`, errInvalidSyntheticsJSON},
		{`1234`, errInvalidSyntheticsJSON},
		{`{}`, errInvalidSyntheticsJSON},
		{`""`, errInvalidSyntheticsJSON},

		// Invalid data types for each field in turn.
		{`["version",1234,"resource","job","monitor"]`, errInvalidSyntheticsVersion},
		{`[1,"account","resource","job","monitor"]`, errInvalidSyntheticsAccountID},
		{`[1,1234,0,"job","monitor"]`, errInvalidSyntheticsResourceID},
		{`[1,1234,"resource",-1,"monitor"]`, errInvalidSyntheticsJobID},
		{`[1,1234,"resource","job",false]`, errInvalidSyntheticsMonitorID},
	} {
		synthetics := &SyntheticsHeader{}

		if err := json.Unmarshal([]byte(tc.input), synthetics); err != tc.err {
			t.Errorf("given %s: error expected to be %v; got %v", tc.input, tc.err, err)
		}
	}
}

func TestSyntheticsUnmarshalValid(t *testing.T) {
	for _, test := range []struct {
		json       string
		synthetics SyntheticsHeader
	}{
		{
			json: `[1,1234,"resource","job","monitor"]`,
			synthetics: SyntheticsHeader{
				Version:    1,
				AccountID:  1234,
				ResourceID: "resource",
				JobID:      "job",
				MonitorID:  "monitor",
			},
		},
	} {
		// Test unmarshalling.
		synthetics := &SyntheticsHeader{}
		if err := json.Unmarshal([]byte(test.json), synthetics); err != nil {
			t.Errorf("given %s: error expected to be nil; got %v", test.json, err)
		}

		if test.synthetics.Version != synthetics.Version {
			t.Errorf("given %s: Version expected to be %d; got %d", test.json, test.synthetics.Version, synthetics.Version)
		}

		if test.synthetics.AccountID != synthetics.AccountID {
			t.Errorf("given %s: AccountID expected to be %d; got %d", test.json, test.synthetics.AccountID, synthetics.AccountID)
		}

		if test.synthetics.ResourceID != synthetics.ResourceID {
			t.Errorf("given %s: ResourceID expected to be %s; got %s", test.json, test.synthetics.ResourceID, synthetics.ResourceID)
		}

		if test.synthetics.JobID != synthetics.JobID {
			t.Errorf("given %s: JobID expected to be %s; got %s", test.json, test.synthetics.JobID, synthetics.JobID)
		}

		if test.synthetics.MonitorID != synthetics.MonitorID {
			t.Errorf("given %s: MonitorID expected to be %s; got %s", test.json, test.synthetics.MonitorID, synthetics.MonitorID)
		}
	}
}
