package gostub

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestStubTime(t *testing.T) {
	var timeNow = time.Now

	var fakeTime = time.Date(2015, 7, 1, 0, 0, 0, 0, time.UTC)
	StubFunc(&timeNow, fakeTime)
	expectVal(t, fakeTime, timeNow())
}

func TestReturnErr(t *testing.T) {
	var osRemove = os.Remove

	StubFunc(&osRemove, nil)
	expectVal(t, nil, osRemove("test"))

	e := errors.New("err")
	StubFunc(&osRemove, e)
	expectVal(t, e, osRemove("test"))
}

func TestStubHostname(t *testing.T) {
	var osHostname = os.Hostname

	StubFunc(&osHostname, "fakehost", nil)
	hostname, err := osHostname()
	expectVal(t, "fakehost", hostname)
	expectVal(t, nil, err)

	var errNoHost = errors.New("no hostname")
	StubFunc(&osHostname, "", errNoHost)
	hostname, err = osHostname()
	expectVal(t, "", hostname)
	expectVal(t, errNoHost, err)
}

func TestStubReturnFunc(t *testing.T) {
	var retFunc = func() func() error {
		return func() error {
			return errors.New("err")
		}
	}

	var errInception = errors.New("in limbo")
	StubFunc(&retFunc, func() error {
		return errInception
	})
	expectVal(t, errInception, retFunc()())
}

func TestStubFuncFail(t *testing.T) {
	var osHostname = os.Hostname
	var s string

	tests := []struct {
		desc     string
		toStub   interface{}
		stubVals []interface{}
		wantErr  string
	}{
		{
			desc:     "toStub is not a function",
			toStub:   &s,
			stubVals: []interface{}{"fakehost", nil},
			wantErr:  "to stub must be a pointer to a function",
		},
		{
			desc:     "toStub is not a pointer",
			toStub:   osHostname,
			stubVals: []interface{}{"fakehost", nil},
			wantErr:  "to stub must be a pointer to a function",
		},
		{
			desc:     "wrong number of stubVals",
			toStub:   &osHostname,
			stubVals: []interface{}{"fakehost"},
			wantErr:  "func type has 2 return values, but only 1 stub values provided",
		},
	}

	for _, tt := range tests {
		func() {
			defer expectPanic(t, tt.desc, tt.wantErr)
			StubFunc(tt.toStub, tt.stubVals...)
		}()
	}
}

func TestMultipleStubFuncs(t *testing.T) {
	var f1 = func() int {
		return 100
	}
	var f2 = func() int {
		return 200
	}
	var f3 = func() int {
		return 300
	}

	stubs := StubFunc(&f1, 1).StubFunc(&f2, 2)
	expectVal(t, f1(), 1)
	expectVal(t, f2(), 2)

	stubs.StubFunc(&f3, 3)
	expectVal(t, f3(), 3)

	stubs.Reset()
	expectVal(t, f1(), 100)
	expectVal(t, f2(), 200)
	expectVal(t, f3(), 300)
}
