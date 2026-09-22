package handler

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhouzhuojie/conditions"
)

type fakeJevClient struct {
	answers       map[string]JevAnswer
	err           error
	calls         int
	lastState     any
	lastQuestions map[string]entity.JevConstraintSpec
}

func (f *fakeJevClient) SystemOne(_ context.Context, state any, questions map[string]entity.JevConstraintSpec) (*JevResponse, error) {
	f.calls++
	f.lastState = state
	f.lastQuestions = questions
	if f.err != nil {
		return nil, f.err
	}
	return &JevResponse{Answers: f.answers}, nil
}

func jevF64(v float64) *float64 { return &v }

// jevTestFlag builds the fixture flag with a single Jev constraint and prepares it.
func jevTestFlag(t *testing.T, c entity.Constraint) entity.Flag {
	t.Helper()
	f := entity.GenFixtureFlag()
	c.ID = 500
	c.SegmentID = 200
	f.Segments[0].Constraints = []entity.Constraint{c}
	require.NoError(t, f.PrepareEvaluation())
	return f
}

// jevConstraint builds a Jev constraint with the given question.
func jevConstraint(t *testing.T, property, operator, value string, q *entity.JevQuestion) entity.Constraint {
	t.Helper()
	c := entity.Constraint{Property: property, Operator: operator, Value: value}
	require.NoError(t, c.SetJevQuestion(q))
	return c
}

// setupJevTest wires a fake client, enables Jev, and resets caches around the test.
func setupJevTest(t *testing.T, client *fakeJevClient) {
	t.Helper()
	stubEnabled := gostub.Stub(&config.Config.JevEnabled, true)
	stubThreshold := gostub.Stub(&config.Config.JevConfidenceThreshold, 0.5)
	stubDebug := gostub.Stub(&config.Config.EvalDebugEnabled, true)
	stubClient := gostub.StubFunc(&NewJevClient, JevClient(client))
	stubLog := gostub.StubFunc(&logEvalResult)
	ResetJevCache()
	t.Cleanup(func() {
		stubEnabled.Reset()
		stubThreshold.Reset()
		stubDebug.Reset()
		stubClient.Reset()
		stubLog.Reset()
		ResetJevCache()
	})
}

func TestJevSyntheticPropertyResolves(t *testing.T) {
	t.Parallel()
	m := map[string]any{"@jev": map[string]any{"noul": 0.9, "plan_tier": "pro", "score": 1.5}}
	cases := map[string]bool{
		`({@jev.noul} >= 0.8)`:        true,
		`({@jev.plan_tier} == "pro")`: true,
		`({@jev.score} >= 1)`:         true,
		`({@jev.score} >= 2)`:         false,
	}
	for exprStr, want := range cases {
		p := conditions.NewParser(strings.NewReader(exprStr))
		expr, err := p.Parse()
		require.NoError(t, err, exprStr)
		got, err := conditions.Evaluate(expr, m)
		require.NoError(t, err, exprStr)
		assert.Equal(t, want, got, exprStr)
	}
}

func TestJevConstraintNoulMatches(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"intent": {Type: entity.JevTypeNoul, Noul: jevF64(0.9)},
	}}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.intent", models.ConstraintOperatorGTE, "0.8",
		&entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "Is this buying intent?"})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{
		EnableDebug:   true,
		FlagID:        100,
		EntityID:      "e1",
		EntityContext: map[string]any{"plan": "pro"},
	})
	assert.NotZero(t, r.VariantID)
	assert.Equal(t, 1, fake.calls)
	require.NotNil(t, fake.lastState)
	if m, ok := fake.lastState.(map[string]any); assert.True(t, ok) {
		assert.Equal(t, "pro", m["plan"])
	}
}

func TestJevConstraintChoiceMatches(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"plan_tier": {Type: entity.JevTypeChoice, Choice: "pro", Confidence: jevF64(0.9)},
	}}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.plan_tier", models.ConstraintOperatorEQ, `"pro"`,
		&entity.JevQuestion{
			Type:         entity.JevTypeChoice,
			Instructions: "Which plan should this account see?",
			Criteria:     map[string]any{"free": "self-serve", "pro": "team usage"},
		})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"plan": "pro"}})
	assert.NotZero(t, r.VariantID)
}

func TestJevConstraintScoreMatches(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"risk": {Type: entity.JevTypeScore, Score: jevF64(1.5), Confidence: jevF64(0.8)},
	}}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.risk", models.ConstraintOperatorGTE, "1",
		&entity.JevQuestion{
			Type:         entity.JevTypeScore,
			Instructions: "How risky is this transaction?",
			Criteria:     []any{"low", "medium", "high"},
		})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"amount": 900}})
	assert.NotZero(t, r.VariantID)
}

func TestJevConstraintLowConfidenceDoesNotMatch(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"plan_tier": {Type: entity.JevTypeChoice, Choice: "pro", Confidence: jevF64(0.2)},
	}}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.plan_tier", models.ConstraintOperatorEQ, `"pro"`,
		&entity.JevQuestion{
			Type:         entity.JevTypeChoice,
			Instructions: "Which plan?",
			Criteria:     map[string]any{"free": "x", "pro": "y"},
		})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{EnableDebug: true, FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"plan": "pro"}})
	assert.Zero(t, r.VariantID)
	require.NotEmpty(t, r.EvalDebugLog.SegmentDebugLogs)
	assert.Contains(t, r.EvalDebugLog.SegmentDebugLogs[0].Msg, "not found")
}

func TestJevConstraintPerConstraintThreshold(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"plan_tier": {Type: entity.JevTypeChoice, Choice: "pro", Confidence: jevF64(0.6)},
	}}
	setupJevTest(t, fake)

	// Global threshold is 0.5; a per-constraint override of 0.8 must reject 0.6.
	c := jevConstraint(t, "@jev.plan_tier", models.ConstraintOperatorEQ, `"pro"`,
		&entity.JevQuestion{
			Type:                entity.JevTypeChoice,
			Instructions:        "Which plan?",
			Criteria:            map[string]any{"free": "x", "pro": "y"},
			ConfidenceThreshold: jevF64(0.8),
		})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"plan": "pro"}})
	assert.Zero(t, r.VariantID)
}

func TestJevConstraintFailClosedOnError(t *testing.T) {
	fake := &fakeJevClient{err: errors.New("connection refused")}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.intent", models.ConstraintOperatorGTE, "0.8",
		&entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "Is this intent?"})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"plan": "pro"}})
	assert.Zero(t, r.VariantID)
	assert.Equal(t, 1, fake.calls)
}

func TestJevConstraintDisabledNoCall(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"intent": {Type: entity.JevTypeNoul, Noul: jevF64(0.99)},
	}}
	setupJevTest(t, fake)
	defer gostub.Stub(&config.Config.JevEnabled, false).Reset()

	c := jevConstraint(t, "@jev.intent", models.ConstraintOperatorGTE, "0.8",
		&entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "Is this intent?"})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"plan": "pro"}})
	assert.Zero(t, r.VariantID)
	assert.Zero(t, fake.calls)
}

func TestJevConstraintCacheHit(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"intent": {Type: entity.JevTypeNoul, Noul: jevF64(0.9)},
	}}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.intent", models.ConstraintOperatorGTE, "0.8",
		&entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "Is this intent?"})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	ctx := models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"plan": "pro"}}
	assert.NotZero(t, EvalFlag(ctx).VariantID)
	assert.NotZero(t, EvalFlag(ctx).VariantID)
	assert.Equal(t, 1, fake.calls, "second evaluation should be served from cache")
}

func TestInjectJevContextKeepsBuiltInKeysAndInjects(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"intent": {Type: entity.JevTypeNoul, Noul: jevF64(0.9)},
	}}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.intent", models.ConstraintOperatorGTE, "0.8",
		&entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "Is this intent?"})
	f := jevTestFlag(t, c)
	require.Len(t, f.FlagEvaluation.JevQuestions, 1)

	out := injectJevContext(models.EvalContext{
		EntityContext: map[string]any{"plan": "pro", "@ts": 123.0, "@http_host": "example.com"},
	}, &f)

	state, ok := fake.lastState.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "pro", state["plan"])
	assert.Equal(t, 123.0, state["@ts"])
	assert.Equal(t, "example.com", state["@http_host"])
	assert.NotContains(t, state, entity.JevContextKey, "@jev must never be fed back to the model")

	injected, ok := out.EntityContext.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 0.9, injected[entity.JevContextKey].(map[string]any)["intent"])
}

// The UI emits these exact value/operator shapes; keep the backend contract covered.

func TestJevConstraintChoiceMultiOptionIN(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"plan_tier": {Type: entity.JevTypeChoice, Choice: "enterprise", Confidence: jevF64(0.9)},
	}}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.plan_tier", models.ConstraintOperatorIN, `["pro","enterprise"]`,
		&entity.JevQuestion{
			Type:         entity.JevTypeChoice,
			Instructions: "Which plan?",
			Criteria:     map[string]any{"free": "x", "pro": "y", "enterprise": "z"},
		})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"plan": "enterprise"}})
	assert.NotZero(t, r.VariantID)
}

func TestJevConstraintChoiceNegatedNOTIN(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"plan_tier": {Type: entity.JevTypeChoice, Choice: "enterprise", Confidence: jevF64(0.9)},
	}}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.plan_tier", models.ConstraintOperatorNOTIN, `["free"]`,
		&entity.JevQuestion{
			Type:         entity.JevTypeChoice,
			Instructions: "Which plan?",
			Criteria:     map[string]any{"free": "x", "enterprise": "z"},
		})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"plan": "enterprise"}})
	assert.NotZero(t, r.VariantID)
}

func TestJevConstraintScaleNegatedLT(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"risk": {Type: entity.JevTypeScore, Score: jevF64(0.5), Confidence: jevF64(0.9)},
	}}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.risk", models.ConstraintOperatorLT, "1",
		&entity.JevQuestion{
			Type:         entity.JevTypeScore,
			Instructions: "How risky?",
			Criteria:     []any{"low", "high"},
		})
	f := jevTestFlag(t, c)

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"amount": 1}})
	assert.NotZero(t, r.VariantID)
}
