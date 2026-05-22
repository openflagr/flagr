package handler

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/datar"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestDatarEndpoints_Summary(t *testing.T) {
	// Reset singleton for test isolation.
	singletonDatarOnce = sync.Once{}
	singletonDatar = nil
	t.Cleanup(func() { singletonDatarOnce = sync.Once{}; singletonDatar = nil })

	defer gostub.Stub(&config.Config.DatarEnabled, true).Reset()
	defer gostub.Stub(&config.Config.DatarFlushInterval, 24*time.Hour).Reset()

	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
	db.AutoMigrate(entity.AutoMigrateTables...)

	// Initialize Datar (starts flush loop with long interval).
	d := GetDatar()
	assert.NotNil(t, d)

	// Create flags.
	assert.NoError(t, db.Exec(`INSERT INTO flags (id, key, enabled, description, created_at, updated_at) VALUES (1, 'test-flag', 1, 'Test flag', datetime('now'), datetime('now'))`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO flags (id, key, enabled, description, created_at, updated_at) VALUES (2, 'other-flag', 0, 'Disabled', datetime('now'), datetime('now'))`).Error)

	// Seed hourly data.
	now := time.Now().UTC().Truncate(time.Hour)
	assert.NoError(t, db.Exec(`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 1, 10, ?, 100)`, now).Error)
	assert.NoError(t, db.Exec(`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 2, 10, ?, 50)`, now).Error)

	// Flag 2 has no hourly data — should still appear with 0 total.

	resp := HandleGetDatarSummary(datar.GetDatarSummaryParams{})
	assert.NotNil(t, resp)

	okResp, ok := resp.(*datar.GetDatarSummaryOK)
	if !ok {
		t.Fatalf("expected *datar.GetDatarSummaryOK, got %T", resp)
	}

	assert.Len(t, okResp.Payload.Flags, 2, "should return both flags")

	// Flag 1 has data.
	f1 := okResp.Payload.Flags[0]
	assert.Equal(t, int64(1), f1.FlagID)
	assert.Equal(t, "test-flag", f1.FlagKey)
	assert.Equal(t, int64(150), f1.TotalEvalCount)
	assert.True(t, f1.Enabled)

	// Flag 2 has no data — 0 count.
	f2 := okResp.Payload.Flags[1]
	assert.Equal(t, int64(2), f2.FlagID)
	assert.Equal(t, int64(0), f2.TotalEvalCount)
}

func TestDatarEndpoints_FlagSummary(t *testing.T) {
	// Reset singleton for test isolation.
	singletonDatarOnce = sync.Once{}
	singletonDatar = nil
	t.Cleanup(func() { singletonDatarOnce = sync.Once{}; singletonDatar = nil })

	defer gostub.Stub(&config.Config.DatarEnabled, true).Reset()
	defer gostub.Stub(&config.Config.DatarFlushInterval, 24*time.Hour).Reset()

	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
	db.AutoMigrate(entity.AutoMigrateTables...)

	d := GetDatar()
	assert.NotNil(t, d)

	// Create flag + variant + segment.
	assert.NoError(t, db.Exec(`INSERT INTO flags (id, key, enabled, description, created_at, updated_at) VALUES (1, 'test-flag', 1, 'Test', datetime('now'), datetime('now'))`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO variants (id, flag_id, key) VALUES (1, 1, 'control')`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO variants (id, flag_id, key) VALUES (2, 1, 'treatment')`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO segments (id, flag_id, description, rank, rollout_percent) VALUES (10, 1, 'US users', 1, 100)`).Error)

	// Seed hourly data across multiple days.
	now := time.Now().UTC().Truncate(time.Hour)
	prev := now.Add(-48 * time.Hour)

	assert.NoError(t, db.Exec(`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 1, 10, ?, 100)`, now).Error)
	assert.NoError(t, db.Exec(`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 2, 10, ?, 50)`, now).Error)
	assert.NoError(t, db.Exec(`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (1, 1, 10, ?, 25)`, prev).Error)

	resp := HandleGetDatarFlagSummary(datar.GetDatarFlagSummaryParams{FlagID: 1})
	assert.NotNil(t, resp)

	okResp, ok := resp.(*datar.GetDatarFlagSummaryOK)
	if !ok {
		t.Fatalf("expected *datar.GetDatarFlagSummaryOK, got %T", resp)
	}

	assert.Equal(t, int64(1), okResp.Payload.FlagID)
	assert.Equal(t, map[string]int64{"1": 125, "2": 50}, okResp.Payload.TrafficByVariant, "variant totals")

	assert.Len(t, okResp.Payload.TrafficBySegment, 1)
	assert.Equal(t, int64(10), okResp.Payload.TrafficBySegment[0].SegmentID)

	assert.Len(t, okResp.Payload.TrafficByDay, 2, "should have 2 daily buckets")
}

func TestDatarEndpoints_NotEnabled(t *testing.T) {
	// Reset singleton for test isolation.
	singletonDatarOnce = sync.Once{}
	singletonDatar = nil
	t.Cleanup(func() { singletonDatarOnce = sync.Once{}; singletonDatar = nil })

	defer gostub.Stub(&config.Config.DatarEnabled, false).Reset()

	// When Datar is not enabled, GetDatar() returns nil and handlers return 503.
	resp := HandleGetDatarSummary(datar.GetDatarSummaryParams{})
	_, ok := resp.(*datar.GetDatarSummaryDefault)
	assert.True(t, ok, "expected 503 response when datar disabled")

	flagResp := HandleGetDatarFlagSummary(datar.GetDatarFlagSummaryParams{FlagID: 1})
	_, ok = flagResp.(*datar.GetDatarFlagSummaryDefault)
	assert.True(t, ok, "expected 503 response when datar disabled")
}

func TestDatarEndpoints_Pagination(t *testing.T) {
	// Reset singleton for test isolation.
	singletonDatarOnce = sync.Once{}
	singletonDatar = nil
	t.Cleanup(func() { singletonDatarOnce = sync.Once{}; singletonDatar = nil })

	defer gostub.Stub(&config.Config.DatarEnabled, true).Reset()
	defer gostub.Stub(&config.Config.DatarFlushInterval, 24*time.Hour).Reset()

	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
	db.AutoMigrate(entity.AutoMigrateTables...)

	d := GetDatar()
	assert.NotNil(t, d)

	// Create flags.
	for i := 1; i <= 5; i++ {
		assert.NoError(t, db.Exec(`INSERT INTO flags (id, key, enabled, description, created_at, updated_at) VALUES (?, ?, 1, 'flag', datetime('now'), datetime('now'))`, i, fmt.Sprintf("flag-%d", i)).Error)
	}

	// Pagination test: limit=2, offset=1 should return 2 flags (ids 2-3).
	limit := int32(2)
	offset := int32(1)

	resp := HandleGetDatarSummary(datar.GetDatarSummaryParams{Limit: &limit, Offset: &offset})
	assert.NotNil(t, resp)

	okResp, ok := resp.(*datar.GetDatarSummaryOK)
	if !ok {
		t.Fatalf("expected *datar.GetDatarSummaryOK, got %T", resp)
	}

	assert.Len(t, okResp.Payload.Flags, 2)
	assert.Equal(t, int64(2), okResp.Payload.Flags[0].FlagID)
	assert.Equal(t, int64(3), okResp.Payload.Flags[1].FlagID)
}
