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


}

func TestQueryFlagSummaryBreakdown(t *testing.T) {
	db := newTestDB(t)
	createFlag(t, db, 1, "f1", "flag1", true)

	e := New(db, true, time.Hour)
	defer e.Shutdown()

	now := time.Now().UTC().Truncate(time.Hour)
	prev := now.Add(-48 * time.Hour)

	assert.NoError(t, db.Exec(`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 1, 10, ?, 100)`, now).Error)
	assert.NoError(t, db.Exec(`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 2, 10, ?, 50)`, now).Error)
	assert.NoError(t, db.Exec(`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 1, 10, ?, 25)`, prev).Error)

	summary, err := e.QueryFlagSummaryBreakdown(1, prev.Add(-time.Hour), now.Add(time.Hour))
	assert.NoError(t, err)
	assert.NotNil(t, summary)


	// Variant 1 (100+25=125) > variant 2 (50) → variant 1 first.
	if assert.Len(t, summary.Variants, 2) {
		assert.Equal(t, int64(1), summary.Variants[0].VariantID)
		assert.Equal(t, int64(125), summary.Variants[0].Count)
		assert.Equal(t, int64(2), summary.Variants[1].VariantID)
		assert.Equal(t, int64(50), summary.Variants[1].Count)
	}

	// Single segment.
	assert.Len(t, summary.Segments, 1)
	assert.Equal(t, int64(10), summary.Segments[0].SegmentID)
	assert.Equal(t, int64(175), summary.Segments[0].Count)

	// Two unique days.
	assert.Len(t, summary.Days, 2)
	assert.Equal(t, now.Format("2006-01-02"), summary.Days[1].Date, "most recent day last")
}

func TestQueryFlagSummaryBreakdown_NilEngine(t *testing.T) {
	var e *Engine
	summary, err := e.QueryFlagSummaryBreakdown(1, time.Now(), time.Now())
	assert.Error(t, err)
	assert.Nil(t, summary)
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
	if len(rows) != 1 {
		t.Fatalf("expected 1 row (only flags with traffic), got %d", len(rows))
	}
	assert.Equal(t, int64(1), rows[0].FlagID)
	assert.Equal(t, int64(150), rows[0].TotalEvalCount)
	assert.True(t, rows[0].Enabled)
}

func TestQuerySummary_NoData(t *testing.T) {
	db := newTestDB(t)
	createFlag(t, db, 1, "f1", "flag1", true)

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	// No events for any flag → empty result (only flags with traffic are returned).
	rows, err := e.QuerySummary(time.Now().Add(-7*24*time.Hour), time.Now(), 100, 0)
	assert.NoError(t, err)
	assert.Empty(t, rows)
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

	// Seed events for flags 1-4 with descending counts.
	now := time.Now().UTC().Truncate(time.Hour)
	for i := 1; i <= 4; i++ {
		if err := db.Exec(
			`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (?, 1, 1, ?, ?)`,
			i, now, 100-i,
		).Error; err != nil {
			t.Fatal(err)
		}
	}

	// Page 2 (limit=2, offset=1): positions 1-2 sorted by count desc.
	// Flag 1(99), Flag 2(98), Flag 3(97), Flag 4(96) → returns [Flag 2(98), Flag 3(97)].
	rows, err := e.QuerySummary(now.Add(-time.Hour), now.Add(time.Hour), 2, 1)
	assert.NoError(t, err)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	assert.Equal(t, int64(2), rows[0].FlagID, "flag 2 has 98 count")
	assert.Equal(t, int64(98), rows[0].TotalEvalCount)
	assert.Equal(t, int64(3), rows[1].FlagID, "flag 3 has 97 count")
	assert.Equal(t, int64(97), rows[1].TotalEvalCount)
}

func TestQuerySummary_NilEngine(t *testing.T) {
	var e *Engine
	rows, err := e.QuerySummary(time.Now(), time.Now(), 100, 0)
	assert.Error(t, err)
	assert.Nil(t, rows)
}

func TestQuerySummary_DBError(t *testing.T) {
	db := newTestDB(t)
	// Drop the events table to cause a query error.
	if err := db.Exec("DROP TABLE datar_hourly_events").Error; err != nil {
		t.Fatal(err)
	}

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	_, err := e.QuerySummary(time.Now(), time.Now(), 100, 0)
	assert.Error(t, err, "should fail with missing events table")
}

func TestQueryFlagSummary_DBError(t *testing.T) {
	db := newTestDB(t)
	if err := db.Exec("DROP TABLE datar_hourly_events").Error; err != nil {
		t.Fatal(err)
	}

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	_, err := e.QueryFlagSummary(1, time.Now(), time.Now())
	assert.Error(t, err, "should fail with missing events table")
}

func TestFlush_DBError(t *testing.T) {
	db := newTestDB(t)
	createFlag(t, db, 1, "f1", "test", true)

	e := New(db, true, time.Hour)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	e.Record(1, 1, 1)

	// Drop the events table — flushAggregates will now fail.
	if err := db.Exec("DROP TABLE datar_hourly_events").Error; err != nil {
		t.Fatal(err)
	}

	// Shutdown returns the flush error when it can't write to DB.
	err := e.Shutdown()
	assert.Error(t, err, "should fail flushing to dropped table")

}

func TestFlushLoop_TickerFires(t *testing.T) {
	db := newTestDB(t)
	createFlag(t, db, 1, "f1", "test", true)

	e := New(db, true, 1*time.Millisecond)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	defer e.Shutdown()

	e.Record(1, 1, 1)

	// Give the ticker time to fire at least once.
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, e.Len(), "buffer should be drained by ticker-triggered flush")
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