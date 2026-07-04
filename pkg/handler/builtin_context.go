package handler

import (
	"net/http"
	"strings"
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
// with the same name to prevent spoofing.
func InjectBuiltInContext(entityContext any, r *http.Request) any {
	if !config.Config.InjectedContextEnabled {
		return entityContext
	}

	ctx, ok := entityContext.(map[string]any)
	if !ok {
		ctx = make(map[string]any)
	}

	now := time.Now().UTC()
	ctx[BuiltInKeyTs] = now.Unix()
	ctx[BuiltInKeyTsHour] = now.Hour()
	ctx[BuiltInKeyTsWeekday] = int(now.Weekday())
	ctx[BuiltInKeyTsMonth] = int(now.Month())

	if r != nil {
		injectHTTPHeaders(ctx, r)
	}

	return ctx
}

// injectHTTPHeaders injects matching HTTP headers as @http_* context keys.
func injectHTTPHeaders(ctx map[string]any, r *http.Request) {
	exactHeaders := config.Config.InjectedContextHTTPHeaders
	prefixes := config.Config.InjectedContextHTTPHeaderPrefixes

	// Build normalized exact header set (case-insensitive lookup)
	exactSet := make(map[string]bool, len(exactHeaders))
	for _, h := range exactHeaders {
		exactSet[strings.ToLower(strings.TrimSpace(h))] = true
	}

	// Clone headers and inject Host if configured (Host is in r.Host, not r.Header)
	headers := r.Header.Clone()
	if _, ok := exactSet["host"]; ok && r.Host != "" {
		headers["Host"] = []string{r.Host}
	}

	for name, values := range headers {
		// Skip empty values
		if len(values) == 0 || (len(values) == 1 && values[0] == "") {
			continue
		}

		// Check exact match (case-insensitive)
		matched := exactSet[strings.ToLower(name)]

		// Check prefix match
		if !matched {
			for _, prefix := range prefixes {
				if strings.HasPrefix(name, prefix) {
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

