package internal

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

var (
	now       = time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	strLen512 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	strLen255 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

// Tests use a single key-value pair in params to ensure deterministic JSON
// ordering.

func TestCreateCustomEventSuccess(t *testing.T) {
	event, err := CreateCustomEvent("myEvent", map[string]interface{}{"alpha": 1}, now)
	if nil != err {
		t.Fatal(err)
	}
	js, err := json.Marshal(event)
	if nil != err {
		t.Fatal(err)
	}
	if string(js) != `[{"type":"myEvent","timestamp":1.41713646e+09},{"alpha":1},{}]` {
		t.Fatal(string(js))
	}
}

func TestInvalidEventTypeCharacter(t *testing.T) {
	event, err := CreateCustomEvent("myEvent!", map[string]interface{}{"alpha": 1}, now)
	if err != ErrEventTypeRegex {
		t.Fatal(err)
	}
	if nil != event {
		t.Fatal(event)
	}
}

func TestLongEventType(t *testing.T) {
	event, err := CreateCustomEvent(strLen512, map[string]interface{}{"alpha": 1}, now)
	if err != errEventTypeLength {
		t.Fatal(err)
	}
	if nil != event {
		t.Fatal(event)
	}
}

func TestNilParams(t *testing.T) {
	event, err := CreateCustomEvent("myEvent", nil, now)
	if nil != err {
		t.Fatal(err)
	}
	js, err := json.Marshal(event)
	if nil != err {
		t.Fatal(err)
	}
	if string(js) != `[{"type":"myEvent","timestamp":1.41713646e+09},{},{}]` {
		t.Fatal(string(js))
	}
}

func TestMissingEventType(t *testing.T) {
	event, err := CreateCustomEvent("", map[string]interface{}{"alpha": 1}, now)
	if err != ErrEventTypeRegex {
		t.Fatal(err)
	}
	if nil != event {
		t.Fatal(event)
	}
}

func TestEmptyParams(t *testing.T) {
	event, err := CreateCustomEvent("myEvent", map[string]interface{}{}, now)
	if nil != err {
		t.Fatal(err)
	}
	js, err := json.Marshal(event)
	if nil != err {
		t.Fatal(err)
	}
	if string(js) != `[{"type":"myEvent","timestamp":1.41713646e+09},{},{}]` {
		t.Fatal(string(js))
	}
}

func TestTruncatedStringValue(t *testing.T) {
	event, err := CreateCustomEvent("myEvent", map[string]interface{}{"alpha": strLen512}, now)
	if nil != err {
		t.Fatal(err)
	}
	js, err := json.Marshal(event)
	if nil != err {
		t.Fatal(err)
	}
	if string(js) != `[{"type":"myEvent","timestamp":1.41713646e+09},{"alpha":"`+strLen255+`"},{}]` {
		t.Fatal(string(js))
	}
}

func TestInvalidValueType(t *testing.T) {
	event, err := CreateCustomEvent("myEvent", map[string]interface{}{"alpha": []string{}}, now)
	if _, ok := err.(ErrInvalidAttributeType); !ok {
		t.Fatal(err)
	}
	if nil != event {
		t.Fatal(event)
	}
}

func TestInvalidCustomAttributeKey(t *testing.T) {
	event, err := CreateCustomEvent("myEvent", map[string]interface{}{strLen512: 1}, now)
	if nil == err {
		t.Fatal(err)
	}
	if _, ok := err.(invalidAttributeKeyErr); !ok {
		t.Fatal(err)
	}
	if nil != event {
		t.Fatal(event)
	}
}

func TestTooManyAttributes(t *testing.T) {
	params := make(map[string]interface{})
	for i := 0; i < customEventAttributeLimit+1; i++ {
		params[strconv.Itoa(i)] = i
	}
	event, err := CreateCustomEvent("myEvent", params, now)
	if errNumAttributes != err {
		t.Fatal(err)
	}
	if nil != event {
		t.Fatal(event)
	}
}

func TestCustomEventAttributeTypes(t *testing.T) {
	testcases := []struct {
		val interface{}
		js  string
	}{
		{"string", `"string"`},
		{true, `true`},
		{false, `false`},
		{nil, `null`},
		{uint8(1), `1`},
		{uint16(1), `1`},
		{uint32(1), `1`},
		{uint64(1), `1`},
		{int8(1), `1`},
		{int16(1), `1`},
		{int32(1), `1`},
		{int64(1), `1`},
		{float32(1), `1`},
		{float64(1), `1`},
		{uint(1), `1`},
		{int(1), `1`},
		{uintptr(1), `1`},
	}

	for _, tc := range testcases {
		event, err := CreateCustomEvent("myEvent", map[string]interface{}{"key": tc.val}, now)
		if nil != err {
			t.Fatal(err)
		}
		js, err := json.Marshal(event)
		if nil != err {
			t.Fatal(err)
		}
		if string(js) != `[{"type":"myEvent","timestamp":1.41713646e+09},{"key":`+tc.js+`},{}]` {
			t.Fatal(string(js))
		}
	}
}

func TestCustomParamsCopied(t *testing.T) {
	params := map[string]interface{}{"alpha": 1}
	event, err := CreateCustomEvent("myEvent", params, now)
	if nil != err {
		t.Fatal(err)
	}
	// Attempt to change the params after the event created:
	params["zip"] = "zap"
	js, err := json.Marshal(event)
	if nil != err {
		t.Fatal(err)
	}
	if string(js) != `[{"type":"myEvent","timestamp":1.41713646e+09},{"alpha":1},{}]` {
		t.Fatal(string(js))
	}
}

func TestMultipleAttributeJSON(t *testing.T) {
	params := map[string]interface{}{"alpha": 1, "beta": 2}
	event, err := CreateCustomEvent("myEvent", params, now)
	if nil != err {
		t.Fatal(err)
	}
	js, err := json.Marshal(event)
	if nil != err {
		t.Fatal(err)
	}
	// Params order may not be deterministic, so we simply test that the
	// JSON created is valid.
	var valid interface{}
	if err := json.Unmarshal(js, &valid); nil != err {
		t.Error(string(js))
	}
}
