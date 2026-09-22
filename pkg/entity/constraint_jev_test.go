package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstraintJevValidate(t *testing.T) {
	t.Parallel()

	newChoice := func() *JevQuestion {
		return &JevQuestion{
			Type:         JevTypeChoice,
			Instructions: "Which plan?",
			Criteria:     map[string]any{"free": "no procurement", "pro": "card payment"},
		}
	}
	newScore := func() *JevQuestion {
		return &JevQuestion{
			Type:         JevTypeScore,
			Instructions: "How risky?",
			Criteria:     []any{"low", "medium", "high"},
		}
	}

	cases := []struct {
		name      string
		property  string
		operator  string
		value     string
		question  *JevQuestion
		wantError bool
	}{
		{
			name:     "noul ok",
			property: "@jev.intent",
			operator: "GTE",
			value:    "0.7",
			question: &JevQuestion{Type: JevTypeNoul, Instructions: "Is this buying intent?"},
		},
		{
			name:     "choice ok",
			property: "@jev.plan",
			operator: "EQ",
			value:    `"pro"`,
			question: newChoice(),
		},
		{
			name:     "score ok",
			property: "@jev.risk",
			operator: "LT",
			value:    "1",
			question: newScore(),
		},
		{
			name:      "invalid type",
			property:  "@jev.intent",
			operator:  "GTE",
			value:     "0.7",
			question:  &JevQuestion{Type: "classify", Instructions: "x"},
			wantError: true,
		},
		{
			name:      "missing prefix",
			property:  "intent",
			operator:  "GTE",
			value:     "0.7",
			question:  &JevQuestion{Type: JevTypeNoul, Instructions: "x"},
			wantError: true,
		},
		{
			name:      "choice missing criteria",
			property:  "@jev.plan",
			operator:  "EQ",
			value:     `"pro"`,
			question:  &JevQuestion{Type: JevTypeChoice, Instructions: "x"},
			wantError: true,
		},
		{
			name:      "choice empty criteria",
			property:  "@jev.plan",
			operator:  "EQ",
			value:     `"pro"`,
			question:  &JevQuestion{Type: JevTypeChoice, Instructions: "x", Criteria: map[string]any{}},
			wantError: true,
		},
		{
			name:      "score too few levels",
			property:  "@jev.risk",
			operator:  "GT",
			value:     "0",
			question:  &JevQuestion{Type: JevTypeScore, Instructions: "x", Criteria: []any{"only"}},
			wantError: true,
		},
		{
			name:      "score too many levels",
			property:  "@jev.risk",
			operator:  "GT",
			value:     "0",
			question:  &JevQuestion{Type: JevTypeScore, Instructions: "x", Criteria: []any{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}},
			wantError: true,
		},
		{
			name:      "choice with numeric operator",
			property:  "@jev.plan",
			operator:  "GTE",
			value:     `"pro"`,
			question:  newChoice(),
			wantError: true,
		},
		{
			name:      "noul with set operator",
			property:  "@jev.intent",
			operator:  "IN",
			value:     `["x"]`,
			question:  &JevQuestion{Type: JevTypeNoul, Instructions: "x"},
			wantError: true,
		},
		{
			name:      "threshold above one",
			property:  "@jev.plan",
			operator:  "EQ",
			value:     `"pro"`,
			question:  &JevQuestion{Type: JevTypeChoice, Instructions: "x", Criteria: map[string]any{"pro": "x"}, ConfidenceThreshold: f64(1.5)},
			wantError: true,
		},
		{
			name:     "threshold boundary ok",
			property: "@jev.plan",
			operator: "EQ",
			value:    `"pro"`,
			question: &JevQuestion{Type: JevTypeChoice, Instructions: "x", Criteria: map[string]any{"pro": "x"}, ConfidenceThreshold: f64(1)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := Constraint{Property: tc.property, Operator: tc.operator, Value: tc.value}
			// SetJevQuestion validates the question; Validate also checks the
			// constraint-level property prefix and match operator.
			err := c.SetJevQuestion(tc.question)
			if err == nil {
				err = c.Validate()
			}
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConstraintSetJevQuestionRoundTrip(t *testing.T) {
	t.Parallel()
	c := Constraint{}
	require.NoError(t, c.SetJevQuestion(&JevQuestion{
		Type:                JevTypeScore,
		Instructions:        map[string]any{"question": "How relevant?"},
		Criteria:            []any{"none", "some", "deep"},
		ConfidenceThreshold: f64(0.4),
	}))
	assert.True(t, c.IsJev())
	assert.NotEmpty(t, c.JevJSON)

	named := Constraint{Property: "@jev.risk"}
	require.NoError(t, named.SetJevQuestion(&JevQuestion{Type: JevTypeNoul, Instructions: "x"}))
	assert.Equal(t, "risk", named.JevName())

	q, err := c.JevQuestion()
	require.NoError(t, err)
	require.NotNil(t, q)
	assert.Equal(t, JevTypeScore, q.Type)
	assert.Equal(t, map[string]any{"question": "How relevant?"}, q.Instructions)
	assert.Equal(t, []any{"none", "some", "deep"}, q.Criteria)
	require.NotNil(t, q.ConfidenceThreshold)
	assert.Equal(t, 0.4, *q.ConfidenceThreshold)

	// Clearing removes the stored question.
	require.NoError(t, c.SetJevQuestion(nil))
	assert.False(t, c.IsJev())
	assert.Empty(t, c.JevJSON)
}

func TestSetJevQuestionRejectsInvalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		q    *JevQuestion
	}{
		{"unknown type", &JevQuestion{Type: "classify", Instructions: "x"}},
		{"missing instructions", &JevQuestion{Type: JevTypeNoul}},
		{"empty instructions", &JevQuestion{Type: JevTypeNoul, Instructions: "  "}},
		{"choice without criteria", &JevQuestion{Type: JevTypeChoice, Instructions: "x"}},
		{"choice empty criteria", &JevQuestion{Type: JevTypeChoice, Instructions: "x", Criteria: map[string]any{}}},
		{"score too few levels", &JevQuestion{Type: JevTypeScore, Instructions: "x", Criteria: []any{"only"}}},
		{"threshold out of range", &JevQuestion{Type: JevTypeChoice, Instructions: "x", Criteria: map[string]any{"a": "b"}, ConfidenceThreshold: f64(1.5)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := Constraint{}
			assert.Error(t, c.SetJevQuestion(tc.q))
			assert.Empty(t, c.JevJSON, "invalid question must not be written")
		})
	}
}

func TestConstraintJevQuestionRejectsMalformedJSON(t *testing.T) {
	t.Parallel()
	c := Constraint{Property: "@jev.risk", JevJSON: "{not json"}
	_, err := c.JevQuestion()
	assert.Error(t, err)
	assert.Error(t, c.Validate())
}

func TestFlagCollectJevQuestions(t *testing.T) {
	t.Parallel()
	f := GenFixtureFlag()
	first := Constraint{Property: "@jev.intent", Operator: "GTE", Value: "0.7"}
	require.NoError(t, first.SetJevQuestion(&JevQuestion{Type: JevTypeNoul, Instructions: "first wins"}))
	duplicate := Constraint{Property: "@jev.intent", Operator: "GTE", Value: "0.9"}
	require.NoError(t, duplicate.SetJevQuestion(&JevQuestion{Type: JevTypeNoul, Instructions: "dropped"}))
	plain := Constraint{Property: "dl_state", Operator: "EQ", Value: `"CA"`}
	f.Segments[0].Constraints = []Constraint{first, duplicate, plain}
	require.NoError(t, f.PrepareEvaluation())

	require.Len(t, f.FlagEvaluation.JevQuestions, 1)
	spec := f.FlagEvaluation.JevQuestions["intent"]
	assert.Equal(t, "first wins", spec.Instructions)
}

func TestDuplicateJevQuestionProperties(t *testing.T) {
	t.Parallel()
	jev := func(property string) Constraint {
		c := Constraint{Property: property, Operator: "GTE", Value: "0.5"}
		require.NoError(t, c.SetJevQuestion(&JevQuestion{Type: JevTypeScore, Instructions: "x", Criteria: []any{"a", "b"}}))
		return c
	}
	plain := Constraint{Property: "dl_state", Operator: "EQ", Value: `"CA"`}

	assert.Empty(t, DuplicateJevQuestionProperties([]Constraint{plain}))
	assert.Empty(t, DuplicateJevQuestionProperties([]Constraint{jev("@jev.a"), jev("@jev.b"), plain}))
	assert.Equal(t, []string{"@jev.risk", "@jev.risk2"}, DuplicateJevQuestionProperties([]Constraint{
		jev("@jev.risk"), jev("@jev.risk2"), jev("@jev.risk"), jev("@jev.risk2"),
	}))
}

func f64(v float64) *float64 { return &v }
