# Exposure Logging for A/B Testing

Explicit client-reported exposure (impression) logging complements `POST /evaluation` (assignment). Clients evaluate first, cache the result locally, then call `POST /exposures` when the user actually sees the variant.

## Motivation

- **Evaluation** = server-side bucketing decision.
- **Exposure** = user saw the experiment surface (may be later, batched on unload).
- Warehouse consumers filter on `recordSource: "exposure"` on the same Kafka `evalResult` wire shape.

## API

- `POST /api/v1/exposures`
- Body: `{ "exposures": [ ... ] }`, `minItems: 1`
- `FLAGR_EXPOSURE_BATCH_SIZE` default `100` (request-level 400 if exceeded)

### Per-row validation (catalog D)

| Rule | Behavior |
|------|----------|
| `entityID` | Required |
| Flag | At least one of `flagID` / `flagKey`; if both, must resolve to same flag |
| Variant | Optional; if `variantID` / `variantKey` set, must exist on flag |
| `flagSnapshotID` | Optional; if set, pass through on the record; if omitted, use flag `SnapshotID` from eval cache (no DB validation; downstream joins) |
| Disabled flag | Allowed |
| `timestamp` | Client RFC3339 if valid, else server now |
| Context | Merge `entityContext` + `metadata` → `evalContext.entityContext` |
| `entityType` | Flag `entityType` overrides client (same as eval) |

### Response

- **200** partial accept: `loggedCount`, `errors[]` with `{ index, message }`
- **400** request-level: empty body, missing `exposures`, batch size exceeded

### Recording gate

Same as eval: `FLAGR_RECORDER_ENABLED` + `dataRecordsEnabled` on flag. Otherwise **200** with `loggedCount: 0`.

## Data recorder (not eval metrics)

- **Do not** call `logEvalResult` (no eval Prometheus / eval Statsd `evaluation` metric).
- **Datar** skips rows with `recordSource == "exposure"` (aggregate eval assignments only).
- Build sparse **`EvalResult`** with `recordSource: "exposure"`, `segmentID: 0`, optional variant. Eval assignments use `recordSource: "evaluation"` from `BlankResult` (symmetric wire contract).
- Exposure Statsd: `exposure.ingest` (tags: status, FlagID, …), `exposure.recorded` when written to the data recorder.
- `dataRecordEnabled(flag)` in `data_recorder.go` — shared gate for eval (`logEvalResult`) and exposure `AsyncRecord`.

## Auth

Same middleware as `POST /evaluation`. No dedicated rate limit v1.

## Eval-only mode

`json_file` / `json_http` eval-only `Setup` registers evaluation only; **`POST /exposures` is not available** on those nodes.

## Implementation notes

- Single `pkg/handler/exposure.go`: `PostExposures`, `buildExposureDataRecord` (returns result + flag), `resolveExposureFlag`, `resolveExposureVariant`, `mergeJSONIntoMap`.
- Naming: **`dataRecord`** / **`buildExposureDataRecord`** / “data recorders” (not “pipeline”).
- No thin wrappers around `AsyncRecord`; exposures never call `logEvalResult`.
- Ingest is **cache-only** for flags/variants (no re-segmentation, no snapshot DB lookup); **one cache lookup per accepted row** (no second `GetByFlagKeyOrID` in the loop).

## Files

```
swagger/exposure.yaml
swagger/index.yaml          — paths, definitions, recordSource on evalResult
pkg/config/env.go           — ExposureBatchSize
pkg/handler/exposure.go
pkg/handler/exposure_test.go
pkg/handler/handler.go        — setupExposure
pkg/handler/data_recorder.go   — dataRecordEnabled gate
pkg/handler/data_recorder_datar.go — skip exposure
pkg/handler/data_recorder_kafka.go — skip exposure on kafka Statsd only
docs/flagr_exposure.md
integration_tests/integration_test.go
browser/flagr-ui/.../FlagConfigCard.vue — tooltip
README.md, docs/_sidebar.md
```

## Decisions log

Grill-me session (2026-06-25): synthetic `EvalResult` on same Kafka topic (option 3), variant optional, partial batch accept, separate exposure metrics, catalog validation (D), batch partial 200.

Post-implementation:

- Dropped `recordPipelineEvent` / shared validation modules; direct `AsyncRecord` in eval and exposure.
- Docs/tests aligned with data-recorder vocabulary; operator sections in `flagr_exposure.md`.
- **`flagSnapshotID`**: pass-through only (removed `getDB` validation) — warehouse validates snapshot/flag pairing.

## Code quality review (2026-06-25, updated post thermo-nuclear follow-up)

**Verdict:** Approve for merge after structural follow-up (applied on branch).

**Addressed in follow-up**

- `buildExposureDataRecord` returns `(EvalResult, *entity.Flag, error)` — removed duplicate eval-cache lookup in `PostExposures`.
- `dataRecordEnabled(flag)` — single gate for eval and exposure recorder writes.
- `resolveExposureFlag(ec, row)` — flag ID/key reconciliation extracted for readability.
- `resolveExposureVariant` — single-pass when both ID and key are set.
- Rejection Statsd uses client `flagID` / `flagKey` when present.
- Eval assignments: `recordSource: "evaluation"` on `BlankResult` (symmetric Kafka contract).

**Strengths (unchanged)**

- Direct `AsyncRecord` — correct separation from eval metrics.
- `recordSource` policy in canonical recorders (Datar skip, Kafka Statsd skip).
- No file >1k lines; impressions ≠ assignments.

**Optional later**

- Recorder-enabled integration test for Kafka payload / `recordSource`.
- Tighten swagger `dataRecordsEnabled` “metrics pipeline” wording (pre-existing).