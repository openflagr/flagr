package gostub

import "testing"

// Variables used in stubbing.
var v1, v2, v3, v4 int

// resetVars is used to reset the variables used in stubbing tests to their default values.
func resetVars() {
	v1 = 100
	v2 = 200
	v3 = 300
	v4 = 400
}

func TestStub(t *testing.T) {
	resetVars()

	stubs := Stub(&v1, 1)

	if v1 != 1 {
		t.Errorf("expected")
	}
	expectVal(t, v1, 1)
	stubs.Reset()
	expectVal(t, v1, 100)
}

func TestRestub(t *testing.T) {
	resetVars()

	stubs := Stub(&v1, 1)
	expectVal(t, v1, 1)
	stubs.Stub(&v1, 2)
	expectVal(t, v1, 2)
	stubs.Reset()
	expectVal(t, v1, 100)
}

func TestResetSingle(t *testing.T) {
	resetVars()

	stubs := Stub(&v1, 1).Stub(&v2, 2)
	expectVal(t, v1, 1)
	expectVal(t, v2, 2)

	stubs.ResetSingle(&v1)
	expectVal(t, v1, 100)
	expectVal(t, v2, 2)

	stubs.Reset()
	expectVal(t, v1, 100)
	expectVal(t, v2, 200)
}

func TestResetSingleNotStubbed(t *testing.T) {
	resetVars()

	stubs := Stub(&v1, 1)
	expectVal(t, v1, 1)

	defer expectPanic(t, "ResetSingle unstubbed variable", "not been stubbed")
	stubs.ResetSingle(&v2)
}

func TestResetTwice(t *testing.T) {
	resetVars()

	stubs := Stub(&v1, 1)
	expectVal(t, v1, 1)

	stubs.Reset()
	expectVal(t, v1, 100)

	stubs.Stub(&v1, 2)
	expectVal(t, v1, 2)

	stubs.Reset()
	expectVal(t, v1, 100)
}

func TestMultipleStubs(t *testing.T) {
	resetVars()

	stubs := Stub(&v1, 1).Stub(&v2, 2).Stub(&v3, 3)
	expectVal(t, v1, 1)
	expectVal(t, v2, 2)
	expectVal(t, v3, 3)
	expectVal(t, v4, 400)

	stubs.Stub(&v4, 4)
	expectVal(t, v4, 4)

	stubs.Reset()
	expectVal(t, v1, 100)
	expectVal(t, v2, 200)
	expectVal(t, v3, 300)
	expectVal(t, v4, 400)
}

func TestVarNotPtr(t *testing.T) {
	defer expectPanic(t, "Stub non-pointer", "expected to be a pointer")
	Stub(v1, 1)
}

func TestTypeMismatch(t *testing.T) {
	defer expectPanic(t, "Stub wrong type", "not assignable")
	Stub(&v1, "test")
}
