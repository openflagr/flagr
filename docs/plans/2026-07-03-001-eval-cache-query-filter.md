# Plan: EvalCache Query Filtering

**Issue:** [#628](https://github.com/openflagr/flagr/issues/628)
**Branch:** `feat/eval-cache-query-filter`
**Date:** 2026-07-03
**Based on:** [PR #630](https://github.com/openflagr/flagr/pull/630) by @iamafanasyev

## Goal

Add query parameters to `GET /api/v1/export/eval_cache/json` so users can filter flags from the in-memory EvalCache without hitting the database. Also expose the export endpoint in eval-only mode (replicas).

## Motivation

Flag DBs grow over time (deprecated, disabled, test flags). The export API currently dumps everything, causing network bloat and client-side filtering. EvalCache filtering provides a lightweight alternative that works on replicas (eval-only mode) too.

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Swagger type | `type: array` + `collectionFormat: csv` | Proper OpenAPI semantics, go-swagger auto-parses to `[]int64`/`[]string` |
| Filter location | Inside `export()` under cache lock | Uses O(1) cache lookups (idCache, keyCache) |
| Filter precedence | ids/keys override enabled/tags | ids/keys are O(1) lookups, no need to iterate all flags |
| enabled/tags semantics | AND across both | Consistent with FindFlags behavior |
| Tag semantics | ANY by default, `all=true` for ALL | Matches postEvaluation's flagTagsOperator |
| Eval-only mode | Expose export endpoint | Reduces master bottleneck for replicas |

## API Reference

### Endpoint

```
GET /api/v1/export/eval_cache/json
```

### Query Parameters

| Param | Type | Default | Description |
|---|---|---|---|
| `ids` | `int64[]` (CSV) | — | Flag IDs to include. **Highest precedence** — when provided, all other params are ignored. |
| `keys` | `string[]` (CSV) | — | Flag keys to include. **Second precedence** — when provided, enabled/tags are ignored. |
| `enabled` | `boolean` | — | Filter by enabled status (omit to return all). Combined with tags via AND. |
| `tags` | `string[]` (CSV) | — | Tag values to filter by. Combined with enabled via AND. |
| `all` | `boolean` | `false` | Use ALL semantics for tags (default: ANY). |

### Examples

```bash
# Get all flags (unchanged behavior)
GET /api/v1/export/eval_cache/json

# Get only enabled flags
GET /api/v1/export/eval_cache/json?enabled=true

# Get flags with specific IDs
GET /api/v1/export/eval_cache/json?ids=1,2,3

# Get flags with specific keys
GET /api/v1/export/eval_cache/json?keys=feature-a,feature-b

# Get flags with either "foo" or "bar" tag (ANY semantics)
GET /api/v1/export/eval_cache/json?tags=foo,bar

# Get flags with BOTH "foo" AND "bar" tags (ALL semantics)
GET /api/v1/export/eval_cache/json?tags=foo,bar&all=true

# Combined: enabled flags with specific tags
GET /api/v1/export/eval_cache/json?enabled=true&tags=foo,bar

# IDs override everything (even if enabled=false or tags don't match)
GET /api/v1/export/eval_cache/json?ids=1,2&enabled=false&tags=nonexistent
# Returns flags 1 and 2 regardless of their enabled state or tags
```

### Precedence Rules

1. **`ids`** — Highest precedence. When provided, `keys`, `enabled`, and `tags` are ignored. Lookup is O(1) via `idCache`.
2. **`keys`** — Second precedence. When provided, `enabled` and `tags` are ignored. Lookup is O(1) via `keyCache`.
3. **`enabled` + `tags`** — Combined with AND semantics. Both conditions must match.

### Response

```json
{
  "Flags": [
    {
      "ID": 1,
      "Key": "my-feature",
      "Enabled": true,
      "Tags": [{"Value": "team-a"}],
      "Segments": [...],
      "Variants": [...]
    }
  ]
}
```

## Implementation

### Files Changed

| File | Change |
|---|---|
| `swagger/export_eval_cache_json.yaml` | Added query parameters with array types |
| `swagger_gen/` | Regenerated with new param types |
| `pkg/handler/eval_cache_fetcher.go` | Filter logic inside `export()` method |
| `pkg/handler/handler.go` | Extracted `setupExportEvalCache`, added to eval-only mode |
| `pkg/handler/fixture.go` | Added `GenFixtureFlagWithTags`, `GenFixtureEvalCacheWithFlags` |
| `pkg/handler/export_test.go` | Unit tests for filtering |
| `integration_tests/integration_test.go` | Integration tests for query params |
| `browser/flagr-ui/e2e/flag-detail.spec.ts` | Fixed flaky toast assertion |

### Key Implementation Details

**Filter logic in `export()`:**

```go
func (ec *EvalCache) export(query export.GetExportEvalCacheJSONParams) EvalCacheJSON {
    // Build O(1) lookup sets for ids/keys
    var targetIDs map[int64]struct{}
    if len(query.Ids) > 0 {
        targetIDs = make(map[int64]struct{}, len(query.Ids))
        for _, id := range query.Ids {
            targetIDs[id] = struct{}{}
        }
    }
    // ... similar for keys, tags ...

    ec.cacheMutex.RLock()
    defer ec.cacheMutex.RUnlock()

    for _, f := range ec.cache.idCache {
        // ids filter: highest precedence, O(1) lookup
        if targetIDs != nil {
            if _, ok := targetIDs[int64(f.ID)]; ok {
                fs = append(fs, *f)
            }
            continue
        }
        // keys filter: second precedence, O(1) lookup
        if targetKeys != nil {
            if _, ok := targetKeys[f.Key]; ok {
                fs = append(fs, *f)
            }
            continue
        }
        // enabled + tags: AND semantics
        if query.Enabled != nil && *query.Enabled != f.Enabled {
            continue
        }
        if hasTags != nil && !hasTags(f) {
            continue
        }
        fs = append(fs, *f)
    }
    return EvalCacheJSON{Flags: fs}
}
```

**Eval-only mode registration:**

```go
func Setup(api *operations.FlagrAPI) {
    if config.Config.EvalOnlyMode {
        setupHealth(api)
        setupEvaluation(api)
        setupExportEvalCache(api)  // NEW: expose export on replicas
        return
    }
    // ... full mode setup ...
}
```

## Testing

### Unit Tests (`TestExportEvalCacheQuery`)

- No params returns all flags
- Filter by IDs (single and multiple)
- Filter by keys (single and multiple)
- Filter by enabled (true/false)
- Filter by tags (ANY and ALL semantics)
- IDs override enabled and tags
- Keys override enabled and tags
- Combined enabled AND tags
- No match returns empty

### Integration Tests (`TestIntegration_EvalCacheExportQuery`)

- Create flags with specific enabled states
- Add tags to flags
- Test all query parameter combinations
- Verify precedence rules (ids override)
- Verify empty results for non-existent IDs

### E2E Tests

- Fixed flaky `can toggle flag enabled state` test (removed transient toast assertion)

## Backward Compatibility

All new parameters are optional with zero-values that reproduce current behavior (return all flags). No breaking changes.

## Usage Scenarios

1. **Selective sync** — Clients sync only enabled flags: `?enabled=true`
2. **Targeted evaluation** — Replicas export only specific flags: `?ids=1,2,3`
3. **Tag-based filtering** — Export flags for a specific team: `?tags=team-a`
4. **Hybrid filtering** — Enabled flags with specific tags: `?enabled=true&tags=production`
