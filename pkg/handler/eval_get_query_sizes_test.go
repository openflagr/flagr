package handler

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/stretchr/testify/require"
)

// getEvalRawQueryLen returns len("json="+QueryEscape(JSON)) for a GET /evaluation request.
func getEvalRawQueryLen(t *testing.T, body any) int {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	return len("json=" + url.QueryEscape(string(raw)))
}

// getEvalBatchRawQueryLen same for batch body.
func getEvalBatchRawQueryLen(t *testing.T, body any) int {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	return len("json=" + url.QueryEscape(string(raw)))
}

// TestGetEvalQuerySizesDocumentsTypicalPayloads records raw query lengths for shapes used in
// integration_tests and pkg/handler fixtures. Default FLAGR_EVAL_GET_MAX_URL_BYTES=8192.
// Integration limit probe uses ~8033-char entityContext.blob at the cap (see TestIntegration_GetEvaluation_QueryURLBytesLimit).
func TestGetEvalQuerySizesDocumentsTypicalPayloads(t *testing.T) {
	const maxDefault = 8192

	cases := []struct {
		name string
		body any
	}{
		{
			name: "integration TestIntegration_Evaluation / datarEvalBody",
			body: map[string]any{
				"flagID": int64(2), "entityID": "eval-entity-1", "entityType": "user",
				"entityContext": map[string]any{"tier": "premium"},
			},
		},
		{
			name: "integration get-eval by flagID",
			body: map[string]any{
				"entityID": "get-eval-entity", "entityType": "user",
				"entityContext": map[string]any{"tier": "premium"},
				"flagID":        int64(2),
			},
		},
		{
			name: "integration batch one entity two fields",
			body: map[string]any{
				"entities": []map[string]any{{
					"entityID": "batch-1", "entityType": "user",
					"entityContext": map[string]any{"region": "us-west", "age": 30},
				}},
				"flagIDs": []int64{1, 2, 3, 4, 5},
			},
		},
		{
			name: "handler eval_get_test EvalContext",
			body: models.EvalContext{
				EntityID: "e1", FlagID: 100,
				EntityContext: map[string]any{"dl_state": "CA"},
			},
		},
		{
			name: "benchmark nested user",
			body: map[string]any{
				"entityID": "bench", "entityType": "user", "flagID": int64(1),
				"entityContext": map[string]any{
					"user": map[string]any{"name": "Alice", "age": 30},
				},
			},
		},
		{
			name: "benchmark multi-field context",
			body: map[string]any{
				"entityID": "bench", "entityType": "user", "flagID": int64(1),
				"entityContext": map[string]any{
					"region": "us-west", "age": 25, "tier": "premium",
				},
			},
		},
		{
			name: "enableDebug true",
			body: map[string]any{
				"entityID": "dbg", "entityType": "user", "flagID": int64(1),
				"enableDebug":   true,
				"entityContext": map[string]any{"state": "CA"},
			},
		},
	}

	t.Logf("FLAGR_EVAL_GET_MAX_URL_BYTES default=%d (raw query string only, not full URL)", maxDefault)
	t.Logf("At cap (integration probe): rawQueryLen=8192 with entityContext.blob len=8033 ASCII")

	for _, tc := range cases {
		n := getEvalRawQueryLen(t, tc.body)
		t.Logf("%s: rawQueryLen=%d (~%.1f%% of default cap)", tc.name, n, 100*float64(n)/float64(maxDefault))
		require.Less(t, n, maxDefault, "fixture should fit default GET cap; use POST if you exceed this")
	}

	// Batch: tag eval with several entities (still small vs cap)
	batchN := getEvalBatchRawQueryLen(t, map[string]any{
		"entities": []map[string]any{
			{"entityID": "u1", "entityType": "user", "entityContext": map[string]any{"region": "us-west"}},
			{"entityID": "u2", "entityType": "user", "entityContext": map[string]any{"region": "eu"}},
			{"entityID": "u3", "entityType": "user", "entityContext": map[string]any{"region": "ap"}},
		},
		"flagTags": []string{"web", "mobile"}, "flagTagsOperator": "ANY",
	})
	t.Logf("integration-style batch (3 entities, flagTags): rawQueryLen=%d", batchN)
	require.Less(t, batchN, maxDefault)

	// Rough headroom: how much ASCII blob fits after a minimal evalContext skeleton
	skeleton := map[string]any{
		"entityID": "x", "entityType": "user", "flagID": int64(1),
		"entityContext": map[string]any{"blob": ""},
	}
	base := getEvalRawQueryLen(t, skeleton)
	headroom := maxDefault - base
	t.Logf("minimal skeleton rawQueryLen=%d; ~%d bytes of query budget left for entityContext.blob (ASCII, before encoding overhead)", base, headroom)
}
