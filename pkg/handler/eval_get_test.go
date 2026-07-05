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
	orig := config.Config.EvalGetMaxURLBytes
	config.Config.EvalGetMaxURLBytes = 10
	defer func() { config.Config.EvalGetMaxURLBytes = orig }()

	e := NewEval()
	req := &http.Request{URL: &url.URL{RawQuery: "json=%7B%7D"}}
	resp := e.GetEvaluation(evaluation.GetEvaluationParams{HTTPRequest: req, JSON: "{}"})
	_, isDef := resp.(*evaluation.GetEvaluationDefault)
	assert.True(t, isDef)
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
