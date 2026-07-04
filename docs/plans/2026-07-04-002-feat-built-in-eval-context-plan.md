# feat: Built-in Evaluation Context — Server-Side & HTTP Request Context Injection

**Date:** 2026-07-04
**Status:** as-built

## Summary

Inject server-side and HTTP request metadata into `entityContext` as built-in context keys, enabling time-based scheduling and request-aware targeting as ordinary constraint comparisons. Core keys use `@ts` prefix (UTC time), configurable `@http_`-prefixed header keys. Naming follows nginx conventions adapted for the conditions library.

## Problem Frame

Flagr's `entityContext` is 100% client-provided. The eval engine does `map[key]` lookup against it. This means:
- **No time-based targeting** — can't schedule "enable after July 15" without a scheduler goroutine
- **No deployment awareness** — can't scope flags by staging/production without a separate env model
- **No request metadata** — can't target by IP, host, or proxy-injected headers

The conventional approach (separate data models, scheduler infrastructure) is heavy. Instead, built-in keys are additional entries in the `entityContext` map, evaluated by the same `conditions.Evaluate(expr, map)` path.

## Key Technical Decisions

1. **No `$` prefix** — The conditions library (`github.com/zhouzhuojie/conditions`) does not support `$` in variable names. Keys use plain names: `@ts`, `@ts_hour`, `http_x_environment`.

2. **Two key families** — Core keys use `@ts` prefix (`@ts`, `@ts_hour`, `@ts_weekday`, `@ts_month`). Header-injected keys use `@http_` prefix (`http_x_environment`). Client-provided keys have no prefix.

3. **Server overrides client** — Built-in keys overwrite client-provided values with the same name to prevent spoofing.

4. **Single instance, multi-environment** — One Flagr instance, one database. Differentiation via HTTP headers (`http_host`, `http_x_environment`). No separate deployments.

5. **`params.HTTPRequest` already available** — go-swagger generates `PostEvaluationParams` with `HTTPRequest *http.Request`. Zero handler registration changes needed.

6. **Numeric time values** — `@ts` is Unix epoch seconds (int64). The conditions library's `>=`/`<=` operators only work with numbers, not strings.

## Configuration

```yaml
# Global kill switch — disables ALL injection (default: false)
FLAGR_INJECTED_CONTEXT_ENABLED: false

# HTTP headers to expose as http_* context keys (comma-separated)
FLAGR_INJECTED_CONTEXT_HTTP_HEADERS: ""

# Header prefixes to auto-inject (comma-separated)
FLAGR_INJECTED_CONTEXT_HTTP_HEADER_PREFIXES: ""
```

**Always injected (not configurable):** `@ts`, `@ts_hour`, `@ts_weekday`, `@ts_month`

## Scope Boundaries

**In scope:**
- Built-in context injection in eval handler
- Core keys: `@ts`, `@ts_hour`, `@ts_weekday`, `@ts_month`
- Header keys: configurable via `HTTP_HEADERS` and `HTTP_HEADER_PREFIXES`
- Unit tests and integration tests
- UI: `ts_*` value hints showing human-readable UTC datetime

**Out of scope:**
- Environment as a database entity
- API key → environment mapping
- Identity/traits persistence
- Eval cache changes
- SDK changes

**Deferred:**
- `flagr_namespace` / `flagr_version` — add later if multi-region needed
- GeoIP lookup from `http_cf_connecting_ip`

---

## Built-in Context Keys — Complete Reference

### Core Keys (always present, `@ts` prefix, all UTC)

| Key | Source | Description | Sample Values | Use Cases | Sample Constraints |
|---|---|---|---|---|---|
| `@ts` | `time.Now().UTC().Unix()` | Current server time as Unix epoch seconds (int64). Updated per-request. | `1751666400`, `1752537600` | Schedule activation/deprecation; active time windows | `{@ts} GTE 1752537600` |
| `@ts_hour` | `time.Now().UTC().Hour()` | Hour of day in UTC (0–23). Derived from `@ts`. | `0`, `9`, `14`, `23` | Business hours targeting | `{@ts_hour} GTE 9 AND {@ts_hour} LT 17` |
| `@ts_weekday` | `time.Now().UTC().Weekday()` | Day of week in UTC (0=Sunday, 6=Saturday). Derived from `@ts`. | `0`, `1`, `5`, `6` | Weekday/weekend targeting | `{@ts_weekday} GTE 1 AND {@ts_weekday} LE 5` |
| `@ts_month` | `time.Now().UTC().Month()` | Month in UTC (1–12). Derived from `@ts`. | `1`, `7`, `12` | Seasonal/holiday features | `{@ts_month} EQ 12` |

### Header-Injected Keys (`@http_` prefix, configurable)

**Naming rule:** Header `X-Foo-Bar` → context key `http_x_foo_bar` (lowercase, `-` → `_`, prefix `@http_`). Config values are case-insensitive.

| Key | Source Header | Description | Sample Values | Use Cases | Sample Constraints |
|---|---|---|---|---|---|
| `http_host` | `Host` | HTTP Host header value. Special case: read from `r.Host`, not `r.Header`. Add `Host` to `HTTP_HEADERS` to enable. | `flagr.staging.internal`, `flagr.prod.example.com` | Single-instance staging/prod via hostname | `{@http_host} CONTAINS "staging"` |
| `http_x_environment` | `X-Environment` | Environment name injected by proxy/LB. | `staging`, `production`, `canary` | Scope flags to environment | `{@http_x_environment} EQ "production"` |
| `http_x_tenant_id` | `X-Tenant-ID` | Tenant identifier for multi-tenant deployments. | `acme-corp`, `tenant-42` | Tenant-specific feature gating | `{@http_x_tenant_id} EQ "acme-corp"` |
| `http_cf_ipcountry` | `CF-IPCountry` | Two-letter country code from Cloudflare. | `US`, `CN`, `JP`, `DE` | Country-based targeting | `{@http_cf_ipcountry} IN "US,CA,GB"` |
| `http_cf_connecting_ip` | `CF-Connecting-IP` | Real client IP from Cloudflare. | `203.0.113.50` | Specific client IP targeting | `{@http_cf_connecting_ip} EQ "203.0.113.50"` |
| `http_x_region` | `X-Region` | Deployment region. | `us-east-1`, `eu-west-1` | Region-specific features | `{@http_x_region} EQ "us-east-1"` |
| `http_x_canary` | `X-Canary` | Canary deployment flag. | `true`, `false` | Canary-only features | `{@http_x_canary} EQ "true"` |
| `http_x_user_tier` | `X-User-Tier` | User tier from auth proxy. | `free`, `premium`, `enterprise` | Premium feature gating | `{@http_x_user_tier} IN "premium,enterprise"` |

---

## Configuration Examples

**Minimal (core keys only):**
```yaml
FLAGR_INJECTED_CONTEXT_ENABLED: false
```
→ `@ts`, `@ts_hour`, `@ts_weekday`, `@ts_month` only (when enabled).

**Single-instance staging/prod:**
```yaml
FLAGR_INJECTED_CONTEXT_ENABLED: true
FLAGR_INJECTED_CONTEXT_HTTP_HEADERS: "X-Environment"
```
→ Core keys + `http_x_environment`.

**Cloudflare deployment:**
```yaml
FLAGR_INJECTED_CONTEXT_ENABLED: true
FLAGR_INJECTED_CONTEXT_HTTP_HEADERS: "X-Environment,X-Tenant-ID"
FLAGR_INJECTED_CONTEXT_HTTP_HEADER_PREFIXES: "CF-"
```
→ Core keys + `http_x_environment`, `http_x_tenant_id`, `http_cf_ipcountry`, `http_cf_connecting_ip`.

---

## Implementation

### Files Created/Modified

| File | Change |
|---|---|
| `pkg/config/env.go` | Added `InjectedContextEnabled`, `InjectedContextHTTPHeaders`, `InjectedContextHTTPHeaderPrefixes` config vars |
| `pkg/handler/builtin_context.go` | **New** — `InjectBuiltInContext()` function, key constants, header injection logic |
| `pkg/handler/builtin_context_test.go` | **New** — 12 unit tests for injection logic |
| `pkg/handler/builtin_context_integration_test.go` | **New** — 10 integration tests for eval with built-in context |
| `pkg/handler/eval.go` | Modified `PostEvaluation` and `PostEvaluationBatch` to call `InjectBuiltInContext` |

### Key Implementation Details

- `InjectBuiltInContext(entityContext any, r *http.Request) map[string]any` — main entry point
- `toMap(entityContext any) map[string]any` — safely converts `any` to map
- `injectHTTPHeaders(ctx map[string]any, r *http.Request)` — header injection with exact/prefix matching
- Config values normalized to lowercase for case-insensitive matching
- Empty header values skipped
- Multi-value headers joined with `, `
- `Host` header special-cased via `r.Host` (not in `r.Header`)

---

## Performance Impact

Benchmarked on Linux arm64 (Go 1.26.3), realistic HTTP request with 15 headers:

| Scenario | ns/op | vs 328µs p50 |
|---|---|---|
| Baseline (no injection) | 23 | — |
| Core keys only | 376 | +0.11% |
| Full injection (15 headers) | 3,976 | +1.2% |

**Full injection adds ~4µs to a 328µs eval — 1.2% overhead. Negligible.**

---

## Test Results

- 12 unit tests: all pass
- 10 integration tests: all pass
- Existing eval tests: unchanged, all pass

## Sources & Research

- nginx variables: https://nginx.org/en/docs/http/ngx_http_core_module.html#variables
- Conditions library: `github.com/zhouzhuojie/conditions` — does not support `$` in variable names
- Flagr eval engine: `pkg/handler/eval.go`
- Handler registration: `params.HTTPRequest` already in swagger params
