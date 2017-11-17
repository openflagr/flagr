package internal

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

type AttributeTestcase struct {
	Testname string `json:"testname"`
	Config   struct {
		AttributesEnabled        bool     `json:"attributes.enabled"`
		AttributesInclude        []string `json:"attributes.include"`
		AttributesExclude        []string `json:"attributes.exclude"`
		BrowserAttributesEnabled bool     `json:"browser_monitoring.attributes.enabled"`
		BrowserAttributesInclude []string `json:"browser_monitoring.attributes.include"`
		BrowserAttributesExclude []string `json:"browser_monitoring.attributes.exclude"`
		ErrorAttributesEnabled   bool     `json:"error_collector.attributes.enabled"`
		ErrorAttributesInclude   []string `json:"error_collector.attributes.include"`
		ErrorAttributesExclude   []string `json:"error_collector.attributes.exclude"`
		EventsAttributesEnabled  bool     `json:"transaction_events.attributes.enabled"`
		EventsAttributesInclude  []string `json:"transaction_events.attributes.include"`
		EventsAttributesExclude  []string `json:"transaction_events.attributes.exclude"`
		TracerAttributesEnabled  bool     `json:"transaction_tracer.attributes.enabled"`
		TracerAttributesInclude  []string `json:"transaction_tracer.attributes.include"`
		TracerAttributesExclude  []string `json:"transaction_tracer.attributes.exclude"`
	} `json:"config"`
	Key                  string   `json:"input_key"`
	InputDestinations    []string `json:"input_default_destinations"`
	ExpectedDestinations []string `json:"expected_destinations"`
}

var (
	destTranslate = map[string]destinationSet{
		"attributes":         DestAll,
		"transaction_events": destTxnEvent,
		"transaction_tracer": destTxnTrace,
		"error_collector":    destError,
		"browser_monitoring": destBrowser,
	}
)

func destinationsFromArray(dests []string) destinationSet {
	d := destNone
	for _, s := range dests {
		if x, ok := destTranslate[s]; ok {
			d |= x
		}
	}
	return d
}

func destToString(d destinationSet) string {
	if 0 == d {
		return "none"
	}
	out := ""
	for _, ds := range []struct {
		Name string
		Dest destinationSet
	}{
		{Name: "event", Dest: destTxnEvent},
		{Name: "trace", Dest: destTxnTrace},
		{Name: "error", Dest: destError},
		{Name: "browser", Dest: destBrowser},
	} {
		if 0 != d&ds.Dest {
			if "" == out {
				out = ds.Name
			} else {
				out = out + "," + ds.Name
			}
		}
	}
	return out
}

func runAttributeTestcase(t *testing.T, js json.RawMessage) {
	var tc AttributeTestcase

	tc.Config.AttributesEnabled = true
	tc.Config.BrowserAttributesEnabled = false
	tc.Config.ErrorAttributesEnabled = true
	tc.Config.EventsAttributesEnabled = true
	tc.Config.TracerAttributesEnabled = true

	if err := json.Unmarshal(js, &tc); nil != err {
		t.Error(err)
		return
	}

	input := AttributeConfigInput{
		Attributes: AttributeDestinationConfig{
			Enabled: tc.Config.AttributesEnabled,
			Include: tc.Config.AttributesInclude,
			Exclude: tc.Config.AttributesExclude,
		},
		ErrorCollector: AttributeDestinationConfig{
			Enabled: tc.Config.ErrorAttributesEnabled,
			Include: tc.Config.ErrorAttributesInclude,
			Exclude: tc.Config.ErrorAttributesExclude,
		},
		TransactionEvents: AttributeDestinationConfig{
			Enabled: tc.Config.EventsAttributesEnabled,
			Include: tc.Config.EventsAttributesInclude,
			Exclude: tc.Config.EventsAttributesExclude,
		},
		browserMonitoring: AttributeDestinationConfig{
			Enabled: tc.Config.BrowserAttributesEnabled,
			Include: tc.Config.BrowserAttributesInclude,
			Exclude: tc.Config.BrowserAttributesExclude,
		},
		TransactionTracer: AttributeDestinationConfig{
			Enabled: tc.Config.TracerAttributesEnabled,
			Include: tc.Config.TracerAttributesInclude,
			Exclude: tc.Config.TracerAttributesExclude,
		},
	}

	cfg := CreateAttributeConfig(input)

	inputDests := destinationsFromArray(tc.InputDestinations)
	expectedDests := destinationsFromArray(tc.ExpectedDestinations)

	out := applyAttributeConfig(cfg, tc.Key, inputDests)

	if out != expectedDests {
		t.Error(tc.Testname, destToString(expectedDests),
			destToString(out))
	}
}

func TestCrossAgentAttributes(t *testing.T) {
	var tcs []json.RawMessage

	err := crossagent.ReadJSON("attribute_configuration.json", &tcs)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range tcs {
		runAttributeTestcase(t, tc)
	}
}

func TestWriteAttributeValueJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	w := jsonFieldsWriter{buf: buf}

	buf.WriteByte('{')
	writeAttributeValueJSON(&w, "a", nil)
	writeAttributeValueJSON(&w, "a", `escape\me!`)
	writeAttributeValueJSON(&w, "a", true)
	writeAttributeValueJSON(&w, "a", false)
	writeAttributeValueJSON(&w, "a", uint8(1))
	writeAttributeValueJSON(&w, "a", uint16(2))
	writeAttributeValueJSON(&w, "a", uint32(3))
	writeAttributeValueJSON(&w, "a", uint64(4))
	writeAttributeValueJSON(&w, "a", uint(5))
	writeAttributeValueJSON(&w, "a", uintptr(6))
	writeAttributeValueJSON(&w, "a", int8(-1))
	writeAttributeValueJSON(&w, "a", int16(-2))
	writeAttributeValueJSON(&w, "a", int32(-3))
	writeAttributeValueJSON(&w, "a", int64(-4))
	writeAttributeValueJSON(&w, "a", int(-5))
	writeAttributeValueJSON(&w, "a", float32(1.5))
	writeAttributeValueJSON(&w, "a", float64(4.56))
	buf.WriteByte('}')

	expect := CompactJSONString(`{
		"a":null,
		"a":"escape\\me!",
		"a":true,
		"a":false,
		"a":1,
		"a":2,
		"a":3,
		"a":4,
		"a":5,
		"a":6,
		"a":-1,
		"a":-2,
		"a":-3,
		"a":-4,
		"a":-5,
		"a":1.5,
		"a":4.56
		}`)
	js := buf.String()
	if js != expect {
		t.Error(js, expect)
	}
}

func TestUserAttributeValLength(t *testing.T) {
	cfg := CreateAttributeConfig(sampleAttributeConfigInput)
	attrs := NewAttributes(cfg)

	atLimit := strings.Repeat("a", attributeValueLengthLimit)
	tooLong := atLimit + "a"

	err := AddUserAttribute(attrs, `escape\me`, tooLong, DestAll)
	if err != nil {
		t.Error(err)
	}
	js := userAttributesStringJSON(attrs, DestAll, nil)
	if `{"escape\\me":"`+atLimit+`"}` != js {
		t.Error(js)
	}
}

func TestUserAttributeKeyLength(t *testing.T) {
	cfg := CreateAttributeConfig(sampleAttributeConfigInput)
	attrs := NewAttributes(cfg)

	lengthyKey := strings.Repeat("a", attributeKeyLengthLimit+1)
	err := AddUserAttribute(attrs, lengthyKey, 123, DestAll)
	if _, ok := err.(invalidAttributeKeyErr); !ok {
		t.Error(err)
	}
	js := userAttributesStringJSON(attrs, DestAll, nil)
	if `{}` != js {
		t.Error(js)
	}
}

func TestNumUserAttributesLimit(t *testing.T) {
	cfg := CreateAttributeConfig(sampleAttributeConfigInput)
	attrs := NewAttributes(cfg)

	for i := 0; i < attributeUserLimit; i++ {
		s := strconv.Itoa(i)
		err := AddUserAttribute(attrs, s, s, DestAll)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := AddUserAttribute(attrs, "cant_add_me", 123, DestAll)
	if _, ok := err.(userAttributeLimitErr); !ok {
		t.Fatal(err)
	}

	js := userAttributesStringJSON(attrs, DestAll, nil)
	var out map[string]string
	err = json.Unmarshal([]byte(js), &out)
	if nil != err {
		t.Fatal(err)
	}
	if len(out) != attributeUserLimit {
		t.Error(len(out))
	}
	if strings.Contains(js, "cant_add_me") {
		t.Fatal(js)
	}

	// Now test that replacement works when the limit is reached.
	err = AddUserAttribute(attrs, "0", "BEEN_REPLACED", DestAll)
	if nil != err {
		t.Fatal(err)
	}
	js = userAttributesStringJSON(attrs, DestAll, nil)
	if !strings.Contains(js, "BEEN_REPLACED") {
		t.Fatal(js)
	}
}

func TestExtraAttributesIncluded(t *testing.T) {
	cfg := CreateAttributeConfig(sampleAttributeConfigInput)
	attrs := NewAttributes(cfg)

	err := AddUserAttribute(attrs, "a", 1, DestAll)
	if nil != err {
		t.Error(err)
	}
	js := userAttributesStringJSON(attrs, DestAll, map[string]interface{}{"b": 2})
	if `{"b":2,"a":1}` != js {
		t.Error(js)
	}
}

func TestExtraAttributesPrecedence(t *testing.T) {
	cfg := CreateAttributeConfig(sampleAttributeConfigInput)
	attrs := NewAttributes(cfg)

	err := AddUserAttribute(attrs, "a", 1, DestAll)
	if nil != err {
		t.Error(err)
	}
	js := userAttributesStringJSON(attrs, DestAll, map[string]interface{}{"a": 2})
	if `{"a":2}` != js {
		t.Error(js)
	}
}
