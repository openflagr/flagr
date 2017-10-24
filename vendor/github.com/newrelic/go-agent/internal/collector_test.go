package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/newrelic/go-agent/internal/logger"
)

func TestLicenseInvalid(t *testing.T) {
	r := CompactJSONString(`{
		"exception":{
			"message":"Invalid license key, please contact support@newrelic.com",
			"error_type":"NewRelic::Agent::LicenseException"
		}
	}`)
	reply, err := parseResponse([]byte(r))
	if reply != nil {
		t.Fatal(string(reply))
	}
	if !IsLicenseException(err) {
		t.Fatal(err)
	}
}

func TestRedirectSuccess(t *testing.T) {
	r := `{"return_value":"staging-collector-101.newrelic.com"}`
	reply, err := parseResponse([]byte(r))
	if nil != err {
		t.Fatal(err)
	}
	if string(reply) != `"staging-collector-101.newrelic.com"` {
		t.Fatal(string(reply))
	}
}

func TestEmptyHash(t *testing.T) {
	reply, err := parseResponse([]byte(`{}`))
	if nil != err {
		t.Fatal(err)
	}
	if nil != reply {
		t.Fatal(string(reply))
	}
}

func TestReturnValueNull(t *testing.T) {
	reply, err := parseResponse([]byte(`{"return_value":null}`))
	if nil != err {
		t.Fatal(err)
	}
	if "null" != string(reply) {
		t.Fatal(string(reply))
	}
}

func TestReplyNull(t *testing.T) {
	reply, err := parseResponse(nil)

	if nil == err || err.Error() != `unexpected end of JSON input` {
		t.Fatal(err)
	}
	if nil != reply {
		t.Fatal(string(reply))
	}
}

func TestConnectSuccess(t *testing.T) {
	inner := `{
	"agent_run_id":"599551769342729",
	"product_level":40,
	"js_agent_file":"",
	"cross_process_id":"12345#12345",
	"collect_errors":true,
	"url_rules":[
		{
			"each_segment":false,
			"match_expression":".*\\.(txt|udl|plist|css)$",
			"eval_order":1000,
			"replace_all":false,
			"ignore":false,
			"terminate_chain":true,
			"replacement":"\/*.\\1"
		},
		{
			"each_segment":true,
			"match_expression":"^[0-9][0-9a-f_,.-]*$",
			"eval_order":1001,
			"replace_all":false,
			"ignore":false,
			"terminate_chain":false,
			"replacement":"*"
		}
	],
	"messages":[
		{
			"message":"Reporting to staging",
			"level":"INFO"
		}
	],
	"data_report_period":60,
	"collect_traces":true,
	"sampling_rate":0,
	"js_agent_loader":"",
	"encoding_key":"the-encoding-key",
	"apdex_t":0.5,
	"collect_analytics_events":true,
	"trusted_account_ids":[49402]
}`
	outer := `{"return_value":` + inner + `}`
	reply, err := parseResponse([]byte(outer))

	if nil != err {
		t.Fatal(err)
	}
	if string(reply) != inner {
		t.Fatal(string(reply))
	}
}

func TestClientError(t *testing.T) {
	r := `{"exception":{"message":"something","error_type":"my_error"}}`
	reply, err := parseResponse([]byte(r))
	if nil == err || err.Error() != "my_error: something" {
		t.Fatal(err)
	}
	if nil != reply {
		t.Fatal(string(reply))
	}
}

func TestForceRestartException(t *testing.T) {
	// NOTE: This string was generated manually, not taken from the actual
	// collector.
	r := CompactJSONString(`{
		"exception":{
			"message":"something",
			"error_type":"NewRelic::Agent::ForceRestartException"
		}
	}`)
	reply, err := parseResponse([]byte(r))
	if reply != nil {
		t.Fatal(string(reply))
	}
	if !IsRestartException(err) {
		t.Fatal(err)
	}
}

func TestForceDisconnectException(t *testing.T) {
	// NOTE: This string was generated manually, not taken from the actual
	// collector.
	r := CompactJSONString(`{
		"exception":{
			"message":"something",
			"error_type":"NewRelic::Agent::ForceDisconnectException"
		}
	}`)
	reply, err := parseResponse([]byte(r))
	if reply != nil {
		t.Fatal(string(reply))
	}
	if !IsDisconnect(err) {
		t.Fatal(err)
	}
}

func TestRuntimeError(t *testing.T) {
	// NOTE: This string was generated manually, not taken from the actual
	// collector.
	r := `{"exception":{"message":"something","error_type":"RuntimeError"}}`
	reply, err := parseResponse([]byte(r))
	if reply != nil {
		t.Fatal(string(reply))
	}
	if !IsRuntime(err) {
		t.Fatal(err)
	}
}

func TestUnknownError(t *testing.T) {
	r := `{"exception":{"message":"something","error_type":"unknown_type"}}`
	reply, err := parseResponse([]byte(r))
	if reply != nil {
		t.Fatal(string(reply))
	}
	if nil == err || err.Error() != "unknown_type: something" {
		t.Fatal(err)
	}
}

func TestUrl(t *testing.T) {
	cmd := RpmCmd{
		Name:      "foo_method",
		Collector: "example.com",
	}
	cs := RpmControls{
		UseTLS:       true,
		License:      "123abc",
		Client:       nil,
		Logger:       nil,
		AgentVersion: "1",
	}

	out := rpmURL(cmd, cs)
	u, err := url.Parse(out)
	if err != nil {
		t.Fatalf("url.Parse(%q) = %q", out, err)
	}

	got := u.Query().Get("license_key")
	if got != cs.License {
		t.Errorf("got=%q cmd.License=%q", got, cs.License)
	}
	if u.Scheme != "https" {
		t.Error(u.Scheme)
	}
}

const (
	redirectBody   = `{"return_value":"special_collector"}`
	connectBody    = `{"return_value":{"agent_run_id":"my_agent_run_id"}}`
	disconnectBody = `{"exception":{"error_type":"NewRelic::Agent::ForceDisconnectException"}}`
	licenseBody    = `{"exception":{"error_type":"NewRelic::Agent::LicenseException"}}`
	malformedBody  = `{"return_value":}}`
)

func makeResponse(code int, body string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
	}
}

type endpointResult struct {
	response *http.Response
	err      error
}

type connectMockRoundTripper struct {
	redirect endpointResult
	connect  endpointResult
}

func (m connectMockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	cmd := r.URL.Query().Get("method")
	switch cmd {
	case cmdRedirect:
		return m.redirect.response, m.redirect.err
	case cmdConnect:
		return m.connect.response, m.connect.err
	default:
		return nil, fmt.Errorf("unknown cmd: %s", cmd)
	}
}

func (m connectMockRoundTripper) CancelRequest(req *http.Request) {}

func testConnectHelper(transport http.RoundTripper) (*AppRun, error) {
	cs := RpmControls{
		UseTLS:       true,
		License:      "12345",
		Client:       &http.Client{Transport: transport},
		Logger:       logger.ShimLogger{},
		AgentVersion: "1",
	}

	return ConnectAttempt([]byte(`"connect-json"`), "redirect-host", cs)
}

func TestConnectAttemptSuccess(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, redirectBody)},
		connect:  endpointResult{response: makeResponse(200, connectBody)},
	})
	if nil == run || nil != err {
		t.Fatal(run, err)
	}
	if run.Collector != "special_collector" {
		t.Error(run.Collector)
	}
	if run.RunID != "my_agent_run_id" {
		t.Error(run)
	}
}

func TestConnectAttemptDisconnectOnRedirect(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, disconnectBody)},
		connect:  endpointResult{response: makeResponse(200, connectBody)},
	})
	if nil != run {
		t.Error(run)
	}
	if !IsDisconnect(err) {
		t.Fatal(err)
	}
}

func TestConnectAttemptDisconnectOnConnect(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, redirectBody)},
		connect:  endpointResult{response: makeResponse(200, disconnectBody)},
	})
	if nil != run {
		t.Error(run)
	}
	if !IsDisconnect(err) {
		t.Fatal(err)
	}
}

func TestConnectAttemptLicenseExceptionOnRedirect(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, licenseBody)},
		connect:  endpointResult{response: makeResponse(200, connectBody)},
	})
	if nil != run {
		t.Error(run)
	}
	if !IsLicenseException(err) {
		t.Fatal(err)
	}
}

func TestConnectAttemptLicenseExceptionOnConnect(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, redirectBody)},
		connect:  endpointResult{response: makeResponse(200, licenseBody)},
	})
	if nil != run {
		t.Error(run)
	}
	if !IsLicenseException(err) {
		t.Fatal(err)
	}
}

func TestConnectAttemptInvalidJSON(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, redirectBody)},
		connect:  endpointResult{response: makeResponse(200, malformedBody)},
	})
	if nil != run {
		t.Error(run)
	}
	if nil == err {
		t.Fatal("missing error")
	}
}

func TestConnectAttemptCollectorNotString(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, `{"return_value":123}`)},
		connect:  endpointResult{response: makeResponse(200, connectBody)},
	})
	if nil != run {
		t.Error(run)
	}
	if nil == err {
		t.Fatal("missing error")
	}
}

func TestConnectAttempt401(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, redirectBody)},
		connect:  endpointResult{response: makeResponse(401, connectBody)},
	})
	if nil != run {
		t.Error(run)
	}
	if err != ErrUnauthorized {
		t.Fatal(err)
	}
}

func TestConnectAttempt413(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, redirectBody)},
		connect:  endpointResult{response: makeResponse(413, connectBody)},
	})
	if nil != run {
		t.Error(run)
	}
	if err != ErrPayloadTooLarge {
		t.Fatal(err)
	}
}

func TestConnectAttempt415(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, redirectBody)},
		connect:  endpointResult{response: makeResponse(415, connectBody)},
	})
	if nil != run {
		t.Error(run)
	}
	if err != ErrUnsupportedMedia {
		t.Fatal(err)
	}
}

func TestConnectAttemptUnexpectedCode(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, redirectBody)},
		connect:  endpointResult{response: makeResponse(404, connectBody)},
	})
	if nil != run {
		t.Error(run)
	}
	if _, ok := err.(unexpectedStatusCodeErr); !ok {
		t.Fatal(err)
	}
}

func TestConnectAttemptUnexpectedError(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, redirectBody)},
		connect:  endpointResult{err: errors.New("unexpected error")},
	})
	if nil != run {
		t.Error(run)
	}
	if nil == err {
		t.Fatal("missing error")
	}
}

func TestConnectAttemptMissingRunID(t *testing.T) {
	run, err := testConnectHelper(connectMockRoundTripper{
		redirect: endpointResult{response: makeResponse(200, redirectBody)},
		connect:  endpointResult{response: makeResponse(200, `{"return_value":{}}`)},
	})
	if nil != run {
		t.Error(run)
	}
	if nil == err {
		t.Fatal("missing error")
	}
}
