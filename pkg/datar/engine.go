package datar

import (
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openflagr/flagr/pkg/entity"
)

const flushRetries = 3

var errNilEngine = errors.New("datar: engine is nil")

// FlushKey identifies one aggregate dimension set for the in-memory buffer.
type FlushKey struct {
	FlagID    int64
	VariantID int64
	SegmentID int64
	Hour      time.Time // Truncated to the hour so struct equality works as a sync.Map key.
}

// SummaryRow is a single flag's aggregate summary for the list view.
type SummaryRow struct {
	FlagID         int64
	FlagKey        string
	Enabled        bool
	Description    string
	TotalEvalCount int64
	LastEvaluated  string
}

// VariantEntry is one variant's aggregated count.
type VariantEntry struct {
	VariantID int64
	Count     int64
}

// SegmentEntry is one segment's aggregated count.
type SegmentEntry struct {
	SegmentID int64
	Count     int64
}

// DayEntry is one calendar day's aggregated count.
type DayEntry struct {
	Date  string // YYYY-MM-DD
	Count int64
}

// FlagSummaryBreakdown is the pre-aggregated breakdown for a single flag.
type FlagSummaryBreakdown struct {
	FlagID   int64
	Variants []VariantEntry
	Segments []SegmentEntry
	Days     []DayEntry
}

// Engine is the complete Datar analytics engine.
// It aggregates evaluation counts in-memory, periodically flushes to the DB,
// and serves aggregate queries — all in one self-contained struct.
type Engine struct {
	buffer sync.Map // FlushKey → *int32

	db          *gorm.DB
	addEvalExpr string

	flushInterval time.Duration
	closeCh       chan struct{}
	wg            sync.WaitGroup
	shutdownOnce  sync.Once
	closed        atomic.Bool
}

// New creates an Engine and starts its flush loop.
// Returns nil when enabled is false.
func New(db *gorm.DB, enabled bool, flushInterval time.Duration) *Engine {
	if !enabled {
		return nil
	}

	addEvalExpr := "datar_hourly_events.eval_count + EXCLUDED.eval_count"
	if db.Name() == "mysql" {
		addEvalExpr = "datar_hourly_events.eval_count + VALUES(eval_count)"
	}

	e := &Engine{
		db:            db,
		addEvalExpr:   addEvalExpr,
		flushInterval: flushInterval,
		closeCh:       make(chan struct{}),
	}

	e.wg.Add(1)
	go e.flushLoop()
	logrus.Info("Datar: started aggregate analytics")
	return e
}

// Record increments the counter for the given EvalResult.
// Safe to call from concurrent goroutines. Safe on nil receiver.
func (e *Engine) Record(flagID, variantID, segmentID int64) {
	if e == nil || e.closed.Load() {
		return
	}

	key := FlushKey{
		FlagID:    flagID,
		VariantID: variantID,
		SegmentID: segmentID,
		Hour:      time.Now().Truncate(time.Hour),
	}

	actual, _ := e.buffer.LoadOrStore(key, new(int32))
	atomic.AddInt32(actual.(*int32), 1)
}

// SnapshotAndReset drains the buffer and returns frozen counts.
// The returned map is safe to read without holding any lock.
func (e *Engine) SnapshotAndReset() map[FlushKey]int32 {
	if e == nil {
		return nil
	}
	result := make(map[FlushKey]int32)
	e.buffer.Range(func(k, v any) bool {
		result[k.(FlushKey)] = atomic.LoadInt32(v.(*int32))
		e.buffer.Delete(k)
		return true
	})
	return result
}

// Len returns the number of distinct keys in the buffer.
func (e *Engine) Len() int {
	if e == nil {
		return 0
	}
	n := 0
	e.buffer.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}

// QuerySummary returns all flags with traffic totals in the given time range.
// Only includes flags that have actual evaluation traffic in the window.
func (e *Engine) QuerySummary(from, to time.Time, limit, offset int) ([]SummaryRow, error) {
	if e == nil {
		return nil, errNilEngine
	}
	sub := e.db.Model(&entity.HourlyEvent{}).
		Select("flag_id, SUM(eval_count) AS total_count, MAX(updated_at) AS last_evaluated_at").
		Where("bucket_hour >= ? AND bucket_hour < ?", from, to).
		Group("flag_id")

	var rows []SummaryRow
	err := e.db.Model(&entity.Flag{}).
		Select("flags.id AS flag_id, flags.key AS flag_key, flags.enabled, flags.description, agg.total_count AS total_eval_count, agg.last_evaluated_at AS last_evaluated_at").
		Joins("JOIN (?) AS agg ON agg.flag_id = flags.id", sub).
		Order("agg.total_count DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		logrus.WithError(err).Error("Datar: QuerySummary failed")
		return nil, err
	}
	return rows, nil
}

// QueryFlagSummaryBreakdown returns the pre-aggregated breakdown for a single flag.
// It queries raw event rows and aggregates them into variant, segment, and day buckets,
// sorted by count descending (variants, segments) or by date ascending (days).
func (e *Engine) QueryFlagSummaryBreakdown(flagID int64, from, to time.Time) (*FlagSummaryBreakdown, error) {
	if e == nil {
		return nil, errNilEngine
	}

	var rawRows []struct {
		VariantID  int64
		SegmentID  int64
		BucketHour string
		EvalCount  int32
	}
	err := e.db.Model(&entity.HourlyEvent{}).
		Select("variant_id, segment_id, bucket_hour, eval_count").
		Where("flag_id = ? AND bucket_hour >= ? AND bucket_hour < ?", flagID, from, to).
		Scan(&rawRows).Error
	if err != nil {
		logrus.WithError(err).Error("Datar: QueryFlagSummaryBreakdown failed")
		return nil, err
	}

	// Aggregate into three bucket types in a single pass.
	variantTotals := make(map[int64]int64)
	segIDs := make(map[int64]int64)
	dayCounts := make(map[string]int64)

	for _, r := range rawRows {
		variantTotals[r.VariantID] += int64(r.EvalCount)
		if r.SegmentID > 0 {
			segIDs[r.SegmentID] += int64(r.EvalCount)
		}
		if len(r.BucketHour) >= 10 {
			dayCounts[r.BucketHour[:10]] += int64(r.EvalCount)
		}
	}

	variants := make([]VariantEntry, 0, len(variantTotals))
	for id, count := range variantTotals {
		variants = append(variants, VariantEntry{VariantID: id, Count: count})
	}
	sort.Slice(variants, func(i, j int) bool {
		return variants[i].Count > variants[j].Count
	})

	segs := make([]SegmentEntry, 0, len(segIDs))
	for id, count := range segIDs {
		segs = append(segs, SegmentEntry{SegmentID: id, Count: count})
	}
	sort.Slice(segs, func(i, j int) bool {
		return segs[i].Count > segs[j].Count
	})

	days := make([]DayEntry, 0, len(dayCounts))
	for dateStr, count := range dayCounts {
		days = append(days, DayEntry{Date: dateStr, Count: count})
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date < days[j].Date
	})

	return &FlagSummaryBreakdown{
		FlagID:   flagID,
		Variants: variants,
		Segments: segs,
		Days:     days,
	}, nil
}

// Shutdown stops the flush loop and flushes remaining in-memory counts to the DB.
func (e *Engine) Shutdown() error {
	if e == nil {
		return nil
	}

	var shutdownErr error
	e.shutdownOnce.Do(func() {
		logrus.Info("Datar: shutting down")
		e.closed.Store(true)
		close(e.closeCh)
		e.wg.Wait()

		agg := e.SnapshotAndReset()
		if len(agg) > 0 {
			logrus.WithField("keys", len(agg)).Info("Datar: flushing remaining aggregates on shutdown")
			if err := e.flushWithRetry(agg); err != nil {
				logrus.WithError(err).Error("Datar: shutdown flush failed, data may be lost")
				shutdownErr = err
				return
			}
		}
		logrus.Info("Datar: shutdown complete")
	})
	return shutdownErr
}

// flushLoop runs on a goroutine, flushing at the configured interval.
func (e *Engine) flushLoop() {
	defer e.wg.Done()
	ticker := time.NewTicker(e.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.closeCh:
			return
		case <-ticker.C:
			e.flush()
		}
	}
}

func (e *Engine) flush() {
	agg := e.SnapshotAndReset()
	if len(agg) == 0 {
		return
	}
	logrus.WithField("keys", len(agg)).Debug("Datar: flushing aggregates")
	if err := e.flushWithRetry(agg); err != nil {
		logrus.WithError(err).Error("Datar: flush failed after retries, data in this cycle is lost")
	}
}

// flushWithRetry attempts to flush aggregates up to flushRetries times
// before giving up. This is best-effort: if the container restarts,
// in-flight aggregates are lost regardless.
func (e *Engine) flushWithRetry(agg map[FlushKey]int32) error {
	var err error
	for attempt := 1; attempt <= flushRetries; attempt++ {
		if err = e.flushAggregates(agg); err == nil {
			return nil
		}
		logrus.WithError(err).WithField("attempt", attempt).Warn("Datar: flush attempt failed")
	}
	return err
}

// flushAggregates writes the snapshot to the DB using additive UPSERT.
func (e *Engine) flushAggregates(agg map[FlushKey]int32) error {
	if len(agg) == 0 {
		return nil
	}
	now := time.Now()
	records := make([]entity.HourlyEvent, 0, len(agg))
	for k, count := range agg {
		records = append(records, entity.HourlyEvent{
			FlagID:     k.FlagID,
			VariantID:  k.VariantID,
			SegmentID:  k.SegmentID,
			BucketHour: k.Hour,
			EvalCount:  count,
			UpdatedAt:  now,
		})
	}
	return e.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "flag_id"},
			{Name: "variant_id"},
			{Name: "segment_id"},
			{Name: "bucket_hour"},
		},
		DoUpdates: clause.Set{{
			Column: clause.Column{Name: "eval_count"},
			Value:  gorm.Expr(e.addEvalExpr),
		}, {
			Column: clause.Column{Name: "updated_at"},
			Value:  gorm.Expr("CURRENT_TIMESTAMP"),
		}},
	}).Create(&records).Error
}
