package utilization

import (
	"net/http"
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

func TestCrossAgentGCP(t *testing.T) {
	var testCases []testCase

	err := crossagent.ReadJSON("utilization_vendor_specific/gcp.json", &testCases)
	if err != nil {
		t.Fatalf("reading gcp.json failed: %v", err)
	}

	for _, testCase := range testCases {
		client := &http.Client{
			Transport: &mockTransport{
				t:         t,
				responses: testCase.URIs,
			},
		}

		gcp, err := getGCP(client)

		if testCase.ExpectedVendorsHash.GCP == nil {
			if err == nil {
				t.Fatalf("%s: expected error; got nil", testCase.TestName)
			}
		} else {
			if err != nil {
				t.Fatalf("%s: expected no error; got %v", testCase.TestName, err)
			}

			if gcp.ID != testCase.ExpectedVendorsHash.GCP.ID {
				t.Fatalf("%s: ID incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.GCP.ID, gcp.ID)
			}

			if gcp.MachineType != testCase.ExpectedVendorsHash.GCP.MachineType {
				t.Fatalf("%s: MachineType incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.GCP.MachineType, gcp.MachineType)
			}

			if gcp.Name != testCase.ExpectedVendorsHash.GCP.Name {
				t.Fatalf("%s: Name incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.GCP.Name, gcp.Name)
			}

			if gcp.Zone != testCase.ExpectedVendorsHash.GCP.Zone {
				t.Fatalf("%s: Zone incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.GCP.Zone, gcp.Zone)
			}
		}
	}
}

func TestStripGCPPrefix(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"foo/bar", "bar"},
		{"/foo/bar", "bar"},
		{"/foo/bar/", ""},
		{"foo", "foo"},
		{"", ""},
	}

	for _, tc := range testCases {
		actual := stripGCPPrefix(tc.input)
		if tc.expected != actual {
			t.Fatalf("input: %s; expected: %s; actual: %s", tc.input, tc.expected, actual)
		}
	}
}
