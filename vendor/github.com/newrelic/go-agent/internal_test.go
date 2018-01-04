package newrelic

import (
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/newrelic/go-agent/internal"
)

var (
	singleCount = []float64{1, 0, 0, 0, 0, 0, 0}
	webMetrics  = []internal.WantMetric{
		{Name: "WebTransaction/Go/hello", Scope: "", Forced: true, Data: nil},
		{Name: "WebTransaction", Scope: "", Forced: true, Data: nil},
		{Name: "HttpDispatcher", Scope: "", Forced: true, Data: nil},
		{Name: "Apdex", Scope: "", Forced: true, Data: nil},
		{Name: "Apdex/Go/hello", Scope: "", Forced: false, Data: nil},
	}
	webErrorMetrics = append([]internal.WantMetric{
		{Name: "Errors/all", Scope: "", Forced: true, Data: singleCount},
		{Name: "Errors/allWeb", Scope: "", Forced: true, Data: singleCount},
		{Name: "Errors/WebTransaction/Go/hello", Scope: "", Forced: true, Data: singleCount},
	}, webMetrics...)
	backgroundMetrics = []internal.WantMetric{
		{Name: "OtherTransaction/Go/hello", Scope: "", Forced: true, Data: nil},
		{Name: "OtherTransaction/all", Scope: "", Forced: true, Data: nil},
	}
	backgroundErrorMetrics = append([]internal.WantMetric{
		{Name: "Errors/all", Scope: "", Forced: true, Data: singleCount},
		{Name: "Errors/allOther", Scope: "", Forced: true, Data: singleCount},
		{Name: "Errors/OtherTransaction/Go/hello", Scope: "", Forced: true, Data: singleCount},
	}, backgroundMetrics...)
)

// compatibleResponseRecorder wraps ResponseRecorder to ensure consistent behavior
// between different versions of Go.
//
// Unfortunately, there was a behavior change in go1.6:
//
// "The net/http/httptest package's ResponseRecorder now initializes a default
// Content-Type header using the same content-sniffing algorithm as in
// http.Server."
type compatibleResponseRecorder struct {
	*httptest.ResponseRecorder
	wroteHeader bool
}

func newCompatibleResponseRecorder() *compatibleResponseRecorder {
	return &compatibleResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

func (rw *compatibleResponseRecorder) Header() http.Header {
	return rw.ResponseRecorder.Header()
}

func (rw *compatibleResponseRecorder) Write(buf []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(200)
		rw.wroteHeader = true
	}
	return rw.ResponseRecorder.Write(buf)
}

func (rw *compatibleResponseRecorder) WriteHeader(code int) {
	rw.wroteHeader = true
	rw.ResponseRecorder.WriteHeader(code)
}

var (
	sampleLicense = "0123456789012345678901234567890123456789"
	validParams   = map[string]interface{}{"zip": 1, "zap": 2}
)

var (
	helloResponse    = []byte("hello")
	helloPath        = "/hello"
	helloQueryParams = "?secret=hideme"
	helloRequest     = func() *http.Request {
		r, err := http.NewRequest("GET", helloPath+helloQueryParams, nil)
		if nil != err {
			panic(err)
		}

		r.Header.Add(`Accept`, `text/plain`)
		r.Header.Add(`Content-Type`, `text/html; charset=utf-8`)
		r.Header.Add(`Content-Length`, `753`)
		r.Header.Add(`Host`, `my_domain.com`)
		r.Header.Add(`User-Agent`, `Mozilla/5.0`)
		r.Header.Add(`Referer`, `http://en.wikipedia.org/zip?secret=password`)

		return r
	}()
)

func TestNewApplicationNil(t *testing.T) {
	cfg := NewConfig("appname", "wrong length")
	cfg.Enabled = false
	app, err := NewApplication(cfg)
	if nil == err {
		t.Error("error expected when license key is short")
	}
	if nil != app {
		t.Error("app expected to be nil when error is returned")
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.Write(helloResponse)
}

func testApp(replyfn func(*internal.ConnectReply), cfgfn func(*Config), t testing.TB) expectApp {
	cfg := NewConfig("my app", "0123456789012345678901234567890123456789")

	if nil != cfgfn {
		cfgfn(&cfg)
	}

	app, err := newTestApp(replyfn, cfg)
	if nil != err {
		t.Fatal(err)
	}
	return app
}

func TestRecordCustomEventSuccess(t *testing.T) {
	app := testApp(nil, nil, t)
	err := app.RecordCustomEvent("myType", validParams)
	if nil != err {
		t.Error(err)
	}
	app.ExpectCustomEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"type":      "myType",
			"timestamp": internal.MatchAnything,
		},
		UserAttributes: validParams,
	}})
}

func TestRecordCustomEventHighSecurityEnabled(t *testing.T) {
	cfgfn := func(cfg *Config) { cfg.HighSecurity = true }
	app := testApp(nil, cfgfn, t)
	err := app.RecordCustomEvent("myType", validParams)
	if err != errHighSecurityEnabled {
		t.Error(err)
	}
	app.ExpectCustomEvents(t, []internal.WantEvent{})
}

func TestRecordCustomEventEventsDisabled(t *testing.T) {
	cfgfn := func(cfg *Config) { cfg.CustomInsightsEvents.Enabled = false }
	app := testApp(nil, cfgfn, t)
	err := app.RecordCustomEvent("myType", validParams)
	if err != errCustomEventsDisabled {
		t.Error(err)
	}
	app.ExpectCustomEvents(t, []internal.WantEvent{})
}

func TestRecordCustomEventBadInput(t *testing.T) {
	app := testApp(nil, nil, t)
	err := app.RecordCustomEvent("????", validParams)
	if err != internal.ErrEventTypeRegex {
		t.Error(err)
	}
	app.ExpectCustomEvents(t, []internal.WantEvent{})
}

func TestRecordCustomEventRemoteDisable(t *testing.T) {
	replyfn := func(reply *internal.ConnectReply) { reply.CollectCustomEvents = false }
	app := testApp(replyfn, nil, t)
	err := app.RecordCustomEvent("myType", validParams)
	if err != errCustomEventsRemoteDisabled {
		t.Error(err)
	}
	app.ExpectCustomEvents(t, []internal.WantEvent{})
}

func TestRecordCustomMetricSuccess(t *testing.T) {
	app := testApp(nil, nil, t)
	err := app.RecordCustomMetric("myMetric", 123.0)
	if nil != err {
		t.Error(err)
	}
	expectData := []float64{1, 123.0, 123.0, 123.0, 123.0, 123.0 * 123.0}
	app.ExpectMetrics(t, []internal.WantMetric{
		{Name: "Custom/myMetric", Scope: "", Forced: false, Data: expectData},
	})
}

func TestRecordCustomMetricNameEmpty(t *testing.T) {
	app := testApp(nil, nil, t)
	err := app.RecordCustomMetric("", 123.0)
	if err != errMetricNameEmpty {
		t.Error(err)
	}
}

func TestRecordCustomMetricNaN(t *testing.T) {
	app := testApp(nil, nil, t)
	err := app.RecordCustomMetric("myMetric", math.NaN())
	if err != errMetricNaN {
		t.Error(err)
	}
}

func TestRecordCustomMetricPositiveInf(t *testing.T) {
	app := testApp(nil, nil, t)
	err := app.RecordCustomMetric("myMetric", math.Inf(0))
	if err != errMetricInf {
		t.Error(err)
	}
}

func TestRecordCustomMetricNegativeInf(t *testing.T) {
	app := testApp(nil, nil, t)
	err := app.RecordCustomMetric("myMetric", math.Inf(-1))
	if err != errMetricInf {
		t.Error(err)
	}
}

type sampleResponseWriter struct {
	code    int
	written int
	header  http.Header
}

func (w *sampleResponseWriter) Header() http.Header       { return w.header }
func (w *sampleResponseWriter) Write([]byte) (int, error) { return w.written, nil }
func (w *sampleResponseWriter) WriteHeader(x int)         { w.code = x }

func TestTxnResponseWriter(t *testing.T) {
	// NOTE: Eventually when the ResponseWriter is instrumented, this test
	// should be expanded to make sure that calling ResponseWriter methods
	// after the transaction has ended is not problematic.
	w := &sampleResponseWriter{
		header: make(http.Header),
	}
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", w, nil)
	w.header.Add("zip", "zap")
	if out := txn.Header(); out.Get("zip") != "zap" {
		t.Error(out.Get("zip"))
	}
	w.written = 123
	if out, _ := txn.Write(nil); out != 123 {
		t.Error(out)
	}
	if txn.WriteHeader(503); w.code != 503 {
		t.Error(w.code)
	}
}

func TestTransactionEventWeb(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	err := txn.End()
	if nil != err {
		t.Error(err)
	}
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":             "WebTransaction/Go/hello",
			"nr.apdexPerfZone": "S",
		},
	}})
}

func TestTransactionEventBackground(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.End()
	if nil != err {
		t.Error(err)
	}
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name": "OtherTransaction/Go/hello",
		},
	}})
}

func TestTransactionEventLocallyDisabled(t *testing.T) {
	cfgFn := func(cfg *Config) { cfg.TransactionEvents.Enabled = false }
	app := testApp(nil, cfgFn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	err := txn.End()
	if nil != err {
		t.Error(err)
	}
	app.ExpectTxnEvents(t, []internal.WantEvent{})
}

func TestTransactionEventRemotelyDisabled(t *testing.T) {
	replyfn := func(reply *internal.ConnectReply) { reply.CollectAnalyticsEvents = false }
	app := testApp(replyfn, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	err := txn.End()
	if nil != err {
		t.Error(err)
	}
	app.ExpectTxnEvents(t, []internal.WantEvent{})
}

func myErrorHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("my response"))
	if txn, ok := w.(Transaction); ok {
		txn.NoticeError(myError{})
	}
}

func TestWrapHandleFunc(t *testing.T) {
	app := testApp(nil, nil, t)
	mux := http.NewServeMux()
	mux.HandleFunc(WrapHandleFunc(app, helloPath, myErrorHandler))
	w := newCompatibleResponseRecorder()
	mux.ServeHTTP(w, helloRequest)

	out := w.Body.String()
	if "my response" != out {
		t.Error(out)
	}

	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "WebTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.myError",
		Caller:  "go-agent.myErrorHandler",
		URL:     "/hello",
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "newrelic.myError",
			"error.message":   "my msg",
			"transactionName": "WebTransaction/Go/hello",
		},
	}})
	app.ExpectMetrics(t, webErrorMetrics)
}

func TestWrapHandle(t *testing.T) {
	app := testApp(nil, nil, t)
	mux := http.NewServeMux()
	mux.Handle(WrapHandle(app, helloPath, http.HandlerFunc(myErrorHandler)))
	w := newCompatibleResponseRecorder()
	mux.ServeHTTP(w, helloRequest)

	out := w.Body.String()
	if "my response" != out {
		t.Error(out)
	}

	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "WebTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.myError",
		Caller:  "go-agent.myErrorHandler",
		URL:     "/hello",
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "newrelic.myError",
			"error.message":   "my msg",
			"transactionName": "WebTransaction/Go/hello",
		},
	}})
	app.ExpectMetrics(t, webErrorMetrics)
}

func TestSetName(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("one", nil, nil)
	if err := txn.SetName("hello"); nil != err {
		t.Error(err)
	}
	txn.End()
	if err := txn.SetName("three"); err != errAlreadyEnded {
		t.Error(err)
	}

	app.ExpectMetrics(t, backgroundMetrics)
}

func deferEndPanic(txn Transaction, panicMe interface{}) (r interface{}) {
	defer func() {
		r = recover()
	}()

	defer txn.End()

	panic(panicMe)
}

func TestPanicError(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)

	e := myError{}
	r := deferEndPanic(txn, e)
	if r != e {
		t.Error("panic not propagated", r)
	}

	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   internal.PanicErrorKlass,
		Caller:  "go-agent.(*txn).End",
		URL:     "",
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     internal.PanicErrorKlass,
			"error.message":   "my msg",
			"transactionName": "OtherTransaction/Go/hello",
		},
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}

func TestPanicString(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)

	e := "my string"
	r := deferEndPanic(txn, e)
	if r != e {
		t.Error("panic not propagated", r)
	}

	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my string",
		Klass:   internal.PanicErrorKlass,
		Caller:  "go-agent.(*txn).End",
		URL:     "",
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     internal.PanicErrorKlass,
			"error.message":   "my string",
			"transactionName": "OtherTransaction/Go/hello",
		},
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}

func TestPanicInt(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)

	e := 22
	r := deferEndPanic(txn, e)
	if r != e {
		t.Error("panic not propagated", r)
	}

	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "22",
		Klass:   internal.PanicErrorKlass,
		Caller:  "go-agent.(*txn).End",
		URL:     "",
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     internal.PanicErrorKlass,
			"error.message":   "22",
			"transactionName": "OtherTransaction/Go/hello",
		},
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}

func TestPanicNil(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)

	r := deferEndPanic(txn, nil)
	if nil != r {
		t.Error(r)
	}

	app.ExpectErrors(t, []internal.WantError{})
	app.ExpectErrorEvents(t, []internal.WantEvent{})
	app.ExpectMetrics(t, backgroundMetrics)
}

func TestResponseCodeError(t *testing.T) {
	app := testApp(nil, nil, t)
	w := newCompatibleResponseRecorder()
	txn := app.StartTransaction("hello", w, helloRequest)

	txn.WriteHeader(http.StatusBadRequest)   // 400
	txn.WriteHeader(http.StatusUnauthorized) // 401

	txn.End()

	if http.StatusBadRequest != w.Code {
		t.Error(w.Code)
	}

	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "WebTransaction/Go/hello",
		Msg:     "Bad Request",
		Klass:   "400",
		Caller:  "go-agent.(*txn).WriteHeader",
		URL:     "/hello",
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "400",
			"error.message":   "Bad Request",
			"transactionName": "WebTransaction/Go/hello",
		},
	}})
	app.ExpectMetrics(t, webErrorMetrics)
}

func TestResponseCode404Filtered(t *testing.T) {
	app := testApp(nil, nil, t)
	w := newCompatibleResponseRecorder()
	txn := app.StartTransaction("hello", w, helloRequest)

	txn.WriteHeader(http.StatusNotFound)

	txn.End()

	if http.StatusNotFound != w.Code {
		t.Error(w.Code)
	}

	app.ExpectErrors(t, []internal.WantError{})
	app.ExpectErrorEvents(t, []internal.WantEvent{})
	app.ExpectMetrics(t, webMetrics)
}

func TestResponseCodeCustomFilter(t *testing.T) {
	cfgFn := func(cfg *Config) {
		cfg.ErrorCollector.IgnoreStatusCodes =
			append(cfg.ErrorCollector.IgnoreStatusCodes,
				http.StatusNotFound)
	}
	app := testApp(nil, cfgFn, t)
	w := newCompatibleResponseRecorder()
	txn := app.StartTransaction("hello", w, helloRequest)

	txn.WriteHeader(http.StatusNotFound)

	txn.End()

	app.ExpectErrors(t, []internal.WantError{})
	app.ExpectErrorEvents(t, []internal.WantEvent{})
	app.ExpectMetrics(t, webMetrics)
}

func TestResponseCodeAfterEnd(t *testing.T) {
	app := testApp(nil, nil, t)
	w := newCompatibleResponseRecorder()
	txn := app.StartTransaction("hello", w, helloRequest)

	txn.End()
	txn.WriteHeader(http.StatusBadRequest)

	if http.StatusBadRequest != w.Code {
		t.Error(w.Code)
	}

	app.ExpectErrors(t, []internal.WantError{})
	app.ExpectErrorEvents(t, []internal.WantEvent{})
	app.ExpectMetrics(t, webMetrics)
}

func TestResponseCodeAfterWrite(t *testing.T) {
	app := testApp(nil, nil, t)
	w := newCompatibleResponseRecorder()
	txn := app.StartTransaction("hello", w, helloRequest)

	txn.Write([]byte("zap"))
	txn.WriteHeader(http.StatusBadRequest)

	txn.End()

	if out := w.Body.String(); "zap" != out {
		t.Error(out)
	}

	if http.StatusOK != w.Code {
		t.Error(w.Code)
	}

	app.ExpectErrors(t, []internal.WantError{})
	app.ExpectErrorEvents(t, []internal.WantEvent{})
	app.ExpectMetrics(t, webMetrics)
}

func TestQueueTime(t *testing.T) {
	app := testApp(nil, nil, t)
	req, err := http.NewRequest("GET", helloPath+helloQueryParams, nil)
	req.Header.Add("X-Queue-Start", "1465793282.12345")
	if nil != err {
		t.Fatal(err)
	}
	txn := app.StartTransaction("hello", nil, req)
	txn.NoticeError(myError{})
	txn.End()

	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "WebTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.myError",
		Caller:  "go-agent.TestQueueTime",
		URL:     "/hello",
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "newrelic.myError",
			"error.message":   "my msg",
			"transactionName": "WebTransaction/Go/hello",
			"queueDuration":   internal.MatchAnything,
		},
	}})
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "WebFrontend/QueueTime", Scope: "", Forced: true, Data: nil},
	}, webErrorMetrics...))
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":             "WebTransaction/Go/hello",
			"nr.apdexPerfZone": "F",
			"queueDuration":    internal.MatchAnything,
		},
		AgentAttributes: nil,
	}})
}

func TestIgnore(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	txn.NoticeError(myError{})
	err := txn.Ignore()
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{})
	app.ExpectErrorEvents(t, []internal.WantEvent{})
	app.ExpectMetrics(t, []internal.WantMetric{})
	app.ExpectTxnEvents(t, []internal.WantEvent{})
}

func TestIgnoreAlreadyEnded(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	txn.NoticeError(myError{})
	txn.End()
	err := txn.Ignore()
	if err != errAlreadyEnded {
		t.Error(err)
	}
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "my msg",
		Klass:   "newrelic.myError",
		Caller:  "go-agent.TestIgnoreAlreadyEnded",
		URL:     "",
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "newrelic.myError",
			"error.message":   "my msg",
			"transactionName": "OtherTransaction/Go/hello",
		},
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name": "OtherTransaction/Go/hello",
		},
	}})
}

func TestResponseCodeIsError(t *testing.T) {
	cfg := NewConfig("my app", "0123456789012345678901234567890123456789")

	if is := responseCodeIsError(&cfg, 200); is {
		t.Error(is)
	}
	if is := responseCodeIsError(&cfg, 400); !is {
		t.Error(is)
	}
	if is := responseCodeIsError(&cfg, 404); is {
		t.Error(is)
	}
	if is := responseCodeIsError(&cfg, 503); !is {
		t.Error(is)
	}
}

func TestExternalSegmentURL(t *testing.T) {
	rawURL := "http://url.com"
	req, err := http.NewRequest("GET", "http://request.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	responsereq, err := http.NewRequest("GET", "http://response.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := &http.Response{Request: responsereq}

	// empty segment
	u, err := externalSegmentURL(ExternalSegment{})
	host := internal.HostFromURL(u)
	if nil != err || nil != u || "" != host {
		t.Error(u, err, internal.HostFromURL(u))
	}
	// segment only containing url
	u, err = externalSegmentURL(ExternalSegment{URL: rawURL})
	host = internal.HostFromURL(u)
	if nil != err || host != "url.com" {
		t.Error(u, err, internal.HostFromURL(u))
	}
	// segment only containing request
	u, err = externalSegmentURL(ExternalSegment{Request: req})
	host = internal.HostFromURL(u)
	if nil != err || "request.com" != host {
		t.Error(host)
	}
	// segment only containing response
	u, err = externalSegmentURL(ExternalSegment{Response: response})
	host = internal.HostFromURL(u)
	if nil != err || "response.com" != host {
		t.Error(host)
	}
	// segment containing request and response
	u, err = externalSegmentURL(ExternalSegment{
		Request:  req,
		Response: response,
	})
	host = internal.HostFromURL(u)
	if nil != err || "response.com" != host {
		t.Error(host)
	}
	// segment containing url, request, and response
	u, err = externalSegmentURL(ExternalSegment{
		URL:      rawURL,
		Request:  req,
		Response: response,
	})
	host = internal.HostFromURL(u)
	if nil != err || "url.com" != host {
		t.Error(err, host)
	}
}

func TestZeroSegmentsSafe(t *testing.T) {
	s := Segment{}
	s.End()

	StartSegmentNow(nil)

	ds := DatastoreSegment{}
	ds.End()

	es := ExternalSegment{}
	es.End()

	StartSegment(nil, "").End()

	StartExternalSegment(nil, nil).End()
}

func TestTraceSegmentDefer(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	func() {
		defer StartSegment(txn, "segment").End()
	}()
	txn.End()
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Custom/segment", Scope: "", Forced: false, Data: nil},
		{Name: "Custom/segment", Scope: scope, Forced: false, Data: nil},
	}, webMetrics...))
}

func TestTraceSegmentNilErr(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	err := StartSegment(txn, "segment").End()
	if nil != err {
		t.Error(err)
	}
	txn.End()
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Custom/segment", Scope: "", Forced: false, Data: nil},
		{Name: "Custom/segment", Scope: scope, Forced: false, Data: nil},
	}, webMetrics...))
}

func TestTraceSegmentOutOfOrder(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := StartSegment(txn, "s1")
	s2 := StartSegment(txn, "s1")
	err1 := s1.End()
	err2 := s2.End()
	if nil != err1 {
		t.Error(err1)
	}
	if nil == err2 {
		t.Error(err2)
	}
	txn.End()
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Custom/s1", Scope: "", Forced: false, Data: nil},
		{Name: "Custom/s1", Scope: scope, Forced: false, Data: nil},
	}, webMetrics...))
}

func TestTraceSegmentEndedBeforeStartSegment(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	txn.End()
	s := StartSegment(txn, "segment")
	err := s.End()
	if err != errAlreadyEnded {
		t.Error(err)
	}
	app.ExpectMetrics(t, webMetrics)
}

func TestTraceSegmentEndedBeforeEndSegment(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s := StartSegment(txn, "segment")
	txn.End()
	err := s.End()
	if err != errAlreadyEnded {
		t.Error(err)
	}

	app.ExpectMetrics(t, webMetrics)
}

func TestTraceSegmentPanic(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	func() {
		defer func() {
			recover()
		}()

		func() {
			defer StartSegment(txn, "f1").End()

			func() {
				t := StartSegment(txn, "f2")

				func() {
					defer StartSegment(txn, "f3").End()

					func() {
						StartSegment(txn, "f4")

						panic(nil)
					}()
				}()

				t.End()
			}()
		}()
	}()

	txn.End()
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Custom/f1", Scope: "", Forced: false, Data: nil},
		{Name: "Custom/f1", Scope: scope, Forced: false, Data: nil},
		{Name: "Custom/f3", Scope: "", Forced: false, Data: nil},
		{Name: "Custom/f3", Scope: scope, Forced: false, Data: nil},
	}, webMetrics...))
}

func TestTraceSegmentNilTxn(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s := Segment{Name: "hello"}
	err := s.End()
	if err != nil {
		t.Error(err)
	}
	txn.End()
	app.ExpectMetrics(t, webMetrics)
}

func TestTraceDatastore(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s := DatastoreSegment{}
	s.StartTime = txn.StartSegmentNow()
	s.Product = DatastoreMySQL
	s.Collection = "my_table"
	s.Operation = "SELECT"
	err := s.End()
	if nil != err {
		t.Error(err)
	}
	txn.NoticeError(myError{})
	txn.End()
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Datastore/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/operation/MySQL/SELECT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/my_table/SELECT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/my_table/SELECT", Scope: scope, Forced: false, Data: nil},
	}, webErrorMetrics...))
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":       "newrelic.myError",
			"error.message":     "my msg",
			"transactionName":   "WebTransaction/Go/hello",
			"databaseCallCount": 1,
			"databaseDuration":  internal.MatchAnything,
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":              "WebTransaction/Go/hello",
			"nr.apdexPerfZone":  "F",
			"databaseCallCount": 1,
			"databaseDuration":  internal.MatchAnything,
		},
	}})
}

func TestTraceDatastoreBackground(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	s := DatastoreSegment{
		StartTime:  txn.StartSegmentNow(),
		Product:    DatastoreMySQL,
		Collection: "my_table",
		Operation:  "SELECT",
	}
	err := s.End()
	if nil != err {
		t.Error(err)
	}
	txn.NoticeError(myError{})
	txn.End()
	scope := "OtherTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Datastore/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/allOther", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/MySQL/allOther", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/operation/MySQL/SELECT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/my_table/SELECT", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/statement/MySQL/my_table/SELECT", Scope: scope, Forced: false, Data: nil},
	}, backgroundErrorMetrics...))
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":       "newrelic.myError",
			"error.message":     "my msg",
			"transactionName":   "OtherTransaction/Go/hello",
			"databaseCallCount": 1,
			"databaseDuration":  internal.MatchAnything,
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":              "OtherTransaction/Go/hello",
			"databaseCallCount": 1,
			"databaseDuration":  internal.MatchAnything,
		},
	}})
}

func TestTraceDatastoreMissingProductOperationCollection(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s := DatastoreSegment{
		StartTime: txn.StartSegmentNow(),
	}
	err := s.End()
	if nil != err {
		t.Error(err)
	}
	txn.NoticeError(myError{})
	txn.End()
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "Datastore/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/Unknown/all", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/Unknown/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "Datastore/operation/Unknown/other", Scope: "", Forced: false, Data: nil},
		{Name: "Datastore/operation/Unknown/other", Scope: scope, Forced: false, Data: nil},
	}, webErrorMetrics...))
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":       "newrelic.myError",
			"error.message":     "my msg",
			"transactionName":   "WebTransaction/Go/hello",
			"databaseCallCount": 1,
			"databaseDuration":  internal.MatchAnything,
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":              "WebTransaction/Go/hello",
			"nr.apdexPerfZone":  "F",
			"databaseCallCount": 1,
			"databaseDuration":  internal.MatchAnything,
		},
	}})
}

func TestTraceDatastoreNilTxn(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	var s DatastoreSegment
	s.Product = DatastoreMySQL
	s.Collection = "my_table"
	s.Operation = "SELECT"
	err := s.End()
	if nil != err {
		t.Error(err)
	}
	txn.NoticeError(myError{})
	txn.End()
	app.ExpectMetrics(t, webErrorMetrics)
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "newrelic.myError",
			"error.message":   "my msg",
			"transactionName": "WebTransaction/Go/hello",
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":             "WebTransaction/Go/hello",
			"nr.apdexPerfZone": "F",
		},
	}})
}

func TestTraceDatastoreTxnEnded(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	txn.NoticeError(myError{})
	s := DatastoreSegment{
		StartTime:  txn.StartSegmentNow(),
		Product:    DatastoreMySQL,
		Collection: "my_table",
		Operation:  "SELECT",
	}
	txn.End()
	err := s.End()
	if errAlreadyEnded != err {
		t.Error(err)
	}
	app.ExpectMetrics(t, webErrorMetrics)
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "newrelic.myError",
			"error.message":   "my msg",
			"transactionName": "WebTransaction/Go/hello",
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":             "WebTransaction/Go/hello",
			"nr.apdexPerfZone": "F",
		},
	}})
}

func TestTraceExternal(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s := ExternalSegment{
		StartTime: txn.StartSegmentNow(),
		URL:       "http://example.com/",
	}
	err := s.End()
	if nil != err {
		t.Error(err)
	}
	txn.NoticeError(myError{})
	txn.End()
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "External/all", Scope: "", Forced: true, Data: nil},
		{Name: "External/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "External/example.com/all", Scope: "", Forced: false, Data: nil},
		{Name: "External/example.com/all", Scope: scope, Forced: false, Data: nil},
	}, webErrorMetrics...))
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":       "newrelic.myError",
			"error.message":     "my msg",
			"transactionName":   "WebTransaction/Go/hello",
			"externalCallCount": 1,
			"externalDuration":  internal.MatchAnything,
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":              "WebTransaction/Go/hello",
			"nr.apdexPerfZone":  "F",
			"externalCallCount": 1,
			"externalDuration":  internal.MatchAnything,
		},
	}})
}

func TestTraceExternalBadURL(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s := ExternalSegment{
		StartTime: txn.StartSegmentNow(),
		URL:       ":example.com/",
	}
	err := s.End()
	if nil == err {
		t.Error(err)
	}
	txn.NoticeError(myError{})
	txn.End()
	app.ExpectMetrics(t, webErrorMetrics)
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "newrelic.myError",
			"error.message":   "my msg",
			"transactionName": "WebTransaction/Go/hello",
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":             "WebTransaction/Go/hello",
			"nr.apdexPerfZone": "F",
		},
	}})
}

func TestTraceExternalBackground(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	s := ExternalSegment{
		StartTime: txn.StartSegmentNow(),
		URL:       "http://example.com/",
	}
	err := s.End()
	if nil != err {
		t.Error(err)
	}
	txn.NoticeError(myError{})
	txn.End()
	scope := "OtherTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "External/all", Scope: "", Forced: true, Data: nil},
		{Name: "External/allOther", Scope: "", Forced: true, Data: nil},
		{Name: "External/example.com/all", Scope: "", Forced: false, Data: nil},
		{Name: "External/example.com/all", Scope: scope, Forced: false, Data: nil},
	}, backgroundErrorMetrics...))
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":       "newrelic.myError",
			"error.message":     "my msg",
			"transactionName":   "OtherTransaction/Go/hello",
			"externalCallCount": 1,
			"externalDuration":  internal.MatchAnything,
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":              "OtherTransaction/Go/hello",
			"externalCallCount": 1,
			"externalDuration":  internal.MatchAnything,
		},
	}})
}

func TestTraceExternalMissingURL(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s := ExternalSegment{
		StartTime: txn.StartSegmentNow(),
	}
	err := s.End()
	if nil != err {
		t.Error(err)
	}
	txn.NoticeError(myError{})
	txn.End()
	scope := "WebTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "External/all", Scope: "", Forced: true, Data: nil},
		{Name: "External/allWeb", Scope: "", Forced: true, Data: nil},
		{Name: "External/unknown/all", Scope: "", Forced: false, Data: nil},
		{Name: "External/unknown/all", Scope: scope, Forced: false, Data: nil},
	}, webErrorMetrics...))
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":       "newrelic.myError",
			"error.message":     "my msg",
			"transactionName":   "WebTransaction/Go/hello",
			"externalCallCount": 1,
			"externalDuration":  internal.MatchAnything,
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":              "WebTransaction/Go/hello",
			"nr.apdexPerfZone":  "F",
			"externalCallCount": 1,
			"externalDuration":  internal.MatchAnything,
		},
	}})
}

func TestTraceExternalNilTxn(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	txn.NoticeError(myError{})
	var s ExternalSegment
	err := s.End()
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectMetrics(t, webErrorMetrics)
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "newrelic.myError",
			"error.message":   "my msg",
			"transactionName": "WebTransaction/Go/hello",
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":             "WebTransaction/Go/hello",
			"nr.apdexPerfZone": "F",
		},
	}})
}

func TestTraceExternalTxnEnded(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	txn.NoticeError(myError{})
	s := ExternalSegment{
		StartTime: txn.StartSegmentNow(),
		URL:       "http://example.com/",
	}
	txn.End()
	err := s.End()
	if err != errAlreadyEnded {
		t.Error(err)
	}
	app.ExpectMetrics(t, webErrorMetrics)
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "newrelic.myError",
			"error.message":   "my msg",
			"transactionName": "WebTransaction/Go/hello",
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":             "WebTransaction/Go/hello",
			"nr.apdexPerfZone": "F",
		},
	}})
}

func TestRoundTripper(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	url := "http://example.com/"
	client := &http.Client{}
	inner := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// TODO test that request headers have been set here.
		if r.URL.String() != url {
			t.Error(r.URL.String())
		}
		return nil, errors.New("hello")
	})
	client.Transport = NewRoundTripper(txn, inner)
	resp, err := client.Get(url)
	if resp != nil || err == nil {
		t.Error(resp, err.Error())
	}
	txn.NoticeError(myError{})
	txn.End()
	scope := "OtherTransaction/Go/hello"
	app.ExpectMetrics(t, append([]internal.WantMetric{
		{Name: "External/all", Scope: "", Forced: true, Data: nil},
		{Name: "External/allOther", Scope: "", Forced: true, Data: nil},
		{Name: "External/example.com/all", Scope: "", Forced: false, Data: nil},
		{Name: "External/example.com/all", Scope: scope, Forced: false, Data: nil},
	}, backgroundErrorMetrics...))
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":       "newrelic.myError",
			"error.message":     "my msg",
			"transactionName":   "OtherTransaction/Go/hello",
			"externalCallCount": 1,
			"externalDuration":  internal.MatchAnything,
		},
	}})
	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":              "OtherTransaction/Go/hello",
			"externalCallCount": 1,
			"externalDuration":  internal.MatchAnything,
			"nr.guid":           internal.MatchAnything,
			"nr.tripId":         internal.MatchAnything,
			"nr.pathHash":       internal.MatchAnything,
		},
	}})
}

func TestTraceBelowThreshold(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	txn.End()
	app.ExpectTxnTraces(t, []internal.WantTxnTrace{})
}

func TestTraceBelowThresholdBackground(t *testing.T) {
	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	txn.End()
	app.ExpectTxnTraces(t, []internal.WantTxnTrace{})
}

func TestTraceNoSegments(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.TransactionTracer.Threshold.IsApdexFailing = false
		cfg.TransactionTracer.Threshold.Duration = 0
		cfg.TransactionTracer.SegmentThreshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	txn.End()
	app.ExpectTxnTraces(t, []internal.WantTxnTrace{{
		MetricName:  "WebTransaction/Go/hello",
		CleanURL:    "/hello",
		NumSegments: 0,
	}})
}

func TestTraceDisabledLocally(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.TransactionTracer.Threshold.IsApdexFailing = false
		cfg.TransactionTracer.Threshold.Duration = 0
		cfg.TransactionTracer.SegmentThreshold = 0
		cfg.TransactionTracer.Enabled = false
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	txn.End()
	app.ExpectTxnTraces(t, []internal.WantTxnTrace{})
}

func TestTraceDisabledRemotely(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.TransactionTracer.Threshold.IsApdexFailing = false
		cfg.TransactionTracer.Threshold.Duration = 0
		cfg.TransactionTracer.SegmentThreshold = 0
	}
	replyfn := func(reply *internal.ConnectReply) {
		reply.CollectTraces = false
	}
	app := testApp(replyfn, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	txn.End()
	app.ExpectTxnTraces(t, []internal.WantTxnTrace{})
}

func TestTraceWithSegments(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.TransactionTracer.Threshold.IsApdexFailing = false
		cfg.TransactionTracer.Threshold.Duration = 0
		cfg.TransactionTracer.SegmentThreshold = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := StartSegment(txn, "s1")
	s1.End()
	s2 := ExternalSegment{
		StartTime: StartSegmentNow(txn),
		URL:       "http://example.com",
	}
	s2.End()
	s3 := DatastoreSegment{
		StartTime:  StartSegmentNow(txn),
		Product:    DatastoreMySQL,
		Collection: "my_table",
		Operation:  "SELECT",
	}
	s3.End()
	txn.End()
	app.ExpectTxnTraces(t, []internal.WantTxnTrace{{
		MetricName:  "WebTransaction/Go/hello",
		CleanURL:    "/hello",
		NumSegments: 3,
	}})
}

func TestTraceSegmentsBelowThreshold(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.TransactionTracer.Threshold.IsApdexFailing = false
		cfg.TransactionTracer.Threshold.Duration = 0
		cfg.TransactionTracer.SegmentThreshold = 1 * time.Hour
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, helloRequest)
	s1 := StartSegment(txn, "s1")
	s1.End()
	s2 := ExternalSegment{
		StartTime: StartSegmentNow(txn),
		URL:       "http://example.com",
	}
	s2.End()
	s3 := DatastoreSegment{
		StartTime:  StartSegmentNow(txn),
		Product:    DatastoreMySQL,
		Collection: "my_table",
		Operation:  "SELECT",
	}
	s3.End()
	txn.End()
	app.ExpectTxnTraces(t, []internal.WantTxnTrace{{
		MetricName:  "WebTransaction/Go/hello",
		CleanURL:    "/hello",
		NumSegments: 0,
	}})
}
