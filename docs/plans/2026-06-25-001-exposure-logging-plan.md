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
| `flagSnapshotID` | If set, must exist for that flag (stale OK); if omitted, use flag `SnapshotID` from eval cache |
| Disabled flag | Allowed |
| `timestamp` | Client RFC3339 if valid, else server now |
| Context | Merge `entityContext` + `metadata` → `evalContext.entityContext` |
| `entityType` | Flag `entityType` overrides client (same as eval) |

### Response

- **200** partial accept: `loggedCount`, `errors[]` with `{ index, message }`
- **400** request-level: empty body, missing `exposures`, batch size exceeded

### Recording gate

Same as eval: `FLAGR_RECORDER_ENABLED` + `dataRecordsEnabled` on flag. Otherwise **200** with `loggedCount: 0`.

## Pipeline (not eval)

- **Do not** call `logEvalResult` (no eval Prometheus / eval Statsd `evaluation` metric).
- **Datar** skips rows with `recordSource == "exposure"`.
- Build **synthetic sparse `EvalResult`** with `recordSource: "exposure"`, `segmentID: 0`, optional variant.
- Exposure Statsd: `exposure.ingest` (tags: status, FlagID, …), `exposure.recorded` when written to pipeline.
- `GetDataRecorder().AsyncRecord(synthetic)` when gates pass.

## Auth

Same middleware as `POST /evaluation`. No dedicated rate limit v1.

## Files

```
swagger/exposure.yaml
swagger/index.yaml          — paths, definitions, recordSource on evalResult
pkg/config/env.go           — ExposureBatchSize
pkg/handler/exposure.go
pkg/handler/exposure_test.go
pkg/handler/handler.go        — setupExposure
pkg/handler/data_recorder_datar.go — skip exposure
docs/flagr_exposure.md
integration_tests/integration_test.go
browser/flagr-ui/.../FlagConfigCard.vue — tooltip
README.md, docs/_sidebar.md
```

## Decisions log

See grill-me session (2026-06-25): synthetic EvalResult on same Kafka topic (option 3), variant optional, partial batch accept, separate metrics.