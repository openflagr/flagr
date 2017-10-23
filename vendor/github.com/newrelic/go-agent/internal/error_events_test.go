package internal

import (
	"encoding/json"
	"testing"
	"time"
)

func testErrorEventJSON(t *testing.T, e *ErrorEvent, expect string) {
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

var (
	sampleErrorData = ErrorData{
		Klass: "*errors.errorString",
		Msg:   "hello",
		When:  time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
	}
)

func TestErrorEventMarshal(t *testing.T) {
	testErrorEventJSON(t, &ErrorEvent{
		ErrorData: sampleErrorData,
		TxnEvent: TxnEvent{
			FinalName: "myName",
			Duration:  3 * time.Second,
			Attrs:     nil,
		},
	}, `[
		{
			"type":"TransactionError",
			"error.class":"*errors.errorString",
			"error.message":"hello",
			"timestamp":1.41713646e+09,
			"transactionName":"myName",
			"duration":3
		},
		{},
		{}
	]`)
	testErrorEventJSON(t, &ErrorEvent{
		ErrorData: sampleErrorData,
		TxnEvent: TxnEvent{
			FinalName: "myName",
			Duration:  3 * time.Second,
			Queuing:   5 * time.Second,
			Attrs:     nil,
		},
	}, `[
		{
			"type":"TransactionError",
			"error.class":"*errors.errorString",
			"error.message":"hello",
			"timestamp":1.41713646e+09,
			"transactionName":"myName",
			"duration":3,
			"queueDuration":5
		},
		{},
		{}
	]`)
	testErrorEventJSON(t, &ErrorEvent{
		ErrorData: sampleErrorData,
		TxnEvent: TxnEvent{
			FinalName: "myName",
			Duration:  3 * time.Second,
			Queuing:   5 * time.Second,
			DatastoreExternalTotals: DatastoreExternalTotals{
				externalCallCount:  22,
				externalDuration:   1122334 * time.Millisecond,
				datastoreCallCount: 33,
				datastoreDuration:  5566778 * time.Millisecond,
			},
		},
	}, `[
		{
			"type":"TransactionError",
			"error.class":"*errors.errorString",
			"error.message":"hello",
			"timestamp":1.41713646e+09,
			"transactionName":"myName",
			"duration":3,
			"queueDuration":5,
			"externalCallCount":22,
			"externalDuration":1122.334,
			"databaseCallCount":33,
			"databaseDuration":5566.778
		},
		{},
		{}
	]`)
}

func TestErrorEventAttributes(t *testing.T) {
	aci := sampleAttributeConfigInput
	aci.ErrorCollector.Exclude = append(aci.ErrorCollector.Exclude, "zap")
	aci.ErrorCollector.Exclude = append(aci.ErrorCollector.Exclude, hostDisplayName)
	cfg := CreateAttributeConfig(aci)
	attr := NewAttributes(cfg)
	attr.Agent.HostDisplayName = "exclude me"
	attr.Agent.RequestMethod = "GET"
	AddUserAttribute(attr, "zap", 123, DestAll)
	AddUserAttribute(attr, "zip", 456, DestAll)

	testErrorEventJSON(t, &ErrorEvent{
		ErrorData: sampleErrorData,
		TxnEvent: TxnEvent{
			FinalName: "myName",
			Duration:  3 * time.Second,
			Attrs:     attr,
		},
	}, `[
		{
			"type":"TransactionError",
			"error.class":"*errors.errorString",
			"error.message":"hello",
			"timestamp":1.41713646e+09,
			"transactionName":"myName",
			"duration":3
		},
		{
			"zip":456
		},
		{
			"request.method":"GET"
		}
	]`)
}
