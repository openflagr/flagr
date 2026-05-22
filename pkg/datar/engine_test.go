package datar

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := entity.NewTestDB()
	if err := db.AutoMigrate(entity.AutoMigrateTables...); err != nil {
		t.Fatal(err)
	}
	return db
}

func createFlag(t *testing.T, db *gorm.DB, id int64, key, desc string, enabled bool) {
	t.Helper()
	e := 0
	if enabled {
		e = 1
	}
	if err := db.Exec(
		`INSERT INTO flags (id, key, enabled, description, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`,
		id, key, e, desc,
	).Error; err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// Engine lifecycle
// ---------------------------------------------------------------------------

func TestNew_NilWhenDisabled(t *testing.T) {
	e := New(newTestDB(t), false, time.Second)
	assert.Nil(t, e)
}

func TestNew_Enabled(t *testing.T) {
	e := New(newTestDB(t), true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	assert.True(t, e.enabled)
	e.Shutdown()
}

func TestNew_NilMethodsAreSafe(t *testing.T) {
	var e *Engine
	assert.NotPanics(t, func() {
		e.Record(1, 1, 1)
		assert.Equal(t, 0, e.Len())
		assert.Nil(t, e.SnapshotAndReset())
		assert.NoError(t, e.Shutdown())
	})
}

// ---------------------------------------------------------------------------
// Record & buffer
// ---------------------------------------------------------------------------

func TestRecord_Increments(t *testing.T) {
	e := New(newTestDB(t), true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	e.Record(1, 1, 10)
	e.Record(1, 1, 10)
	e.Record(1, 2, 10)
	e.Record(2, 1, 20)

	assert.Equal(t, 3, e.Len()) // 3 unique (flag,variant,segment) combinations

	agg := e.SnapshotAndReset()
	assert.Equal(t, int32(2), agg[FlushKey{FlagID: 1, VariantID: 1, SegmentID: 10, Hour: time.Now().Truncate(time.Hour)}])
	assert.Equal(t, int32(1), agg[FlushKey{FlagID: 1, VariantID: 2, SegmentID: 10, Hour: time.Now().Truncate(time.Hour)}])
	assert.Equal(t, int32(1), agg[FlushKey{FlagID: 2, VariantID: 1, SegmentID: 20, Hour: time.Now().Truncate(time.Hour)}])
}

func TestRecord_AfterCloseIsNoop(t *testing.T) {
	e := New(newTestDB(t), true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	e.closed.Store(true)
	e.Record(1, 1, 1)
	assert.Equal(t, 0, e.Len())
}

func TestRecord_Concurrent(t *testing.T) {
	e := New(newTestDB(t), true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			e.Record(1, int64(n%3), int64(n%5))
		}(i)
	}
	wg.Wait()

	agg := e.SnapshotAndReset()
	total := int32(0)
	for _, v := range agg {
		total += v
	}
	assert.Equal(t, int32(100), total)
}

func TestSnapshotAndReset_EmptyBuffer(t *testing.T) {
	e := New(newTestDB(t), true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	agg := e.SnapshotAndReset()
	assert.Empty(t, agg)
	assert.Equal(t, 0, e.Len())
}

func TestLen_AfterRecordAndSnapshot(t *testing.T) {
	e := New(newTestDB(t), true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	e.Record(1, 1, 1)
	assert.Equal(t, 1, e.Len())
	e.SnapshotAndReset()
	assert.Equal(t, 0, e.Len())
}

// ---------------------------------------------------------------------------
// Shutdown
// ---------------------------------------------------------------------------

func TestShutdown_Idempotent(t *testing.T) {
	e := New(newTestDB(t), true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	assert.NoError(t, e.Shutdown())
	assert.NoError(t, e.Shutdown()) // second call is no-op
}

func TestShutdown_FlushesRemaining(t *testing.T) {
	db := newTestDB(t)
	createFlag(t, db, 1, "test", "flag", true)

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}

	e.Record(1, 1, 10)
	e.Record(1, 2, 10)

	assert.NoError(t, e.Shutdown())

	// Data should be in the DB now.
	rows, err := e.QueryFlagSummary(1, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
}

func TestShutdown_NilEngine(t *testing.T) {
	var e *Engine
	assert.NoError(t, e.Shutdown())
}

// ---------------------------------------------------------------------------
// Query flag summary
// ---------------------------------------------------------------------------

func TestQueryFlagSummary(t *testing.T) {
	db := newTestDB(t)
	createFlag(t, db, 1, "f1", "flag1", true)

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	now := time.Now().UTC().Truncate(time.Hour)
	prev := now.Add(-48 * time.Hour)

	if err := db.Exec(
		`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 1, 10, ?, 100)`, now,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(
		`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 2, 10, ?, 50)`, now,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(
		`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 1, 10, ?, 25)`, prev,
	).Error; err != nil {
		t.Fatal(err)
	}

	rows, err := e.QueryFlagSummary(1, prev.Add(-time.Hour), now.Add(time.Hour))
	assert.NoError(t, err)
	assert.Len(t, rows, 3)

	// Test aggregation helper directly.
	variants, segs, days := aggregateFlagSummary(rows)
	if len(variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(variants))
	}
	assert.Equal(t, int64(1), variants[0].ID)
	assert.Equal(t, int64(125), variants[0].Count)
	assert.Equal(t, int64(2), variants[1].ID)
	assert.Equal(t, int64(50), variants[1].Count)

	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	assert.Equal(t, int64(10), segs[0].ID)

	if len(days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(days))
	}
}

func TestQueryFlagSummary_NoData(t *testing.T) {
	db := newTestDB(t)
	createFlag(t, db, 1, "f1", "flag1", true)

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	rows, err := e.QueryFlagSummary(1, time.Now().Add(-7*24*time.Hour), time.Now())
	assert.NoError(t, err)
	assert.Empty(t, rows)
}

func TestQueryFlagSummary_NonexistentFlag(t *testing.T) {
	db := newTestDB(t)
	createFlag(t, db, 1, "f1", "flag1", true)

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	rows, err := e.QueryFlagSummary(999, time.Now().Add(-7*24*time.Hour), time.Now())
	assert.NoError(t, err)
	assert.Empty(t, rows)
}

// ---------------------------------------------------------------------------
// Query summary
// ---------------------------------------------------------------------------

func TestQuerySummary(t *testing.T) {
	db := newTestDB(t)
	createFlag(t, db, 1, "f1", "flag1", true)
	createFlag(t, db, 2, "f2", "flag2", false)

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	now := time.Now().UTC().Truncate(time.Hour)
	if err := db.Exec(
		`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 1, 10, ?, 100)`, now,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(
		`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 2, 10, ?, 50)`, now,
	).Error; err != nil {
		t.Fatal(err)
	}

	rows, err := e.QuerySummary(now.Add(-time.Hour), now.Add(time.Hour), 100, 0)
	assert.NoError(t, err)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	assert.Equal(t, int64(1), rows[0].FlagID)
	assert.Equal(t, int64(150), rows[0].TotalEvalCount)
	assert.True(t, rows[0].Enabled)

	assert.Equal(t, int64(2), rows[1].FlagID)
	assert.Equal(t, int64(0), rows[1].TotalEvalCount)
	assert.False(t, rows[1].Enabled)
}

func TestQuerySummary_NoData(t *testing.T) {
	db := newTestDB(t)
	createFlag(t, db, 1, "f1", "flag1", true)

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	rows, err := e.QuerySummary(time.Now().Add(-7*24*time.Hour), time.Now(), 100, 0)
	assert.NoError(t, err)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	assert.Equal(t, int64(0), rows[0].TotalEvalCount)
}

func TestQuerySummary_Pagination(t *testing.T) {
	db := newTestDB(t)

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	for i := 1; i <= 5; i++ {
		createFlag(t, db, int64(i), fmt.Sprintf("f-%d", i), fmt.Sprintf("flag-%d", i), true)
	}

	rows, err := e.QuerySummary(time.Now().Add(-7*24*time.Hour), time.Now().Add(time.Hour), 2, 1)
	assert.NoError(t, err)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	// Flags are ordered by total_eval_count DESC — all are 0, so any order.
	assert.Contains(t, []int64{2, 3, 4, 5}, rows[0].FlagID)
}

func TestQuerySummary_NilEngine(t *testing.T) {
	var e *Engine
	rows, err := e.QuerySummary(time.Now(), time.Now(), 100, 0)
	assert.Error(t, err)
	assert.Nil(t, rows)
}

func TestQuerySummary_DBError(t *testing.T) {
	db := newTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.Close()

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	_, err = e.QuerySummary(time.Now(), time.Now(), 100, 0)
	assert.Error(t, err, "should fail on closed DB")
}

func TestQueryFlagSummary_DBError(t *testing.T) {
	db := newTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.Close()

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	_, err = e.QueryFlagSummary(1, time.Now(), time.Now())
	assert.Error(t, err, "should fail on closed DB")
}


// flush / flushAggregates / flushLoop
// ---------------------------------------------------------------------------

func TestFlush_Direct(t *testing.T) {
	e := New(newTestDB(t), true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	e.flush() // flush with no data — no-op
	assert.Equal(t, 0, e.Len())

	e.Record(1, 1, 1)
	e.Record(1, 1, 1)
	e.Record(1, 2, 1)

	e.flush()

	assert.Equal(t, 0, e.Len(), "buffer should be empty after flush")
}

func TestFlushAggregates_Empty(t *testing.T) {
	e := New(newTestDB(t), true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	assert.NoError(t, e.flushAggregates(nil))
	assert.NoError(t, e.flushAggregates(map[FlushKey]int32{}))
}
// ---------------------------------------------------------------------------
// aggregateFlagSummary
// ---------------------------------------------------------------------------

func TestAggregateFlagSummary_Empty(t *testing.T) {
	variants, segs, days := aggregateFlagSummary(nil)
	assert.Empty(t, variants)
	assert.Empty(t, segs)
	assert.Empty(t, days)

	variants, segs, days = aggregateFlagSummary([]RawEvent{})
	assert.Empty(t, variants)
	assert.Empty(t, segs)
	assert.Empty(t, days)
}

func TestAggregateFlagSummary_SortsCorrectly(t *testing.T) {
	rows := []RawEvent{
		{VariantID: 1, SegmentID: 10, BucketHour: "2024-01-01T01:00:00Z", EvalCount: 10},
		{VariantID: 1, SegmentID: 10, BucketHour: "2024-01-01T02:00:00Z", EvalCount: 5},
		{VariantID: 2, SegmentID: 10, BucketHour: "2024-01-01T01:00:00Z", EvalCount: 20},
		{VariantID: 2, SegmentID: 20, BucketHour: "2024-01-01T01:00:00Z", EvalCount: 30},
	}

	variants, segs, days := aggregateFlagSummary(rows)

	// variants sorted by count desc: 2(50), 1(15)
	if len(variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(variants))
	}
	assert.Equal(t, int64(2), variants[0].ID)
	assert.Equal(t, int64(50), variants[0].Count)
	assert.Equal(t, int64(1), variants[1].ID)
	assert.Equal(t, int64(15), variants[1].Count)

	// segments sorted by count desc: 10(35), 20(30)
	if len(segs) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segs))
	}
	assert.Equal(t, int64(10), segs[0].ID)
	assert.Equal(t, int64(35), segs[0].Count)
	assert.Equal(t, int64(20), segs[1].ID)
	assert.Equal(t, int64(30), segs[1].Count)

	// days sorted by date asc
	if len(days) != 1 {
		t.Fatalf("expected 1 day, got %d", len(days))
	}
	assert.Equal(t, "2024-01-01", days[0].Date)
	assert.Equal(t, int64(65), days[0].Count)
}

func TestAggregateFlagSummary_ZeroSegmentID(t *testing.T) {
	rows := []RawEvent{
		{VariantID: 1, SegmentID: 0, BucketHour: "2024-01-01T00:00:00Z", EvalCount: 10},
		{VariantID: 1, SegmentID: 5, BucketHour: "2024-01-01T00:00:00Z", EvalCount: 20},
	}
	_, segs, _ := aggregateFlagSummary(rows)
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	assert.Equal(t, int64(5), segs[0].ID)
}

func TestAggregateFlagSummary_ShortBucketHour(t *testing.T) {
	rows := []RawEvent{
		{VariantID: 1, SegmentID: 1, BucketHour: "short", EvalCount: 10},
	}
	_, _, days := aggregateFlagSummary(rows)
	assert.Empty(t, days)
}

func TestAggregateFlagSummary_DaySort(t *testing.T) {
	rows := []RawEvent{
		{VariantID: 1, SegmentID: 1, BucketHour: "2024-01-03T00:00:00Z", EvalCount: 10},
		{VariantID: 1, SegmentID: 1, BucketHour: "2024-01-01T00:00:00Z", EvalCount: 20},
		{VariantID: 1, SegmentID: 1, BucketHour: "2024-01-02T00:00:00Z", EvalCount: 30},
	}
	_, _, days := aggregateFlagSummary(rows)
	if len(days) != 3 {
		t.Fatalf("expected 3 days, got %d", len(days))
	}
	assert.Equal(t, "2024-01-01", days[0].Date)
	assert.Equal(t, "2024-01-02", days[1].Date)
	assert.Equal(t, "2024-01-03", days[2].Date)
}
