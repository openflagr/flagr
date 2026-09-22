package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJevEndToEndWithMockServer drives the full evaluation path against an HTTP
// mock that implements the System One contract (the same contract as the hosted
// API and the open-source oido-systemone / jeff drop-in servers). No fake client
// is used: this exercises the real HTTP client, request/response shapes, answer
// injection, and conditions evaluation.
func TestJevEndToEndWithMockServer(t *testing.T) {
	var mu sync.Mutex
	var received jevRequestPayload
	var authHeader string
	var calls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/systemone" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		mu.Lock()
		calls++
		authHeader = r.Header.Get("Authorization")
		mu.Unlock()

		var body jevRequestPayload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mu.Lock()
		received = body
		mu.Unlock()

		answers := make(map[string]JevAnswer, len(body.Questions))
		for name, q := range body.Questions {
			switch q.Type {
			case entity.JevTypeNoul:
				answers[name] = JevAnswer{Type: entity.JevTypeNoul, Noul: jevF64(0.92)}
			case entity.JevTypeChoice:
				choice := ""
				if criteria, ok := q.Criteria.(map[string]any); ok {
					for option := range criteria {
						if option == "pro" {
							choice = "pro"
						}
					}
				}
				answers[name] = JevAnswer{Type: entity.JevTypeChoice, Choice: choice, Confidence: jevF64(0.9)}
			case entity.JevTypeScore:
				answers[name] = JevAnswer{Type: entity.JevTypeScore, Score: jevF64(2.0), Confidence: jevF64(0.85)}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-typesafe-request-id", "req-123")
		_ = json.NewEncoder(w).Encode(JevResponse{Model: "mock-jev-1.13", Answers: answers, LatencyMs: jevF64(11.0)})
	}))
	defer server.Close()

	stubBase := gostub.Stub(&config.Config.JevBaseURL, server.URL)
	stubEnabled := gostub.Stub(&config.Config.JevEnabled, true)
	stubModel := gostub.Stub(&config.Config.JevModel, "jev-latest")
	stubKey := gostub.Stub(&config.Config.JevAPIKey, "test-key")
	stubThreshold := gostub.Stub(&config.Config.JevConfidenceThreshold, 0.5)
	stubDebug := gostub.Stub(&config.Config.EvalDebugEnabled, true)
	stubLog := gostub.StubFunc(&logEvalResult)
	ResetJevCache()
	t.Cleanup(func() {
		stubBase.Reset()
		stubEnabled.Reset()
		stubModel.Reset()
		stubKey.Reset()
		stubThreshold.Reset()
		stubDebug.Reset()
		stubLog.Reset()
		ResetJevCache()
	})

	// One segment ANDing all three question types.
	f := entity.GenFixtureFlag()
	noul := jevConstraint(t, "@jev.buying_intent", models.ConstraintOperatorGTE, "0.8",
		&entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "Is this account showing buying intent?"})
	choice := jevConstraint(t, "@jev.plan_tier", models.ConstraintOperatorEQ, `"pro"`,
		&entity.JevQuestion{
			Type:         entity.JevTypeChoice,
			Instructions: "Which plan should this account see?",
			Criteria:     map[string]any{"free": "self-serve", "pro": "team usage", "enterprise": "procurement"},
		})
	score := jevConstraint(t, "@jev.risk", models.ConstraintOperatorLT, "2.5",
		&entity.JevQuestion{
			Type:         entity.JevTypeScore,
			Instructions: "How risky is this transaction?",
			Criteria:     []any{"low", "medium", "high"},
		})
	noul.ID, choice.ID, score.ID = 500, 501, 502
	f.Segments[0].Constraints = []entity.Constraint{noul, choice, score}
	require.NoError(t, f.PrepareEvaluation())
	require.Len(t, f.FlagEvaluation.JevQuestions, 3)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	result := EvalFlag(models.EvalContext{
		EnableDebug:   true,
		FlagID:        100,
		EntityID:      "e1",
		EntityContext: map[string]any{"plan": "pro", "amount": 900, "@ts": 123.0},
	})

	// The segment matched, so a variant was rolled out.
	assert.NotZero(t, result.VariantID)

	// The mock received a single batched call with all three questions.
	mu.Lock()
	assert.Equal(t, 1, calls)
	assert.Equal(t, "Bearer test-key", authHeader)
	assert.Equal(t, "jev-latest", received.Model)
	mu.Unlock()

	require.Len(t, received.Questions, 3)
	assert.Equal(t, "noul", received.Questions["buying_intent"].Type)
	assert.Equal(t, "choice", received.Questions["plan_tier"].Type)
	assert.Equal(t, "score", received.Questions["risk"].Type)

	// Server-injected built-in context is included; only @jev is excluded.
	state, ok := received.State.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "pro", state["plan"])
	assert.Equal(t, "e1", state["entityID"])
	assert.Contains(t, state, "@ts")
	assert.NotContains(t, state, entity.JevContextKey)

	// Debug output carries the Jev request and response.
	require.NotEmpty(t, result.EvalDebugLog.SegmentDebugLogs)
	jevDebug, ok := result.EvalDebugLog.SegmentDebugLogs[0].Jev.(*JevDebug)
	require.True(t, ok)
	assert.Contains(t, jevDebug.Questions, "buying_intent")
	assert.Contains(t, jevDebug.Questions, "plan_tier")
	assert.Contains(t, jevDebug.Answers, "risk")
	assert.Equal(t, "mock-jev-1.13", jevDebug.Model)
	require.NotNil(t, jevDebug.State)
	assert.GreaterOrEqual(t, jevDebug.LatencyMs, 0.0)
	require.NotNil(t, jevDebug.ServerLatencyMs)
	assert.Equal(t, 11.0, *jevDebug.ServerLatencyMs)

	// The injected answers are also visible on the eval context.
	evalCtx, ok := result.EvalContext.EntityContext.(map[string]any)
	require.True(t, ok)
	assert.Contains(t, evalCtx, entity.JevContextKey)
}

// TestJevEndToEndMockServerFallThrough verifies fail-closed behavior when the
// System One endpoint returns an error: the segment does not match.
func TestJevEndToEndMockServerFallThrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"detail":{"error_type":"server_error","message":"boom"}}`))
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()
	defer gostub.Stub(&config.Config.JevEnabled, true).Reset()
	defer gostub.StubFunc(&logEvalResult).Reset()
	ResetJevCache()
	defer ResetJevCache()

	c := jevConstraint(t, "@jev.intent", models.ConstraintOperatorGTE, "0.8",
		&entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "Is this intent?"})
	f := jevTestFlag(t, c)
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	result := EvalFlag(models.EvalContext{
		FlagID:        100,
		EntityID:      "e1",
		EntityContext: map[string]any{"plan": "pro"},
	})
	assert.Zero(t, result.VariantID)
}
