package internal

import (
	"testing"
	"time"
)

func TestCreateFinalMetrics(t *testing.T) {
	now := time.Now()

	h := NewHarvest(now)
	h.CreateFinalMetrics()
	ExpectMetrics(t, h.Metrics, []WantMetric{
		{instanceReporting, "", true, []float64{1, 0, 0, 0, 0, 0}},
		{customEventsSeen, "", true, []float64{0, 0, 0, 0, 0, 0}},
		{customEventsSent, "", true, []float64{0, 0, 0, 0, 0, 0}},
		{txnEventsSeen, "", true, []float64{0, 0, 0, 0, 0, 0}},
		{txnEventsSent, "", true, []float64{0, 0, 0, 0, 0, 0}},
		{errorEventsSeen, "", true, []float64{0, 0, 0, 0, 0, 0}},
		{errorEventsSent, "", true, []float64{0, 0, 0, 0, 0, 0}},
	})

	h = NewHarvest(now)
	h.Metrics = newMetricTable(0, now)
	h.CustomEvents = newCustomEvents(1)
	h.TxnEvents = newTxnEvents(1)
	h.ErrorEvents = newErrorEvents(1)

	h.Metrics.addSingleCount("drop me!", unforced)

	customE, err := CreateCustomEvent("my event type", map[string]interface{}{"zip": 1}, time.Now())
	if nil != err {
		t.Fatal(err)
	}
	h.CustomEvents.Add(customE)
	h.CustomEvents.Add(customE)

	txnE := &TxnEvent{}
	h.TxnEvents.AddTxnEvent(txnE)
	h.TxnEvents.AddTxnEvent(txnE)

	h.ErrorEvents.Add(&ErrorEvent{})
	h.ErrorEvents.Add(&ErrorEvent{})

	h.CreateFinalMetrics()
	ExpectMetrics(t, h.Metrics, []WantMetric{
		{instanceReporting, "", true, []float64{1, 0, 0, 0, 0, 0}},
		{customEventsSeen, "", true, []float64{2, 0, 0, 0, 0, 0}},
		{customEventsSent, "", true, []float64{1, 0, 0, 0, 0, 0}},
		{txnEventsSeen, "", true, []float64{2, 0, 0, 0, 0, 0}},
		{txnEventsSent, "", true, []float64{1, 0, 0, 0, 0, 0}},
		{errorEventsSeen, "", true, []float64{2, 0, 0, 0, 0, 0}},
		{errorEventsSent, "", true, []float64{1, 0, 0, 0, 0, 0}},
		{supportabilityDropped, "", true, []float64{1, 0, 0, 0, 0, 0}},
	})
}

func TestEmptyPayloads(t *testing.T) {
	h := NewHarvest(time.Now())
	payloads := h.Payloads()
	for _, p := range payloads {
		d, err := p.Data("agentRunID", time.Now())
		if d != nil || err != nil {
			t.Error(d, err)
		}
	}
}

func TestMergeFailedHarvest(t *testing.T) {
	start1 := time.Now()
	start2 := start1.Add(1 * time.Minute)
	h := NewHarvest(start1)
	h.Metrics.addCount("zip", 1, forced)
	h.TxnEvents.AddTxnEvent(&TxnEvent{
		FinalName: "finalName",
		Start:     time.Now(),
		Duration:  1 * time.Second,
	})
	customEventParams := map[string]interface{}{"zip": 1}
	ce, err := CreateCustomEvent("myEvent", customEventParams, time.Now())
	if nil != err {
		t.Fatal(err)
	}
	h.CustomEvents.Add(ce)
	h.ErrorEvents.Add(&ErrorEvent{
		ErrorData: ErrorData{
			Klass: "klass",
			Msg:   "msg",
			When:  time.Now(),
		},
		TxnEvent: TxnEvent{
			FinalName: "finalName",
			Duration:  1 * time.Second,
		},
	})

	ers := NewTxnErrors(10)
	ers.Add(ErrorData{
		When:  time.Now(),
		Msg:   "msg",
		Klass: "klass",
		Stack: GetStackTrace(0),
	})
	MergeTxnErrors(&h.ErrorTraces, ers, TxnEvent{
		FinalName: "finalName",
		CleanURL:  "requestURI",
		Attrs:     nil,
	})

	if start1 != h.Metrics.metricPeriodStart {
		t.Error(h.Metrics.metricPeriodStart)
	}
	if 0 != h.Metrics.failedHarvests {
		t.Error(h.Metrics.failedHarvests)
	}
	if 0 != h.CustomEvents.events.failedHarvests {
		t.Error(h.CustomEvents.events.failedHarvests)
	}
	if 0 != h.TxnEvents.events.failedHarvests {
		t.Error(h.TxnEvents.events.failedHarvests)
	}
	if 0 != h.ErrorEvents.events.failedHarvests {
		t.Error(h.ErrorEvents.events.failedHarvests)
	}
	ExpectMetrics(t, h.Metrics, []WantMetric{
		{"zip", "", true, []float64{1, 0, 0, 0, 0, 0}},
	})
	ExpectCustomEvents(t, h.CustomEvents, []WantEvent{{
		Intrinsics: map[string]interface{}{
			"type":      "myEvent",
			"timestamp": MatchAnything,
		},
		UserAttributes: customEventParams,
	}})
	ExpectErrorEvents(t, h.ErrorEvents, []WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "klass",
			"error.message":   "msg",
			"transactionName": "finalName",
		},
	}})
	ExpectTxnEvents(t, h.TxnEvents, []WantEvent{{
		Intrinsics: map[string]interface{}{
			"name": "finalName",
		},
	}})
	ExpectErrors(t, h.ErrorTraces, []WantError{{
		TxnName: "finalName",
		Msg:     "msg",
		Klass:   "klass",
		Caller:  "internal.TestMergeFailedHarvest",
		URL:     "requestURI",
	}})

	nextHarvest := NewHarvest(start2)
	if start2 != nextHarvest.Metrics.metricPeriodStart {
		t.Error(nextHarvest.Metrics.metricPeriodStart)
	}
	payloads := h.Payloads()
	for _, p := range payloads {
		p.MergeIntoHarvest(nextHarvest)
	}

	if start1 != nextHarvest.Metrics.metricPeriodStart {
		t.Error(nextHarvest.Metrics.metricPeriodStart)
	}
	if 1 != nextHarvest.Metrics.failedHarvests {
		t.Error(nextHarvest.Metrics.failedHarvests)
	}
	if 1 != nextHarvest.CustomEvents.events.failedHarvests {
		t.Error(nextHarvest.CustomEvents.events.failedHarvests)
	}
	if 1 != nextHarvest.TxnEvents.events.failedHarvests {
		t.Error(nextHarvest.TxnEvents.events.failedHarvests)
	}
	if 1 != nextHarvest.ErrorEvents.events.failedHarvests {
		t.Error(nextHarvest.ErrorEvents.events.failedHarvests)
	}
	ExpectMetrics(t, nextHarvest.Metrics, []WantMetric{
		{"zip", "", true, []float64{1, 0, 0, 0, 0, 0}},
	})
	ExpectCustomEvents(t, nextHarvest.CustomEvents, []WantEvent{{
		Intrinsics: map[string]interface{}{
			"type":      "myEvent",
			"timestamp": MatchAnything,
		},
		UserAttributes: customEventParams,
	}})
	ExpectErrorEvents(t, nextHarvest.ErrorEvents, []WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "klass",
			"error.message":   "msg",
			"transactionName": "finalName",
		},
	}})
	ExpectTxnEvents(t, nextHarvest.TxnEvents, []WantEvent{{
		Intrinsics: map[string]interface{}{
			"name": "finalName",
		},
	}})
	ExpectErrors(t, nextHarvest.ErrorTraces, []WantError{})
}

func TestCreateTxnMetrics(t *testing.T) {
	txnErr := &ErrorData{}
	txnErrors := []*ErrorData{txnErr}
	webName := "WebTransaction/zip/zap"
	backgroundName := "OtherTransaction/zip/zap"
	args := &TxnData{}
	args.Duration = 123 * time.Second
	args.Exclusive = 109 * time.Second
	args.ApdexThreshold = 2 * time.Second

	args.FinalName = webName
	args.IsWeb = true
	args.Errors = txnErrors
	args.Zone = ApdexTolerating
	metrics := newMetricTable(100, time.Now())
	CreateTxnMetrics(args, metrics)
	ExpectMetrics(t, metrics, []WantMetric{
		{webName, "", true, []float64{1, 123, 109, 123, 123, 123 * 123}},
		{webRollup, "", true, []float64{1, 123, 109, 123, 123, 123 * 123}},
		{dispatcherMetric, "", true, []float64{1, 123, 0, 123, 123, 123 * 123}},
		{"Errors/all", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"Errors/allWeb", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"Errors/" + webName, "", true, []float64{1, 0, 0, 0, 0, 0}},
		{apdexRollup, "", true, []float64{0, 1, 0, 2, 2, 0}},
		{"Apdex/zip/zap", "", false, []float64{0, 1, 0, 2, 2, 0}},
	})

	args.FinalName = webName
	args.IsWeb = true
	args.Errors = nil
	args.Zone = ApdexTolerating
	metrics = newMetricTable(100, time.Now())
	CreateTxnMetrics(args, metrics)
	ExpectMetrics(t, metrics, []WantMetric{
		{webName, "", true, []float64{1, 123, 109, 123, 123, 123 * 123}},
		{webRollup, "", true, []float64{1, 123, 109, 123, 123, 123 * 123}},
		{dispatcherMetric, "", true, []float64{1, 123, 0, 123, 123, 123 * 123}},
		{apdexRollup, "", true, []float64{0, 1, 0, 2, 2, 0}},
		{"Apdex/zip/zap", "", false, []float64{0, 1, 0, 2, 2, 0}},
	})

	args.FinalName = backgroundName
	args.IsWeb = false
	args.Errors = txnErrors
	args.Zone = ApdexNone
	metrics = newMetricTable(100, time.Now())
	CreateTxnMetrics(args, metrics)
	ExpectMetrics(t, metrics, []WantMetric{
		{backgroundName, "", true, []float64{1, 123, 109, 123, 123, 123 * 123}},
		{backgroundRollup, "", true, []float64{1, 123, 109, 123, 123, 123 * 123}},
		{"Errors/all", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"Errors/allOther", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"Errors/" + backgroundName, "", true, []float64{1, 0, 0, 0, 0, 0}},
	})

	args.FinalName = backgroundName
	args.IsWeb = false
	args.Errors = nil
	args.Zone = ApdexNone
	metrics = newMetricTable(100, time.Now())
	CreateTxnMetrics(args, metrics)
	ExpectMetrics(t, metrics, []WantMetric{
		{backgroundName, "", true, []float64{1, 123, 109, 123, 123, 123 * 123}},
		{backgroundRollup, "", true, []float64{1, 123, 109, 123, 123, 123 * 123}},
	})

}
