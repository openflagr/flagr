package gostub

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

const productionData = "production data"

var dataReader io.Reader = bytes.NewReader([]byte(productionData))

func TestAssignableStub(t *testing.T) {
	const testData = "test data"
	stubs := Stub(&dataReader, strings.NewReader(testData))
	defer stubs.Reset()

	got, err := ioutil.ReadAll(dataReader)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if !bytes.Equal(got, []byte(testData)) {
		t.Errorf("Got unexpected data:\n  got %v\n want %v", string(got), string(testData))
	}
}

func TestUnassignableStub(t *testing.T) {
	defer expectPanic(t, "Stub wrong type", "not assignable")
	var noInterface interface{} = "test"
	Stub(&dataReader, noInterface)
}
