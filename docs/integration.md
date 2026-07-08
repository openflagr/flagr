# Integration guide

Your first call to Flagr is a single HTTP request: give it a flag and an entity, get back a variant. Everything else in this guide is that same idea at larger scale — the same call across many flags, across many entities, or paired with an impression for A/B analysis. The base URL is `https://<flagr-host>/api/v1`, with `FLAGR_WEB_PREFIX` prepended if you set it.

This page is the practical "how do I call it" walkthrough. The invariants those calls rely on (eval vs exposure, recording gates, cache lag, eval-only nodes) live in [Behavioral contracts](contracts.md) and are not repeated here. For the concepts behind flags, segments, and bucketing, read the [Overview](flagr_overview.md). For bringing up the server itself, see [Self-hosting](flagr_self_host.md). The full REST surface is in the [API reference](https://openflagr.github.io/flagr/api_docs).

### Eval vs exposure :id=eval-vs-exposure

`POST /evaluation` assigns a variant. `POST /exposures` records that the user actually saw it. UI experiments need both; server-side branching usually needs only eval. See [Eval vs exposure](contracts.md#eval-vs-exposure).

## Endpoints you need

Most integrations touch evaluation (POST or GET), batch eval, optional exposures, and health. POST is the default; GET carries the same JSON in `json=` when browsers or caches need a read-shaped URL ([use cases](flagr_use_cases.md#get-evaluation-browser-friendly)).

| Call | Method | When |
|------|--------|------|
| Assign variant | `POST /evaluation` | **Primary** — servers, SDKs, rich `entityContext` |
| Assign variant (browser) | `GET /evaluation?json=…` | Same JSON as POST in one query param — [Use cases — GET evaluation](flagr_use_cases.md#get-evaluation-browser-friendly) |
| Assign many | `POST /evaluation/batch` | Many entities and/or many flags (or tag filter) |
| Assign many (browser) | `GET /evaluation/batch?json=…` | Batch body in `json=` — same limits as POST |
| Log impression | `POST /exposures` | After the user **sees** the treatment (UI experiments) |
| Liveness | `GET /health` | Probes |

If you run eval-only replicas off a `json_file` or `json_http` source, the surface narrows to evaluation (POST and GET), health, and `GET /export/eval_cache/json` — see [contracts — eval-only](contracts.md#eval-only) for what is intentionally absent on those nodes.

## Request model

Every evaluation request carries the same three ideas. An **`entityID`** identifies who you are evaluating — a user id, device id, account id, anything stable. That stability is what makes evaluation sticky: the same `entityID` against an unchanged flag always returns the same variant. An optional **`entityType`** labels the entity in logs and downstream records so "user-42" and "device-42" are not confused.

The third idea is **`entityContext`**, a free-form JSON object that segment constraints match against. Property names may use dots to reach nested values (`user.tier` resolves to `{"user":{"tier":"pro"}}`); the full resolution rules are in [Overview — constraint property access](flagr_overview.md#constraint-property-access). When **`FLAGR_INJECTED_CONTEXT_ENABLED=true`**, Flagr also merges server-side keys (`@ts`, `@ts_hour`, configurable `@http_*` from request headers) into that map before constraints run — so you can target by time or proxy headers without changing app code. See [Built-in context injection](flagr_injected_context.md).

Finally, the flag itself is resolved with **`flagID`** or **`flagKey`**. Either is enough; supply whichever is convenient in the calling code.

## Single evaluation

The simplest useful call is one entity against one flag. Send a `POST /evaluation` with the entity context and the flag you want evaluated:

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

When your targeting rules need nested context — say a `user` object with its own fields — the same request shape carries it, and constraints can reach into it with dotted paths:

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

The response carries everything your code needs to branch and everything your analytics need to join. Treat the fields below as the contract between your application, your experiment pipeline, and Flagr itself.

| Field | Use |
|-------|-----|
| `variantKey` | Branch in app code; empty ⇒ no assignment ([contracts — EvalCache](contracts.md#evalcache-freshness)) |
| `variantID` | Stable id for exposures and analytics |
| `variantAttachment` | JSON config for this variant (dynamic configuration) |
| `flagSnapshotID` | Pass through on `POST /exposures` for warehouse joins |
| `evalContext` | Echo of entity + matched segment metadata |
| `enableDebug` + `evalDebugLog` | Segment walk — [Debug console](flagr_debugging.md) |

The one rule to internalize: branch on `variantKey`, and treat an empty value as "no assignment" rather than an error. A blank result means the flag was missing, disabled, or had no matching segment — all normal evaluation outcomes, not failures ([contracts — EvalCache](contracts.md#evalcache-freshness)).

## Batch evaluation

The moment a single page needs more than one flag — a cart that branches on checkout, payment, and shipping flags simultaneously, or a bootstrap that resolves a user's entire flag set at startup — a round trip per flag becomes wasteful. `POST /api/v1/evaluation/batch` collapses those into one request. You send a list of entities and a selection of flags, and you get back one result per entity per flag.

Select flags by ID:

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

…or by tag when you want every flag carrying a label without naming them individually. `flagTagsOperator: "ANY"` matches flags that carry at least one listed tag; switch to **`ALL`** when every listed tag must be present:

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

The server caps total work per request — `len(entities) * (len(flagIDs) + len(flagKeys) + tags)` — behind `FLAGR_EVAL_BATCH_SIZE`, where `0` means unlimited. Duplicate IDs and keys are deduplicated before that count, so you cannot inflate a batch by repeating the same flag.

One gotcha worth naming: after you create a flag in CI or the UI, it is not visible to evaluation until the cache reloads. Integration tests handle this by polling with **`waitForEvalReady`** until a real eval returns a variant — the same approach you want in any automated flow that creates and then asserts on a flag ([EvalCache freshness](contracts.md#evalcache-freshness)).

## UI experiment loop

A UI experiment is the one place where evaluation alone is not enough. The server can assign a variant, but that says nothing about whether the user ever *saw* it. Counting an assigned-but-unrendered user in your experiment denominator is how A/B numbers go wrong. The healthy loop is three steps, in order:

1. **`POST /evaluation`** — cache `variantKey`, `variantID`, and `flagSnapshotID` from the response.
2. **Render** only when `variantKey` is non-empty. This is where rollout percentages and holdouts are respected — a matched segment with a low rollout can still produce no variant.
3. **`POST /exposures`** when the surface is actually visible — on mount, on entering the viewport, or batched on unload if you are deferring.

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

The exposure request is validated against the cache — no constraints are re-evaluated — so pass through the `flagSnapshotID` you got at eval time; it is what lets your warehouse join impressions back to the exact flag configuration that produced them. For the request shape, validation rules, and batch limits, read [Exposure logging](flagr_exposure.md). For where those rows go next — Kafka, Kinesis, Pub/Sub, and the warehouse SQL that turns them into A/B lift — read [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md).

## Server-side only (no UI)

Not every flag drives a UI. A backend that picks a timeout, a routing weight, or a feature path just needs the variant and its config — no impression, no denominator. In that case call **`POST /evaluation`** per request (or batch at the edge), read **`variantAttachment`** for whatever configuration the variant carries, and branch. You do not need `POST /exposures` unless you are running a formal A/B test with a warehouse denominator; pure server-side branching never does.

## Client libraries

If you would rather not hand-roll HTTP, maintained clients wrap the REST surface for the common languages:

| Language | Package |
|----------|---------|
| Go | [goflagr](https://github.com/openflagr/goflagr) |
| JavaScript | [jsflagr](https://github.com/openflagr/jsflagr) |
| Python | [pyflagr](https://github.com/openflagr/pyflagr) |
| Ruby | [rbflagr](https://github.com/openflagr/rbflagr) |

## Where to go next

| Goal | Doc |
|------|-----|
| Flag/segment concepts | [Overview](flagr_overview.md), [Use cases](flagr_use_cases.md) |
| Browser GET eval | [Use cases — GET evaluation](flagr_use_cases.md#get-evaluation-browser-friendly) |
| Time/header targeting | [Built-in context injection](flagr_injected_context.md) |
| Run Flagr | [Self-hosting](flagr_self_host.md) |
| Env vars | [Environment variables](flagr_env.md) |
| GitOps flags | [JSON flag source](flagr_json_flag_spec.md) |
| Wrong variant | [Debug console](flagr_debugging.md) |
| Change Flagr itself | [Contributing](CONTRIBUTING.md) |