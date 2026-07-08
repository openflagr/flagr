# Flagr use cases

Feature flags, A/B tests, and dynamic config often get sold as three products. In Flagr they are one **flag** and one evaluation call (`POST /api/v1/evaluation` by default). You start with a kill switch, add segments for rollouts, split variants for experiments, and hang JSON on variants when you need config. Same client call the whole way.

A rollout cares who is on. An experiment cares what they saw and whether they converted. Flagr assigns; your app renders; [exposure logging](flagr_exposure.md) and your warehouse handle the rest.

| If you are building… | Start here |
|----------------------|------------|
| On/off or gradual rollout | [Feature flagging](#feature-flagging) |
| UI or server A/B | [Experimenting](#experimenting-ab-testing) |
| Tunables without redeploys | [Dynamic configuration](#dynamic-configuration) |
| Browser or cache-friendly eval | [GET evaluation](#get-evaluation-browser-friendly) |

> **Note:** Examples use API field names `variantID`, `variantKey`, and `variantAttachment` (camelCase), as returned by `POST` or `GET /evaluation`.

## Feature flagging

The smallest useful question is whether a code path runs. A **kill switch** flips that answer from the UI without a redeploy. The same flag scales to **targeted rollouts**: staff first, one region, then everyone. Deploy and release stop being the same event.

### Boolean on/off template

Most teams begin with two variants (`on` / `off`) and one segment at 100% `on`:

```
Variants
  - on
  - off

Segment
  - Constraints: none (everyone)
  - Rollout Percent: 100%
  - Distribution
    - on: 100%
    - off: 0%
```

Evaluate once per entity and branch on `variantKey`:

```js
const result = await flagr.postEvaluation(entity);

if (result.variantKey === "on") {
  // enabled for this entity
} else {
  // off, or no assignment (empty variantKey)
}
```

That pattern is convention, not a special flag type. You can grow the same flag without changing the call site:

- **Audience** — constraints on `entityContext` (`state == "CA"`, `tier == "beta"`).
- **Gradual rollout** — lower segment rollout percent (10% → 50% → 100%).
- **Experiment** — more variants and a split distribution.
- **Config** — JSON on each variant via **Attachment**.

The flag-level **`enabled`** switch is separate. When `enabled` is `false`, evaluation returns blank before segments run (`PUT /api/v1/flags/{id}/enabled` or the UI toggle). Use it as a global off for the whole flag.

![feature flagging setting demo](/images/demo_ff.png)

To fork segments, variants, and tags, use **`POST /api/v1/flags/{flagID}/duplicate`** or **Duplicate Flag** in the UI. The clone gets a new key and ` (cloned)` in the description unless you override `key` / `description` in the body.

## Experimenting — A/B testing

Add a second "on-like" variant and the question becomes *which experience wins?* Same flag, same `POST /evaluation`, more variants and a finer distribution. Flagr gives you sticky assignment (`entityID` + unchanged flag → same `variantKey`). Significance and conversion math stay in your stack; Flagr can emit events if you wire [recorders](flagr_eval_exposure_pipeline.md).

### Control and treatment (naming)

`control` and `treatment` are **your** labels, not Flagr types. Every variant is equal in the evaluator. In analytics, control is usually the baseline; treatments are alternatives. Name a baseline `control` so warehouse queries stay obvious.

> **Note:** Flagr does not require a `control` variant. It only records the `variantKey` you configured.

```js
const result = await flagr.postEvaluation(entity);

if (result.variantKey === "control") {
  // baseline checkout
} else if (result.variantKey === "treatment1") {
  // single-page checkout
} else if (result.variantKey === "treatment2") {
  // accordion checkout
}
```

> **Warning:** Segments run in order. The **first** segment whose constraints all match wins. Put narrow rules above catch-alls.

Example layout:

```
Variants
  - control
  - treatment1
  - treatment2
  - treatment3

Segment                         // state == "CA"
  - Constraints (state == "CA")
  - Rollout Percent: 20%
  - Distribution
    - control:    25%
    - treatment1: 25%
    - treatment2: 25%
    - treatment3: 25%

Segment                         // state == "NY" AND age >= 21
  - Constraints (state == "NY" AND age >= 21)
  - Rollout Percent: 100%
  - Distribution
    - control:    50%
    - treatment1:  0%
    - treatment2: 25%
    - treatment3: 25%
```

![ab testing setting demo 1](/images/demo_exp1.png)
![ab testing setting demo 2](/images/demo_exp2.png)

### Measuring outcomes

`POST /evaluation` alone does not give you a conversion denominator. Log an impression when the user **sees** the treatment ([Exposure logging](flagr_exposure.md)), then pipe rows through [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md) or your own consumer. For eval volume only, [Datar](flagr_datar.md) is enough.

## Dynamic configuration

Sometimes the flag carries a value, not a branch: cache TTL, button copy, retry timeout. Each variant's **Attachment** is arbitrary JSON next to `variantKey`. One evaluation call; no second config service.

```js
const result = await flagr.postEvaluation(entity);
const colorHex = result.variantAttachment["color_hex"];
```

Example flag:

```
Variants
  - green
    - attachment: {"color_hex": "#42b983"}
  - red
    - attachment: {"color_hex": "#ff0000"}

Segment
  - Constraints: null
  - Rollout Percent: 100%
  - Distribution
    - green: 100%
    - red: 0%
```

![dynamic configuration demo](/images/demo_dynamic_configuration.png)

> **Note:** Before [v1.1.3](https://github.com/openflagr/flagr/releases/tag/1.1.3), attachments were `string:string` maps. Current Flagr uses `map[string]any` for arbitrary JSON.

## GET evaluation (browser-friendly) :id=get-evaluation-browser-friendly

**POST** stays the default for servers, SDKs, and large `entityContext`. **GET** is the same `evalContext` or batch body, URL-encoded in a single `json=` query param, when you need a CORS-simple request or HTTP caching in the browser. Motivation: [issue #613](https://github.com/openflagr/flagr/issues/613); shape discussion: [PR #631](https://github.com/openflagr/flagr/pull/631). The [Debug Console](flagr_debugging.md) still uses POST only.

| | POST (primary) | GET (secondary) |
|--|----------------|-----------------|
| Best for | Backends, SDKs, rich context, batch | Browser `fetch` without preflight, preload, shared links |
| Payload | JSON body | `?json=` (length capped; see below) |
| Privacy | Body rarely logged in URLs | Full request often in access logs, `Referer`, history |
| Caching | Not by default | By full URL when safe |

Use GET only with small, non-sensitive context and stable JSON serialization. When unsure, POST.

### Wire format

| Method | Path | `json` decodes to |
|--------|------|-------------------|
| `GET` | `/api/v1/evaluation` | `evalContext` (same as POST body) |
| `GET` | `/api/v1/evaluation/batch` | `evaluationBatchRequest` |

```javascript
const ctx = {
  entityID: 'user-1',
  entityType: 'user',
  entityContext: { tier: 'premium' },
  flagID: 42,
};

await fetch('/api/v1/evaluation', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(ctx),
});

const url = `/api/v1/evaluation?json=${encodeURIComponent(JSON.stringify(ctx))}`;
await fetch(url);
```

Batch: `GET /api/v1/evaluation/batch?json=${encodeURIComponent(JSON.stringify(batchRequest))}`.

### Security and validation

Everything lives in **`json=`**, so treat URLs like sensitive query strings. Avoid secrets and PII you would not put in a GET. Watch **caching**: personalized URLs can return the wrong assignment if a CDN or browser caches them. Skip `enableDebug: true` on GET URLs you might cache or forward.

After decode, GET runs the same **`Validate` / `ContextValidate`** as POST on `evalContext` and batch bodies. GET also enforces raw query length ≤ **`FLAGR_EVAL_GET_MAX_URL_BYTES`** (default **8192**; `0` disables). Invalid JSON → **400**. Unknown flags and non-matching segments still → **200** with empty or partial results. POST may surface **422** from go-swagger on bind; GET uses **400** with a `json is not valid …` message.

Review auth whitelists in [Environment variables](flagr_env.md): evaluation is often open by default.

### URL length (`FLAGR_EVAL_GET_MAX_URL_BYTES`)

Flagr counts the **raw query string** only (`json=…` after encoding), not the full URL. Over the cap → **400** with a message to use POST.

Typical fixtures in this repo land around **~100–250** bytes for a single eval and **~200–500** for a modest batch (a few percent of the 8192 default). The default is a guardrail aligned with common **~8 KB** request-line limits (e.g. Apache `LimitRequestLine` 8190), not a sizing target for normal traffic. For public browser-only pages, **2–4 KB total URL** is still a sane conservative design limit on old clients and proxies.

| Fixture shape | Raw query length | Share of 8192 |
|---------------|------------------|---------------|
| Integration eval (`tier: premium`, `flagID`) | ~155 | ~2% |
| Handler test (`dl_state: CA`) | ~113 | ~1% |
| Nested `entityContext.user` | ~181 | ~2% |
| Batch: 1 entity, 5 `flagIDs` | ~218 | ~3% |
| Batch: 3 entities + `flagTags` | ~483 | ~6% |
| At cap (integration probe) | 8192 | 100% |

GET fits segment fields (`state`, `tier`, `region`), not multi-kilobyte blobs in `entityContext`. Integration tests pin the boundary: **8033** ASCII chars in `entityContext.blob` → raw query **8192** (200); one byte more → **400**.

Other hops (nginx header buffers, ALB 16 KB request line, Go `MaxHeaderBytes` 1 MiB) can fail earlier or later depending on your edge. If you raise `FLAGR_EVAL_GET_MAX_URL_BYTES`, check ingress and load balancers before relying on longer GET URLs. Reproduce sizes with `go test ./pkg/handler -run TestGetEvalQuerySizesDocumentsTypicalPayloads -v`.

When you cache GET responses, serialize JSON consistently (key order, no pretty-print) and only cache non-personalized URLs.

```bash
curl -sS -X POST "http://localhost:18000/api/v1/evaluation" \
  -H 'Content-Type: application/json' \
  -d '{"entityID":"user-1","entityType":"user","entityContext":{"tier":"premium"},"flagKey":"my-feature"}'
```

Response fields remain camelCase: `variantKey`, `variantAttachment`, `evalContext`.

## Where to go next

| Goal | Doc |
|------|-----|
| HTTP details and batch | [Integration guide](integration.md) |
| Eval vs exposure, cache lag | [Behavioral contracts](contracts.md) |
| Segment and bucketing concepts | [Overview](flagr_overview.md) |
| Env vars (`FLAGR_EVAL_GET_MAX_URL_BYTES`, etc.) | [Environment variables](flagr_env.md) |