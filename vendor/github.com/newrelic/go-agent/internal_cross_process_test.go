package newrelic

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/newrelic/go-agent/internal"
	"github.com/newrelic/go-agent/internal/cat"
)

var (
	crossProcessReplyFn = func(reply *internal.ConnectReply) {
		reply.EncodingKey = "encoding_key"
		reply.CrossProcessID = "12345#67890"
		reply.TrustedAccounts = map[int]struct{}{
			12345: struct{}{},
		}
	}
)

func inboundHeaders(t *testing.T) http.Header {
	app := testApp(crossProcessReplyFn, nil, t)
	clientTxn := app.StartTransaction("client", nil, nil)
	req, err := http.NewRequest("GET", "newrelic.com", nil)
	if nil != err {
		t.Fatal(err)
	}
	StartExternalSegment(clientTxn, req)
	if "" == req.Header.Get(cat.NewRelicIDName) {
		t.Fatal(req.Header.Get(cat.NewRelicIDName))
	}
	if "" == req.Header.Get(cat.NewRelicTxnName) {
		t.Fatal(req.Header.Get(cat.NewRelicTxnName))
	}
	return req.Header
}

var (
	inboundCrossProcessRequest = func() *http.Request {
		app := testApp(crossProcessReplyFn, nil, nil)
		clientTxn := app.StartTransaction("client", nil, nil)
		req, err := http.NewRequest("GET", "newrelic.com", nil)
		StartExternalSegment(clientTxn, req)
		if "" == req.Header.Get(cat.NewRelicIDName) {
			panic("missing cat header NewRelicIDName: " + req.Header.Get(cat.NewRelicIDName))
		}
		if "" == req.Header.Get(cat.NewRelicTxnName) {
			panic("missing cat header NewRelicTxnName: " + req.Header.Get(cat.NewRelicTxnName))
		}
		if nil != err {
			panic(err)
		}
		return req
	}()
	catIntrinsics = map[string]interface{}{
		"name":                        "WebTransaction/Go/hello",
		"nr.pathHash":                 "fa013f2a",
		"nr.guid":                     internal.MatchAnything,
		"nr.referringTransactionGuid": internal.MatchAnything,
		"nr.referringPathHash":        "41c04f7d",
		"nr.apdexPerfZone":            "S",
		"client_cross_process_id":     "12345#67890",
		"nr.tripId":                   internal.MatchAnything,
	}
)

func TestCrossProcessWriteHeaderSuccess(t *testing.T) {
	// Test that the CAT response header is present when the consumer uses
	// txn.WriteHeader.
	app := testApp(crossProcessReplyFn, nil, t)
	w := httptest.NewRecorder()
	txn := app.StartTransaction("hello", w, inboundCrossProcessRequest)
	txn.WriteHeader(200)
	txn.End()

	if "" == w.Header().Get(cat.NewRelicAppDataName) {
		t.Error(w.Header().Get(cat.NewRelicAppDataName))
	}

	app.ExpectMetrics(t, webMetrics)
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: catIntrinsics,
		AgentAttributes: map[string]interface{}{
			"request.method":   "GET",
			"httpResponseCode": 200,
		},
		UserAttributes: map[string]interface{}{},
	}})
}

func TestCrossProcessWriteSuccess(t *testing.T) {
	// Test that the CAT response header is present when the consumer uses
	// txn.Write.
	app := testApp(crossProcessReplyFn, nil, t)
	w := httptest.NewRecorder()
	txn := app.StartTransaction("hello", w, inboundCrossProcessRequest)
	txn.Write([]byte("response text"))
	txn.End()

	if "" == w.Header().Get(cat.NewRelicAppDataName) {
		t.Error(w.Header().Get(cat.NewRelicAppDataName))
	}

	app.ExpectMetrics(t, webMetrics)
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: catIntrinsics,
		// Do not test attributes here:  In Go 1.5
		// response.headers.contentType will be not be present.
		AgentAttributes: nil,
		UserAttributes:  map[string]interface{}{},
	}})
}
