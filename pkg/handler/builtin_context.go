package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/openflagr/flagr/pkg/config"
)

// Built-in context key constants.
// @ prefix is used because the conditions library explicitly supports it as a
// variable prefix character, providing namespace isolation from client context.
const (
	BuiltInKeyTs      = "@ts"
	BuiltInKeyTsHour  = "@ts_hour"
	BuiltInKeyTsWeekday = "@ts_weekday"
	BuiltInKeyTsMonth = "@ts_month"
)

// httpHeaderPrefix is the prefix used for HTTP header context keys.
const httpHeaderPrefix = "@http_"

// InjectBuiltInContext enriches entityContext with server-side and HTTP request
// metadata. Core keys (@ts, @ts_hour, @ts_weekday, @ts_month) are always injected
// when enabled. HTTP headers matching the configured lists are injected as
// @http_* keys.
// Server-injected keys (@ts_* and @http_*) overwrite any client-provided values
func InjectBuiltInContext(entityContext any, r *http.Request) any {
	if !config.Config.InjectedContextEnabled {
		return entityContext
	}

	ctx, ok := entityContext.(map[string]any)
	if !ok {
		ctx = make(map[string]any)
	}

	now := time.Now().UTC()
	// All values are float64 to survive JSON round-tripping (int64 → float64).
	ctx[BuiltInKeyTs] = float64(now.Unix())
	ctx[BuiltInKeyTsHour] = float64(now.Hour())
	ctx[BuiltInKeyTsWeekday] = float64(now.Weekday())
	ctx[BuiltInKeyTsMonth] = float64(now.Month())

	if r != nil {
		injectHTTPHeaders(ctx, r)
	}

	return ctx
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
		key := httpHeaderPrefix + strings.ToLower(strings.ReplaceAll(name, "-", "_"))

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

