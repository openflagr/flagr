package internal

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

var (
	emptyStackTrace = make([]uintptr, 0)
)

func TestErrorTraceMarshal(t *testing.T) {
	he := &tracedError{
		ErrorData: ErrorData{
			When:  time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
			Stack: emptyStackTrace,
			Msg:   "my_msg",
			Klass: "my_class",
		},
		TxnEvent: TxnEvent{
			FinalName: "my_txn_name",
			CleanURL:  "my_request_uri",
			Attrs:     nil,
		},
	}
	js, err := json.Marshal(he)
	if nil != err {
		t.Error(err)
	}
	expect := CompactJSONString(`
	[
		1.41713646e+12,
		"my_txn_name",
		"my_msg",
		"my_class",
		{
			"agentAttributes":{},
			"userAttributes":{},
			"intrinsics":{},
			"stack_trace":[],
			"request_uri":"my_request_uri"
		}
	]`)
	if string(js) != expect {
		t.Error(string(js))
	}
}

func TestErrorTraceAttributes(t *testing.T) {
	aci := sampleAttributeConfigInput
	aci.ErrorCollector.Exclude = append(aci.ErrorCollector.Exclude, "zap")
	aci.ErrorCollector.Exclude = append(aci.ErrorCollector.Exclude, hostDisplayName)
	cfg := CreateAttributeConfig(aci)
	attr := NewAttributes(cfg)
	attr.Agent.HostDisplayName = "exclude me"
	attr.Agent.RequestMethod = "GET"
	AddUserAttribute(attr, "zap", 123, DestAll)
	AddUserAttribute(attr, "zip", 456, DestAll)

	he := &tracedError{
		ErrorData: ErrorData{
			When:  time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
			Stack: nil,
			Msg:   "my_msg",
			Klass: "my_class",
		},
		TxnEvent: TxnEvent{
			FinalName: "my_txn_name",
			CleanURL:  "my_request_uri",
			Attrs:     attr,
		},
	}
	js, err := json.Marshal(he)
	if nil != err {
		t.Error(err)
	}
	expect := CompactJSONString(`
	[
		1.41713646e+12,
		"my_txn_name",
		"my_msg",
		"my_class",
		{
			"agentAttributes":{"request.method":"GET"},
			"userAttributes":{"zip":456},
			"intrinsics":{},
			"request_uri":"my_request_uri"
		}
	]`)
	if string(js) != expect {
		t.Error(string(js))
	}
}

func TestErrorsLifecycle(t *testing.T) {
	ers := NewTxnErrors(5)

	when := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	ers.Add(TxnErrorFromResponseCode(when, 400))
	ers.Add(TxnErrorFromPanic(when, errors.New("oh no panic")))
	ers.Add(TxnErrorFromPanic(when, 123))
	ers.Add(TxnErrorFromPanic(when, 123))

	he := newHarvestErrors(3)
	MergeTxnErrors(&he, ers, TxnEvent{
		FinalName: "txnName",
		CleanURL:  "requestURI",
		Attrs:     nil,
	})
	js, err := he.Data("agentRunID", time.Now())
	if nil != err {
		t.Error(err)
	}
	expect := CompactJSONString(`
[
   "agentRunID",
   [
      [
         1.41713646e+12,
         "txnName",
         "Bad Request",
         "400",
         {
            "agentAttributes":{},
            "userAttributes":{},
            "intrinsics":{},
            "request_uri":"requestURI"
         }
      ],
      [
         1.41713646e+12,
         "txnName",
         "oh no panic",
         "panic",
         {
            "agentAttributes":{},
            "userAttributes":{},
            "intrinsics":{},
            "request_uri":"requestURI"
         }
      ],
      [
         1.41713646e+12,
         "txnName",
         "123",
         "panic",
         {
            "agentAttributes":{},
            "userAttributes":{},
            "intrinsics":{},
            "request_uri":"requestURI"
         }
      ]
   ]
]`)
	if string(js) != expect {
		t.Error(string(js), expect)
	}
}

func BenchmarkErrorsJSON(b *testing.B) {
	when := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	max := 20
	ers := NewTxnErrors(max)

	for i := 0; i < max; i++ {
		ers.Add(ErrorData{
			When:  time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
			Msg:   "error message",
			Klass: "error class",
		})
	}

	cfg := CreateAttributeConfig(sampleAttributeConfigInput)
	attr := NewAttributes(cfg)
	attr.Agent.RequestMethod = "GET"
	AddUserAttribute(attr, "zip", 456, DestAll)

	he := newHarvestErrors(max)
	MergeTxnErrors(&he, ers, TxnEvent{
		FinalName: "WebTransaction/Go/hello",
		CleanURL:  "/url",
		Attrs:     attr,
	})

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		js, err := he.Data("agentRundID", when)
		if nil != err || nil == js {
			b.Fatal(err, js)
		}
	}
}
