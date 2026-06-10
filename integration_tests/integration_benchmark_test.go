//go:build integration

package flagr_integration

import (
	"encoding/json"
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
	flagIDsJSON, _ := json.Marshal(seedFlagIDs[:5])
	body := fmt.Sprintf(
		`{"entities":[{"entityID":"bench","entityType":"user","entityContext":{"region":"us-west","age":30}}],"flagIDs":%s}`,
		flagIDsJSON,
	)
	benchEvalRaw(b, "/api/v1/evaluation/batch", body)
}

func BenchmarkEvalBatchByTags(b *testing.B) {
	body := `{"entities":[{"entityID":"bench","entityType":"user","entityContext":{"region":"us-west"}}],"flagTags":["int_test"],"flagTagsOperator":"ANY"}`
	benchEvalRaw(b, "/api/v1/evaluation/batch", body)
}

func BenchmarkEvalBatchLarge(b *testing.B) {
	if len(seedFlagIDs) < 10 {
		b.Skip("need at least 10 seeded flags")
	}
	flagIDsJSON, _ := json.Marshal(seedFlagIDs[:10])
	body := fmt.Sprintf(
		`{"entities":[{"entityID":"bench0","entityType":"user","entityContext":{"region":"us-west"}},{"entityID":"bench1","entityType":"user","entityContext":{"region":"us-east"}},{"entityID":"bench2","entityType":"user","entityContext":{"region":"eu-west"}},{"entityID":"bench3","entityType":"user","entityContext":{"region":"ap-northeast"}},{"entityID":"bench4","entityType":"user","entityContext":{"region":"us-west"}}],"flagIDs":%s}`,
		flagIDsJSON,
	)
	benchEvalRaw(b, "/api/v1/evaluation/batch", body)
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
	for i := 0; i < b.N; i++ {
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

// benchEvalRaw benchmarks an endpoint using a pre-serialized JSON string body.
func benchEvalRaw(b *testing.B, path, body string) {
	b.Helper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
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
