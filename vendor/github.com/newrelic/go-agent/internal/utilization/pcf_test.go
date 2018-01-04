package utilization

import (
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

func TestCrossAgentPCF(t *testing.T) {
	var testCases []testCase

	err := crossagent.ReadJSON("utilization_vendor_specific/pcf.json", &testCases)
	if err != nil {
		t.Fatalf("reading pcf.json failed: %v", err)
	}

	for _, testCase := range testCases {
		pcf, err := getPCF(func(key string) string {
			resp := testCase.EnvVars[key]
			if resp.Timeout {
				return ""
			}
			return resp.Response
		})

		if testCase.ExpectedVendorsHash.PCF == nil {
			if err == nil {
				t.Fatalf("%s: expected error; got nil", testCase.TestName)
			}
		} else {
			if err != nil {
				t.Fatalf("%s: expected no error; got %v", testCase.TestName, err)
			}

			if pcf.InstanceGUID != testCase.ExpectedVendorsHash.PCF.InstanceGUID {
				t.Fatalf("%s: InstanceGUID incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.PCF.InstanceGUID, pcf.InstanceGUID)
			}

			if pcf.InstanceIP != testCase.ExpectedVendorsHash.PCF.InstanceIP {
				t.Fatalf("%s: InstanceIP incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.PCF.InstanceIP, pcf.InstanceIP)
			}

			if pcf.MemoryLimit != testCase.ExpectedVendorsHash.PCF.MemoryLimit {
				t.Fatalf("%s: MemoryLimit incorrect; expected: %s; got: %s", testCase.TestName, testCase.ExpectedVendorsHash.PCF.MemoryLimit, pcf.MemoryLimit)
			}
		}
	}
}
