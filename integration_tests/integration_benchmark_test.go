//go:build integration

package flagr_integration

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// HTTP benchmark helpers
// ---------------------------------------------------------------------------

func benchEval(b *testing.B, path, body string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+path, "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkEvalByFlagID(b *testing.B) {
	if len(seedFlagIDs) == 0 {
		b.Skip("no seeded flags")
	}
	body := fmt.Sprintf(`{"flagID":%d,"entityID":"bench-user","entityType":"user","entityContext":{"region":"us-west","age":30}}`, seedFlagIDs[0])
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalByFlagKey(b *testing.B) {
	if len(seedFlagKeys) == 0 {
		b.Skip("no seeded flags")
	}
	body := fmt.Sprintf(`{"flagKey":"%s","entityID":"bench-user","entityType":"user","entityContext":{"region":"us-west","age":30}}`, seedFlagKeys[0])
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalNoMatch(b *testing.B) {
	if len(seedFlagIDs) == 0 {
		b.Skip("no seeded flags")
	}
	body := fmt.Sprintf(`{"flagID":%d,"entityID":"bench-user","entityType":"user","entityContext":{"region":"nonexistent"}}`, seedFlagIDs[0])
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalDisabledFlag(b *testing.B) {
	body := `{"flagKey":"int_flag_disabled","entityID":"bench-user","entityType":"user"}`
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalBatchByIDs(b *testing.B) {
	if len(seedFlagIDs) < 5 {
		b.Skip("need at least 5 seeded flags")
	}
	body := fmt.Sprintf(
		`{"entities":[{"entityID":"bench","entityType":"user","entityContext":{"region":"us-west","age":30}}],"flagIDs":%s}`,
		jsonInts(seedFlagIDs[:5]),
	)
	benchEval(b, "/api/v1/evaluation/batch", body)
}

func BenchmarkEvalBatchByTags(b *testing.B) {
	body := `{"entities":[{"entityID":"bench","entityType":"user","entityContext":{"region":"us-west"}}],"flagTags":["int_test"],"flagTagsOperator":"ANY"}`
	benchEval(b, "/api/v1/evaluation/batch", body)
}

func BenchmarkEvalBatchLarge(b *testing.B) {
	if len(seedFlagIDs) < 10 {
		b.Skip("need at least 10 seeded flags")
	}
	body := fmt.Sprintf(
		`{"entities":[{"entityID":"bench0","entityType":"user","entityContext":{"region":"us-west"}},{"entityID":"bench1","entityType":"user","entityContext":{"region":"us-east"}},{"entityID":"bench2","entityType":"user","entityContext":{"region":"eu-west"}},{"entityID":"bench3","entityType":"user","entityContext":{"region":"ap-northeast"}},{"entityID":"bench4","entityType":"user","entityContext":{"region":"us-west"}}],"flagIDs":%s}`,
		jsonInts(seedFlagIDs[:10]),
	)
	benchEval(b, "/api/v1/evaluation/batch", body)
}

func BenchmarkEvalEQ(b *testing.B) {
	body := `{"flagKey":"int_flag_EQ_01","entityID":"bench","entityType":"user","entityContext":{"region":"us-west"}}`
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalIN(b *testing.B) {
	body := `{"flagKey":"int_flag_IN_01","entityID":"bench","entityType":"user","entityContext":{"region":"us-west"}}`
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalRegex(b *testing.B) {
	body := `{"flagKey":"int_flag_EREG_01","entityID":"bench","entityType":"user","entityContext":{"email":"user@company.com"}}`
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalMultiConstraint(b *testing.B) {
	body := `{"flagKey":"int_flag_complex_and","entityID":"bench","entityType":"user","entityContext":{"region":"us-west","age":25,"tier":"premium"}}`
	benchEval(b, "/api/v1/evaluation", body)
}

func BenchmarkEvalNestedContext(b *testing.B) {
	body := `{"flagKey":"int_flag_complex_and","entityID":"bench","entityType":"user","entityContext":{"user":{"name":"Alice","age":30,"tier":"premium"},"region":"us-west"}}`
	benchEval(b, "/api/v1/evaluation", body)
}
