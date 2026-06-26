package handler

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/datar"
	"github.com/openflagr/flagr/pkg/entity"
	datarapi "github.com/openflagr/flagr/swagger_gen/restapi/operations/datar"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestParseTimeRange(t *testing.T) {
	from := strfmt.DateTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	to := strfmt.DateTime(time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC))

	// Both nil → uses defaults.
	s, e := parseTimeRange(nil, nil)
	assert.WithinDuration(t, time.Now().UTC(), e, time.Second)
	assert.WithinDuration(t, time.Now().UTC().Add(-7*24*time.Hour), s, time.Second)

	// With explicit from/to.
	s, e = parseTimeRange(&from, &to)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), s)
	assert.Equal(t, time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), e)
}

func TestDatarEndpoints_QueryError(t *testing.T) {
	defer ResetDatar()

	defer gostub.Stub(&config.Config.RecorderType, []string{"datar"}).Reset()
	defer gostub.Stub(&config.Config.RecorderEnabled, true).Reset()
	defer gostub.Stub(&config.Config.RecorderDatarFlushInterval, 24*time.Hour).Reset()

	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
	db.AutoMigrate(entity.AutoMigrateTables...)

	engine := datar.New(db, true, 24*time.Hour)
	defer engine.Shutdown()

	if err := db.Exec("DROP TABLE datar_hourly_events").Error; err != nil {
		t.Fatal(err)
	}

	resp := respondDatarSummary(engine, datarapi.GetDatarSummaryParams{})
	_, ok := resp.(*datarapi.GetDatarSummaryDefault)
	assert.True(t, ok, "expected 500 when query fails, got %T", resp)

	flagResp := respondDatarFlagSummary(engine, datarapi.GetDatarFlagSummaryParams{FlagID: 1})
	_, ok = flagResp.(*datarapi.GetDatarFlagSummaryDefault)
	assert.True(t, ok, "expected 500 when query fails, got %T", flagResp)
}

func TestDatarEndpoints_Summary(t *testing.T) {
	defer ResetDatar()

	defer gostub.Stub(&config.Config.RecorderType, []string{"datar"}).Reset()
	defer gostub.Stub(&config.Config.RecorderEnabled, true).Reset()
	defer gostub.Stub(&config.Config.RecorderDatarFlushInterval, 24*time.Hour).Reset()

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
	resp := HandleGetDatarSummary(datarapi.GetDatarSummaryParams{})
	assert.NotNil(t, resp)

	okResp, ok := resp.(*datarapi.GetDatarSummaryOK)
	if !ok {
		t.Fatalf("expected *datarapi.GetDatarSummaryOK, got %T", resp)
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

	defer gostub.Stub(&config.Config.RecorderType, []string{"datar"}).Reset()
	defer gostub.Stub(&config.Config.RecorderEnabled, true).Reset()
	defer gostub.Stub(&config.Config.RecorderDatarFlushInterval, 24*time.Hour).Reset()

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

	resp := HandleGetDatarFlagSummary(datarapi.GetDatarFlagSummaryParams{FlagID: 1})
	assert.NotNil(t, resp)

	okResp, ok := resp.(*datarapi.GetDatarFlagSummaryOK)
	if !ok {
		t.Fatalf("expected *datarapi.GetDatarFlagSummaryOK, got %T", resp)
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

	defer gostub.Stub(&config.Config.RecorderType, []string{"kafka"}).Reset()

	// When Datar is not enabled, GetDatar() returns nil and handlers return 503.
	resp := HandleGetDatarSummary(datarapi.GetDatarSummaryParams{})
	_, ok := resp.(*datarapi.GetDatarSummaryDefault)
	assert.True(t, ok, "expected 503 response when datar disabled")

	flagResp := HandleGetDatarFlagSummary(datarapi.GetDatarFlagSummaryParams{FlagID: 1})
	_, ok = flagResp.(*datarapi.GetDatarFlagSummaryDefault)
	assert.True(t, ok, "expected 503 response when datar disabled")
}

func TestDatarEndpoints_Pagination(t *testing.T) {
	defer ResetDatar()

	defer gostub.Stub(&config.Config.RecorderType, []string{"datar"}).Reset()
	defer gostub.Stub(&config.Config.RecorderEnabled, true).Reset()
	defer gostub.Stub(&config.Config.RecorderDatarFlushInterval, 24*time.Hour).Reset()

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

	resp := HandleGetDatarSummary(datarapi.GetDatarSummaryParams{Limit: &limit, Offset: &offset})
	assert.NotNil(t, resp)

	okResp, ok := resp.(*datarapi.GetDatarSummaryOK)
	if !ok {
		t.Fatalf("expected *datarapi.GetDatarSummaryOK, got %T", resp)
	}

	if assert.Len(t, okResp.Payload.Flags, 2) {
		flags := okResp.Payload.Flags
		assert.Equal(t, int64(2), flags[0].FlagID, "flag 2 has 98 count")
		assert.Equal(t, int64(98), flags[0].TotalEvalCount)
		assert.Equal(t, int64(3), flags[1].FlagID, "flag 3 has 97 count")
		assert.Equal(t, int64(97), flags[1].TotalEvalCount)
	}
}

// ---------------------------------------------------------------------------
// slices.Contains (replaces hasDatar)
// ---------------------------------------------------------------------------

func TestSlicesContainsDatar(t *testing.T) {
	tests := []struct {
		name  string
		types []string
		want  bool
	}{
		{"nil slice", nil, false},
		{"empty slice", []string{}, false},
		{"single datar", []string{"datar"}, true},
		{"single kafka", []string{"kafka"}, false},
		{"multiple with datar", []string{"kafka", "datar"}, true},
		{"multiple without datar", []string{"kafka", "pubsub"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Contains(tt.types, "datar")
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Converter edge cases
// ---------------------------------------------------------------------------

func TestToSwaggerSummaryFlag_WithValidLastEvaluated(t *testing.T) {
	r := datar.SummaryRow{
		FlagID:         42,
		FlagKey:        "my-flag",
		Enabled:        true,
		TotalEvalCount: 1000,
		LastEvaluated:  "2024-06-15T10:30:00Z",
	}
	f := toSwaggerSummaryFlag(r)
	assert.Equal(t, int64(42), f.FlagID)
	assert.Equal(t, "my-flag", f.FlagKey)
	assert.Equal(t, int64(1000), f.TotalEvalCount)
	assert.False(t, f.LastEvaluatedAt.IsZero(), "should parse RFC3339 date")
}

func TestToSwaggerSummaryFlag_EmptyLastEvaluated(t *testing.T) {
	r := datar.SummaryRow{FlagID: 1, LastEvaluated: ""}
	f := toSwaggerSummaryFlag(r)
	assert.True(t, f.LastEvaluatedAt.IsZero(), "empty string should leave zero value")
}

func TestToSwaggerDay_InvalidDate(t *testing.T) {
	d := datar.DayEntry{Day: "not-a-date", Count: 10}
	entry := toSwaggerDay(d)
	assert.Nil(t, entry, "unparseable date should return nil")
}

func TestToSwaggerDay_ValidDate(t *testing.T) {
	d := datar.DayEntry{Day: "2024-01-15", Count: 42}
	entry := toSwaggerDay(d)
	assert.NotNil(t, entry)
	assert.Equal(t, int64(42), entry.Count)
}
