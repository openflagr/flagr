package datar

import (
	"fmt"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// testDB creates a fresh in-memory SQLite DB for testing.
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := entity.NewSQLiteDB(":memory:")
	assert.NotNil(t, db)
	return db
}

// seedHourly inserts test data into the datar_hourly_events table using SQLite-compatible upsert.
func seedHourly(t *testing.T, db *gorm.DB, rows []entity.HourlyEvent) {
	t.Helper()
	for _, r := range rows {
		err := db.Exec(`
			INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(flag_id, variant_id, segment_id, bucket_hour)
			DO UPDATE SET eval_count = datar_hourly_events.eval_count + excluded.eval_count
		`, r.FlagID, r.VariantID, r.SegmentID, r.BucketHour, r.EvalCount).Error
		assert.NoError(t, err)
	}
}

func TestFlushAggregates(t *testing.T) {
	db := testDB(t)
	s := NewTestStore(db)

	now := time.Now().UTC().Truncate(time.Hour)
	agg := map[FlushKey]int32{
		{FlagID: 1, VariantID: 1, SegmentID: 10, Hour: now}: 100,
		{FlagID: 1, VariantID: 2, SegmentID: 10, Hour: now}: 200,
		{FlagID: 2, VariantID: 1, SegmentID: 0, Hour: now}:  50,
	}

	err := s.FlushAggregates(agg)
	assert.NoError(t, err)

	// Verify via query
	var count int64
	db.Model(&entity.HourlyEvent{}).Count(&count)
	assert.Equal(t, int64(3), count)

	// Verify additive: flush the same keys again
	err = s.FlushAggregates(agg)
	assert.NoError(t, err)

	db.Model(&entity.HourlyEvent{}).Count(&count)
	assert.Equal(t, int64(3), count, "should still have 3 rows after additive upsert")

	var totalSum int64
	db.Model(&entity.HourlyEvent{}).Select("COALESCE(SUM(eval_count), 0)").Scan(&totalSum)
	assert.Equal(t, int64(700), totalSum, "counts should double after second flush")
}

func TestFlushAggregates_Empty(t *testing.T) {
	db := testDB(t)
	s := NewTestStore(db)
	err := s.FlushAggregates(nil)
	assert.NoError(t, err)
	err = s.FlushAggregates(map[FlushKey]int32{})
	assert.NoError(t, err)
}

func TestQueryFlagSummary(t *testing.T) {
	db := testDB(t)
	s := NewTestStore(db)

	now := time.Now().UTC().Truncate(time.Hour)
	prev := now.Add(-1 * time.Hour)

	// Seed segments table.
	assert.NoError(t, db.Exec("INSERT INTO segments (id, description, flag_id, rank, rollout_percent) VALUES (10, 'US users', 1, 1, 100)").Error)
	assert.NoError(t, db.Exec("INSERT INTO segments (id, description, flag_id, rank, rollout_percent) VALUES (20, 'EU users', 1, 2, 100)").Error)

	seedHourly(t, db, []entity.HourlyEvent{
		{FlagID: 1, VariantID: 1, SegmentID: 10, BucketHour: now, EvalCount: 50},
		{FlagID: 1, VariantID: 2, SegmentID: 10, BucketHour: now, EvalCount: 30},
		{FlagID: 1, VariantID: 1, SegmentID: 20, BucketHour: now, EvalCount: 20},
		{FlagID: 1, VariantID: 1, SegmentID: 10, BucketHour: prev, EvalCount: 10},
	})

	from := prev.Add(-time.Hour)
	to := now.Add(time.Hour)
	rows, err := s.QueryFlagSummary(1, from, to)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(rows), "should return all 4 raw event rows")
}

func TestQueryFlagSummary_NoData(t *testing.T) {
	db := testDB(t)
	s := NewTestStore(db)

	from := time.Now().UTC().Add(-24 * time.Hour)
	to := time.Now().UTC().Add(time.Hour)
	rows, err := s.QueryFlagSummary(999, from, to)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(rows), "non-existent flag should return empty")
}



func TestQuerySummary(t *testing.T) {
	db := testDB(t)
	s := NewTestStore(db)

	// Create a flag in the flags table (use SQLite-compatible datetime).
	assert.NoError(t, db.Exec("INSERT INTO flags (id, key, enabled, description, created_at, updated_at) VALUES (1, 'test-flag', 1, 'Test flag', datetime('now'), datetime('now'))").Error)

	now := time.Now().UTC().Truncate(time.Hour)
	seedHourly(t, db, []entity.HourlyEvent{
		{FlagID: 1, VariantID: 1, SegmentID: 0, BucketHour: now, EvalCount: 100},
		{FlagID: 1, VariantID: 2, SegmentID: 0, BucketHour: now, EvalCount: 50},
	})

	from := now.Add(-24 * time.Hour)
	to := now.Add(time.Hour)
	rows, err := s.QuerySummary(from, to, 10, 0)
	assert.NoError(t, err)
	if assert.Equal(t, 1, len(rows)) {
		assert.Equal(t, "test-flag", rows[0].FlagKey)
		assert.Equal(t, int64(150), rows[0].TotalEvalCount)
		assert.True(t, rows[0].Enabled)
	}
}

func TestModels_Migration(t *testing.T) {
	db := testDB(t)
	assert.True(t, db.Migrator().HasTable(&entity.HourlyEvent{}))
}

func TestFlushAggregates_ZeroCount(t *testing.T) {
	db := testDB(t)
	s := NewTestStore(db)

	now := time.Now().UTC().Truncate(time.Hour)
	agg := map[FlushKey]int32{
		{FlagID: 1, VariantID: 1, SegmentID: 10, Hour: now}: 0,
	}

	err := s.FlushAggregates(agg)
	assert.NoError(t, err)

	var count int64
	db.Model(&entity.HourlyEvent{}).Count(&count)
	assert.Equal(t, int64(1), count, "zero-count key should still create a row")

	var totalSum int64
	db.Model(&entity.HourlyEvent{}).Select("COALESCE(SUM(eval_count), 0)").Scan(&totalSum)
	assert.Equal(t, int64(0), totalSum, "zero count should persist as 0")
}

func TestFlushAggregates_MultipleBuckets(t *testing.T) {
	db := testDB(t)
	s := NewTestStore(db)

	now := time.Now().UTC().Truncate(time.Hour)
	h1 := now
	h2 := now.Add(-1 * time.Hour)
	h3 := now.Add(-2 * time.Hour)

	agg := map[FlushKey]int32{
		{FlagID: 1, VariantID: 1, SegmentID: 10, Hour: h1}: 10,
		{FlagID: 1, VariantID: 1, SegmentID: 10, Hour: h2}: 20,
		{FlagID: 1, VariantID: 1, SegmentID: 10, Hour: h3}: 30,
	}

	err := s.FlushAggregates(agg)
	assert.NoError(t, err)

	var count int64
	db.Model(&entity.HourlyEvent{}).Count(&count)
	assert.Equal(t, int64(3), count)

	var totalSum int64
	db.Model(&entity.HourlyEvent{}).Select("COALESCE(SUM(eval_count), 0)").Scan(&totalSum)
	assert.Equal(t, int64(60), totalSum)
}

func TestQuerySummary_Empty(t *testing.T) {
	db := testDB(t)
	s := NewTestStore(db)

	// Create a flag with no hourly data.
	assert.NoError(t, db.Exec("INSERT INTO flags (id, key, enabled, description, created_at, updated_at) VALUES (1, 'empty-flag', 0, 'No data', datetime('now'), datetime('now'))").Error)

	from := time.Now().UTC().Add(-24 * time.Hour)
	to := time.Now().UTC().Add(time.Hour)
	rows, err := s.QuerySummary(from, to, 10, 0)
	assert.NoError(t, err)
	if assert.Equal(t, 1, len(rows)) {
		assert.Equal(t, int64(0), rows[0].TotalEvalCount, "flag with no data should show 0 count")
	}
}

func TestQuerySummary_Pagination(t *testing.T) {
	db := testDB(t)
	s := NewTestStore(db)

	for i := 1; i <= 3; i++ {
		assert.NoError(t, db.Exec("INSERT INTO flags (id, key, enabled, description, created_at, updated_at) VALUES (?, ?, 1, 'flag', datetime('now'), datetime('now'))", i, fmt.Sprintf("flag-%d", i)).Error)
	}

	from := time.Now().UTC().Add(-24 * time.Hour)
	to := time.Now().UTC().Add(time.Hour)

	rows, err := s.QuerySummary(from, to, 2, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(rows), "limit=2 should return 2 rows")

	rows, err = s.QuerySummary(from, to, 10, 2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rows), "offset=2 should skip first 2")
}


