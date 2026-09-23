package handler

import (
	"context"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/sirupsen/logrus"
)

// Entity identity keys added to the System One state. They are not part of
// entityContext, so they are attached explicitly.
const (
	jevEntityIDKey   = "entityID"
	jevEntityTypeKey = "entityType"
)

// buildJevEnricher resolves a flag's `jev` config into a runtime enricher backed
// by one batched System One call. The whole accumulated evaluation context
// (including @ts*/@http_*) is the model state; only the enricher's own `@jev_`
// properties are masked out by the pipeline.
func buildJevEnricher(configJSON string) (*enricher, error) {
	cfg, err := DecodeJevConfig(configJSON)
	if err != nil {
		return nil, err
	}

	return &enricher{
		namespace:  entity.EnricherNamespaceJev,
		scope:      scopeFlag,
		prefix:     JevPropertyPrefix,
		properties: cfg.Properties(),
		enabled:    config.Config.InjectedContextJevBaseURL != "",
		run: func(in enrichInput) (map[string]any, error) {
			if len(cfg.Questions) == 0 {
				return nil, nil
			}
			ctx := in.ctx
			if ctx == nil {
				ctx = context.Background()
			}
			state := jevState(in.context, in.entityID, in.entityType)
			resp, err := NewJevClient().SystemOne(ctx, state, cfg.Questions)
			if err != nil {
				logrus.WithError(err).
					Warn("jev enricher failed; its properties will be absent")
				return nil, err
			}
			out := make(map[string]any, len(resp.Answers))
			for name, answer := range resp.Answers {
				spec, ok := cfg.Questions[name]
				if !ok {
					continue
				}
				if value, ok := jevAnswerValue(spec, answer); ok {
					out[JevProperty(name)] = value
				}
			}
			if len(out) == 0 {
				return nil, nil
			}
			return out, nil
		},
	}, nil
}

// jevState builds the System One state: the accumulated context plus entity
// identity (canonical identity wins). Any `@jev_` property is dropped
// defensively so answers are never fed back to the model.
func jevState(in map[string]any, entityID, entityType string) map[string]any {
	out := make(map[string]any, len(in)+2)
	for k, v := range in {
		if IsJevProperty(k) {
			continue
		}
		out[k] = v
	}
	if entityID != "" {
		out[jevEntityIDKey] = entityID
	}
	if entityType != "" {
		out[jevEntityTypeKey] = entityType
	}
	return out
}
