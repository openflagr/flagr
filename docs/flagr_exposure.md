# Exposure logging

Evaluation assigns a variant. Exposure records that the user **saw** it. Do not treat eval volume as experiment participants. Full rule: [contracts: eval vs exposure](contracts.md#eval-vs-exposure).

Client flow: `POST /evaluation` → render → `POST /exposures`. Wire format and warehouse SQL: [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md). Architecture: [Overview](flagr_overview.md#architecture).

## Endpoint

`POST /api/v1/exposures`

```json
{
  "exposures": [
    {
      "flagKey": "my-feature",
      "variantKey": "treatment",
      "entityID": "user-123",
      "entityType": "user",
      "entityContext": { "country": "US", "page": "/checkout" },
      "flagSnapshotID": 42,
      "timestamp": "2026-06-25T12:00:00Z"
    }
  ]
}
```

Each row also accepts `flagID` and `variantID` instead of (or alongside) key-based fields.

### `entityContext`

Same object as evaluation, stored on the record as `evalContext.entityContext`. Segment attributes (`country`, `tier`) and impression fields (`page`, `component`) share one map. There is no separate `metadata` field.

> **Warning:** Exposure ingest does **not** re-run segment constraints. Context is copied only.

Max batch: `FLAGR_EXPOSURE_BATCH_SIZE` (default **100**, `pkg/config/env.go`).

Integration tests: **`TestIntegration_Exposures`** requires this route on current images ([`integration_tests/README.md`](https://github.com/openflagr/flagr/blob/main/integration_tests/README.md)).

## Response

- **200** - `loggedCount`, `message`, optional `errors[]` (`index`, `message`) per row.
- **400** - empty body or batch over limit.

`loggedCount` follows [recording gates](contracts.md#recording-gates). Valid rows can return **200** with `loggedCount: 0`.

## Recording and metrics

Exposure rows use the same streaming pipeline as eval with `recordSource: "exposure"`. **Datar** does not count exposures. Statsd: `exposure.ingest`, `exposure.recorded` (tags: `status`, `FlagID`, `FlagKey`).

Eval-only nodes have **no** exposure API. See [contracts: eval-only](contracts.md#eval-only).

## Validation

| Field | Rule |
|-------|------|
| `entityID` | Required |
| `flagID` / `flagKey` | At least one; if both, must resolve to the same flag |
| `variantID` / `variantKey` | Optional; if set, must exist on the flag |
| `flagSnapshotID` | Optional; omitted → current eval-cache snapshot |
| Disabled flags | Allowed |

## Authentication

`/exposures` is on the default auth whitelist with `/evaluation` (`FLAGR_JWT_AUTH_WHITELIST_PATHS` / `FLAGR_BASIC_AUTH_WHITELIST_PATHS` in `pkg/config/env.go`). Remove `/api/v1/exposures` from those lists when impressions must be authenticated.

> **Warning:** An open `/exposures` endpoint lets callers post fabricated impressions. Rate-limit at your edge; require auth when integrity matters. Flagr has no in-process rate limiter for this route.

Design notes: [`docs/plans/2026-06-25-001-exposure-logging-plan.md`](https://github.com/openflagr/flagr/blob/main/docs/plans/2026-06-25-001-exposure-logging-plan.md).
