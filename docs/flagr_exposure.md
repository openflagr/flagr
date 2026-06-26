# Exposure Logging

Exposure logging records when a user **actually saw** a flag or experiment surface, separate from `POST /evaluation` (assignment). Typical flow: evaluate to get a variant, cache the result client-side, then call `POST /exposures` when the UI renders.

## Endpoint

`POST /api/v1/exposures`

Request body:

```json
{
  "exposures": [
    {
      "flagKey": "my-feature",
      "variantKey": "treatment",
      "entityID": "user-123",
      "entityType": "user",
      "entityContext": { "country": "US" },
      "metadata": { "page": "/checkout" },
      "flagSnapshotID": 42,
      "timestamp": "2026-06-25T12:00:00Z"
    }
  ]
}
```

Single impressions use an array of one item. Maximum batch size: `FLAGR_EXPOSURE_BATCH_SIZE` (default **100**).

### Typical client flow

1. `POST /evaluation` — get variant assignment.
2. Render the experiment surface in the UI.
3. `POST /exposures` — report that the user saw it (batch on page unload if needed).


## Response

- **200** with `loggedCount`, `message`, and optional `errors[]` (`index`, `message`) for per-row validation failures.
- **400** for empty body or batch over limit.

`loggedCount` is the number of rows written to the data recorder. If the recorder is off or the flag’s `dataRecordsEnabled` is false, valid rows still return **200** with `loggedCount: 0`.

## Recording

Same gate as evaluation data records:

1. `FLAGR_RECORDER_ENABLED=true`
2. Recorder type configured (e.g. `kafka`)
3. Per-flag `dataRecordsEnabled: true`

Recorded rows use the same `evalResult` JSON shape as evaluations (same Kafka topic when Kafka is configured), with `recordSource: "exposure"`. **Datar does not count exposures.** Exposure ingest uses separate Statsd metrics (`exposure.ingest`, `exposure.recorded`).

## How this connects to evaluation and `AsyncRecord`

**Evaluation** runs bucketing, then `logEvalResult` (eval Statsd/Prometheus). Results use `recordSource: "evaluation"`. When `dataRecordsEnabled` and `FLAGR_RECORDER_ENABLED`, it calls `GetDataRecorder().AsyncRecord(evalResult)`.

**Exposure** skips `logEvalResult`. For valid rows it builds an exposure `evalResult` with `recordSource: "exposure"` and, when the same gates pass, calls `GetDataRecorder().AsyncRecord` directly.

`AsyncRecord` fans out to configured **data recorders** (Kafka, Kinesis, Pub/Sub, Datar). **Datar** ignores rows with `recordSource == exposure`. Exposure uses Statsd `exposure.ingest` / `exposure.recorded`.

### Eval-only mode

`json_file` / `json_http` nodes with eval-only setup register **evaluation** only; **`POST /exposures` is not available** on those deployments.

### Kafka / warehouse consumers

Filter or branch on `recordSource`:

- `evaluation` — rows from the evaluation API path (`BlankResult`), including not-found/disabled responses; not necessarily a successful variant assignment
- `exposure` — client-reported impression

Do not treat exposure rows as assignments for experiment analysis.

### Statsd (exposure ingest)

Separate from eval `evaluation` metric: `exposure.ingest` (tags: `status=accepted|rejected|recorded`, `FlagID`, `FlagKey`) and `exposure.recorded` when a row is passed to `AsyncRecord`.

## Validation

| Field | Rule |
|-------|------|
| `entityID` | Required |
| `flagID` / `flagKey` | At least one; both must match the same flag |
| `variantID` / `variantKey` | Optional; if set, must exist on the flag |
| `flagSnapshotID` | Optional; if set, recorded as sent (downstream can join to snapshots); if omitted, uses current eval-cache snapshot |
| Disabled flags | Allowed |

Auth matches `POST /evaluation`.

See [plan](../plans/2026-06-25-001-exposure-logging-plan.md) for design decisions.