package datar

import (
	"sync"
	"testing"
	"time"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/stretchr/testify/assert"
)

func TestAggregator_Record(t *testing.T) {
	a := NewAggregator()

	r := &models.EvalResult{FlagID: 1, VariantID: 2, SegmentID: 3}
	a.Record(r)
	a.Record(r)
	a.Record(r)

	assert.Equal(t, 1, a.Len(), "should have 1 key")

	snap := a.SnapshotAndReset()
	assert.Equal(t, 1, len(snap), "snapshot should have 1 key")
	assert.Equal(t, int32(3), snap[flushKeyFromResult(r)], "should count 3")

	assert.Equal(t, 0, a.Len(), "buffer should be empty after reset")
}

func TestAggregator_MultipleKeys(t *testing.T) {
	a := NewAggregator()

	keys := []*models.EvalResult{
		{FlagID: 1, VariantID: 1, SegmentID: 1},
		{FlagID: 1, VariantID: 2, SegmentID: 1},
		{FlagID: 2, VariantID: 1, SegmentID: 0},
	}
	for _, k := range keys {
		a.Record(k)
	}
	a.Record(keys[0])
	a.Record(keys[0])

	snap := a.SnapshotAndReset()
	assert.Equal(t, 3, len(snap), "should have 3 unique keys")
	assert.Equal(t, int32(3), snap[flushKeyFromResult(keys[0])])
	assert.Equal(t, int32(1), snap[flushKeyFromResult(keys[1])])
	assert.Equal(t, int32(1), snap[flushKeyFromResult(keys[2])])
}

func TestAggregator_Concurrent(t *testing.T) {
	a := NewAggregator()

	var wg sync.WaitGroup
	n := 100
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := &models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0}
			a.Record(r)
		}()
	}
	wg.Wait()

	snap := a.SnapshotAndReset()
	assert.Equal(t, int32(n), snap[flushKeyFromResult(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})])
}

func TestAggregator_Closed(t *testing.T) {
	a := NewAggregator()
	a.Close()

	a.Record(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})
	assert.Equal(t, 0, a.Len(), "should not record after close")
}

func TestAggregator_SnapshotAndReset(t *testing.T) {
	a := NewAggregator()

	// Add some counts
	a.Record(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})
	a.Record(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})

	snap1 := a.SnapshotAndReset()
	assert.Equal(t, int32(2), snap1[flushKeyFromResult(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})])

	// After reset, new records go to fresh buffer
	a.Record(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})
	snap2 := a.SnapshotAndReset()
	assert.Equal(t, int32(1), snap2[flushKeyFromResult(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})],
		"should only have new records after reset")
}

func flushKeyFromResult(r *models.EvalResult) FlushKey {
	return FlushKey{
		FlagID:    r.FlagID,
		VariantID: r.VariantID,
		SegmentID: r.SegmentID,
		Hour:      time.Now().Truncate(time.Hour),
	}
}

func TestAggregator_NilEvalResult(t *testing.T) {
	a := NewAggregator()
	assert.NotPanics(t, func() {
		a.Record(nil)
	}, "Record(nil) should not panic")
	assert.Equal(t, 0, a.Len(), "nil record should not create a key")
}

func TestAggregator_SnapshotAndResetOnEmpty(t *testing.T) {
	a := NewAggregator()
	snap := a.SnapshotAndReset()
	assert.NotNil(t, snap, "snapshot should be non-nil map")
	assert.Equal(t, 0, len(snap), "empty aggregator should return empty map")
}

func TestAggregator_DoubleClose(t *testing.T) {
	a := NewAggregator()
	a.Close()
	assert.NotPanics(t, func() {
		a.Close()
	}, "Close() twice should not panic")
	a.Record(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})
	assert.Equal(t, 0, a.Len(), "should not record after close")
}

func TestAggregator_ZeroKey(t *testing.T) {
	a := NewAggregator()
	a.Record(&models.EvalResult{FlagID: 0, VariantID: 0, SegmentID: 0})
	assert.Equal(t, 1, a.Len(), "zero-value key should still count")

	snap := a.SnapshotAndReset()
	k := FlushKey{FlagID: 0, VariantID: 0, SegmentID: 0, Hour: time.Now().Truncate(time.Hour)}
	assert.Equal(t, int32(1), snap[k], "zero-key should be present")
}

func TestAggregator_RecordAfterSnapshot(t *testing.T) {
	a := NewAggregator()

	a.Record(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})
	snap1 := a.SnapshotAndReset()

	// Record after snapshot into new buffer
	a.Record(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})
	a.Record(&models.EvalResult{FlagID: 2, VariantID: 1, SegmentID: 0})

	snap2 := a.SnapshotAndReset()
	assert.Equal(t, 2, len(snap2), "second snapshot should have new records")
	assert.Equal(t, int32(1), snap1[flushKeyFromResult(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})],
		"first snapshot should be frozen")
	assert.Equal(t, int32(1), snap2[flushKeyFromResult(&models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 0})],
		"second snapshot should have its own count for same key")
}

func BenchmarkRecord_ExistingKey(b *testing.B) {
	a := NewAggregator()
	r := &models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 1}
	a.Record(r) // warm up — create key

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Record(r)
	}
}

func BenchmarkRecord_NewKey(b *testing.B) {
	a := NewAggregator()
	keys := make([]*models.EvalResult, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = &models.EvalResult{FlagID: int64(i % 1000), VariantID: int64(i % 10), SegmentID: int64(i % 100)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Record(keys[i])
	}
}

func BenchmarkSnapshotAndReset(b *testing.B) {
	a := NewAggregator()
	// Seed with 1000 keys.
	for i := 0; i < 1000; i++ {
		a.Record(&models.EvalResult{FlagID: int64(i), VariantID: 1, SegmentID: 0})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.SnapshotAndReset()
	}
}

func BenchmarkConcurrentRecord(b *testing.B) {
	a := NewAggregator()
	r := &models.EvalResult{FlagID: 1, VariantID: 1, SegmentID: 1}
	a.Record(r)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.Record(r)
		}
	})
}
