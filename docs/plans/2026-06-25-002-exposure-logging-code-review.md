# Exposure logging — thermo-nuclear code quality review (2026-06-25)

Scope: `feat/exposure-logging` branch changes (~240-line `exposure.go`, recorder hooks, docs, tests).

## Verdict

**Approve with follow-ups addressed in this pass** (docs/README/tests). No file >1k lines; no unjustified wrapper layer after `1dba3ef3`.

## What already landed well (code judo)

- **Direct `AsyncRecord`** instead of `logEvalResult` + thin `recordPipelineEvent` — correct boundary: impressions are not assignments.
- **`recordSource` in Datar/Kafka** — one enum check in canonical recorders, not scattered eval special cases.
- **Single handler file** — validation + `dataRecord` build colocated; helpers are real (variant resolve, JSON merge), not pass-throughs.
- **Vocabulary** — `dataRecord` / `buildExposureDataRecord` / “data recorders” in docs aligns with `DataRecorder` API.

## Findings addressed in this pass

| Priority | Issue | Action |
|----------|--------|--------|
| P2 | Docs said “pipeline” while code says `DataRecorder` | README, `flagr_exposure.md`, plan, overview, env |
| P2 | `buildExposureDataRecord` ~70% coverage (snapshot DB, eval-only, merge) | Unit tests with `getDB` stub + `SaveFlagSnapshot` |
| P3 | Eval-only exposure omission undocumented in tests | Documented in `flagr_exposure.md` (handler test omitted: `Setup` starts eval cache) |
| P3 | Missing operator sections (Kafka filter, Statsd, eval-only) | `flagr_exposure.md` sections |

## Residual / optional (not blockers)

1. **`buildExposureDataRecord` size (~85 lines)** — If exposure grows (tokens, assignment binding), extract **validation** vs **record assembly** into two functions in the same file; not worth splitting for v1.
2. **Flag lookup branch (id + key)** — Slightly busy `switch`; a small `resolveExposureFlag(ec, row)` would shrink `buildExposureDataRecord` only if a third caller appears.
3. **`mergeJSONIntoMap` + `map[string]interface{}`** — Matches eval/wire JSON patterns elsewhere; typed context model would be cross-cutting, out of scope.
4. **Integration test does not assert `recordSource` on Kafka** — Would need recorder-enabled integration fixture; unit + `TestDatarRecorder_SkipsExposure` sufficient for v1.
5. **Swagger `dataRecordsEnabled` still says “metrics pipeline”** — Pre-existing wording on flag models; regenerate with `make swagger` if you tighten copy in `swagger/index.yaml` (exposure-only `recordSource` text already updated).

## Approval bar checklist

- [x] No structural regression in shared eval path
- [x] No file-size explosion
- [x] No new spaghetti in `eval.go` / recorders (isolated `recordSource` checks)
- [x] Abstractions earn their keep post-refactor
- [x] Docs/tests updated for operator and consumer contracts