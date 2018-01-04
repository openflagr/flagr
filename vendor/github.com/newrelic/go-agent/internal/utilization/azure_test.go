package utilization

import (
	"net/http"
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

func TestCrossAgentAzure(t *testing.T) {
	var testCases []testCase

	err := crossagent.ReadJSON("utilization_vendor_specific/azure.json", &testCases)
	if err != nil {
		t.Fatalf("reading azure.json failed: %v", err)
	}

	for _, testCase := range testCases {
		client := &http.Client{
			Transport: &mockTransport{
				t:         t,
				responses: testCase.URIs,
			},
		}

		azure, err := getAzure(client)

		if testCase.ExpectedVendorsHash.Azure == nil {
			if err == nil {
				t.Fatalf("%s: expected error; got nil", testCase.TestName)
			}
		} else {
			if err != nil {
				t.Fatalf("%s: expected no error; got %v", testCase.TestName, err)
			}

			if azure.Location != testCase.ExpectedVendorsHash.Azure.Location {
				t.Fatalf("%s: Location incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.Azure.Location, azure.Location)
			}

			if azure.Name != testCase.ExpectedVendorsHash.Azure.Name {
				t.Fatalf("%s: Name incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.Azure.Name, azure.Name)
			}

			if azure.VMID != testCase.ExpectedVendorsHash.Azure.VMID {
				t.Fatalf("%s: VMID incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.Azure.VMID, azure.VMID)
			}

			if azure.VMSize != testCase.ExpectedVendorsHash.Azure.VMSize {
				t.Fatalf("%s: VMSize incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.Azure.VMSize, azure.VMSize)
			}
		}
	}
}
