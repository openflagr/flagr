package handler

import (
	"fmt"
	"testing"
	"time"
	"github.com/go-openapi/strfmt"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/datar"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestParseTimeRange(t *testing.T) {
	from := strfmt.DateTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	to := strfmt.DateTime(time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC))

	// Both nil → uses defaults.
	s, e := parseTimeRange(nil, nil, 7)
	assert.WithinDuration(t, time.Now().UTC(), e, time.Second)
	assert.WithinDuration(t, time.Now().UTC().Add(-7*24*time.Hour), s, time.Second)

	// With explicit from/to.
	s, e = parseTimeRange(&from, &to, 7)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), s)
	assert.Equal(t, time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), e)
}

func TestDatarEndpoints_QueryError(t *testing.T) {
	defer ResetDatar()

	defer gostub.Stub(&config.Config.DatarEnabled, true).Reset()
	defer gostub.Stub(&config.Config.DatarFlushInterval, 24*time.Hour).Reset()

	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
	db.AutoMigrate(entity.AutoMigrateTables...)

	// Initialize engine.
	_ = GetDatar()

	// Drop the hourly events table so queries fail.
	if err := db.Exec("DROP TABLE datar_hourly_events").Error; err != nil {
		t.Fatal(err)
	}

	// Summary endpoint should return 500.
	resp := HandleGetDatarSummary(datar.GetDatarSummaryParams{})
	_, ok := resp.(*datar.GetDatarSummaryDefault)
	assert.True(t, ok, "expected 500 when query fails")

	// Flag summary endpoint should return 500.
	flagResp := HandleGetDatarFlagSummary(datar.GetDatarFlagSummaryParams{FlagID: 1})
	_, ok = flagResp.(*datar.GetDatarFlagSummaryDefault)
	assert.True(t, ok, "expected 500 when query fails")
}

func TestDatarEndpoints_Summary(t *testing.T) {
	defer ResetDatar()

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

	// Flag 2 has no hourly data — only flags with traffic appear.
	resp := HandleGetDatarSummary(datar.GetDatarSummaryParams{})
	assert.NotNil(t, resp)

	okResp, ok := resp.(*datar.GetDatarSummaryOK)
	if !ok {
		t.Fatalf("expected *datar.GetDatarSummaryOK, got %T", resp)
	}

	assert.Len(t, okResp.Payload.Flags, 1, "only flag 1 has traffic data")
	f1 := okResp.Payload.Flags[0]
	assert.Equal(t, int64(1), f1.FlagID)
	assert.Equal(t, "test-flag", f1.FlagKey)
	assert.Equal(t, int64(150), f1.TotalEvalCount)
	assert.True(t, f1.Enabled)
}

func TestDatarEndpoints_FlagSummary(t *testing.T) {
	defer ResetDatar()

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
	if assert.Len(t, okResp.Payload.TrafficByVariant, 2, "variant totals") {
		assert.Equal(t, int64(1), okResp.Payload.TrafficByVariant[0].VariantID)
		assert.Equal(t, int64(125), okResp.Payload.TrafficByVariant[0].Count, "variant 1 count")
		assert.Equal(t, int64(2), okResp.Payload.TrafficByVariant[1].VariantID)
		assert.Equal(t, int64(50), okResp.Payload.TrafficByVariant[1].Count, "variant 2 count")
	}

	assert.Len(t, okResp.Payload.TrafficBySegment, 1)
	assert.Equal(t, int64(10), okResp.Payload.TrafficBySegment[0].SegmentID)

	assert.Len(t, okResp.Payload.TrafficByDay, 2, "should have 2 daily buckets")
}

func TestDatarEndpoints_NotEnabled(t *testing.T) {
	defer ResetDatar()

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
	defer ResetDatar()

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

	// Seed events for flags 1-4 with descending counts.
	now := time.Now().UTC().Truncate(time.Hour)
	for i := 1; i <= 4; i++ {
		assert.NoError(t, db.Exec(
			`INSERT INTO datar_hourly_events (flag_id, variant_id, segment_id, bucket_hour, eval_count) VALUES (?, 1, 1, ?, ?)`,
			i, now, 100-i,
		).Error)
	}

	// Page 2 (limit=2, offset=1): positions 1-2 by count desc.
	// Flag 1(99), Flag 2(98), Flag 3(97), Flag 4(96) → [Flag 2(98), Flag 3(97)].
	limit := int32(2)
	offset := int32(1)

	resp := HandleGetDatarSummary(datar.GetDatarSummaryParams{Limit: &limit, Offset: &offset})
	assert.NotNil(t, resp)

	okResp, ok := resp.(*datar.GetDatarSummaryOK)
	if !ok {
		t.Fatalf("expected *datar.GetDatarSummaryOK, got %T", resp)
	}

	if assert.Len(t, okResp.Payload.Flags, 2) {
		flags := okResp.Payload.Flags
		assert.Equal(t, int64(2), flags[0].FlagID, "flag 2 has 98 count")
		assert.Equal(t, int64(98), flags[0].TotalEvalCount)
		assert.Equal(t, int64(3), flags[1].FlagID, "flag 3 has 97 count")
		assert.Equal(t, int64(97), flags[1].TotalEvalCount)
	}
}
