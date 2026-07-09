# Integration guide

Your first Flagr call is one HTTP request: flag + entity → variant. Everything else on this page is the same idea at scale: many flags, many entities, or an impression after render.

Base URL: `https://<flagr-host>/api/v1` (prepend `FLAGR_WEB_PREFIX` if you set one).

Invariants live in [Behavioral contracts](contracts.md) (eval vs exposure, segment evaluation, blank vs stream, recording gates, cache lag, eval-only). Concepts and bucketing: [Overview](flagr_overview.md). Deploy: [Self-hosting](flagr_self_host.md). REST: [API reference](https://openflagr.github.io/flagr/api_docs).

### Eval vs exposure :id=eval-vs-exposure

`POST /evaluation` assigns a variant. `POST /exposures` records that the user saw it. UI experiments need both. Pure server-side branching usually needs only eval. Details: [Eval vs exposure](contracts.md#eval-vs-exposure).

## Endpoints you need

| Call | Method | When |
|------|--------|------|
| Assign variant | `POST /evaluation` | **Primary** - servers, SDKs, rich `entityContext` |
| Assign variant (browser) | `GET /evaluation?json=…` | Same JSON as POST in one query param - [GET evaluation](flagr_use_cases.md#get-evaluation-browser-friendly) |
| Assign many | `POST /evaluation/batch` | Many entities and/or flags (or tag filter) |
| Assign many (browser) | `GET /evaluation/batch?json=…` | Batch body in `json=` - same limits as POST |
| Log impression | `POST /exposures` | After the user **sees** the treatment |
| Liveness | `GET /health` | Probes |

Eval-only replicas (`json_file` / `json_http`) expose evaluation, health, and `GET /api/v1/export/eval_cache/json` only. See [contracts: eval-only](contracts.md#eval-only).

## Request model

Three ideas on every eval request:

1. **`entityID`** - who you are evaluating (user, device, account). Stability is stickiness: same client-sent ID + unchanged flag → same variant. Omit it and the server injects a random id (non-sticky).
2. **`entityType`** (optional) - labels the entity in logs and records so `user-42` and `device-42` stay distinct.
3. **`entityContext`** (optional) - free-form JSON that constraints match. Dotted paths reach nested values (`user.tier` → `{"user":{"tier":"pro"}}`); rules: [Overview: constraint property access](flagr_overview.md#constraint-property-access).

With **`FLAGR_INJECTED_CONTEXT_ENABLED=true`**, Flagr merges server-side keys (`@ts`, `@ts_hour`, configured `@http_*` from request headers) into that map before constraints run. App code does not need to send those keys. See [Built-in context injection](flagr_injected_context.md).

Resolve the flag with **`flagID`** or **`flagKey`**. Either is enough.

## Single evaluation

```bash
curl -sS -X POST 'http://localhost:18000/api/v1/evaluation' \
  -H 'content-type: application/json' \
  -d '{
    "entityID": "user-42",
    "entityType": "user",
    "entityContext": {
      "region": "us-west",
      "age": 30,
      "tier": "premium"
    },
    "flagID": 1
  }'
```

Nested context (constraints use dotted paths):

```json
{
  "entityID": "user-42",
  "entityType": "user",
  "entityContext": {
    "user": { "name": "Alice", "age": 30 }
  },
  "flagKey": "my-feature"
}
```

### Response fields to use

| Field | Use |
|-------|-----|
| `variantKey` | Branch in app code; empty ⇒ no assignment ([EvalCache](contracts.md#evalcache-freshness)) |
| `variantID` | Stable id for exposures and analytics |
| `variantAttachment` | JSON config for this variant |
| `flagSnapshotID` | Pass through on `POST /exposures` for warehouse joins |
| `evalContext` | Echo of entity + match metadata |
| `enableDebug` + `evalDebugLog` | Segment walk - [Debug console](flagr_debugging.md) |

Branch on `variantKey`. Empty means missing flag, disabled flag, or no assigned variant: normal outcomes, not transport failures.

## Batch evaluation

One page, many flags? Use `POST /api/v1/evaluation/batch`. One request, one result per entity per selected flag.

By ID:

```json
{
  "entities": [
    {
      "entityID": "a",
      "entityType": "user",
      "entityContext": { "region": "us-west", "age": 30 }
    }
  ],
  "flagIDs": [1, 2]
}
```

By tag (`ANY` = at least one listed tag; `ALL` = every listed tag):

```json
{
  "entities": [
    {
      "entityID": "a",
      "entityType": "user",
      "entityContext": { "region": "us-west" }
    }
  ],
  "flagTags": ["int_test"],
  "flagTagsOperator": "ANY"
}
```

Work cap: `len(entities) * (len(flagIDs) + len(flagKeys) + tags estimate)` against `FLAGR_EVAL_BATCH_SIZE` (`0` = unlimited). Duplicate IDs/keys are deduped before the count.

**CI gotcha:** a flag you just created is not evaluable until EvalCache reloads. Poll with a real eval (this repo's **`waitForEvalReady`**) until you see a variant. See [EvalCache freshness](contracts.md#evalcache-freshness).

## UI experiment loop

Assignment alone is not a denominator. Count people who **saw** the treatment ([contracts: eval vs exposure](contracts.md#eval-vs-exposure)).

1. **`POST /evaluation`** - cache `variantKey`, `variantID`, `flagSnapshotID`.
2. **Render** only when `variantKey` is non-empty. Low rollout on a matched segment can leave the key empty and does **not** fall through to later segments ([segment evaluation](contracts.md#segment-evaluation)).
3. **`POST /exposures`** when the surface is visible (mount, viewport, or unload batch).

```bash
curl -sS -X POST 'http://localhost:18000/api/v1/exposures' \
  -H 'content-type: application/json' \
  -d '{
    "exposures": [{
      "flagID": 1,
      "variantID": 2,
      "variantKey": "treatment",
      "entityID": "user-42",
      "entityType": "user",
      "flagSnapshotID": 42,
      "entityContext": { "page": "/checkout" }
    }]
  }'
```

Exposure validates against the cache; it does **not** re-run constraints. Pass the `flagSnapshotID` from eval so the warehouse can join impressions to the config that produced them. Full shape: [Exposure logging](flagr_exposure.md). Downstream: [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md).

## Server-side only (no UI)

Timeouts, routing weights, feature paths: call **`POST /evaluation`** (or batch), read **`variantAttachment`**, branch. Skip `POST /exposures` unless you need a formal A/B denominator in a warehouse.

## Client libraries

| Language | Package |
|----------|---------|
| Go | [goflagr](https://github.com/openflagr/goflagr) |
| JavaScript | [jsflagr](https://github.com/openflagr/jsflagr) |
| Python | [pyflagr](https://github.com/openflagr/pyflagr) |
| Ruby | [rbflagr](https://github.com/openflagr/rbflagr) |

## Where to go next

| Goal | Doc |
|------|-----|
| Flag / segment concepts | [Overview](flagr_overview.md), [Use cases](flagr_use_cases.md) |
| Browser GET eval | [Use cases: GET evaluation](flagr_use_cases.md#get-evaluation-browser-friendly) |
| Time / header targeting | [Built-in context injection](flagr_injected_context.md) |
| Run Flagr | [Self-hosting](flagr_self_host.md) |
| Env vars | [Environment variables](flagr_env.md) |
| GitOps flags | [JSON flag source](flagr_json_flag_spec.md) |
| Wrong variant | [Debug console](flagr_debugging.md) |
| Change Flagr itself | [Contributing](CONTRIBUTING.md) |
