//go:build integration

package flagr_integration

import (
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkEvalByFlagID(b *testing.B) {
	if len(seedFlagIDs) == 0 {
		b.Skip("no seeded flags")
	}
	body := map[string]any{
		"flagID":     seedFlagIDs[0],
		"entityID":   "bench-user",
		"entityType": "user",
		"entityContext": map[string]any{
			"region": "us-west",
			"age":    30,
		},
	}
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalByFlagKey(b *testing.B) {
	if len(seedFlagKeys) == 0 {
		b.Skip("no seeded flags")
	}
	body := map[string]any{
		"flagKey":    seedFlagKeys[0],
		"entityID":   "bench-user",
		"entityType": "user",
		"entityContext": map[string]any{
			"region": "us-west",
			"age":    30,
		},
	}
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalNoMatch(b *testing.B) {
	if len(seedFlagIDs) == 0 {
		b.Skip("no seeded flags")
	}
	body := map[string]any{
		"flagID":     seedFlagIDs[0],
		"entityID":   "bench-user",
		"entityType": "user",
		"entityContext": map[string]any{
			"region": "nonexistent",
		},
	}
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalDisabledFlag(b *testing.B) {
	body := map[string]any{
		"flagKey":    "int_flag_disabled",
		"entityID":   "bench-user",
		"entityType": "user",
	}
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalBatchByIDs(b *testing.B) {
	if len(seedFlagIDs) < 5 {
		b.Skip("need at least 5 seeded flags")
	}
	body := map[string]any{
		"entities": []map[string]any{
			{"entityID": "bench", "entityType": "user", "entityContext": map[string]any{"region": "us-west", "age": 30}},
		},
		"flagIDs": seedFlagIDs[:5],
	}
	benchEval(b, "/api/v1/evaluation/batch", body)
}

func BenchmarkEvalBatchByTags(b *testing.B) {
	body := map[string]any{
		"entities": []map[string]any{
			{"entityID": "bench", "entityType": "user", "entityContext": map[string]any{"region": "us-west"}},
		},
		"flagTags":         []string{"int_test"},
		"flagTagsOperator": "ANY",
	}
	benchEval(b, "/api/v1/evaluation/batch", body)
}

func BenchmarkEvalBatchLarge(b *testing.B) {
	if len(seedFlagIDs) < 10 {
		b.Skip("need at least 10 seeded flags")
	}
	entities := make([]map[string]any, 5)
	regions := []string{"us-west", "us-east", "eu-west", "ap-northeast", "us-west"}
	for i := range entities {
		entities[i] = map[string]any{
			"entityID":   fmt.Sprintf("bench%d", i),
			"entityType": "user",
			"entityContext": map[string]any{
				"region": regions[i],
			},
		}
	}
	body := map[string]any{
		"entities": entities,
		"flagIDs":  seedFlagIDs[:10],
	}
	benchEval(b, "/api/v1/evaluation/batch", body)
}

func BenchmarkEvalEQ(b *testing.B) {
	body := map[string]any{
		"flagKey":    "int_flag_EQ_01",
		"entityID":   "bench",
		"entityType": "user",
		"entityContext": map[string]any{
			"region": "us-west",
		},
	}
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalIN(b *testing.B) {
	body := map[string]any{
		"flagKey":    "int_flag_IN_01",
		"entityID":   "bench",
		"entityType": "user",
		"entityContext": map[string]any{
			"region": "us-west",
		},
	}
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalRegex(b *testing.B) {
	body := map[string]any{
		"flagKey":    "int_flag_EREG_01",
		"entityID":   "bench",
		"entityType": "user",
		"entityContext": map[string]any{
			"email": "user@company.com",
		},
	}
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalMultiConstraint(b *testing.B) {
	body := map[string]any{
		"flagKey":    "int_flag_complex_and",
		"entityID":   "bench",
		"entityType": "user",
		"entityContext": map[string]any{
			"region": "us-west",
			"age":    25,
			"tier":   "premium",
		},
	}
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalNestedContext(b *testing.B) {
	body := map[string]any{
		"flagKey":    "int_flag_complex_and",
		"entityID":   "bench",
		"entityType": "user",
		"entityContext": map[string]any{
			"user": map[string]any{
				"name": "Alice",
				"age":  30,
				"tier": "premium",
			},
			"region": "us-west",
		},
	}
	benchEval(b, "/api/v1/evaluation", body)
}

// benchEval benchmarks an endpoint using structured body (marshal-then-send).
func benchEval(b *testing.B, path string, body any) {
	b.Helper()
	b.ResetTimer()
	for b.Loop() {
		resp, err := doReq("POST", path, body)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			b.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	}
}
