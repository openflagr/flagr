package gostub_test

import (
	"fmt"
	"os"

	"github.com/prashantv/gostub"
)

func ExampleStubFunc() {
	var osHostname = os.Hostname

	defer gostub.StubFunc(&osHostname, "fakehost", nil).Reset()
	host, err := osHostname()

	fmt.Println("Host:", host, "err:", err)
	// Output:
	// Host: fakehost err: <nil>
}

func ExampleStub() {
	var counter = 100

	defer gostub.Stub(&counter, 200).Reset()
	fmt.Println("Counter:", counter)
	// Output:
	// Counter: 200
}
