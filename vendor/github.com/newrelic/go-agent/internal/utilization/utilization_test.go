package utilization

import (
	"encoding/json"
	"testing"

	"github.com/newrelic/go-agent/internal/logger"
)

func TestJSONMarshalling(t *testing.T) {
	ramMib := uint64(1024)
	processors := 16
	u := Data{
		MetadataVersion:   metadataVersion,
		LogicalProcessors: 4,
		RAMMib:            &ramMib,
		Hostname:          "localhost",
		Vendors: &vendors{
			AWS: &vendor{
				ID:   "8BADFOOD",
				Type: "t2.micro",
				Zone: "us-west-1",
			},
			Docker: &vendor{ID: "47cbd16b77c50cbf71401"},
		},
		Config: &override{
			LogicalProcessors: &processors,
		},
	}

	expect := `{
	"metadata_version": 2,
	"logical_processors": 4,
	"total_ram_mib": 1024,
	"hostname": "localhost",
	"vendors": {
		"aws": {
			"id": "8BADFOOD",
			"type": "t2.micro",
			"zone": "us-west-1"
		},
		"docker": {
			"id": "47cbd16b77c50cbf71401"
		}
	},
	"config": {
		"logical_processors": 16
	}
}`

	j, err := json.MarshalIndent(u, "", "\t")
	if err != nil {
		t.Error(err)
	}
	if string(j) != expect {
		t.Error(string(j), expect)
	}

	// Test that we marshal not-present values to nil.
	u.RAMMib = nil
	u.Hostname = ""
	u.Config = nil
	expect = `{
	"metadata_version": 2,
	"logical_processors": 4,
	"total_ram_mib": null,
	"hostname": "",
	"vendors": {
		"aws": {
			"id": "8BADFOOD",
			"type": "t2.micro",
			"zone": "us-west-1"
		},
		"docker": {
			"id": "47cbd16b77c50cbf71401"
		}
	}
}`

	j, err = json.MarshalIndent(u, "", "\t")
	if err != nil {
		t.Error(err)
	}
	if string(j) != expect {
		t.Error(string(j), expect)
	}
}

func TestUtilizationHash(t *testing.T) {
	config := []Config{
		{DetectAWS: true, DetectDocker: true},
		{DetectAWS: false, DetectDocker: false},
	}
	for _, c := range config {
		u := Gather(c, logger.ShimLogger{})
		js, err := json.Marshal(u)
		if err != nil {
			t.Error(err)
		}
		if u.MetadataVersion == 0 || u.LogicalProcessors == 0 ||
			u.RAMMib == nil || *u.RAMMib == 0 ||
			u.Hostname == "" {
			t.Fatal(u, string(js))
		}
	}
}

func TestOverrideFromConfig(t *testing.T) {
	testcases := []struct {
		config Config
		expect string
	}{
		{Config{}, `null`},
		{Config{LogicalProcessors: 16}, `{"logical_processors":16}`},
		{Config{TotalRAMMIB: 1024}, `{"total_ram_mib":1024}`},
		{Config{BillingHostname: "localhost"}, `{"hostname":"localhost"}`},
		{Config{
			LogicalProcessors: 16,
			TotalRAMMIB:       1024,
			BillingHostname:   "localhost",
		}, `{"logical_processors":16,"total_ram_mib":1024,"hostname":"localhost"}`},
	}

	for _, tc := range testcases {
		ov := overrideFromConfig(tc.config)
		js, err := json.Marshal(ov)
		if nil != err {
			t.Error(tc.expect, err)
			continue
		}
		if string(js) != tc.expect {
			t.Error(tc.expect, string(js))
		}
	}
}
