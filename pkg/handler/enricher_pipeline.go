package handler

import (
	"net/http"
	"strings"

	"github.com/openflagr/flagr/pkg/entity"
)

// enrichContext runs enrichers in order, each seeing the accumulated context
// with its own property prefix masked out. Enricher output overwrites any
// client-provided key of the same name. A disabled enricher is skipped, and a
// failing enricher contributes nothing (fail-closed) while the remaining
// enrichers still run.
func enrichContext(entries []enricher, base map[string]any, in enrichInput) map[string]any {
	if len(entries) == 0 {
		return base
	}
	if base == nil {
		base = make(map[string]any)
	}
	for i := range entries {
		e := &entries[i]
		if !e.enabled {
			continue
		}
		input := in
		input.context = maskPropertyPrefix(base, e.prefix)
		out, err := e.run(input)
		if err != nil {
			continue
		}
		for k, v := range out {
			base[k] = v
		}
	}
	return base
}

// maskPropertyPrefix returns a shallow copy of ctx without keys carrying the
// given property prefix, so an enricher never sees its own prior output.
func maskPropertyPrefix(ctx map[string]any, prefix string) map[string]any {
	if prefix == "" {
		return ctx
	}
	masked := make(map[string]any, len(ctx))
	for k, v := range ctx {
		if strings.HasPrefix(k, prefix) {
			continue
		}
		masked[k] = v
	}
	return masked
}

// InjectBuiltInContext runs the enabled global built-in enrichers (ts, http)
// over the caller's entityContext. It is the request-boundary entry point,
// where the HTTP request is available. Returns the input unchanged when no
// global enricher is enabled, preserving the historical no-op behavior.
func InjectBuiltInContext(entityContext any, r *http.Request) any {
	entries := enabledEnrichers(globalEnrichers())
	if len(entries) == 0 {
		return entityContext
	}
	return enrichContext(entries, toContextMap(entityContext), enrichInput{request: r})
}

// enrichFlagContext runs the flag's enabled flag-scoped enrichers over the
// (already global-enriched) entityContext. It is a no-op when the flag declares
// none or all are disabled, so plain flags pay nothing.
//
// It copies the context before writing: the caller's map is shared across flags
// in batch and tag evaluation, so writing in place would leak one flag's
// enriched properties into the next flag.
func enrichFlagContext(entityContext any, flag *entity.Flag, entityID, entityType string) any {
	entries := enabledEnrichers(flagEnrichers(flag))
	if len(entries) == 0 {
		return entityContext
	}
	return enrichContext(entries, copyContext(entityContext), enrichInput{
		entityID:   entityID,
		entityType: entityType,
	})
}

// toContextMap returns v as an entityContext map, allocating one when the
// caller sent nothing usable. It returns the caller's map, so callers that must
// not write through to shared state should copyContext instead.
func toContextMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok && m != nil {
		return m
	}
	return make(map[string]any)
}

// copyContext returns a shallow copy of v as an entityContext map.
func copyContext(v any) map[string]any {
	src, ok := v.(map[string]any)
	if !ok || src == nil {
		return make(map[string]any)
	}
	dst := make(map[string]any, len(src))
	for k, val := range src {
		dst[k] = val
	}
	return dst
}
