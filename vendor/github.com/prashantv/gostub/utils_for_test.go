package gostub

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func expectVal(t *testing.T, expected interface{}, got interface{}) {
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %v but got %v", expected, got)
	}
}

func expectPanic(t *testing.T, msg string, expectedPanic string) {
	if r := recover(); r != nil {
		got := fmt.Sprint(r)
		if !strings.Contains(got, expectedPanic) {
			t.Errorf("panic message expected to contain %q, got %v", expectedPanic, got)
		}
		return
	}

	t.Errorf("%v expected to panic", msg)
}
