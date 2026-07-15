package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/evaluation"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEvaluation(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()

	e := NewEval()
	ctx := models.EvalContext{
		EntityID:      "e1",
		EntityContext: map[string]any{"dl_state": "CA"},
		FlagID:        100,
	}
	raw, err := json.Marshal(ctx)
	require.NoError(t, err)

	req := &http.Request{URL: &url.URL{RawQuery: "json=" + url.QueryEscape(string(raw))}}
	resp := e.GetEvaluation(evaluation.GetEvaluationParams{HTTPRequest: req, JSON: string(raw)})
	ok, isOK := resp.(*evaluation.GetEvaluationOK)
	require.True(t, isOK, "got %T", resp)
	require.NotNil(t, ok.Payload)
	assert.NotZero(t, ok.Payload.VariantID)
}

func TestGetEvaluation_MissingJSON(t *testing.T) {
	e := NewEval()
	req := &http.Request{URL: &url.URL{}}
	resp := e.GetEvaluation(evaluation.GetEvaluationParams{HTTPRequest: req})
	_, isDef := resp.(*evaluation.GetEvaluationDefault)
	assert.True(t, isDef)
}

func TestGetEvaluation_InvalidJSON(t *testing.T) {
	e := NewEval()
	req := &http.Request{URL: &url.URL{RawQuery: "json=not-json"}}
	resp := e.GetEvaluation(evaluation.GetEvaluationParams{HTTPRequest: req, JSON: "not-json"})
	_, isDef := resp.(*evaluation.GetEvaluationDefault)
	assert.True(t, isDef)
}

func TestGetEvaluation_QueryTooLong(t *testing.T) {
	stubs := gostub.Stub(&config.Config.EvalGetMaxURLBytes, 10)
	defer stubs.Reset()
	e := NewEval()
	raw := `{"flagID":1,"entityID":"e1"}`
	req := &http.Request{URL: &url.URL{RawQuery: "json=" + url.QueryEscape(raw)}}
	resp := e.GetEvaluation(evaluation.GetEvaluationParams{HTTPRequest: req, JSON: raw})
	def, isDef := resp.(*evaluation.GetEvaluationDefault)
	require.True(t, isDef)
	require.NotNil(t, def.Payload)
	assert.Contains(t, *def.Payload.Message, "exceeds maximum")
}

func TestGetEvaluationBatch(t *testing.T) {
	defer gostub.StubFunc(&EvalFlag, &models.EvalResult{FlagID: 1}).Reset()

	e := NewEval()
	batch := models.EvaluationBatchRequest{
		Entities: []*models.EvaluationEntity{
			{EntityID: "e1", EntityContext: map[string]any{"tier": "premium"}},
		},
		FlagIDs: []int64{100},
	}
	raw, err := json.Marshal(batch)
	require.NoError(t, err)

	req := &http.Request{URL: &url.URL{RawQuery: "json=" + url.QueryEscape(string(raw))}}
	resp := e.GetEvaluationBatch(evaluation.GetEvaluationBatchParams{HTTPRequest: req, JSON: string(raw)})
	ok, isOK := resp.(*evaluation.GetEvaluationBatchOK)
	require.True(t, isOK, "got %T", resp)
	require.Len(t, ok.Payload.EvaluationResults, 1)
}

func TestGetEvaluation_SchemaParityWithPOST(t *testing.T) {
	e := NewEval()
	t.Run("flagID below minimum", func(t *testing.T) {
		raw := `{"flagID":-1,"entityID":"e1"}`
		req := &http.Request{URL: &url.URL{RawQuery: "json=" + url.QueryEscape(raw)}}
		resp := e.GetEvaluation(evaluation.GetEvaluationParams{HTTPRequest: req, JSON: raw})
		def, ok := resp.(*evaluation.GetEvaluationDefault)
		require.True(t, ok, "got %T", resp)
		require.NotNil(t, def.Payload)
		assert.Contains(t, *def.Payload.Message, "json is not valid evalContext")
		assert.Contains(t, *def.Payload.Message, "flagID")
	})
	t.Run("invalid flagTagsOperator", func(t *testing.T) {
		raw := `{"flagID":1,"flagTagsOperator":"BOTH"}`
		req := &http.Request{URL: &url.URL{RawQuery: "json=" + url.QueryEscape(raw)}}
		resp := e.GetEvaluation(evaluation.GetEvaluationParams{HTTPRequest: req, JSON: raw})
		def, ok := resp.(*evaluation.GetEvaluationDefault)
		require.True(t, ok)
		assert.Contains(t, *def.Payload.Message, "flagTagsOperator")
	})
}

func TestGetEvaluationBatch_SchemaParityWithPOST(t *testing.T) {
	e := NewEval()
	t.Run("missing entities", func(t *testing.T) {
		raw := `{"flagIDs":[1]}`
		req := &http.Request{URL: &url.URL{RawQuery: "json=" + url.QueryEscape(raw)}}
		resp := e.GetEvaluationBatch(evaluation.GetEvaluationBatchParams{HTTPRequest: req, JSON: raw})
		def, ok := resp.(*evaluation.GetEvaluationBatchDefault)
		require.True(t, ok, "got %T", resp)
		assert.Contains(t, *def.Payload.Message, "entities")
	})
	t.Run("empty entities array", func(t *testing.T) {
		raw := `{"entities":[],"flagIDs":[1]}`
		req := &http.Request{URL: &url.URL{RawQuery: "json=" + url.QueryEscape(raw)}}
		resp := e.GetEvaluationBatch(evaluation.GetEvaluationBatchParams{HTTPRequest: req, JSON: raw})
		def, ok := resp.(*evaluation.GetEvaluationBatchDefault)
		require.True(t, ok)
		assert.Contains(t, *def.Payload.Message, "entities")
	})
}

// Not parallel: "query too long" stubs config.Config.EvalGetMaxURLBytes.
func TestDecodeEvalContextFromGet(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "json=%7B%7D"}}

	t.Run("missing json", func(t *testing.T) {
		_, err := decodeEvalContextFromGet(req, 10, "")
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "missing required query parameter json")
	})

	t.Run("invalid syntax", func(t *testing.T) {
		_, err := decodeEvalContextFromGet(req, 10, "{")
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "json is not valid evalContext")
	})

	t.Run("schema violation", func(t *testing.T) {
		_, err := decodeEvalContextFromGet(req, 10, `{"flagID":-1,"entityID":"e1"}`)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "flagID")
	})

	t.Run("query too long", func(t *testing.T) {
		stubs := gostub.Stub(&config.Config.EvalGetMaxURLBytes, 8)
		defer stubs.Reset()
		_, err := decodeEvalContextFromGet(req, 9, `{"flagID":1}`)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "exceeds maximum")
	})

	t.Run("ok", func(t *testing.T) {
		ec, err := decodeEvalContextFromGet(req, 50, `{"flagID":1,"entityID":"e1"}`)
		require.Nil(t, err)
		assert.Equal(t, int64(1), ec.FlagID)
	})
}

// Not parallel: "query too long" stubs config.Config.EvalGetMaxURLBytes.
func TestDecodeEvaluationBatchFromGet(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "json=%7B%7D"}}

	t.Run("missing entities", func(t *testing.T) {
		_, err := decodeEvaluationBatchFromGet(req, 10, `{"flagIDs":[1]}`)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "entities")
	})

	t.Run("missing json", func(t *testing.T) {
		_, err := decodeEvaluationBatchFromGet(req, 10, "")
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "missing required query parameter json")
	})

	t.Run("query too long", func(t *testing.T) {
		stubs := gostub.Stub(&config.Config.EvalGetMaxURLBytes, 8)
		defer stubs.Reset()
		_, err := decodeEvaluationBatchFromGet(req, 9, `{"entities":[{"entityID":"e1"}]}`)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "exceeds maximum")
	})

	t.Run("ok", func(t *testing.T) {
		b, err := decodeEvaluationBatchFromGet(req, 80, `{"entities":[{"entityID":"e1"}],"flagIDs":[1]}`)
		require.Nil(t, err)
		require.Len(t, b.Entities, 1)
	})
}

func TestValidateSwaggerModelAfterJSON_nil(t *testing.T) {
	t.Parallel()
	err := validateSwaggerModelAfterJSON(nil, "evalContext", nil)
	require.NotNil(t, err)
	assert.Contains(t, *err.Message, "empty object")
}

func TestValidateEvalContextAfterJSON(t *testing.T) {
	t.Parallel()
	req := &http.Request{}

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		ec := models.EvalContext{FlagID: 1, EntityID: "e1"}
		assert.Nil(t, validateEvalContextAfterJSON(req, &ec))
	})

	t.Run("flagID negative", func(t *testing.T) {
		t.Parallel()
		ec := models.EvalContext{FlagID: -1, EntityID: "e1"}
		err := validateEvalContextAfterJSON(req, &ec)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "flagID")
	})

	t.Run("nil context", func(t *testing.T) {
		t.Parallel()
		err := validateEvalContextAfterJSON(req, nil)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "empty object")
	})
}

func TestValidateEvaluationBatchRequestAfterJSON(t *testing.T) {
	t.Parallel()
	req := &http.Request{}

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		b := models.EvaluationBatchRequest{
			Entities: []*models.EvaluationEntity{{EntityID: "e1"}},
			FlagIDs:  []int64{1},
		}
		assert.Nil(t, validateEvaluationBatchRequestAfterJSON(req, &b))
	})

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()
		err := validateEvaluationBatchRequestAfterJSON(req, nil)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "empty object")
	})

	t.Run("missing entities", func(t *testing.T) {
		t.Parallel()
		b := models.EvaluationBatchRequest{FlagIDs: []int64{1}}
		err := validateEvaluationBatchRequestAfterJSON(req, &b)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "entities")
	})
}

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
