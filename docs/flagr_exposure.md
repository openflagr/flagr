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

## Response

- **200** with `loggedCount`, `message`, and optional `errors[]` (`index`, `message`) for per-row validation failures.
- **400** for empty body or batch over limit.

`loggedCount` is the number of rows written to the data recorder. If the recorder is off or the flag’s `dataRecordsEnabled` is false, valid rows still return **200** with `loggedCount: 0`.

## Recording

Same gate as evaluation data records:

1. `FLAGR_RECORDER_ENABLED=true`
2. Recorder type configured (e.g. `kafka`)
3. Per-flag `dataRecordsEnabled: true`

Pipeline events are synthetic `evalResult` JSON on the **same topic** as evaluations, with `recordSource: "exposure"`. **Datar does not count exposures.** Exposure ingest uses separate Statsd metrics (`exposure.ingest`, `exposure.recorded`).

## Validation

| Field | Rule |
|-------|------|
| `entityID` | Required |
| `flagID` / `flagKey` | At least one; both must match the same flag |
| `variantID` / `variantKey` | Optional; if set, must exist on the flag |
| `flagSnapshotID` | If set, must exist for the flag (may be stale); else current cache snapshot is used |
| Disabled flags | Allowed |

Auth matches `POST /evaluation`.

See [plan](../plans/2026-06-25-001-exposure-logging-plan.md) for design decisions.