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
	Model     string                      `json:"model,omitempty"`
	State     any                         `json:"state,omitempty"`
	Questions map[string]JevDebugQuestion `json:"questions,omitempty"`
	Answers   map[string]JevAnswer        `json:"answers,omitempty"`
	Usage     *JevUsage                   `json:"usage,omitempty"`
	// LatencyMs is the client-measured round trip; ServerLatencyMs is the
	// endpoint-reported inference latency when provided. Omitted on cache hits.
	LatencyMs       float64  `json:"latencyMs,omitempty"`
	ServerLatencyMs *float64 `json:"serverLatencyMs,omitempty"`
	Cached          bool     `json:"cached,omitempty"`
	Error           string   `json:"error,omitempty"`
}

// JevDebugQuestion is the request-side view of one question.
type JevDebugQuestion struct {
	Type                string   `json:"type"`
	Instructions        any      `json:"instructions,omitempty"`
	Criteria            any      `json:"criteria,omitempty"`
	ConfidenceThreshold *float64 `json:"confidenceThreshold,omitempty"`
}

func newJevDebug(state any, questions map[string]entity.JevConstraintSpec) *JevDebug {
	debugQuestions := make(map[string]JevDebugQuestion, len(questions))
	for name, spec := range questions {
		debugQuestions[name] = JevDebugQuestion{
			Type:                spec.Type,
			Instructions:        spec.Instructions,
			Criteria:            spec.Criteria,
			ConfidenceThreshold: spec.ConfidenceThreshold,
		}
	}
	return &JevDebug{
		Model:     config.Config.JevModel,
		State:     state,
		Questions: debugQuestions,
	}
}

// injectJevContext resolves the flag's Jev questions and injects the answers
// into entityContext under the `@jev` key. The map is cloned so the injected
// values do not leak across flags in a batch evaluation.
//
// Fail-closed: when Jev is disabled or the call fails, nothing is injected and
// every `@jev.<name>` constraint evaluates false. The returned *JevDebug
// carries the request/response for the eval debug log (nil when Jev is off or
// the flag has no Jev questions).
func injectJevContext(evalContext models.EvalContext, flag *entity.Flag) (models.EvalContext, *JevDebug) {
	questions := flag.FlagEvaluation.JevQuestions
	if !config.Config.JevEnabled || len(questions) == 0 {
		return evalContext, nil
	}

	state := jevState(evalContext.EntityContext, evalContext.EntityID, evalContext.EntityType)
	debug := newJevDebug(state, questions)

	resp, cached, err := resolveJevAnswers(state, questions)
	debug.Cached = cached
	if err != nil {
		debug.Error = err.Error()
		logrus.WithError(err).WithField("flagID", flag.ID).
			Warn("jev evaluation failed; jev constraints will not match")
		return evalContext, debug
	}
	if resp != nil {
		if resp.Model != "" {
			debug.Model = resp.Model
		}
		debug.Answers = resp.Answers
		debug.Usage = resp.Usage
		if !cached {
			debug.LatencyMs = resp.ClientLatencyMs
			debug.ServerLatencyMs = resp.LatencyMs
		}
	}

	injected := make(map[string]any, len(resp.Answers))
	for name, answer := range resp.Answers {
		spec, ok := questions[name]
		if !ok {
			continue
		}
		if value, ok := jevAnswerValue(spec, answer); ok {
			injected[name] = value
		}
	}
	if len(injected) == 0 {
		return evalContext, debug
	}

	base, ok := evalContext.EntityContext.(map[string]any)
	if !ok {
		base = map[string]any{}
	}
	clone := make(map[string]any, len(base)+1)
	for k, v := range base {
		clone[k] = v
	}
	clone[entity.JevContextKey] = injected
	evalContext.EntityContext = clone
	return evalContext, debug
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

// resolveJevAnswers returns cached answers when possible, otherwise calls
// System One and caches the response. The bool reports a cache hit.
func resolveJevAnswers(state any, questions map[string]entity.JevConstraintSpec) (*JevResponse, bool, error) {
	key, err := jevCacheKey(state, questions)
	if err == nil {
		if cached, ok := GetJevCache().Get(key); ok {
			return cached, true, nil
		}
	}

	resp, err := NewJevClient().SystemOne(context.Background(), state, questions)
	if err != nil {
		return nil, false, err
	}
	if key != "" {
		GetJevCache().Set(key, resp)
	}
	return resp, false, nil
}

// jevAnswerValue converts an answer into the value injected for its question.
// choice and score answers below the confidence threshold are rejected, which
// makes the containing constraint fall through.
func jevAnswerValue(spec entity.JevConstraintSpec, answer JevAnswer) (any, bool) {
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

func jevConfidentEnough(spec entity.JevConstraintSpec, confidence *float64) bool {
	if confidence == nil {
		return false
	}
	threshold := config.Config.JevConfidenceThreshold
	if spec.ConfidenceThreshold != nil {
		threshold = *spec.ConfidenceThreshold
	}
	return *confidence >= threshold
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
	filtered.Questions = make(map[string]JevDebugQuestion, len(names))
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
