# Exposure logging

Evaluation answers one question: *which variant should this entity see?*
Exposure logging answers a different one: *did the entity actually see it?*

The distinction matters because evaluation and impression are not the same
event. An evaluation can fire before anything renders — on every navigation,
inside a prefetch, against a flag that is disabled or matches no segment.
Counting those calls as experiment participants inflates denominators and
dilutes A/B results. Exposure logging exists so that the moment a user
*sees* the treatment is reported separately, by the client, after the
surface has rendered. That impression — not the assignment — is the honest
denominator for an experiment.

The rules that govern how these two events are recorded, and what each one
means, are defined in [Behavioral contracts](contracts.md). The wire format
both events share, and how a warehouse turns them into A/B analysis, is
covered in [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md).
For where exposure sits in the wider system, see
[Overview](flagr_overview.md#architecture).

The client flow is a short chain: `POST /evaluation` to get the assignment,
render the surface, then `POST /exposures` when the user sees the treatment.

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

Each row also accepts `flagID` and `variantID` instead of key-based fields.

### `entityContext`

Same object as evaluation — stored on the record as `evalContext.entityContext`. Segment attributes (`country`, `tier`) and impression fields (`page`, `component`) share one map; there is no separate `metadata` field.

> **Warning:** Exposure ingest does **not** re-run segment constraints. Context is copied only.

Max batch: `FLAGR_EXPOSURE_BATCH_SIZE` (default **100**).

Integration tests: **`TestIntegration_Exposures`** requires this route on current images ([`integration_tests/README.md`](https://github.com/openflagr/flagr/blob/main/integration_tests/README.md)).

## Response

- **200** — `loggedCount`, `message`, optional `errors[]` (`index`, `message`) per row.
- **400** — empty body or batch over limit.

`loggedCount` follows [recording gates](contracts.md#recording-gates). Valid rows can return **200** with `loggedCount: 0`.

## Recording and metrics

Exposure rows use the same streaming pipeline as eval with `recordSource: "exposure"`. **Datar** does not count exposures. Statsd: `exposure.ingest`, `exposure.recorded` (tags: `status`, `FlagID`, `FlagKey`).


Eval-only nodes have **no** exposure API — [contracts — eval-only](contracts.md#eval-only).

## Validation

| Field | Rule |
|-------|------|
| `entityID` | Required |
| `flagID` / `flagKey` | At least one; if both, must match same flag |
| `variantID` / `variantKey` | Optional; if set, must exist on flag |
| `flagSnapshotID` | Optional; omitted → current eval-cache snapshot |
| Disabled flags | Allowed |

## Authentication

`/exposures` is on the default auth whitelist with `/evaluation`. Remove `/api/v1/exposures` from `FLAGR_JWT_AUTH_WHITELIST_PATHS` / `FLAGR_BASIC_AUTH_WHITELIST_PATHS` when impressions must be authenticated.

> **Warning:** Open `/exposures` lets callers post fabricated impressions. Rate-limit at your edge; require auth when integrity matters. Flagr has no in-process rate limiter for this route.

Design notes: [`docs/plans/2026-06-25-001-exposure-logging-plan.md`](https://github.com/openflagr/flagr/blob/main/docs/plans/2026-06-25-001-exposure-logging-plan.md).