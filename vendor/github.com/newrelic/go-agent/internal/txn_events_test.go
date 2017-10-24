package internal

import (
	"encoding/json"
	"testing"
	"time"
)

func testTxnEventJSON(t *testing.T, e *TxnEvent, expect string) {
	js, err := json.Marshal(e)
	if nil != err {
		t.Error(err)
		return
	}
	expect = CompactJSONString(expect)
	if string(js) != expect {
		t.Error(string(js), expect)
	}
}

func TestTxnEventMarshal(t *testing.T) {
	testTxnEventJSON(t, &TxnEvent{
		FinalName: "myName",
		Start:     time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
		Duration:  2 * time.Second,
		Zone:      ApdexNone,
		Attrs:     nil,
	}, `[
	{
		"type":"Transaction",
		"name":"myName",
		"timestamp":1.41713646e+09,
		"duration":2
	},
	{},
	{}]`)
	testTxnEventJSON(t, &TxnEvent{
		FinalName: "myName",
		Start:     time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
		Duration:  2 * time.Second,
		Zone:      ApdexFailing,
		Attrs:     nil,
	}, `[
	{
		"type":"Transaction",
		"name":"myName",
		"timestamp":1.41713646e+09,
		"duration":2,
		"nr.apdexPerfZone":"F"
	},
	{},
	{}]`)
	testTxnEventJSON(t, &TxnEvent{
		FinalName: "myName",
		Start:     time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
		Duration:  2 * time.Second,
		Queuing:   5 * time.Second,
		Zone:      ApdexNone,
		Attrs:     nil,
	}, `[
	{
		"type":"Transaction",
		"name":"myName",
		"timestamp":1.41713646e+09,
		"duration":2,
		"queueDuration":5
	},
	{},
	{}]`)
	testTxnEventJSON(t, &TxnEvent{
		FinalName: "myName",
		Start:     time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
		Duration:  2 * time.Second,
		Queuing:   5 * time.Second,
		Zone:      ApdexNone,
		Attrs:     nil,
		DatastoreExternalTotals: DatastoreExternalTotals{
			externalCallCount:  22,
			externalDuration:   1122334 * time.Millisecond,
			datastoreCallCount: 33,
			datastoreDuration:  5566778 * time.Millisecond,
		},
	}, `[
	{
		"type":"Transaction",
		"name":"myName",
		"timestamp":1.41713646e+09,
		"duration":2,
		"queueDuration":5,
		"externalCallCount":22,
		"externalDuration":1122.334,
		"databaseCallCount":33,
		"databaseDuration":5566.778
	},
	{},
	{}]`)
}

func TestTxnEventAttributes(t *testing.T) {
	aci := sampleAttributeConfigInput
	aci.TransactionEvents.Exclude = append(aci.TransactionEvents.Exclude, "zap")
	aci.TransactionEvents.Exclude = append(aci.TransactionEvents.Exclude, hostDisplayName)
	cfg := CreateAttributeConfig(aci)
	attr := NewAttributes(cfg)
	attr.Agent.HostDisplayName = "exclude me"
	attr.Agent.RequestMethod = "GET"
	AddUserAttribute(attr, "zap", 123, DestAll)
	AddUserAttribute(attr, "zip", 456, DestAll)

	testTxnEventJSON(t, &TxnEvent{
		FinalName: "myName",
		Start:     time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
		Duration:  2 * time.Second,
		Zone:      ApdexNone,
		Attrs:     attr,
	}, `[
	{
		"type":"Transaction",
		"name":"myName",
		"timestamp":1.41713646e+09,
		"duration":2
	},
	{
		"zip":456
	},
	{
		"request.method":"GET"
	}]`)
}
