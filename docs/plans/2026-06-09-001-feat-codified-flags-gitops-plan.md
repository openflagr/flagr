# feat: Codified flags — JSON-based GitOps for flag management

**Date:** 2026-06-09
**Status:** implemented

## Summary

Make Flagr's existing JSON export/import good enough for a git-based flag management workflow. Teams can hand-edit flag JSON files, validate them with a standalone tool, and deploy via `json_file` or `json_http` — no running Flagr instance needed to bootstrap.

## What was built

### 1. Auto-ID normalization (`normalizeIDs`)

Hand-edited JSON files can omit all IDs. The system auto-assigns globally unique IDs per entity type (flag, variant, segment, constraint, distribution, tag) on load. Explicit IDs are preserved; auto-assigned IDs start after the highest existing ID.

- `pkg/handler/eval_cache_fetcher.go` — `normalizeIDs()` called from `unmarshalFlags()`
- `pkg/handler/eval_cache_normalize_ids_test.go` — 17 test cases

### 2. JSON flag validator (`ValidateFlags` + `flagr-validate`)

A Go-based validator with deep semantic checks:

- Structural: valid JSON, required fields, key uniqueness
- Semantic: constraint expression parsing (reuses `entity.Constraint.Validate()`), distribution sum = 100, variant references, percentage ranges, attachment JSON validity
- Standalone binary: `cmd/flagr-validate/main.go` — `go build -o flagr-validate ./cmd/flagr-validate/`
- Called during flag loading (logs warnings/errors, doesn't block)
- `pkg/handler/eval_cache_validate.go` — core validation logic
- `pkg/handler/eval_cache_validate_test.go` — 29 test cases

### 3. JSON format documentation

- `docs/flagr_json_flag_spec.md` — schema reference with field tables, operator list, hand-editing guide, and complete example

### 4. GitOps workflow documentation

Updated `docs/flagr_env.md` with two entry paths:
- **Start from scratch:** create `{ "Flags": [] }`, add flags by hand
- **Export from existing:** `curl /api/v1/export/eval_cache/json -o flags.json`

### 5. Eval-only mode improvements

- Export endpoint now available in eval-only mode (`handler.go`)
- `getSnapshotMaxID` no longer panics when no DB is present (`eval_cache.go`)

## What was NOT built (and why)

- **No `?clean=true` export** — the raw export with timestamps is fine. The extra fields don't hurt, and stripping them adds code for marginal benefit.
- **No `json_git` driver** — git clone/fetch inside the server adds operational complexity for a problem solved externally. Point `json_file` at a checkout, or `json_http` at a CI artifact.
- **No new CLI framework** — standalone binary, not a subcommand. The project uses `go-flags` via go-swagger.
- **No schema migration** — premature until there's a v2 format to migrate from.

## Files changed

| File | Change |
|------|--------|
| `pkg/handler/eval_cache_fetcher.go` | `unmarshalFlags()` with validation + `normalizeIDs()`, shared `unmarshalFlags` helper |
| `pkg/handler/eval_cache_validate.go` | `ValidateFlags()` — deep semantic validation |
| `pkg/handler/eval_cache_validate_test.go` | 29 test cases |
| `pkg/handler/eval_cache_normalize_ids_test.go` | 17 test cases |
| `pkg/handler/eval_cache.go` | Fixed `getSnapshotMaxID` for eval-only mode |
| `pkg/handler/handler.go` | Enabled export endpoint in eval-only mode |
| `cmd/flagr-validate/main.go` | Standalone validation binary |
| `docs/flagr_json_flag_spec.md` | Schema reference + example |
| `AGENTS.md` | Added validate binary build command |
| `.gitignore` | Added `flagr-validate` |
