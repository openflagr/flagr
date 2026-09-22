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
			require.NoError(t, c.SetJevQuestion(tc.question))
			err := c.Validate()
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
	named := Constraint{Property: "@jev.risk", JevType: JevTypeNoul}
	assert.Equal(t, "risk", named.JevName())

	q, err := c.JevQuestion()
	require.NoError(t, err)
	require.NotNil(t, q)
	assert.Equal(t, JevTypeScore, q.Type)
	assert.Equal(t, map[string]any{"question": "How relevant?"}, q.Instructions)
	assert.Equal(t, []any{"none", "some", "deep"}, q.Criteria)
	require.NotNil(t, q.ConfidenceThreshold)
	assert.Equal(t, 0.4, *q.ConfidenceThreshold)

	// Clearing removes all Jev fields.
	require.NoError(t, c.SetJevQuestion(nil))
	assert.False(t, c.IsJev())
	assert.Empty(t, c.JevType)
	assert.Nil(t, c.JevConfidenceThreshold)
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

func f64(v float64) *float64 { return &v }
