package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/newrelic/go-agent/internal"
)

func fail(reason string) {
	fmt.Println(reason)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		fail("improper usage: ./rules path/to/reply_file input")
	}

	connectReplyFile := os.Args[1]
	name := os.Args[2]

	data, err := ioutil.ReadFile(connectReplyFile)
	if nil != err {
		fail(fmt.Sprintf("unable to open '%s': %s", connectReplyFile, err))
	}

	var reply internal.ConnectReply
	err = json.Unmarshal(data, &reply)
	if nil != err {
		fail(fmt.Sprintf("unable unmarshal reply: %s", err))
	}

	// Metric Rules
	out := reply.MetricRules.Apply(name)
	fmt.Println("metric rules applied:", out)

	// Url Rules + Txn Name Rules + Segment Term Rules

	out = internal.CreateFullTxnName(name, &reply, true)
	fmt.Println("treated as web txn name:", out)

	out = internal.CreateFullTxnName(name, &reply, false)
	fmt.Println("treated as backround txn name:", out)
}
