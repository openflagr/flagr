package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// orderEnricher records the context it sees and contributes one property.
func orderEnricher(namespace, prefix, key string, seen *[]map[string]any) enricher {
	return enricher{
		namespace: namespace,
		scope:     scopeGlobal,
		prefix:    prefix,
		enabled:   true,
		run: func(in enrichInput) (map[string]any, error) {
			snapshot := make(map[string]any, len(in.context))
			for k, v := range in.context {
				snapshot[k] = v
			}
			*seen = append(*seen, snapshot)
			return map[string]any{key: namespace}, nil
		},
	}
}

func TestEnrichContextOrderMaskAndOverwrite(t *testing.T) {
	t.Parallel()

	var seen []map[string]any
	entries := []enricher{
		orderEnricher("a", "@a_", "@a_x", &seen),
		orderEnricher("b", "@b_", "@b_x", &seen),
	}
	base := map[string]any{
		"client": "v",
		"@a_x":   "spoofed-a",
		"@b_x":   "spoofed-b",
	}
	out := enrichContext(entries, base, enrichInput{})

	// Enricher output overwrites client-provided/spoofed keys.
	assert.Equal(t, "a", out["@a_x"])
	assert.Equal(t, "b", out["@b_x"])
	assert.Equal(t, "v", out["client"])

	// First enricher sees the base without its own prefix but with the other's.
	assert.Equal(t, "spoofed-b", seen[0]["@b_x"])
	assert.NotContains(t, seen[0], "@a_x")
	// Second enricher sees the first enricher's output.
	assert.Equal(t, "a", seen[1]["@a_x"])
	assert.NotContains(t, seen[1], "@b_x")
}

func TestEnrichContextFailClosed(t *testing.T) {
	t.Parallel()

	failing := enricher{
		namespace: "bad",
		prefix:    "@bad_",
		enabled:   true,
		run: func(enrichInput) (map[string]any, error) {
			return map[string]any{"@bad_x": 1}, errors.New("boom")
		},
	}
	ok := orderEnricher("ok", "@ok_", "@ok_x", new([]map[string]any))

	out := enrichContext([]enricher{failing, ok}, map[string]any{}, enrichInput{})
	assert.NotContains(t, out, "@bad_x", "a failed enricher contributes nothing")
	assert.Equal(t, "ok", out["@ok_x"], "later enrichers still run")
}

func TestEnrichContextNoEntriesReturnsBase(t *testing.T) {
	t.Parallel()

	base := map[string]any{"a": 1}
	assert.Equal(t, base, enrichContext(nil, base, enrichInput{}))
}

func TestGlobalEnrichersProperties(t *testing.T) {
	t.Parallel()

	globals := globalEnrichers()
	require.Len(t, globals, 2)
	assert.Equal(t, entity.EnricherNamespaceTs, globals[0].namespace)
	assert.Equal(t, scopeGlobal, globals[0].scope)
	assert.Equal(t, []string{"@ts", "@ts_hour", "@ts_weekday", "@ts_month"}, globals[0].properties)
	assert.Equal(t, entity.EnricherNamespaceHTTP, globals[1].namespace)
}

func TestEffectiveEnrichersMergesFlagAndGlobals(t *testing.T) {
	t.Parallel()

	flag := &entity.Flag{}
	e := entity.Enricher{}
	require.NoError(t, e.SetConfig(entity.EnricherNamespaceJev, &JevEnricherConfig{
		Questions: map[string]JevQuestion{
			"plan_tier": {Type: JevTypeChoice, Instructions: "plan?", Criteria: map[string]any{"pro": "Pro"}},
		},
	}))
	flag.Enrichers = []entity.Enricher{e}
	require.NoError(t, flag.PrepareEvaluation())

	entries := effectiveEnrichers(flag)
	require.Len(t, entries, 3)
	assert.Equal(t, entity.EnricherNamespaceJev, entries[2].namespace)
	assert.Equal(t, scopeFlag, entries[2].scope)
	assert.Equal(t, []string{"@jev_plan_tier"}, entries[2].properties)
}

func TestUnknownEnrichedProperties(t *testing.T) {
	t.Parallel()

	flag := &entity.Flag{}
	flag.Segments = []entity.Segment{{
		Constraints: entity.ConstraintArray{
			{Property: "@ts_hour", Operator: "GTE", Value: "9"},
			{Property: "@http_x_env", Operator: "EQ", Value: "\"prod\""},
			{Property: "@jev_typo", Operator: "GTE", Value: "0.5"},
			{Property: "plain", Operator: "EQ", Value: "1"},
		},
	}}
	require.NoError(t, flag.PrepareEvaluation())

	assert.Equal(t, []string{"@jev_typo"}, unknownEnrichedProperties(flag))
}

// fakeJevClient is a stub System One endpoint.
type fakeJevClient struct {
	resp  *JevResponse
	err   error
	state any
	asks  map[string]JevQuestion
}

func (f *fakeJevClient) SystemOne(_ context.Context, state any, questions map[string]JevQuestion) (*JevResponse, error) {
	f.state = state
	f.asks = questions
	return f.resp, f.err
}

func setJevConfig(t *testing.T, cfg *JevEnricherConfig) *entity.Flag {
	t.Helper()
	e := entity.Enricher{}
	require.NoError(t, e.SetConfig(entity.EnricherNamespaceJev, cfg))
	flag := &entity.Flag{
		EntityType: "user",
		Enrichers:  []entity.Enricher{e},
	}
	require.NoError(t, flag.PrepareEvaluation())
	return flag
}

func TestEnrichFlagContextJevEndToEnd(t *testing.T) {
	origJevURL := config.Config.InjectedContextJevBaseURL
	origNew := NewJevClient
	defer func() {
		config.Config.InjectedContextJevBaseURL = origJevURL
		NewJevClient = origNew
	}()
	config.Config.InjectedContextJevBaseURL = "http://jev.test"

	flag := setJevConfig(t, &JevEnricherConfig{
		Questions: map[string]JevQuestion{
			"plan_tier": {Type: JevTypeChoice, Instructions: "plan?", Criteria: map[string]any{"pro": "Pro"}},
			"churn":     {Type: JevTypeNoul, Instructions: "churn?"},
		},
	})

	fake := &fakeJevClient{resp: &JevResponse{Answers: map[string]JevAnswer{
		"plan_tier": {Type: JevTypeChoice, Choice: "pro", Confidence: ptr(0.9)},
		"churn":     {Type: JevTypeNoul, Noul: ptr(0.8)},
	}}}
	NewJevClient = func() JevClient { return fake }

	base := map[string]any{
		"country":    "US",
		"@ts_hour":   float64(14),
		"@jev_stale": "old",
	}
	out := enrichFlagContext(base, flag, "u-1", "user").(map[string]any)

	assert.Equal(t, "pro", out["@jev_plan_tier"])
	assert.Equal(t, 0.8, out["@jev_churn"])
	assert.Equal(t, "US", out["country"])
	assert.Equal(t, float64(14), out["@ts_hour"])

	// The stale @jev_ property never reaches the model state.
	state := fake.state.(map[string]any)
	assert.NotContains(t, state, "@jev_stale")
	assert.Equal(t, "u-1", state["entityID"])
	assert.Equal(t, "user", state["entityType"])
	assert.Equal(t, float64(14), state["@ts_hour"])
}

func TestEnrichFlagContextJevFailClosed(t *testing.T) {
	origJevURL := config.Config.InjectedContextJevBaseURL
	origNew := NewJevClient
	defer func() {
		config.Config.InjectedContextJevBaseURL = origJevURL
		NewJevClient = origNew
	}()
	config.Config.InjectedContextJevBaseURL = "http://jev.test"

	flag := setJevConfig(t, &JevEnricherConfig{
		Questions: map[string]JevQuestion{
			"plan_tier": {Type: JevTypeChoice, Instructions: "plan?", Criteria: map[string]any{"pro": "Pro"}},
		},
	})
	NewJevClient = func() JevClient { return &fakeJevClient{err: errors.New("timeout")} }

	base := map[string]any{"country": "US"}
	out := enrichFlagContext(base, flag, "u-1", "user").(map[string]any)
	assert.NotContains(t, out, "@jev_plan_tier")
	assert.Equal(t, "US", out["country"])
}

func TestEnrichFlagContextJevDisabled(t *testing.T) {
	origJevURL := config.Config.InjectedContextJevBaseURL
	defer func() { config.Config.InjectedContextJevBaseURL = origJevURL }()
	config.Config.InjectedContextJevBaseURL = ""

	flag := setJevConfig(t, &JevEnricherConfig{
		Questions: map[string]JevQuestion{
			"plan_tier": {Type: JevTypeChoice, Instructions: "plan?", Criteria: map[string]any{"pro": "Pro"}},
		},
	})

	out := enrichFlagContext(map[string]any{"a": 1}, flag, "u", "user").(map[string]any)
	assert.NotContains(t, out, "@jev_plan_tier")
	assert.Equal(t, 1, out["a"])
}

func TestEnabledEnrichersRespectsEnv(t *testing.T) {
	origEnabled := config.Config.InjectedContextEnabled
	origJevURL := config.Config.InjectedContextJevBaseURL
	defer func() {
		config.Config.InjectedContextEnabled = origEnabled
		config.Config.InjectedContextJevBaseURL = origJevURL
	}()

	flag := setJevConfig(t, validJevConfig())

	config.Config.InjectedContextEnabled = false
	config.Config.InjectedContextJevBaseURL = ""
	assert.Empty(t, enabledEnrichers(effectiveEnrichers(flag)))

	config.Config.InjectedContextEnabled = true
	config.Config.InjectedContextJevBaseURL = "http://jev.test"
	assert.Len(t, enabledEnrichers(effectiveEnrichers(flag)), 3)

	config.Config.InjectedContextEnabled = false
	config.Config.InjectedContextJevBaseURL = "http://jev.test"
	got := enabledEnrichers(effectiveEnrichers(flag))
	require.Len(t, got, 1)
	assert.Equal(t, entity.EnricherNamespaceJev, got[0].namespace)
}

func TestVisibleEnrichersKeepsDisabledFlagScoped(t *testing.T) {
	origEnabled := config.Config.InjectedContextEnabled
	origJevURL := config.Config.InjectedContextJevBaseURL
	defer func() {
		config.Config.InjectedContextEnabled = origEnabled
		config.Config.InjectedContextJevBaseURL = origJevURL
	}()

	flag := setJevConfig(t, validJevConfig())

	// Nothing enabled: the flag-scoped declaration stays visible so the UI can
	// still author it, but it is marked disabled and never runs.
	config.Config.InjectedContextEnabled = false
	config.Config.InjectedContextJevBaseURL = ""
	got := visibleEnrichers(effectiveEnrichers(flag))
	require.Len(t, got, 1)
	assert.Equal(t, entity.EnricherNamespaceJev, got[0].namespace)
	assert.False(t, got[0].enabled)
	assert.Empty(t, enabledEnrichers(effectiveEnrichers(flag)))

	models := effectiveEnricherModels(flag)
	require.Len(t, models, 1)
	assert.Equal(t, entity.EnricherNamespaceJev, *models[0].Namespace)
	require.NotNil(t, models[0].Enabled)
	assert.False(t, *models[0].Enabled)
	assert.Contains(t, models[0].Properties, "@jev_plan_tier")

	// Globals appear only once enabled, ahead of the flag-scoped entries.
	config.Config.InjectedContextEnabled = true
	config.Config.InjectedContextJevBaseURL = ""
	namespaces := make([]string, 0, 3)
	for _, e := range effectiveEnricherModels(flag) {
		namespaces = append(namespaces, *e.Namespace)
	}
	assert.Equal(t,
		[]string{entity.EnricherNamespaceTs, entity.EnricherNamespaceHTTP, entity.EnricherNamespaceJev},
		namespaces)
}

func TestEnrichFlagContextDoesNotMutateCallerContext(t *testing.T) {
	origURL := config.Config.InjectedContextJevBaseURL
	origNew := NewJevClient
	defer func() {
		config.Config.InjectedContextJevBaseURL = origURL
		NewJevClient = origNew
	}()
	config.Config.InjectedContextJevBaseURL = "http://jev.test"

	flag := setJevConfig(t, &JevEnricherConfig{
		Questions: map[string]JevQuestion{
			"plan_tier": {Type: JevTypeChoice, Instructions: "plan?", Criteria: map[string]any{"pro": "Pro"}},
		},
	})
	NewJevClient = func() JevClient {
		return &fakeJevClient{resp: &JevResponse{Answers: map[string]JevAnswer{
			"plan_tier": {Type: JevTypeChoice, Choice: "pro", Confidence: ptr(0.9)},
		}}}
	}

	base := map[string]any{"country": "US"}
	out := enrichFlagContext(base, flag, "u-1", "user").(map[string]any)

	assert.Equal(t, "pro", out["@jev_plan_tier"])
	assert.NotContains(t, base, "@jev_plan_tier",
		"flag enrichment must not write through to the shared caller context")

	// Same base reused for a flag without enrichers must stay clean (batch case).
	plain := enrichFlagContext(base, &entity.Flag{}, "u-1", "user").(map[string]any)
	assert.NotContains(t, plain, "@jev_plan_tier")
}

func TestEnrichFlagContextNoEnrichersReturnsBase(t *testing.T) {
	t.Parallel()

	base := map[string]any{"a": 1}
	assert.Equal(t, base, enrichFlagContext(base, &entity.Flag{}, "u", "user"))
}

func TestEnrichFlagContextConfidenceGate(t *testing.T) {
	origJevURL := config.Config.InjectedContextJevBaseURL
	origNew := NewJevClient
	defer func() {
		config.Config.InjectedContextJevBaseURL = origJevURL
		NewJevClient = origNew
	}()
	config.Config.InjectedContextJevBaseURL = "http://jev.test"

	threshold := 0.9
	flag := setJevConfig(t, &JevEnricherConfig{
		Questions: map[string]JevQuestion{
			"plan_tier": {
				Type:                JevTypeChoice,
				Instructions:        "plan?",
				Criteria:            map[string]any{"pro": "Pro"},
				ConfidenceThreshold: &threshold,
			},
		},
	})
	NewJevClient = func() JevClient {
		return &fakeJevClient{resp: &JevResponse{Answers: map[string]JevAnswer{
			"plan_tier": {Type: JevTypeChoice, Choice: "pro", Confidence: ptr(0.5)},
		}}}
	}

	out := enrichFlagContext(map[string]any{}, flag, "u", "user").(map[string]any)
	assert.NotContains(t, out, "@jev_plan_tier", "answer below threshold must be omitted")
}

func ptr[T any](v T) *T { return &v }
