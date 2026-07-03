# Plan: EvalCache Query Filtering

**Issue:** [#628](https://github.com/openflagr/flagr/issues/628)
**Branch:** `feat/eval-cache-query-filter`
**Date:** 2026-07-03

## Goal

Add query parameters to `GET /api/v1/export/eval_cache/json` so users can filter flags from the in-memory EvalCache without hitting the database. Also expose the export endpoint in eval-only mode (replicas).

## Motivation

Flag DBs grow over time (deprecated, disabled, test flags). The export API currently dumps everything, causing network bloat and client-side filtering. EvalCache filtering provides a lightweight alternative that works on replicas (eval-only mode) too.

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Array format | CSV (`ids=1,2`) | Consistent with existing `FindFlags` and PR #630 |
| Filter logic | AND across all filter groups | Consistent with `FindFlags`; `ids`/`keys` are OR within their own group |
| Tag semantics | ANY by default, `all=true` for ALL | Matches PR #630; ANY is more useful default |
| `enabled` param | Optional | Omit = return both enabled and disabled |
| Implementation | Filter at export time (not in EvalCache) | Simple, no cache changes, negligible perf cost |
| No match result | 200 with empty `Flags: []` | Consistent with `FindFlags` |
| Invalid IDs | Skip silently | Resilient, no need for aggressive validation |
| Eval-only mode | Expose export endpoint | Reduces master bottleneck for replicas |

## API Changes

### New Query Parameters

| Param | Type | Default | Description |
|---|---|---|---|
| `ids` | string (CSV) | — | Comma-separated flag IDs to include |
| `keys` | string (CSV) | — | Comma-separated flag keys to include |
| `enabled` | boolean | — | Filter by enabled status (omit = both) |
| `tags` | string (CSV) | — | Comma-separated tag values to filter by |
| `all` | boolean | false | Use ALL semantics for tags (default: ANY) |

### Examples

```bash
# Get all flags (unchanged)
GET /api/v1/export/eval_cache/json

# Get only enabled flags
GET /api/v1/export/eval_cache/json?enabled=true

# Get flags with specific IDs
GET /api/v1/export/eval_cache/json?ids=1,2,3

# Get flags with either "foo" or "bar" tag
GET /api/v1/export/eval_cache/json?tags=foo,bar

# Get flags with BOTH "foo" AND "bar" tags
GET /api/v1/export/eval_cache/json?tags=foo,bar&all=true

# Combined: enabled flags with specific IDs
GET /api/v1/export/eval_cache/json?ids=1,2&enabled=true
```

### Filter Precedence

All filters are AND'd together. Within `ids` and `keys`, values are OR'd:

```
(ids=1 OR id=2) AND (key="one" OR key="two") AND (enabled=true) AND (has tag "foo" OR has tag "bar")
```

## Implementation Steps

### Step 1: Swagger spec update

**File:** `swagger/export_eval_cache_json.yaml`

Add 5 optional query parameters to the GET operation.

### Step 2: Regenerate swagger

```bash
make swagger
```

Regenerates `swagger_gen/` with new param struct fields in `GetExportEvalCacheJSONParams`.

### Step 3: Handler refactor

**File:** `pkg/handler/handler.go`

- Extract `setupExportEvalCache(api)` function that registers only the eval cache JSON endpoint
- Call it from eval-only mode block in `Setup()`
- Reuse it in full-mode `setupExport()`

### Step 4: Filter logic

**File:** `pkg/handler/export.go`

- Add `filterFlags(flags []entity.Flag, params export.GetExportEvalCacheJSONParams) []entity.Flag`
- Update `exportEvalCacheJSONHandler` to pass params through
- Logic:
  1. Parse `ids` CSV → `[]uint` (skip invalid)
  2. Parse `keys` CSV → `[]string`
  3. Parse `tags` CSV → `[]string`
  4. Iterate all flags, apply AND predicates
  5. Return filtered slice wrapped in `EvalCacheJSON`

### Step 5: Tests

**File:** `pkg/handler/export_test.go`

Append 8 test cases:
1. No params → all flags
2. `ids=1,2` → matching IDs only
3. `keys=one,two` → matching keys only
4. `enabled=true` → enabled flags only
5. `tags=foo,bar` (ANY) → flags with either tag
6. `tags=foo,bar&all=true` (ALL) → flags with both tags
7. `ids=1&enabled=true` → combined AND
8. No match → empty array, 200 OK

## Files Touched

- `swagger/export_eval_cache_json.yaml`
- `swagger_gen/` (auto-generated)
- `pkg/handler/handler.go`
- `pkg/handler/export.go`
- `pkg/handler/export_test.go`

## Verification

```bash
make swagger          # regenerate
make test             # unit tests pass
make test-e2e         # full suite
```

## Backward Compatibility

All new parameters are optional with zero-values that reproduce current behavior (return all flags). No breaking changes.
