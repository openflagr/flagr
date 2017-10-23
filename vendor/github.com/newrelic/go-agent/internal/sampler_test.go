package internal

import (
	"testing"
	"time"

	"github.com/newrelic/go-agent/internal/logger"
)

func TestGetSample(t *testing.T) {
	now := time.Now()
	sample := GetSample(now, logger.ShimLogger{})
	if nil == sample {
		t.Fatal(sample)
	}
	if now != sample.when {
		t.Error(now, sample.when)
	}
	if sample.numGoroutine <= 0 {
		t.Error(sample.numGoroutine)
	}
	if sample.numCPU <= 0 {
		t.Error(sample.numCPU)
	}
	if sample.memStats.HeapObjects == 0 {
		t.Error(sample.memStats.HeapObjects)
	}
}

func TestMetricsCreated(t *testing.T) {
	now := time.Now()
	h := NewHarvest(now)

	stats := Stats{
		heapObjects:  5 * 1000,
		numGoroutine: 23,
		allocBytes:   37 * 1024 * 1024,
		user: cpuStats{
			used:     20 * time.Millisecond,
			fraction: 0.01,
		},
		system: cpuStats{
			used:     40 * time.Millisecond,
			fraction: 0.02,
		},
		gcPauseFraction: 3e-05,
		deltaNumGC:      2,
		deltaPauseTotal: 500 * time.Microsecond,
		minPause:        100 * time.Microsecond,
		maxPause:        400 * time.Microsecond,
	}

	stats.MergeIntoHarvest(h)

	ExpectMetrics(t, h.Metrics, []WantMetric{
		{"Memory/Heap/AllocatedObjects", "", true, []float64{1, 5000, 5000, 5000, 5000, 25000000}},
		{"Memory/Physical", "", true, []float64{1, 37, 0, 37, 37, 1369}},
		{"CPU/User Time", "", true, []float64{1, 0.02, 0.02, 0.02, 0.02, 0.0004}},
		{"CPU/System Time", "", true, []float64{1, 0.04, 0.04, 0.04, 0.04, 0.0016}},
		{"CPU/User/Utilization", "", true, []float64{1, 0.01, 0, 0.01, 0.01, 0.0001}},
		{"CPU/System/Utilization", "", true, []float64{1, 0.02, 0, 0.02, 0.02, 0.0004}},
		{"Go/Runtime/Goroutines", "", true, []float64{1, 23, 23, 23, 23, 529}},
		{"GC/System/Pause Fraction", "", true, []float64{1, 3e-05, 0, 3e-05, 3e-05, 9e-10}},
		{"GC/System/Pauses", "", true, []float64{2, 0.0005, 0, 0.0001, 0.0004, 2.5e-7}},
	})
}

func TestMetricsCreatedEmpty(t *testing.T) {
	now := time.Now()
	h := NewHarvest(now)
	stats := Stats{}

	stats.MergeIntoHarvest(h)

	ExpectMetrics(t, h.Metrics, []WantMetric{
		{"Memory/Heap/AllocatedObjects", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"Memory/Physical", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"CPU/User Time", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"CPU/System Time", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"CPU/User/Utilization", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"CPU/System/Utilization", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"Go/Runtime/Goroutines", "", true, []float64{1, 0, 0, 0, 0, 0}},
		{"GC/System/Pause Fraction", "", true, []float64{1, 0, 0, 0, 0, 0}},
	})
}
