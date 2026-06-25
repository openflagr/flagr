package handler

import (
	"sync"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestGetDataRecorder(t *testing.T) {
	singletonDataRecorderOnce = sync.Once{}
	defer gostub.StubFunc(&NewKafkaRecorder, nil).Reset()

	assert.NotPanics(t, func() {
		GetDataRecorder()
	})
}

func TestGetDataRecorderWhenKinesisIsSet(t *testing.T) {
	singletonDataRecorderOnce = sync.Once{}
	defer gostub.StubFunc(&NewKinesisRecorder, nil).Reset()
	config.Config.RecorderType = []string{"kinesis"}

	assert.NotPanics(t, func() {
		GetDataRecorder()
	})

	config.Config.RecorderType = []string{"kafka"}
}

func TestGetDataRecorderWhenPubsubIsSet(t *testing.T) {
	singletonDataRecorderOnce = sync.Once{}
	defer gostub.StubFunc(&NewPubsubRecorder, nil).Reset()
	config.Config.RecorderType = []string{"pubsub"}

	assert.NotPanics(t, func() {
		GetDataRecorder()
	})

	config.Config.RecorderType = []string{"kafka"}
}

func TestGetDataRecorderPanicsWhenRecorderIsInvalid(t *testing.T) {
	singletonDataRecorderOnce = sync.Once{}
	config.Config.RecorderType = []string{"invalid"}
	config.Config.RecorderEnabled = true

	assert.Panics(t, func() {
		GetDataRecorder()
	})

	config.Config.RecorderType = []string{"kafka"}
}

// ---------------------------------------------------------------------------
// fanOutRecorder
// ---------------------------------------------------------------------------

type mockRecorder struct {
	calls []models.EvalResult
}

func (m *mockRecorder) AsyncRecord(r models.EvalResult) {
	m.calls = append(m.calls, r)
}

func (m *mockRecorder) NewDataRecordFrame(_ models.EvalResult) DataRecordFrame {
	return DataRecordFrame{}
}

func TestFanOutRecorder_Empty(t *testing.T) {
	var f fanOutRecorder
	// Empty fan-out: AsyncRecord is a no-op (should not panic).
	assert.NotPanics(t, func() {
		f.AsyncRecord(models.EvalResult{FlagID: 1})
	})
}

func TestFanOutRecorder_Single(t *testing.T) {
	m := &mockRecorder{}
	f := fanOutRecorder{m}
	f.AsyncRecord(models.EvalResult{FlagID: 1, VariantID: 10})

	if assert.Len(t, m.calls, 1) {
		assert.Equal(t, int64(1), m.calls[0].FlagID)
		assert.Equal(t, int64(10), m.calls[0].VariantID)
	}
}

func TestFanOutRecorder_Multiple(t *testing.T) {
	m1, m2 := &mockRecorder{}, &mockRecorder{}
	f := fanOutRecorder{m1, m2}
	f.AsyncRecord(models.EvalResult{FlagID: 42})

	assert.Len(t, m1.calls, 1)
	assert.Len(t, m2.calls, 1)
	assert.Equal(t, int64(42), m1.calls[0].FlagID)
	assert.Equal(t, int64(42), m2.calls[0].FlagID)
}

func TestFanOutRecorder_NewDataRecordFrame(t *testing.T) {
	var f fanOutRecorder
	frame := f.NewDataRecordFrame(models.EvalResult{})
	assert.Equal(t, DataRecordFrame{}, frame)
}

// ---------------------------------------------------------------------------
// GetDataRecorder edge cases
// ---------------------------------------------------------------------------

func TestGetDataRecorderWithEmptyType(t *testing.T) {
	singletonDataRecorderOnce = sync.Once{}
	defer gostub.Stub(&config.Config.RecorderType, []string{}).Reset()

	// Empty RecorderType + Datar not enabled → fanOutRecorder with 0 recorders (no-op).
	assert.NotPanics(t, func() {
		rec := GetDataRecorder()
		rec.AsyncRecord(models.EvalResult{FlagID: 1})
	})
}

func TestGetDataRecorderWithKafkaDefault(t *testing.T) {
	singletonDataRecorderOnce = sync.Once{}
	defer gostub.StubFunc(&NewKafkaRecorder, nil).Reset()

	// Default RecorderType is ["kafka"] → NewKafkaRecorder is called (returns nil).
	assert.NotPanics(t, func() {
		GetDataRecorder()
	})
}

// ---------------------------------------------------------------------------
// datarRecorder
// ---------------------------------------------------------------------------

func TestDatarRecorder_AsyncRecord(t *testing.T) {
	defer ResetDatar()

	defer gostub.Stub(&config.Config.RecorderType, []string{"datar"}).Reset()
	defer gostub.Stub(&config.Config.RecorderEnabled, true).Reset()
	defer gostub.Stub(&config.Config.RecorderDatarFlushInterval, 24*time.Hour).Reset()

	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
	db.AutoMigrate(entity.AutoMigrateTables...)

	r := NewDatarRecorder()
	assert.NotNil(t, r)

	// Record some events.
	r.AsyncRecord(models.EvalResult{FlagID: 1, VariantID: 10, SegmentID: 20})
	r.AsyncRecord(models.EvalResult{FlagID: 1, VariantID: 10, SegmentID: 20})
	r.AsyncRecord(models.EvalResult{FlagID: 2, VariantID: 1, SegmentID: 5})

	// Verify engine buffer has the right counts.
	d := GetDatar()
	assert.NotNil(t, d)
	assert.Equal(t, 2, d.Len(), "2 distinct keys")
}

func TestDatarRecorder_SkipsExposure(t *testing.T) {
	defer ResetDatar()
	defer gostub.Stub(&config.Config.RecorderType, []string{"datar"}).Reset()
	defer gostub.Stub(&config.Config.RecorderEnabled, true).Reset()

	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
	db.AutoMigrate(entity.AutoMigrateTables...)

	_ = GetDatar()
	r := NewDatarRecorder()
	r.AsyncRecord(models.EvalResult{
		FlagID:       1,
		VariantID:    10,
		SegmentID:    0,
		RecordSource: models.EvalResultRecordSourceExposure,
	})
	assert.Equal(t, 0, GetDatar().Len())

	r.AsyncRecord(models.EvalResult{FlagID: 1, VariantID: 10, SegmentID: 20})
	assert.Equal(t, 1, GetDatar().Len())
}

func TestDatarRecorder_NewDataRecordFrame(t *testing.T) {
	r := NewDatarRecorder()
	frame := r.NewDataRecordFrame(models.EvalResult{})
	assert.Equal(t, DataRecordFrame{}, frame, "datar recorder returns empty frame")
}
