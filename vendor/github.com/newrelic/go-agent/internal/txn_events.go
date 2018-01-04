package internal

import (
	"bytes"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// DatastoreExternalTotals contains overview of external and datastore calls
// made during a transaction.
type DatastoreExternalTotals struct {
	externalCallCount  uint64
	externalDuration   time.Duration
	datastoreCallCount uint64
	datastoreDuration  time.Duration
}

// WriteJSON prepares JSON in the format expected by the collector.
func (e *TxnEvent) WriteJSON(buf *bytes.Buffer) {
	w := jsonFieldsWriter{buf: buf}
	buf.WriteByte('[')
	buf.WriteByte('{')
	w.stringField("type", "Transaction")
	w.stringField("name", e.FinalName)
	w.floatField("timestamp", timeToFloatSeconds(e.Start))
	w.floatField("duration", e.Duration.Seconds())
	if ApdexNone != e.Zone {
		w.stringField("nr.apdexPerfZone", e.Zone.label())
	}
	if e.Queuing > 0 {
		w.floatField("queueDuration", e.Queuing.Seconds())
	}
	if e.externalCallCount > 0 {
		w.intField("externalCallCount", int64(e.externalCallCount))
		w.floatField("externalDuration", e.externalDuration.Seconds())
	}
	if e.datastoreCallCount > 0 {
		// Note that "database" is used for the keys here instead of
		// "datastore" for historical reasons.
		w.intField("databaseCallCount", int64(e.datastoreCallCount))
		w.floatField("databaseDuration", e.datastoreDuration.Seconds())
	}
	if e.CrossProcess.Used() {
		if e.CrossProcess.ClientID != "" {
			w.stringField("client_cross_process_id", e.CrossProcess.ClientID)
		}
		if e.CrossProcess.TripID != "" {
			w.stringField("nr.tripId", e.CrossProcess.TripID)
		}
		if e.CrossProcess.PathHash != "" {
			w.stringField("nr.pathHash", e.CrossProcess.PathHash)
		}
		if e.CrossProcess.ReferringPathHash != "" {
			w.stringField("nr.referringPathHash", e.CrossProcess.ReferringPathHash)
		}
		if e.CrossProcess.GUID != "" {
			w.stringField("nr.guid", e.CrossProcess.GUID)
		}
		if e.CrossProcess.ReferringTxnGUID != "" {
			w.stringField("nr.referringTransactionGuid", e.CrossProcess.ReferringTxnGUID)
		}
		if len(e.CrossProcess.AlternatePathHashes) > 0 {
			hashes := make([]string, 0, len(e.CrossProcess.AlternatePathHashes))
			for hash := range e.CrossProcess.AlternatePathHashes {
				hashes = append(hashes, hash)
			}
			sort.Strings(hashes)
			w.stringField("nr.alternatePathHashes", strings.Join(hashes, ","))
		}
		if e.CrossProcess.IsSynthetics() {
			w.stringField("nr.syntheticsResourceId", e.CrossProcess.Synthetics.ResourceID)
			w.stringField("nr.syntheticsJobId", e.CrossProcess.Synthetics.JobID)
			w.stringField("nr.syntheticsMonitorId", e.CrossProcess.Synthetics.MonitorID)
		}
	}
	buf.WriteByte('}')
	buf.WriteByte(',')
	userAttributesJSON(e.Attrs, buf, destTxnEvent, nil)
	buf.WriteByte(',')
	agentAttributesJSON(e.Attrs, buf, destTxnEvent)
	buf.WriteByte(']')
}

// MarshalJSON is used for testing.
func (e *TxnEvent) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 256))

	e.WriteJSON(buf)

	return buf.Bytes(), nil
}

type txnEvents struct {
	events *analyticsEvents
}

func newTxnEvents(max int) *txnEvents {
	return &txnEvents{
		events: newAnalyticsEvents(max),
	}
}

func (events *txnEvents) AddTxnEvent(e *TxnEvent) {
	stamp := eventStamp(rand.Float32())

	// Synthetics events always get priority: normal event stamps are in the
	// range [0.0,1.0), so adding 1 means that a Synthetics event will always
	// win.
	if e.CrossProcess.IsSynthetics() {
		stamp += 1.0
	}

	events.events.addEvent(analyticsEvent{stamp, e})
}

func (events *txnEvents) MergeIntoHarvest(h *Harvest) {
	h.TxnEvents.events.mergeFailed(events.events)
}

func (events *txnEvents) Data(agentRunID string, harvestStart time.Time) ([]byte, error) {
	return events.events.CollectorJSON(agentRunID)
}

func (events *txnEvents) numSeen() float64  { return events.events.NumSeen() }
func (events *txnEvents) numSaved() float64 { return events.events.NumSaved() }
