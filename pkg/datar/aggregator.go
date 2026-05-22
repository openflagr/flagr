package datar

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/openflagr/flagr/swagger_gen/models"
)

// Aggregator holds the in-memory counter map.
// Thread-safe via RWMutex + double-checked locking on the hot path.
type Aggregator struct {
	mu     sync.RWMutex
	buffer map[FlushKey]*int32
	closed atomic.Bool
}

// NewAggregator creates an Aggregator with an empty buffer.
func NewAggregator() *Aggregator {
	return &Aggregator{
		buffer: make(map[FlushKey]*int32),
	}
}

// Record increments the counter for the given EvalResult on the hot path.
// Fast path (~30ns): RLock + atomic increment on existing key.
// Slow path (~100ns): escalate to WLock, create key if missing.
func (a *Aggregator) Record(r *models.EvalResult) {
	if a.closed.Load() {
		return
	}
	if r == nil {
		return
	}

	key := FlushKey{
		FlagID:    r.FlagID,
		VariantID: r.VariantID,
		SegmentID: r.SegmentID,
		Hour:      time.Now().Truncate(time.Hour),
	}

	// Fast path: existing key under RLock.
	a.mu.RLock()
	ptr, ok := a.buffer[key]
	a.mu.RUnlock()
	if ok {
		atomic.AddInt32(ptr, 1)
		return
	}

	// Slow path: new key under WLock with double-check.
	a.mu.Lock()
	if ptr, ok = a.buffer[key]; ok {
		a.mu.Unlock()
		atomic.AddInt32(ptr, 1)
		return
	}
	var zero int32
	a.buffer[key] = &zero
	a.mu.Unlock()
	atomic.AddInt32(&zero, 1)
}

// SnapshotAndReset swaps the buffer under lock and returns frozen counts.
// After this call, the returned map is safe to read without holding the lock,
// and Record() writes go into a fresh buffer.
func (a *Aggregator) SnapshotAndReset() map[FlushKey]int32 {
	a.mu.Lock()
	defer a.mu.Unlock()

	old := a.buffer
	a.buffer = make(map[FlushKey]*int32, len(old))

	result := make(map[FlushKey]int32, len(old))
	for k, ptr := range old {
		result[k] = atomic.LoadInt32(ptr)
	}
	return result
}

// Len returns the number of distinct flush keys in the buffer.
func (a *Aggregator) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.buffer)
}

// Close prevents any further Record calls.
func (a *Aggregator) Close() {
	a.closed.Store(true)
}
