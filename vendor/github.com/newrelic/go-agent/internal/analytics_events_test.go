package internal

import (
	"bytes"
	"strconv"
	"testing"
	"time"
)

var (
	agentRunID = `12345`
)

type intWriter int

func (x intWriter) WriteJSON(buf *bytes.Buffer) {
	buf.WriteString(strconv.Itoa(int(x)))
}

func sampleAnalyticsEvent(stamp int) analyticsEvent {
	return analyticsEvent{
		eventStamp(stamp),
		intWriter(stamp),
	}
}

func TestBasic(t *testing.T) {
	events := newAnalyticsEvents(10)
	events.addEvent(sampleAnalyticsEvent(1))
	events.addEvent(sampleAnalyticsEvent(1))
	events.addEvent(sampleAnalyticsEvent(1))

	json, err := events.CollectorJSON(agentRunID)
	if nil != err {
		t.Fatal(err)
	}

	expected := `["12345",{"reservoir_size":10,"events_seen":3},[1,1,1]]`

	if string(json) != expected {
		t.Error(string(json), expected)
	}
	if 3 != events.numSeen {
		t.Error(events.numSeen)
	}
	if 3 != events.NumSaved() {
		t.Error(events.NumSaved())
	}
}

func TestEmpty(t *testing.T) {
	events := newAnalyticsEvents(10)
	json, err := events.CollectorJSON(agentRunID)
	if nil != err {
		t.Fatal(err)
	}
	if nil != json {
		t.Error(string(json))
	}
	if 0 != events.numSeen {
		t.Error(events.numSeen)
	}
	if 0 != events.NumSaved() {
		t.Error(events.NumSaved())
	}
}

func TestSampling(t *testing.T) {
	events := newAnalyticsEvents(3)
	events.addEvent(sampleAnalyticsEvent(10))
	events.addEvent(sampleAnalyticsEvent(1))
	events.addEvent(sampleAnalyticsEvent(9))
	events.addEvent(sampleAnalyticsEvent(2))
	events.addEvent(sampleAnalyticsEvent(8))
	events.addEvent(sampleAnalyticsEvent(3))

	json, err := events.CollectorJSON(agentRunID)
	if nil != err {
		t.Fatal(err)
	}
	if string(json) != `["12345",{"reservoir_size":3,"events_seen":6},[8,10,9]]` {
		t.Error(string(json))
	}
	if 6 != events.numSeen {
		t.Error(events.numSeen)
	}
	if 3 != events.NumSaved() {
		t.Error(events.NumSaved())
	}
}

func TestMergeEmpty(t *testing.T) {
	e1 := newAnalyticsEvents(10)
	e2 := newAnalyticsEvents(10)
	e1.Merge(e2)
	json, err := e1.CollectorJSON(agentRunID)
	if nil != err {
		t.Fatal(err)
	}
	if nil != json {
		t.Error(string(json))
	}
	if 0 != e1.numSeen {
		t.Error(e1.numSeen)
	}
	if 0 != e1.NumSaved() {
		t.Error(e1.NumSaved())
	}
}

func TestMergeFull(t *testing.T) {
	e1 := newAnalyticsEvents(2)
	e2 := newAnalyticsEvents(3)

	e1.addEvent(sampleAnalyticsEvent(5))
	e1.addEvent(sampleAnalyticsEvent(10))
	e1.addEvent(sampleAnalyticsEvent(15))

	e2.addEvent(sampleAnalyticsEvent(6))
	e2.addEvent(sampleAnalyticsEvent(12))
	e2.addEvent(sampleAnalyticsEvent(18))
	e2.addEvent(sampleAnalyticsEvent(24))

	e1.Merge(e2)
	json, err := e1.CollectorJSON(agentRunID)
	if nil != err {
		t.Fatal(err)
	}
	if string(json) != `["12345",{"reservoir_size":2,"events_seen":7},[18,24]]` {
		t.Error(string(json))
	}
	if 7 != e1.numSeen {
		t.Error(e1.numSeen)
	}
	if 2 != e1.NumSaved() {
		t.Error(e1.NumSaved())
	}
}

func TestAnalyticsEventMergeFailedSuccess(t *testing.T) {
	e1 := newAnalyticsEvents(2)
	e2 := newAnalyticsEvents(3)

	e1.addEvent(sampleAnalyticsEvent(5))
	e1.addEvent(sampleAnalyticsEvent(10))
	e1.addEvent(sampleAnalyticsEvent(15))

	e2.addEvent(sampleAnalyticsEvent(6))
	e2.addEvent(sampleAnalyticsEvent(12))
	e2.addEvent(sampleAnalyticsEvent(18))
	e2.addEvent(sampleAnalyticsEvent(24))

	e1.mergeFailed(e2)

	json, err := e1.CollectorJSON(agentRunID)
	if nil != err {
		t.Fatal(err)
	}
	if string(json) != `["12345",{"reservoir_size":2,"events_seen":7},[18,24]]` {
		t.Error(string(json))
	}
	if 7 != e1.numSeen {
		t.Error(e1.numSeen)
	}
	if 2 != e1.NumSaved() {
		t.Error(e1.NumSaved())
	}
	if 1 != e1.failedHarvests {
		t.Error(e1.failedHarvests)
	}
}

func TestAnalyticsEventMergeFailedLimitReached(t *testing.T) {
	e1 := newAnalyticsEvents(2)
	e2 := newAnalyticsEvents(3)

	e1.addEvent(sampleAnalyticsEvent(5))
	e1.addEvent(sampleAnalyticsEvent(10))
	e1.addEvent(sampleAnalyticsEvent(15))

	e2.addEvent(sampleAnalyticsEvent(6))
	e2.addEvent(sampleAnalyticsEvent(12))
	e2.addEvent(sampleAnalyticsEvent(18))
	e2.addEvent(sampleAnalyticsEvent(24))

	e2.failedHarvests = failedEventsAttemptsLimit

	e1.mergeFailed(e2)

	json, err := e1.CollectorJSON(agentRunID)
	if nil != err {
		t.Fatal(err)
	}
	if string(json) != `["12345",{"reservoir_size":2,"events_seen":3},[10,15]]` {
		t.Error(string(json))
	}
	if 3 != e1.numSeen {
		t.Error(e1.numSeen)
	}
	if 2 != e1.NumSaved() {
		t.Error(e1.NumSaved())
	}
	if 0 != e1.failedHarvests {
		t.Error(e1.failedHarvests)
	}
}

func analyticsEventBenchmarkHelper(b *testing.B, w jsonWriter) {
	events := newAnalyticsEvents(maxTxnEvents)
	event := analyticsEvent{eventStamp(1), w}
	for n := 0; n < maxTxnEvents; n++ {
		events.addEvent(event)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		js, err := events.CollectorJSON(agentRunID)
		if nil != err {
			b.Fatal(err, js)
		}
	}
}

func BenchmarkTxnEventsCollectorJSON(b *testing.B) {
	event := &TxnEvent{
		FinalName: "WebTransaction/Go/zip/zap",
		Start:     time.Now(),
		Duration:  2 * time.Second,
		Queuing:   1 * time.Second,
		Zone:      ApdexSatisfying,
		Attrs:     nil,
	}
	analyticsEventBenchmarkHelper(b, event)
}

func BenchmarkCustomEventsCollectorJSON(b *testing.B) {
	now := time.Now()
	ce, err := CreateCustomEvent("myEventType", map[string]interface{}{
		"string": "myString",
		"bool":   true,
		"int64":  int64(123),
		"nil":    nil,
	}, now)
	if nil != err {
		b.Fatal(err)
	}
	analyticsEventBenchmarkHelper(b, ce)
}

func BenchmarkErrorEventsCollectorJSON(b *testing.B) {
	e := TxnErrorFromResponseCode(time.Now(), 503)
	e.Stack = GetStackTrace(0)

	txnName := "WebTransaction/Go/zip/zap"
	event := &ErrorEvent{
		ErrorData: e,
		TxnEvent: TxnEvent{
			FinalName: txnName,
			Duration:  3 * time.Second,
			Attrs:     nil,
		},
	}
	analyticsEventBenchmarkHelper(b, event)
}
