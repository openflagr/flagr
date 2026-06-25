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
- **Kafka** still produces exposure rows; Statsd `data_recorder.kafka` skips exposure (eval-oriented counter).
- Build sparse **`EvalResult`** with `recordSource: "exposure"`, `segmentID: 0`, optional variant.
- Exposure Statsd: `exposure.ingest` (tags: status, FlagID, …), `exposure.recorded` when written to the data recorder.
- `GetDataRecorder().AsyncRecord(dataRecord)` when `dataRecordsEnabled && RecorderEnabled` (eval and exposure use the same inline gate).

## Auth

Same middleware as `POST /evaluation`. No dedicated rate limit v1.

## Eval-only mode

`json_file` / `json_http` eval-only `Setup` registers evaluation only; **`POST /exposures` is not available** on those nodes.

## Implementation notes

- Single `pkg/handler/exposure.go` (~230 lines): `PostExposures` loop, `buildExposureDataRecord`, `resolveExposureVariant`, `mergeJSONIntoMap`.
- Naming: **`dataRecord`** / **`buildExposureDataRecord`** / “data recorders” (not “pipeline”).
- No thin wrappers around `AsyncRecord`; exposures never call `logEvalResult`.
- Ingest is **cache-only** for flags/variants (no re-segmentation, no snapshot DB lookup).

## Files

```
swagger/exposure.yaml
swagger/index.yaml          — paths, definitions, recordSource on evalResult
pkg/config/env.go           — ExposureBatchSize
pkg/handler/exposure.go
pkg/handler/exposure_test.go
pkg/handler/handler.go        — setupExposure
pkg/handler/eval.go             — RecorderEnabled + dataRecordsEnabled gate on AsyncRecord
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

## Code quality review (2026-06-25)

**Verdict:** Approve. No file >1k lines; boundaries clean (impressions ≠ assignments).

**Strengths**

- Direct `AsyncRecord` — correct separation from eval metrics.
- `recordSource` handled in canonical recorders (Datar skip, Kafka Statsd skip), not scattered in eval.
- Helpers earn their keep (variant resolve, JSON merge); not pass-through wrappers.

**Optional later (not v1 blockers)**

- Split `buildExposureDataRecord` if exposure grows (tokens, assignment binding).
- `resolveExposureFlag(ec, row)` only if a second caller appears.
- Recorder-enabled integration test for Kafka payload / `recordSource`.
- Tighten swagger `dataRecordsEnabled` “metrics pipeline” wording (pre-existing on flag models).