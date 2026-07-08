package handler

import (
	"net/http"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/prashantv/gostub"
)

// BenchmarkEvalSegmentBaseline benchmarks evalSegment without injected context (current behavior).
func BenchmarkEvalSegmentBaseline(b *testing.B) {
	config.Config.InjectedContextEnabled = false
	defer gostub.StubFunc(&logEvalResult).Reset()

	s := entity.GenFixtureSegment()
	s.RolloutPercent = 100
	s.PrepareEvaluation()

	ctx := map[string]any{"dl_state": "CA"}
	ec := models.EvalContext{
		EnableDebug:   false,
		EntityContext: ctx,
		EntityID:      "bench-entity",
		EntityType:    "bench-type",
		FlagID:        100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalSegment(ec, s)
	}
}

// BenchmarkEvalSegmentWithInjection benchmarks evalSegment with injected context enabled (no headers).
func BenchmarkEvalSegmentWithInjection(b *testing.B) {
	config.Config.InjectedContextEnabled = true
	defer func() {
		config.Config.InjectedContextEnabled = false
	}()
	defer gostub.StubFunc(&logEvalResult).Reset()

	s := entity.GenFixtureSegment()
	s.RolloutPercent = 100
	s.PrepareEvaluation()

	injectedCtx := InjectBuiltInContext(map[string]any{"dl_state": "CA"}, nil)
	ec := models.EvalContext{
		EnableDebug:   false,
		EntityContext: injectedCtx,
		EntityID:      "bench-entity",
		EntityType:    "bench-type",
		FlagID:        100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalSegment(ec, s)
	}
}

// BenchmarkEvalSegmentWithHTTPHeaders benchmarks evalSegment with injected context + HTTP headers.
func BenchmarkEvalSegmentWithHTTPHeaders(b *testing.B) {
	config.Config.InjectedContextEnabled = true
	config.Config.InjectedContextHTTPHeaders = []string{"X-Environment", "X-Tenant-ID"}
	config.Config.InjectedContextHTTPHeaderPrefixes = []string{"CF-"}
	defer func() {
		config.Config.InjectedContextEnabled = false
		config.Config.InjectedContextHTTPHeaders = nil
		config.Config.InjectedContextHTTPHeaderPrefixes = nil
	}()
	defer gostub.StubFunc(&logEvalResult).Reset()

	s := entity.GenFixtureSegment()
	s.RolloutPercent = 100
	s.PrepareEvaluation()

	r := &http.Request{
		Header: http.Header{
			"X-Environment": []string{"production"},
			"X-Tenant-ID":   []string{"acme"},
			"CF-IPCountry":  []string{"US"},
			"CF-Ray":        []string{"abc-123"},
			"User-Agent":    []string{"Mozilla/5.0"},
			"Accept":        []string{"application/json"},
		},
		Host: "flagr.prod.example.com",
	}

	injectedCtx := InjectBuiltInContext(map[string]any{"dl_state": "CA"}, r)
	ec := models.EvalContext{
		EnableDebug:   false,
		EntityContext: injectedCtx,
		EntityID:      "bench-entity",
		EntityType:    "bench-type",
		FlagID:        100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalSegment(ec, s)
	}
}

// BenchmarkInjectBuiltInContext benchmarks the injection function itself.
func BenchmarkInjectBuiltInContext(b *testing.B) {
	config.Config.InjectedContextEnabled = true
	config.Config.InjectedContextHTTPHeaders = []string{"X-Environment", "X-Tenant-ID"}
	config.Config.InjectedContextHTTPHeaderPrefixes = []string{"CF-"}
	defer func() {
		config.Config.InjectedContextEnabled = false
		config.Config.InjectedContextHTTPHeaders = nil
		config.Config.InjectedContextHTTPHeaderPrefixes = nil
	}()

	r := &http.Request{
		Header: http.Header{
			"X-Environment": []string{"production"},
			"X-Tenant-ID":   []string{"acme"},
			"CF-IPCountry":  []string{"US"},
			"CF-Ray":        []string{"abc-123"},
		},
		Host: "flagr.prod.example.com",
	}

	ctx := map[string]any{"dl_state": "CA"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InjectBuiltInContext(ctx, r)
	}
}

// BenchmarkInjectBuiltInContextCoreOnly benchmarks injection with only core keys (no headers).
func BenchmarkInjectBuiltInContextCoreOnly(b *testing.B) {
	config.Config.InjectedContextEnabled = true
	defer func() { config.Config.InjectedContextEnabled = false }()

	ctx := map[string]any{"dl_state": "CA"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InjectBuiltInContext(ctx, nil)
	}
}

// BenchmarkTsInjectionCost isolates the cost of time.Now().UTC() calls.
func BenchmarkTsInjectionCost(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		now := time.Now().UTC()
		_ = now.Unix()
		_ = now.Hour()
		_ = int(now.Weekday())
		_ = int(now.Month())
	}
}
