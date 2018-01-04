package newrelic

import (
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/newrelic/go-agent/internal"
	"github.com/newrelic/go-agent/internal/utilization"
)

var (
	fixRegex = regexp.MustCompile(`e\+\d+`)
)

// In Go 1.8 Marshalling of numbers was changed:
// Before: "StackTraceThreshold":5e+08
// After:  "StackTraceThreshold":500000000
func standardizeNumbers(input string) string {
	return fixRegex.ReplaceAllStringFunc(input, func(s string) string {
		n, err := strconv.Atoi(s[2:])
		if nil != err {
			return s
		}
		return strings.Repeat("0", n)
	})
}

func TestCopyConfigReferenceFieldsPresent(t *testing.T) {
	cfg := NewConfig("my appname", "0123456789012345678901234567890123456789")
	cfg.Labels["zip"] = "zap"
	cfg.ErrorCollector.IgnoreStatusCodes = append(cfg.ErrorCollector.IgnoreStatusCodes, 405)
	cfg.Attributes.Include = append(cfg.Attributes.Include, "1")
	cfg.Attributes.Exclude = append(cfg.Attributes.Exclude, "2")
	cfg.TransactionEvents.Attributes.Include = append(cfg.TransactionEvents.Attributes.Include, "3")
	cfg.TransactionEvents.Attributes.Exclude = append(cfg.TransactionEvents.Attributes.Exclude, "4")
	cfg.ErrorCollector.Attributes.Include = append(cfg.ErrorCollector.Attributes.Include, "5")
	cfg.ErrorCollector.Attributes.Exclude = append(cfg.ErrorCollector.Attributes.Exclude, "6")
	cfg.TransactionTracer.Attributes.Include = append(cfg.TransactionTracer.Attributes.Include, "7")
	cfg.TransactionTracer.Attributes.Exclude = append(cfg.TransactionTracer.Attributes.Exclude, "8")
	cfg.Transport = &http.Transport{}
	cfg.Logger = NewLogger(os.Stdout)

	cp := copyConfigReferenceFields(cfg)

	cfg.Labels["zop"] = "zup"
	cfg.ErrorCollector.IgnoreStatusCodes[0] = 201
	cfg.Attributes.Include[0] = "zap"
	cfg.Attributes.Exclude[0] = "zap"
	cfg.TransactionEvents.Attributes.Include[0] = "zap"
	cfg.TransactionEvents.Attributes.Exclude[0] = "zap"
	cfg.ErrorCollector.Attributes.Include[0] = "zap"
	cfg.ErrorCollector.Attributes.Exclude[0] = "zap"
	cfg.TransactionTracer.Attributes.Include[0] = "zap"
	cfg.TransactionTracer.Attributes.Exclude[0] = "zap"

	expect := internal.CompactJSONString(`[
	{
		"pid":123,
		"language":"go",
		"agent_version":"0.2.2",
		"host":"my-hostname",
		"settings":{
			"AppName":"my appname",
			"Attributes":{"Enabled":true,"Exclude":["2"],"Include":["1"]},
			"CrossApplicationTracer":{"Enabled":true},
			"CustomInsightsEvents":{"Enabled":true},
			"DatastoreTracer":{
				"DatabaseNameReporting":{"Enabled":true},
				"InstanceReporting":{"Enabled":true},
				"QueryParameters":{"Enabled":true},
				"SlowQuery":{
					"Enabled":true,
					"Threshold":10000000
				}
			},
			"Enabled":true,
			"ErrorCollector":{
				"Attributes":{"Enabled":true,"Exclude":["6"],"Include":["5"]},
				"CaptureEvents":true,
				"Enabled":true,
				"IgnoreStatusCodes":[404,405]
			},
			"HighSecurity":false,
			"HostDisplayName":"",
			"Labels":{"zip":"zap"},
			"Logger":"*logger.logFile",
			"RuntimeSampler":{"Enabled":true},
			"TransactionEvents":{
				"Attributes":{"Enabled":true,"Exclude":["4"],"Include":["3"]},
				"Enabled":true
			},
			"TransactionTracer":{
				"Attributes":{"Enabled":true,"Exclude":["8"],"Include":["7"]},
				"Enabled":true,
				"SegmentThreshold":2000000,
				"StackTraceThreshold":500000000,
				"Threshold":{
					"Duration":500000000,
					"IsApdexFailing":true
				}
			},
			"Transport":"*http.Transport",
			"UseTLS":true,
			"Utilization":{
				"BillingHostname":"",
				"DetectAWS":true,
				"DetectAzure":true,
				"DetectDocker":true,
				"DetectGCP":true,
				"DetectPCF":true,
				"LogicalProcessors":0,
				"TotalRAMMIB":0
			}
		},
		"app_name":["my appname"],
		"high_security":false,
		"labels":[{"label_type":"zip","label_value":"zap"}],
		"environment":[
			["runtime.Compiler","comp"],
			["runtime.GOARCH","arch"],
			["runtime.GOOS","goos"],
			["runtime.Version","vers"],
			["runtime.NumCPU",8]
		],
		"identifier":"my appname",
		"utilization":{
			"metadata_version":3,
			"logical_processors":16,
			"total_ram_mib":1024,
			"hostname":"my-hostname"
		}
	}]`)

	js, err := configConnectJSONInternal(cp, 123, &utilization.SampleData, internal.SampleEnvironment, "0.2.2")
	if nil != err {
		t.Fatal(err)
	}
	out := standardizeNumbers(string(js))
	if out != expect {
		t.Error(out)
	}
}

func TestCopyConfigReferenceFieldsAbsent(t *testing.T) {
	cfg := NewConfig("my appname", "0123456789012345678901234567890123456789")
	cfg.Labels = nil
	cfg.ErrorCollector.IgnoreStatusCodes = nil

	cp := copyConfigReferenceFields(cfg)

	expect := internal.CompactJSONString(`[
	{
		"pid":123,
		"language":"go",
		"agent_version":"0.2.2",
		"host":"my-hostname",
		"settings":{
			"AppName":"my appname",
			"Attributes":{"Enabled":true,"Exclude":null,"Include":null},
			"CrossApplicationTracer":{"Enabled":true},
			"CustomInsightsEvents":{"Enabled":true},
			"DatastoreTracer":{
				"DatabaseNameReporting":{"Enabled":true},
				"InstanceReporting":{"Enabled":true},
				"QueryParameters":{"Enabled":true},
				"SlowQuery":{
					"Enabled":true,
					"Threshold":10000000
				}
			},
			"Enabled":true,
			"ErrorCollector":{
				"Attributes":{"Enabled":true,"Exclude":null,"Include":null},
				"CaptureEvents":true,
				"Enabled":true,
				"IgnoreStatusCodes":null
			},
			"HighSecurity":false,
			"HostDisplayName":"",
			"Labels":null,
			"Logger":null,
			"RuntimeSampler":{"Enabled":true},
			"TransactionEvents":{
				"Attributes":{"Enabled":true,"Exclude":null,"Include":null},
				"Enabled":true
			},
			"TransactionTracer":{
				"Attributes":{"Enabled":true,"Exclude":null,"Include":null},
				"Enabled":true,
				"SegmentThreshold":2000000,
				"StackTraceThreshold":500000000,
				"Threshold":{
					"Duration":500000000,
					"IsApdexFailing":true
				}
			},
			"Transport":null,
			"UseTLS":true,
			"Utilization":{
				"BillingHostname":"",
				"DetectAWS":true,
				"DetectAzure":true,
				"DetectDocker":true,
				"DetectGCP":true,
				"DetectPCF":true,
				"LogicalProcessors":0,
				"TotalRAMMIB":0
			}
		},
		"app_name":["my appname"],
		"high_security":false,
		"environment":[
			["runtime.Compiler","comp"],
			["runtime.GOARCH","arch"],
			["runtime.GOOS","goos"],
			["runtime.Version","vers"],
			["runtime.NumCPU",8]
		],
		"identifier":"my appname",
		"utilization":{
			"metadata_version":3,
			"logical_processors":16,
			"total_ram_mib":1024,
			"hostname":"my-hostname"
		}
	}]`)

	js, err := configConnectJSONInternal(cp, 123, &utilization.SampleData, internal.SampleEnvironment, "0.2.2")
	if nil != err {
		t.Fatal(err)
	}
	out := standardizeNumbers(string(js))
	if out != expect {
		t.Error(string(js))
	}
}

func TestValidate(t *testing.T) {
	c := Config{
		License: "0123456789012345678901234567890123456789",
		AppName: "my app",
		Enabled: true,
	}
	if err := c.Validate(); nil != err {
		t.Error(err)
	}
	c = Config{
		License: "",
		AppName: "my app",
		Enabled: true,
	}
	if err := c.Validate(); err != errLicenseLen {
		t.Error(err)
	}
	c = Config{
		License: "",
		AppName: "my app",
		Enabled: false,
	}
	if err := c.Validate(); nil != err {
		t.Error(err)
	}
	c = Config{
		License: "wronglength",
		AppName: "my app",
		Enabled: true,
	}
	if err := c.Validate(); err != errLicenseLen {
		t.Error(err)
	}
	c = Config{
		License: "0123456789012345678901234567890123456789",
		AppName: "too;many;app;names",
		Enabled: true,
	}
	if err := c.Validate(); err != errAppNameLimit {
		t.Error(err)
	}
	c = Config{
		License: "0123456789012345678901234567890123456789",
		AppName: "",
		Enabled: true,
	}
	if err := c.Validate(); err != errAppNameMissing {
		t.Error(err)
	}
	c = Config{
		License:      "0123456789012345678901234567890123456789",
		AppName:      "my app",
		Enabled:      true,
		HighSecurity: true,
	}
	if err := c.Validate(); err != errHighSecurityTLS {
		t.Error(err)
	}
	c = Config{
		License:      "0123456789012345678901234567890123456789",
		AppName:      "my app",
		Enabled:      true,
		UseTLS:       true,
		HighSecurity: true,
	}
	if err := c.Validate(); err != nil {
		t.Error(err)
	}
}
