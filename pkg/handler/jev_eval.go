package handler

import (
	"context"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/sirupsen/logrus"
)

// injectJevContext resolves the flag's Jev questions and injects the answers
// into entityContext under the `@jev` key. The map is cloned so the injected
// values do not leak across flags in a batch evaluation.
//
// Fail-closed: when Jev is disabled or the call fails, nothing is injected and
// every `@jev.<name>` constraint evaluates false.
func injectJevContext(evalContext models.EvalContext, flag *entity.Flag) models.EvalContext {
	questions := flag.FlagEvaluation.JevQuestions
	if !config.Config.JevEnabled || len(questions) == 0 {
		return evalContext
	}

	state := jevState(evalContext.EntityContext)
	answers, err := resolveJevAnswers(state, questions)
	if err != nil {
		logrus.WithError(err).WithField("flagID", flag.ID).
			Warn("jev evaluation failed; jev constraints will not match")
		return evalContext
	}

	injected := make(map[string]any, len(answers))
	for name, answer := range answers {
		spec, ok := questions[name]
		if !ok {
			continue
		}
		if value, ok := jevAnswerValue(spec, answer); ok {
			injected[name] = value
		}
	}
	if len(injected) == 0 {
		return evalContext
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
	return evalContext
}

// jevState drops Flagr's own `@jev` answer namespace so answers are never fed
// back into the model, and returns the rest of entityContext (including
// server-injected `@ts*` and `@http_*` keys) as the System One state.
func jevState(entityContext any) any {
	m, ok := entityContext.(map[string]any)
	if !ok {
		return entityContext
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		if k == entity.JevContextKey {
			continue
		}
		out[k] = v
	}
	return out
}

// resolveJevAnswers returns cached answers when possible, otherwise calls
// System One and caches the result.
func resolveJevAnswers(state any, questions map[string]entity.JevConstraintSpec) (map[string]JevAnswer, error) {
	key, err := jevCacheKey(state, questions)
	if err == nil {
		if cached, ok := GetJevCache().Get(key); ok {
			return cached, nil
		}
	}

	resp, err := NewJevClient().SystemOne(context.Background(), state, questions)
	if err != nil {
		return nil, err
	}
	if key != "" {
		GetJevCache().Set(key, resp.Answers)
	}
	return resp.Answers, nil
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

// jevDebugAnswers returns the injected answers for debug logging, or nil.
func jevDebugAnswers(evalContext models.EvalContext) any {
	m, ok := evalContext.EntityContext.(map[string]any)
	if !ok {
		return nil
	}
	return m[entity.JevContextKey]
}
