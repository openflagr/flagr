package handler

import (
	"sync"
	"testing"

	"github.com/openflagr/flagr/pkg/config"
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

func TestGetDataRecorderWithRecorderDisabled(t *testing.T) {
	singletonDataRecorderOnce = sync.Once{}
	defer gostub.StubFunc(&NewKafkaRecorder, nil).Reset()

	// RecorderEnabled is false by default → kafka is listed but not created.
	assert.NotPanics(t, func() {
		GetDataRecorder()
	})
}
