package handler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrFloat(v float64) *float64 { return &v }

// manyChoiceOptions returns one more option than the System One cap.
func manyChoiceOptions() map[string]any {
	m := make(map[string]any, JevMaxChoiceOptions+1)
	for i := 0; i <= JevMaxChoiceOptions; i++ {
		m[fmt.Sprintf("option_%d", i)] = "desc"
	}
	return m
}

func validJevConfig() *JevEnricherConfig {
	return &JevEnricherConfig{
		Questions: map[string]JevQuestion{
			"plan_tier": {
				Type:         JevTypeChoice,
				Instructions: "Which plan is this account on?",
				Criteria:     map[string]any{"free": "Free", "pro": "Pro"},
			},
			"churn_risk": {
				Type:                JevTypeScore,
				Instructions:        "How likely is this account to churn?",
				Criteria:            []any{"very low", "low", "high"},
				ConfidenceThreshold: ptrFloat(0.7),
			},
		},
	}
}

func TestJevPropertyHelpers(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "@jev_plan_tier", JevProperty("plan_tier"))
	assert.True(t, IsJevProperty("@jev_plan_tier"))
	assert.False(t, IsJevProperty("@ts_hour"))
	assert.Equal(t, "plan_tier", JevQuestionName("@jev_plan_tier"))
	assert.Equal(t, "", JevQuestionName("@ts_hour"))
}

func TestJevEnricherConfigProperties(t *testing.T) {
	t.Parallel()

	cfg := validJevConfig()
	assert.Equal(t, []string{"churn_risk", "plan_tier"}, cfg.QuestionNames())
	assert.Equal(t, []string{"@jev_churn_risk", "@jev_plan_tier"}, cfg.Properties())
}

func TestDecodeJevConfig(t *testing.T) {
	t.Parallel()

	cfg, err := DecodeJevConfig(`{"questions":{"q":{"type":"noul","instructions":"x"}}}`)
	require.NoError(t, err)
	assert.Equal(t, JevTypeNoul, cfg.Questions["q"].Type)

	_, err = DecodeJevConfig("")
	assert.Error(t, err)
	_, err = DecodeJevConfig("not json")
	assert.Error(t, err)
	_, err = DecodeJevConfig(`{"questions":{}}`)
	assert.Error(t, err, "a jev enricher must ask at least one question")
}

func TestJevEnricherConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *JevEnricherConfig
		wantErr bool
	}{
		{name: "valid", cfg: validJevConfig()},
		{name: "nil", cfg: nil, wantErr: true},
		{name: "empty questions", cfg: &JevEnricherConfig{}, wantErr: true},
		{
			name: "bad question name",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"has space": {Type: JevTypeNoul, Instructions: "x"},
			}},
			wantErr: true,
		},
		{
			name: "missing instructions",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"q": {Type: JevTypeNoul},
			}},
			wantErr: true,
		},
		{
			name: "bad type",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"q": {Type: "bogus", Instructions: "x"},
			}},
			wantErr: true,
		},
		{
			name: "choice criteria not object",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"q": {Type: JevTypeChoice, Instructions: "x", Criteria: []any{"a"}},
			}},
			wantErr: true,
		},
		{
			name: "score criteria too few levels",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"q": {Type: JevTypeScore, Instructions: "x", Criteria: []any{"only one"}},
			}},
			wantErr: true,
		},
		{
			name: "threshold out of range",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"q": {Type: JevTypeNoul, Instructions: "x", ConfidenceThreshold: ptrFloat(1.5)},
			}},
			wantErr: true,
		},
		{
			// noul's answer is already a 0-1 probability; a confidence gate on top of
			// it would be silently ignored, so it is rejected instead.
			name: "noul threshold unsupported",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"q": {Type: JevTypeNoul, Instructions: "x", ConfidenceThreshold: ptrFloat(0.5)},
			}},
			wantErr: true,
		},
		{
			name: "choice threshold allowed",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"q": {
					Type:                JevTypeChoice,
					Instructions:        "x",
					Criteria:            map[string]any{"a": "A"},
					ConfidenceThreshold: ptrFloat(0.5),
				},
			}},
		},
		{
			name: "noul criteria true/false",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"q": {
					Type:         JevTypeNoul,
					Instructions: "x",
					Criteria:     map[string]any{"true": "about billing", "false": "anything else"},
				},
			}},
		},
		{
			name: "noul criteria bad key",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"q": {Type: JevTypeNoul, Instructions: "x", Criteria: map[string]any{"maybe": "?"}},
			}},
			wantErr: true,
		},
		{
			name: "choice too many options",
			cfg: &JevEnricherConfig{Questions: map[string]JevQuestion{
				"q": {Type: JevTypeChoice, Instructions: "x", Criteria: manyChoiceOptions()},
			}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJevAnswerValue(t *testing.T) {
	t.Parallel()

	threshold := 0.8
	tests := []struct {
		name     string
		spec     JevQuestion
		answer   JevAnswer
		wantVal  any
		wantKeep bool
	}{
		{name: "noul", spec: JevQuestion{Type: JevTypeNoul}, answer: JevAnswer{Noul: ptrFloat(0.9)}, wantVal: 0.9, wantKeep: true},
		{name: "noul missing", spec: JevQuestion{Type: JevTypeNoul}, wantKeep: false},
		{name: "choice", spec: JevQuestion{Type: JevTypeChoice}, answer: JevAnswer{Choice: "pro"}, wantVal: "pro", wantKeep: true},
		{name: "choice empty", spec: JevQuestion{Type: JevTypeChoice}, wantKeep: false},
		{
			name:     "choice below threshold",
			spec:     JevQuestion{Type: JevTypeChoice, ConfidenceThreshold: &threshold},
			answer:   JevAnswer{Choice: "pro", Confidence: ptrFloat(0.5)},
			wantKeep: false,
		},
		{
			name:     "choice above threshold",
			spec:     JevQuestion{Type: JevTypeChoice, ConfidenceThreshold: &threshold},
			answer:   JevAnswer{Choice: "pro", Confidence: ptrFloat(0.9)},
			wantVal:  "pro",
			wantKeep: true,
		},
		{name: "score", spec: JevQuestion{Type: JevTypeScore}, answer: JevAnswer{Score: ptrFloat(3)}, wantVal: 3.0, wantKeep: true},
		{name: "unknown type", spec: JevQuestion{Type: "bogus"}, wantKeep: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val, keep := jevAnswerValue(tt.spec, tt.answer)
			assert.Equal(t, tt.wantKeep, keep)
			if tt.wantKeep {
				assert.Equal(t, tt.wantVal, val)
			}
		})
	}
}
