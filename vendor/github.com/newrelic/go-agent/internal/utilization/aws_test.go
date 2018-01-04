package utilization

import (
	"net/http"
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

func TestCrossAgentAWS(t *testing.T) {
	var testCases []testCase

	err := crossagent.ReadJSON("utilization_vendor_specific/aws.json", &testCases)
	if err != nil {
		t.Fatalf("reading aws.json failed: %v", err)
	}

	for _, testCase := range testCases {
		client := &http.Client{
			Transport: &mockTransport{
				t:         t,
				responses: testCase.URIs,
			},
		}

		aws, err := getAWS(client)

		if testCase.ExpectedVendorsHash.AWS == nil {
			if err == nil {
				t.Fatalf("%s: expected error; got nil", testCase.TestName)
			}
		} else {
			if err != nil {
				t.Fatalf("%s: expected no error; got %v", testCase.TestName, err)
			}

			if aws.InstanceID != testCase.ExpectedVendorsHash.AWS.InstanceID {
				t.Fatalf("%s: instanceId incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.AWS.InstanceID, aws.InstanceID)
			}

			if aws.InstanceType != testCase.ExpectedVendorsHash.AWS.InstanceType {
				t.Fatalf("%s: instanceType incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.AWS.InstanceType, aws.InstanceType)
			}

			if aws.AvailabilityZone != testCase.ExpectedVendorsHash.AWS.AvailabilityZone {
				t.Fatalf("%s: availabilityZone incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.AWS.AvailabilityZone, aws.AvailabilityZone)
			}
		}
	}
}
