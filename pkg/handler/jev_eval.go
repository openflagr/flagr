package handler

import (
	"context"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/sirupsen/logrus"
)

// Jev entity identity keys added to the System One state.
const (
	JevEntityIDKey   = "entityID"
	JevEntityTypeKey = "entityType"
)

// JevDebug captures the System One request and response for the eval debug log.
type JevDebug struct {
	Model     string                        `json:"model,omitempty"`
	State     any                           `json:"state,omitempty"`
	Questions map[string]entity.JevQuestion `json:"questions,omitempty"`
	Answers   map[string]JevAnswer          `json:"answers,omitempty"`
	Usage     *JevUsage                     `json:"usage,omitempty"`
	// LatencyMs is the client-measured round trip; ServerLatencyMs is the
	// endpoint-reported inference latency when provided.
	LatencyMs       float64  `json:"latencyMs,omitempty"`
	ServerLatencyMs *float64 `json:"serverLatencyMs,omitempty"`
	Retries         int      `json:"retries,omitempty"`
	Error           string   `json:"error,omitempty"`
}

func newJevDebug(state any, questions map[string]entity.JevQuestion) *JevDebug {
	return &JevDebug{
		Model:     config.Config.JevModel,
		State:     state,
		Questions: questions,
	}
}

// resolveJevForFlag evaluates the flag's Jev questions in one batched System One
// call and returns the values to compare against, plus the request/response for
// the debug log.
//
// It deliberately does not modify evalContext: the answers are used only while
// evaluating constraints and are never written back into the result context.
//
// Fail-closed: when Jev is disabled or the call fails, the returned map is nil
// and every `@jev.<name>` constraint evaluates false.
func resolveJevForFlag(ctx context.Context, evalContext models.EvalContext, flag *entity.Flag) (map[string]any, *JevDebug) {
	questions := flag.FlagEvaluation.JevQuestions
	if !config.Config.JevEnabled || len(questions) == 0 {
		return nil, nil
	}

	state := jevState(evalContext.EntityContext, evalContext.EntityID, evalContext.EntityType)
	debug := newJevDebug(state, questions)

	resp, err := NewJevClient().SystemOne(ctx, state, questions)
	if err != nil {
		debug.Error = err.Error()
		logrus.WithError(err).WithField("flagID", flag.ID).
			Warn("jev evaluation failed; jev constraints will not match")
		return nil, debug
	}
	if resp.Model != "" {
		debug.Model = resp.Model
	}
	debug.Answers = resp.Answers
	debug.Usage = resp.Usage
	debug.LatencyMs = resp.ClientLatencyMs
	debug.ServerLatencyMs = resp.LatencyMs
	debug.Retries = resp.Retries

	answers := make(map[string]any, len(resp.Answers))
	for name, answer := range resp.Answers {
		spec, ok := questions[name]
		if !ok {
			continue
		}
		if value, ok := jevAnswerValue(spec, answer); ok {
			answers[name] = value
		}
	}
	if len(answers) == 0 {
		return nil, debug
	}
	return answers, debug
}

// jevState builds the System One state from entityContext. It drops Flagr's own
// `@jev` answer namespace (so answers are never fed back to the model) and adds
// the entity identity, which is not part of entityContext. Server-injected
// `@ts*` and `@http_*` keys are kept.
func jevState(entityContext any, entityID, entityType string) any {
	m, ok := entityContext.(map[string]any)
	if !ok {
		m = map[string]any{}
	}
	out := make(map[string]any, len(m)+2)
	for k, v := range m {
		if k == entity.JevContextKey {
			continue
		}
		out[k] = v
	}
	// The canonical entity identity wins over any same-named entityContext key.
	if entityID != "" {
		out[JevEntityIDKey] = entityID
	}
	if entityType != "" {
		out[JevEntityTypeKey] = entityType
	}
	return out
}

// jevAnswerValue converts an answer into the value used for its question.
// choice and score answers below the confidence threshold are rejected, which
// makes the containing constraint fall through.
func jevAnswerValue(spec entity.JevQuestion, answer JevAnswer) (any, bool) {
	switch spec.Type {
	case entity.JevTypeNoul:
		if answer.Noul == nil {
			return nil, false
		}
		return *answer.Noul, true
	case entity.JevTypeChoice:
		if answer.Choice == "" || !jevConfidentEnough(spec, answer.Confidence) {
			return nil, false
		}
		return answer.Choice, true
	case entity.JevTypeScore:
		if answer.Score == nil || !jevConfidentEnough(spec, answer.Confidence) {
			return nil, false
		}
		return *answer.Score, true
	default:
		return nil, false
	}
}

func jevConfidentEnough(spec entity.JevQuestion, confidence *float64) bool {
	if confidence == nil {
		return false
	}
	threshold := config.Config.JevConfidenceThreshold
	if spec.ConfidenceThreshold != nil {
		threshold = *spec.ConfidenceThreshold
	}
	return *confidence >= threshold
}

// withJevContext returns entityContext with the resolved Jev answers merged
// under `@jev`, so constraints can reference them without mutating the caller's
// context. The result context (and data records) stay clean.
func withJevContext(entityContext any, jev map[string]any) map[string]any {
	m, _ := entityContext.(map[string]any)
	merged := make(map[string]any, len(m)+1)
	for k, v := range m {
		merged[k] = v
	}
	merged[entity.JevContextKey] = jev
	return merged
}

// segmentJevQuestionNames returns the `@jev.<name>` questions a segment uses.
func segmentJevQuestionNames(segment entity.Segment) []string {
	names := make([]string, 0, len(segment.Constraints))
	for i := range segment.Constraints {
		if segment.Constraints[i].IsJev() {
			names = append(names, segment.Constraints[i].JevName())
		}
	}
	return names
}

// filterJevDebug narrows a flag-level Jev debug payload to the questions a
// segment actually uses. It returns nil when the segment has no Jev
// constraints, so non-Jev segments do not carry a Jev entry in their debug log.
func filterJevDebug(debug *JevDebug, names []string) *JevDebug {
	if debug == nil || len(names) == 0 {
		return nil
	}
	filtered := *debug
	filtered.Questions = make(map[string]entity.JevQuestion, len(names))
	filtered.Answers = make(map[string]JevAnswer, len(names))
	for _, name := range names {
		if q, ok := debug.Questions[name]; ok {
			filtered.Questions[name] = q
		}
		if a, ok := debug.Answers[name]; ok {
			filtered.Answers[name] = a
		}
	}
	return &filtered
}
