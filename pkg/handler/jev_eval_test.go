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

// setupJevTest wires a fake client and enables Jev around the test.
func setupJevTest(t *testing.T, client *fakeJevClient) {
	t.Helper()
	stubEnabled := gostub.Stub(&config.Config.JevEnabled, true)
	stubThreshold := gostub.Stub(&config.Config.JevConfidenceThreshold, 0.5)
	stubDebug := gostub.Stub(&config.Config.EvalDebugEnabled, true)
	stubClient := gostub.StubFunc(&NewJevClient, JevClient(client))
	stubLog := gostub.StubFunc(&logEvalResult)
	t.Cleanup(func() {
		stubEnabled.Reset()
		stubThreshold.Reset()
		stubDebug.Reset()
		stubClient.Reset()
		stubLog.Reset()
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

func TestResolveJevForFlagDoesNotMutateContext(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"intent": {Type: entity.JevTypeNoul, Noul: jevF64(0.9)},
	}}
	setupJevTest(t, fake)

	c := jevConstraint(t, "@jev.intent", models.ConstraintOperatorGTE, "0.8",
		&entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "Is this intent?"})
	f := jevTestFlag(t, c)
	require.Len(t, f.FlagEvaluation.JevQuestions, 1)

	entityContext := map[string]any{"plan": "pro", "@ts": 123.0, "@http_host": "example.com"}
	evalContext := models.EvalContext{
		EntityID:      "user-42",
		EntityType:    "account",
		EntityContext: entityContext,
	}

	answers, debug := resolveJevForFlag(evalContext, &f)

	// The state sent to Jev carries built-ins + entity identity, never @jev.
	state, ok := fake.lastState.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "pro", state["plan"])
	assert.Equal(t, 123.0, state["@ts"])
	assert.Equal(t, "example.com", state["@http_host"])
	assert.Equal(t, "user-42", state["entityID"])
	assert.Equal(t, "account", state["entityType"])
	assert.NotContains(t, state, entity.JevContextKey, "@jev must never be fed back to the model")

	// Answers are returned separately and the caller's context is untouched.
	assert.Equal(t, 0.9, answers["intent"])
	assert.NotContains(t, entityContext, entity.JevContextKey)

	// The debug payload exposes the request and the response.
	require.NotNil(t, debug)
	assert.Equal(t, "Is this intent?", debug.Questions["intent"].Instructions)
	require.Contains(t, debug.Answers, "intent")
	assert.Empty(t, debug.Error)
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

// A segment without Jev constraints must not carry a Jev entry in its debug
// log, and a Jev segment's entry must be scoped to its own questions.
func TestJevDebugScopedToJevSegments(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"intent": {Type: entity.JevTypeNoul, Noul: jevF64(0.9)},
	}}
	setupJevTest(t, fake)

	// Segment 0: plain constraint that does not match, so eval falls through.
	plainSeg := entity.GenFixtureSegment()
	plainSeg.ID = 200
	plainSeg.Rank = 0
	plainSeg.Constraints = []entity.Constraint{{
		Property: "dl_state", Operator: models.ConstraintOperatorEQ, Value: `"NY"`,
	}}
	plainSeg.Constraints[0].ID = 500

	// Segment 1: Jev constraint that matches.
	jevSeg := entity.GenFixtureSegment()
	jevSeg.ID = 201
	jevSeg.Rank = 1
	jevC := jevConstraint(t, "@jev.intent", models.ConstraintOperatorGTE, "0.8",
		&entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "Is this intent?"})
	jevC.ID = 501
	jevC.SegmentID = 201
	jevSeg.Constraints = []entity.Constraint{jevC}

	f := entity.GenFixtureFlag()
	f.Segments = []entity.Segment{plainSeg, jevSeg}
	require.NoError(t, f.PrepareEvaluation())

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{
		EnableDebug:   true,
		FlagID:        100,
		EntityID:      "e1",
		EntityContext: map[string]any{"dl_state": "CA"},
	})
	require.Len(t, r.EvalDebugLog.SegmentDebugLogs, 2)

	// The plain segment has no Jev debug entry.
	assert.Nil(t, r.EvalDebugLog.SegmentDebugLogs[0].Jev)
	// The Jev segment has one, scoped to its own question.
	jevDebug, ok := r.EvalDebugLog.SegmentDebugLogs[1].Jev.(*JevDebug)
	require.True(t, ok)
	assert.Contains(t, jevDebug.Questions, "intent")
	require.Contains(t, jevDebug.Answers, "intent")
}

// A segment can mix plain and Jev constraints: the plain one resolves from
// entityContext, the Jev one from the answers merged in only for evaluation.
func TestJevConstraintMixedWithPlainConstraint(t *testing.T) {
	fake := &fakeJevClient{answers: map[string]JevAnswer{
		"intent": {Type: entity.JevTypeNoul, Noul: jevF64(0.9)},
	}}
	setupJevTest(t, fake)

	f := entity.GenFixtureFlag()
	plain := entity.Constraint{Property: "dl_state", Operator: models.ConstraintOperatorEQ, Value: `"CA"`}
	plain.ID = 500
	plain.SegmentID = 200
	jevC := jevConstraint(t, "@jev.intent", models.ConstraintOperatorGTE, "0.8",
		&entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "Is this intent?"})
	jevC.ID = 501
	jevC.SegmentID = 200
	f.Segments[0].Constraints = []entity.Constraint{plain, jevC}
	require.NoError(t, f.PrepareEvaluation())

	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

	r := EvalFlag(models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"dl_state": "CA"}})
	assert.NotZero(t, r.VariantID)

	// The result context stays clean — no `@jev`.
	ctx, ok := r.EvalContext.EntityContext.(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, ctx, entity.JevContextKey)
}

// Table-driven coverage of every question type and comparison direction,
// including the choice any-of / none-of logic and the confidence gate.
func TestJevConstraintQuestionTypes(t *testing.T) {
	noul := func() *entity.JevQuestion {
		return &entity.JevQuestion{Type: entity.JevTypeNoul, Instructions: "q?"}
	}
	choice := func() *entity.JevQuestion {
		return &entity.JevQuestion{
			Type:         entity.JevTypeChoice,
			Instructions: "q?",
			Criteria:     map[string]any{"free": "a", "pro": "b", "enterprise": "c"},
		}
	}
	scale := func() *entity.JevQuestion {
		return &entity.JevQuestion{
			Type:         entity.JevTypeScore,
			Instructions: "q?",
			Criteria:     []any{"low", "high"},
		}
	}
	noulAnswer := func(v float64) JevAnswer {
		return JevAnswer{Type: entity.JevTypeNoul, Noul: jevF64(v)}
	}
	choiceAnswer := func(picked string, confidence float64) JevAnswer {
		return JevAnswer{Type: entity.JevTypeChoice, Choice: picked, Confidence: jevF64(confidence)}
	}
	scaleAnswer := func(score, confidence float64) JevAnswer {
		return JevAnswer{Type: entity.JevTypeScore, Score: jevF64(score), Confidence: jevF64(confidence)}
	}

	cases := []struct {
		name      string
		question  *entity.JevQuestion
		operator  string
		value     string
		answer    JevAnswer
		wantMatch bool
	}{
		// noul: P(true) threshold
		{"noul gte match", noul(), models.ConstraintOperatorGTE, "0.8", noulAnswer(0.9), true},
		{"noul gte no match", noul(), models.ConstraintOperatorGTE, "0.8", noulAnswer(0.5), false},
		{"noul lt match", noul(), models.ConstraintOperatorLT, "0.5", noulAnswer(0.2), true},
		{"noul lt no match", noul(), models.ConstraintOperatorLT, "0.5", noulAnswer(0.9), false},

		// choice single: any of (EQ) / none of (NEQ)
		{"choice eq match", choice(), models.ConstraintOperatorEQ, `"pro"`, choiceAnswer("pro", 0.9), true},
		{"choice eq no match", choice(), models.ConstraintOperatorEQ, `"pro"`, choiceAnswer("free", 0.9), false},
		{"choice neq match", choice(), models.ConstraintOperatorNEQ, `"free"`, choiceAnswer("pro", 0.9), true},
		{"choice neq no match", choice(), models.ConstraintOperatorNEQ, `"free"`, choiceAnswer("free", 0.9), false},

		// choice multiple: any of (IN) / none of (NOTIN)
		{"choice in match", choice(), models.ConstraintOperatorIN, `["pro","enterprise"]`, choiceAnswer("enterprise", 0.9), true},
		{"choice in no match", choice(), models.ConstraintOperatorIN, `["pro","enterprise"]`, choiceAnswer("free", 0.9), false},
		{"choice notin match", choice(), models.ConstraintOperatorNOTIN, `["free"]`, choiceAnswer("enterprise", 0.9), true},
		{"choice notin no match", choice(), models.ConstraintOperatorNOTIN, `["free"]`, choiceAnswer("free", 0.9), false},

		// scale: at least (GTE) / below (LT)
		{"scale gte match", scale(), models.ConstraintOperatorGTE, "1", scaleAnswer(1.5, 0.9), true},
		{"scale gte no match", scale(), models.ConstraintOperatorGTE, "1", scaleAnswer(0.5, 0.9), false},
		{"scale lt match", scale(), models.ConstraintOperatorLT, "1", scaleAnswer(0.5, 0.9), true},
		{"scale lt no match", scale(), models.ConstraintOperatorLT, "1", scaleAnswer(1.5, 0.9), false},

		// confidence gate (choice / scale); noul has no separate confidence
		{"choice low confidence", choice(), models.ConstraintOperatorEQ, `"pro"`, choiceAnswer("pro", 0.2), false},
		{"scale low confidence", scale(), models.ConstraintOperatorGTE, "0", scaleAnswer(1.5, 0.2), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeJevClient{answers: map[string]JevAnswer{"q": tc.answer}}
			setupJevTest(t, fake)

			c := jevConstraint(t, "@jev.q", tc.operator, tc.value, tc.question)
			f := jevTestFlag(t, c)
			defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCacheWithFlags([]entity.Flag{f})).Reset()

			r := EvalFlag(models.EvalContext{FlagID: 100, EntityID: "e1", EntityContext: map[string]any{"x": 1}})
			if tc.wantMatch {
				assert.NotZero(t, r.VariantID, tc.name)
			} else {
				assert.Zero(t, r.VariantID, tc.name)
			}
		})
	}
}
