package handler

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
)

// Built-in context key constants.
// @ prefix is used because the conditions library explicitly supports it as a
// variable prefix character, providing namespace isolation from client context.
const (
	BuiltInKeyTs        = "@ts"
	BuiltInKeyTsHour    = "@ts_hour"
	BuiltInKeyTsWeekday = "@ts_weekday"
	BuiltInKeyTsMonth   = "@ts_month"
)

// httpHeaderPrefix is the prefix used for HTTP header context keys.
const httpHeaderPrefix = "@http_"

// tsProperties is the exact property set the ts enricher contributes.
var tsProperties = []string{
	BuiltInKeyTs,
	BuiltInKeyTsHour,
	BuiltInKeyTsWeekday,
	BuiltInKeyTsMonth,
}

// buildTsEnricher builds the server-time enricher. It contributes the bare @ts
// and the @ts_* derivations; values are float64 so they survive JSON
// round-trips. Gated by FLAGR_INJECTED_CONTEXT_ENABLED.
func buildTsEnricher(string) (*enricher, error) {
	return &enricher{
		namespace:  entity.EnricherNamespaceTs,
		scope:      scopeGlobal,
		prefix:     BuiltInKeyTs,
		properties: tsProperties,
		enabled:    config.Config.InjectedContextEnabled,
		run: func(in enrichInput) (map[string]any, error) {
			now := time.Now().UTC()
			return map[string]any{
				BuiltInKeyTs:        float64(now.Unix()),
				BuiltInKeyTsHour:    float64(now.Hour()),
				BuiltInKeyTsWeekday: float64(now.Weekday()),
				BuiltInKeyTsMonth:   float64(now.Month()),
			}, nil
		},
	}, nil
}

// buildHTTPEnricher builds the request-header enricher. It contributes
// @http_<header> keys for the headers allowed by
// FLAGR_INJECTED_CONTEXT_HTTP_HEADERS / _PREFIXES.
func buildHTTPEnricher(string) (*enricher, error) {
	return &enricher{
		namespace:  entity.EnricherNamespaceHTTP,
		scope:      scopeGlobal,
		prefix:     httpHeaderPrefix,
		properties: httpEnricherProperties(),
		enabled:    config.Config.InjectedContextEnabled,
		run: func(in enrichInput) (map[string]any, error) {
			if in.request == nil {
				return nil, nil
			}
			out := make(map[string]any)
			injectHTTPHeaders(out, in.request)
			return out, nil
		},
	}, nil
}

// httpEnricherProperties lists the exact configured header properties for the
// picker. Prefix-matched headers are intentionally omitted (request-dependent).
// It reads config directly rather than getHeaderMatchSets so it never primes
// the sync.Once match cache before a request is available.
func httpEnricherProperties() []string {
	props := make([]string, 0, len(config.Config.InjectedContextHTTPHeaders))
	seen := map[string]bool{}
	for _, header := range config.Config.InjectedContextHTTPHeaders {
		trimmed := strings.TrimSpace(header)
		if trimmed == "" {
			continue
		}
		p := httpHeaderPrefix + normalizeHeaderKey(trimmed)
		if seen[p] {
			continue
		}
		seen[p] = true
		props = append(props, p)
	}
	sort.Strings(props)
	return props
}

// normalizeHeaderKey maps a header name to its context key fragment: lowercase
// with "-" replaced by "_".
func normalizeHeaderKey(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "-", "_"))
}

// injectHTTPHeaders injects matching HTTP headers as @http_* context keys.
func injectHTTPHeaders(ctx map[string]any, r *http.Request) {
	exactSet, prefixSet := getHeaderMatchSets()

	// Check Host separately (it's in r.Host, not r.Header)
	if _, ok := exactSet["host"]; ok && r.Host != "" {
		ctx[httpHeaderPrefix+"host"] = r.Host
	}

	// Iterate r.Header directly — no Clone needed
	for name, values := range r.Header {
		// Skip empty values
		if len(values) == 0 || (len(values) == 1 && values[0] == "") {
			continue
		}

		lower := strings.ToLower(name)

		// Check exact match (case-insensitive)
		matched := exactSet[lower]

		// Check prefix match (also case-insensitive)
		if !matched {
			for prefix := range prefixSet {
				if strings.HasPrefix(lower, prefix) {
					matched = true
					break
				}
			}
		}

		if !matched {
			continue
		}

		// Build context key: lowercase, replace - with _, prefix @http_
		key := httpHeaderPrefix + normalizeHeaderKey(name)

		// Join multi-value headers with ", "
		if len(values) == 1 {
			ctx[key] = values[0]
		} else {
			ctx[key] = strings.Join(values, ", ")
		}
	}
}

var (
	headerMatchOnce sync.Once
	headerExactSet  map[string]bool
	headerPrefixSet map[string]bool
)

// getHeaderMatchSets returns cached, normalized header matching sets.
// Exact set uses lowercase keys for case-insensitive lookup.
// Prefix set uses lowercase keys for case-insensitive prefix matching.
func getHeaderMatchSets() (map[string]bool, map[string]bool) {
	headerMatchOnce.Do(func() {
		exactHeaders := config.Config.InjectedContextHTTPHeaders
		prefixes := config.Config.InjectedContextHTTPHeaderPrefixes

		headerExactSet = make(map[string]bool, len(exactHeaders))
		for _, h := range exactHeaders {
			if trimmed := strings.TrimSpace(h); trimmed != "" {
				headerExactSet[strings.ToLower(trimmed)] = true
			}
		}

		headerPrefixSet = make(map[string]bool, len(prefixes))
		for _, p := range prefixes {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				headerPrefixSet[strings.ToLower(trimmed)] = true
			}
		}
	})
	return headerExactSet, headerPrefixSet
}

// ResetHeaderMatchCache resets the cached header match sets.
// Use only in tests that change config values between calls.
func ResetHeaderMatchCache() {
	headerMatchOnce = sync.Once{}
	headerExactSet = nil
	headerPrefixSet = nil
}
