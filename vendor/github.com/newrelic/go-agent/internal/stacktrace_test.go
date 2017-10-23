package internal

import (
	"encoding/json"
	"testing"
)

func TestGetStackTrace(t *testing.T) {
	stack := GetStackTrace(0)
	js, err := json.Marshal(stack)
	if nil != err {
		t.Fatal(err)
	}
	if nil == js {
		t.Fatal(string(js))
	}
}

func TestLongStackTrace(t *testing.T) {
	st := StackTrace(make([]uintptr, maxStackTraceFrames+20))
	js, err := json.Marshal(st)
	if nil != err {
		t.Fatal(err)
	}
	expect := `[
	{},{},{},{},{},{},{},{},{},{},
	{},{},{},{},{},{},{},{},{},{},
	{},{},{},{},{},{},{},{},{},{},
	{},{},{},{},{},{},{},{},{},{},
	{},{},{},{},{},{},{},{},{},{},
	{},{},{},{},{},{},{},{},{},{},
	{},{},{},{},{},{},{},{},{},{},
	{},{},{},{},{},{},{},{},{},{},
	{},{},{},{},{},{},{},{},{},{},
	{},{},{},{},{},{},{},{},{},{}
	]`
	if string(js) != CompactJSONString(expect) {
		t.Error(string(js))
	}
}
