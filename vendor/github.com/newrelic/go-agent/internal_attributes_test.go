package newrelic

import (
	"errors"
	"net/http"
	"testing"

	"github.com/newrelic/go-agent/internal"
)

func TestUserAttributeBasics(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.TransactionTracer.Threshold.IsApdexFailing = false
		cfg.TransactionTracer.Threshold.Duration = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, nil)

	txn.NoticeError(errors.New("zap"))

	if err := txn.AddAttribute(`int\key`, 1); nil != err {
		t.Error(err)
	}
	if err := txn.AddAttribute(`str\key`, `zip\zap`); nil != err {
		t.Error(err)
	}
	err := txn.AddAttribute("invalid_value", struct{}{})
	if _, ok := err.(internal.ErrInvalidAttributeType); !ok {
		t.Error(err)
	}
	txn.End()
	if err := txn.AddAttribute("already_ended", "zap"); err != errAlreadyEnded {
		t.Error(err)
	}

	agentAttributes := map[string]interface{}{}
	userAttributes := map[string]interface{}{`int\key`: 1, `str\key`: `zip\zap`}

	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name": "OtherTransaction/Go/hello",
		},
		AgentAttributes: agentAttributes,
		UserAttributes:  userAttributes,
	}})
	app.ExpectErrors(t, []internal.WantError{{
		TxnName:         "OtherTransaction/Go/hello",
		Msg:             "zap",
		Klass:           "*errors.errorString",
		Caller:          "go-agent.TestUserAttributeBasics",
		URL:             "",
		AgentAttributes: agentAttributes,
		UserAttributes:  userAttributes,
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "*errors.errorString",
			"error.message":   "zap",
			"transactionName": "OtherTransaction/Go/hello",
		},
		AgentAttributes: agentAttributes,
		UserAttributes:  userAttributes,
	}})
	app.ExpectTxnTraces(t, []internal.WantTxnTrace{{
		MetricName:      "OtherTransaction/Go/hello",
		CleanURL:        "",
		NumSegments:     0,
		AgentAttributes: agentAttributes,
		UserAttributes:  userAttributes,
	}})
}

func TestUserAttributeConfiguration(t *testing.T) {
	cfgfn := func(cfg *Config) {
		cfg.TransactionEvents.Attributes.Exclude = []string{"only_errors", "only_txn_traces"}
		cfg.ErrorCollector.Attributes.Exclude = []string{"only_txn_events", "only_txn_traces"}
		cfg.TransactionTracer.Attributes.Exclude = []string{"only_txn_events", "only_errors"}
		cfg.Attributes.Exclude = []string{"completed_excluded"}
		cfg.TransactionTracer.Threshold.IsApdexFailing = false
		cfg.TransactionTracer.Threshold.Duration = 0
	}
	app := testApp(nil, cfgfn, t)
	txn := app.StartTransaction("hello", nil, nil)

	txn.NoticeError(errors.New("zap"))

	if err := txn.AddAttribute("only_errors", 1); nil != err {
		t.Error(err)
	}
	if err := txn.AddAttribute("only_txn_events", 2); nil != err {
		t.Error(err)
	}
	if err := txn.AddAttribute("only_txn_traces", 3); nil != err {
		t.Error(err)
	}
	if err := txn.AddAttribute("completed_excluded", 4); nil != err {
		t.Error(err)
	}
	txn.End()

	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name": "OtherTransaction/Go/hello",
		},
		AgentAttributes: map[string]interface{}{},
		UserAttributes:  map[string]interface{}{"only_txn_events": 2},
	}})
	app.ExpectErrors(t, []internal.WantError{{
		TxnName:         "OtherTransaction/Go/hello",
		Msg:             "zap",
		Klass:           "*errors.errorString",
		Caller:          "go-agent.TestUserAttributeConfiguration",
		URL:             "",
		AgentAttributes: map[string]interface{}{},
		UserAttributes:  map[string]interface{}{"only_errors": 1},
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "*errors.errorString",
			"error.message":   "zap",
			"transactionName": "OtherTransaction/Go/hello",
		},
		AgentAttributes: map[string]interface{}{},
		UserAttributes:  map[string]interface{}{"only_errors": 1},
	}})
	app.ExpectTxnTraces(t, []internal.WantTxnTrace{{
		MetricName:      "OtherTransaction/Go/hello",
		CleanURL:        "",
		NumSegments:     0,
		AgentAttributes: map[string]interface{}{},
		UserAttributes:  map[string]interface{}{"only_txn_traces": 3},
	}})
}

// Second attributes have priority.
func mergeAttributes(a1, a2 map[string]interface{}) map[string]interface{} {
	a := make(map[string]interface{})
	for k, v := range a1 {
		a[k] = v
	}
	for k, v := range a2 {
		a[k] = v
	}
	return a
}

var (
	// Agent attributes expected in txn events from usualAttributeTestTransaction.
	agent1 = map[string]interface{}{
		AttributeHostDisplayName:       `my\host\display\name`,
		AttributeResponseCode:          `404`,
		AttributeResponseContentType:   `text/plain; charset=us-ascii`,
		AttributeResponseContentLength: 345,
		AttributeRequestMethod:         "GET",
		AttributeRequestAccept:         "text/plain",
		AttributeRequestContentType:    "text/html; charset=utf-8",
		AttributeRequestContentLength:  753,
		AttributeRequestHost:           "my_domain.com",
	}
	// Agent attributes expected in errors and traces from usualAttributeTestTransaction.
	agent2 = mergeAttributes(agent1, map[string]interface{}{
		AttributeRequestUserAgent: "Mozilla/5.0",
		AttributeRequestReferer:   "http://en.wikipedia.org/zip",
	})
	// User attributes expected from usualAttributeTestTransaction.
	user1 = map[string]interface{}{
		"myStr": "hello",
	}
)

func agentAttributeTestcase(t testing.TB, cfgfn func(cfg *Config), e AttributeExpect) {
	app := testApp(nil, func(cfg *Config) {
		cfg.HostDisplayName = `my\host\display\name`
		cfg.TransactionTracer.Threshold.IsApdexFailing = false
		cfg.TransactionTracer.Threshold.Duration = 0
		if nil != cfgfn {
			cfgfn(cfg)
		}
	}, t)
	w := newCompatibleResponseRecorder()
	txn := app.StartTransaction("hello", w, helloRequest)
	txn.NoticeError(errors.New("zap"))

	hdr := txn.Header()
	hdr.Set("Content-Type", `text/plain; charset=us-ascii`)
	hdr.Set("Content-Length", `345`)

	txn.WriteHeader(404)
	txn.AddAttribute("myStr", "hello")

	txn.End()

	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":             "WebTransaction/Go/hello",
			"nr.apdexPerfZone": "F",
		},
		AgentAttributes: e.TxnEvent.Agent,
		UserAttributes:  e.TxnEvent.User,
	}})
	app.ExpectErrors(t, []internal.WantError{{
		TxnName:         "WebTransaction/Go/hello",
		Msg:             "zap",
		Klass:           "*errors.errorString",
		Caller:          "go-agent.agentAttributeTestcase",
		URL:             "/hello",
		AgentAttributes: e.Error.Agent,
		UserAttributes:  e.Error.User,
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "*errors.errorString",
			"error.message":   "zap",
			"transactionName": "WebTransaction/Go/hello",
		},
		AgentAttributes: e.Error.Agent,
		UserAttributes:  e.Error.User,
	}})
	app.ExpectTxnTraces(t, []internal.WantTxnTrace{{
		MetricName:      "WebTransaction/Go/hello",
		CleanURL:        "/hello",
		NumSegments:     0,
		AgentAttributes: e.TxnTrace.Agent,
		UserAttributes:  e.TxnTrace.User,
	}})
}

type UserAgent struct {
	User  map[string]interface{}
	Agent map[string]interface{}
}

type AttributeExpect struct {
	TxnEvent UserAgent
	Error    UserAgent
	TxnTrace UserAgent
}

func TestAgentAttributes(t *testing.T) {
	agentAttributeTestcase(t, nil, AttributeExpect{
		TxnEvent: UserAgent{
			Agent: agent1,
			User:  user1},
		Error: UserAgent{
			Agent: agent2,
			User:  user1},
	})
}

func TestAttributesDisabled(t *testing.T) {
	agentAttributeTestcase(t, func(cfg *Config) {
		cfg.Attributes.Enabled = false
	}, AttributeExpect{
		TxnEvent: UserAgent{
			Agent: map[string]interface{}{},
			User:  map[string]interface{}{}},
		Error: UserAgent{
			Agent: map[string]interface{}{},
			User:  map[string]interface{}{}},
		TxnTrace: UserAgent{
			Agent: map[string]interface{}{},
			User:  map[string]interface{}{}},
	})
}

func TestDefaultResponseCode(t *testing.T) {
	app := testApp(nil, nil, t)
	w := newCompatibleResponseRecorder()
	txn := app.StartTransaction("hello", w, &http.Request{})
	txn.Write([]byte("hello"))
	txn.End()

	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":             "WebTransaction/Go/hello",
			"nr.apdexPerfZone": "S",
		},
		AgentAttributes: map[string]interface{}{AttributeResponseCode: 200},
		UserAttributes:  map[string]interface{}{},
	}})
}

func TestNoResponseCode(t *testing.T) {
	app := testApp(nil, nil, t)
	w := newCompatibleResponseRecorder()
	txn := app.StartTransaction("hello", w, &http.Request{})
	txn.End()

	app.ExpectTxnEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"name":             "WebTransaction/Go/hello",
			"nr.apdexPerfZone": "S",
		},
		AgentAttributes: map[string]interface{}{},
		UserAttributes:  map[string]interface{}{},
	}})
}

func TestTxnEventAttributesDisabled(t *testing.T) {
	agentAttributeTestcase(t, func(cfg *Config) {
		cfg.TransactionEvents.Attributes.Enabled = false
	}, AttributeExpect{
		TxnEvent: UserAgent{
			Agent: map[string]interface{}{},
			User:  map[string]interface{}{}},
		Error: UserAgent{
			Agent: agent2,
			User:  user1},
		TxnTrace: UserAgent{
			Agent: agent2,
			User:  user1},
	})
}

func TestErrorAttributesDisabled(t *testing.T) {
	agentAttributeTestcase(t, func(cfg *Config) {
		cfg.ErrorCollector.Attributes.Enabled = false
	}, AttributeExpect{
		TxnEvent: UserAgent{
			Agent: agent1,
			User:  user1},
		Error: UserAgent{
			Agent: map[string]interface{}{},
			User:  map[string]interface{}{}},
		TxnTrace: UserAgent{
			Agent: agent2,
			User:  user1},
	})
}

func TestTxnTraceAttributesDisabled(t *testing.T) {
	agentAttributeTestcase(t, func(cfg *Config) {
		cfg.TransactionTracer.Attributes.Enabled = false
	}, AttributeExpect{
		TxnEvent: UserAgent{
			Agent: agent1,
			User:  user1},
		Error: UserAgent{
			Agent: agent2,
			User:  user1},
		TxnTrace: UserAgent{
			Agent: map[string]interface{}{},
			User:  map[string]interface{}{}},
	})
}

var (
	allAgentAttributeNames = []string{
		AttributeResponseCode,
		AttributeRequestMethod,
		AttributeRequestAccept,
		AttributeRequestContentType,
		AttributeRequestContentLength,
		AttributeRequestHost,
		AttributeResponseContentType,
		AttributeResponseContentLength,
		AttributeHostDisplayName,
		AttributeRequestUserAgent,
		AttributeRequestReferer,
	}
)

func TestAgentAttributesExcluded(t *testing.T) {
	agentAttributeTestcase(t, func(cfg *Config) {
		cfg.Attributes.Exclude = allAgentAttributeNames
	}, AttributeExpect{
		TxnEvent: UserAgent{
			Agent: map[string]interface{}{},
			User:  user1},
		Error: UserAgent{
			Agent: map[string]interface{}{},
			User:  user1},
		TxnTrace: UserAgent{
			Agent: map[string]interface{}{},
			User:  user1},
	})
}

func TestAgentAttributesExcludedFromErrors(t *testing.T) {
	agentAttributeTestcase(t, func(cfg *Config) {
		cfg.ErrorCollector.Attributes.Exclude = allAgentAttributeNames
	}, AttributeExpect{
		TxnEvent: UserAgent{
			Agent: agent1,
			User:  user1},
		Error: UserAgent{
			Agent: map[string]interface{}{},
			User:  user1},
		TxnTrace: UserAgent{
			Agent: agent2,
			User:  user1},
	})
}

func TestAgentAttributesExcludedFromTxnEvents(t *testing.T) {
	agentAttributeTestcase(t, func(cfg *Config) {
		cfg.TransactionEvents.Attributes.Exclude = allAgentAttributeNames
	}, AttributeExpect{
		TxnEvent: UserAgent{
			Agent: map[string]interface{}{},
			User:  user1},
		Error: UserAgent{
			Agent: agent2,
			User:  user1},
		TxnTrace: UserAgent{
			Agent: agent2,
			User:  user1},
	})
}

func TestAgentAttributesExcludedFromTxnTraces(t *testing.T) {
	agentAttributeTestcase(t, func(cfg *Config) {
		cfg.TransactionTracer.Attributes.Exclude = allAgentAttributeNames
	}, AttributeExpect{
		TxnEvent: UserAgent{
			Agent: agent1,
			User:  user1},
		Error: UserAgent{
			Agent: agent2,
			User:  user1},
		TxnTrace: UserAgent{
			Agent: map[string]interface{}{},
			User:  user1},
	})
}
