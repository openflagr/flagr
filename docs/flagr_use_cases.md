# Flagr use cases

Feature flags, A/B tests, and dynamic config look like three products. In Flagr they are one **flag** and one **`POST /evaluation`**. This page follows that single primitive as it grows: from a kill switch that flips a code path on or off, through a targeted rollout, into an A/B experiment that measures which variant wins, and finally into a dynamic-config carrier that ships values instead of booleans. Every step adds capability to the same flag — no new service, no new client call, no schema migration. The shape of the decision changes; the call does not.

A kill switch cares about *who is on*. An experiment cares about *what they saw* and whether it converted — same instrumentation, different analytics ([exposures](flagr_exposure.md) + your warehouse).


> **Note:** The pseudocode below uses the actual API response field names —
> `variantID`, `variantKey`, and `variantAttachment` (camelCase), as returned
> by `POST /evaluation` or `GET /evaluation` (same JSON body).

## GET evaluation (browser-friendly) :id=get-evaluation-browser-friendly

**POST `/api/v1/evaluation`** is the **primary** integration path: server-side apps, official SDKs, large `entityContext`, batch workloads, and anything that must not live in a URL. **GET** is **secondary** — the same `evalContext` / `evaluationBatchRequest` JSON, but carried in one query parameter so browsers and caches can treat evaluation like a normal read.

### Why GET exists (and why POST stays default)

Browsers treat `POST` + `Content-Type: application/json` as a **non-simple** request (CORS preflight). That blocks patterns like a static page calling Flagr directly, `<link rel="preload">` warming eval URLs, or sharing a bookmarkable eval link. [Issue #613](https://github.com/openflagr/flagr/issues/613) tracks that motivation; [PR #631](https://github.com/openflagr/flagr/pull/631) explored the API shape. Flagr ships GET as an **opt-in wire variant** with POST semantics, not a replacement for SDK/server traffic.

| | **POST (primary)** | **GET (secondary)** |
|--|-------------------|---------------------|
| **Best for** | App servers, SDKs, large or rich `entityContext`, batch eval, Debug Console | Browser `fetch` without preflight, preload, CDN/browser HTTP cache |
| **Payload** | JSON body — no URL length ceiling from Flagr | Single `json` query param (percent-encoded) — bounded by proxies and `FLAGR_EVAL_GET_MAX_URL_BYTES` |
| **Privacy** | Body usually absent from access logs and `Referer` | Full eval input often appears in **URLs** (logs, analytics, `Referer`, history, shared links) |
| **Caching** | Not cacheable by default | Cacheable by full URL — good for repeat reads, risky if URLs embed sensitive context |
| **Auth** | Same middleware as GET; body not in query strings | Same routes; whitelists often keep `/api/v1/evaluation` open — see [Environment Variables](flagr_env.md) |

**Use POST** when in doubt. **Use GET** only when you need CORS-simple or HTTP caching and your eval payload is small, non-sensitive, and stable when serialized.

### Wire format

| Method | Path | `json` decodes to |
|--------|------|-------------------|
| `GET` | `/api/v1/evaluation` | `evalContext` (same as POST body) |
| `GET` | `/api/v1/evaluation/batch` | `evaluationBatchRequest` (same as POST body) |

Responses match POST (`evalResult` / `evaluationBatchResponse`). The [Debug Console](flagr_debugging.md) still uses POST only.

```javascript
const ctx = {
  entityID: 'user-1',
  entityType: 'user',
  entityContext: { tier: 'premium' },
  flagID: 42,
};

// POST — preferred from backends and SDKs
await fetch('/api/v1/evaluation', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(ctx),
});

// GET — same object; only when browser/cache constraints require it
const url = `/api/v1/evaluation?json=${encodeURIComponent(JSON.stringify(ctx))}`;
await fetch(url);
```

Batch: `GET /api/v1/evaluation/batch?json=${encodeURIComponent(JSON.stringify(batchRequest))}`.

### Security and validation

Packing the whole request into **`json=`** concentrates user input in one place. That is convenient but **more exposed** than a POST body:

- **Logging and observability** — load balancers, reverse proxies, and APM often log **query strings**. Treat GET eval URLs like credentials-adjacent data: do not put secrets, tokens, or PII you would not put in a GET URL.
- **`Referer` and leakage** — if your page links to third parties, long eval URLs can leak context via the `Referer` header unless you use strict referrer policy.
- **Caching** — shared CDNs or browser caches key on the full URL. A stable `json` serialization is required for hits; accidental caching of personalized eval URLs can serve one user’s assignment to another. Scope cache TTLs deliberately.
- **`enableDebug: true`** — debug logs in the response; avoid on GET URLs that might be cached or forwarded.

**What Flagr validates (GET aligned with POST after `json` decode):**

- **Bind layer (all methods):** GET requires non-empty `json` query string (swagger bind); POST requires body and runs consume + validate on the body.
- **GET handlers (after `json.Unmarshal`):** same **`Validate` + `ContextValidate`** as POST `BindRequest` on `evalContext` / `evaluationBatchRequest` (e.g. `flagID` ≥ 1 when set, `flagTagsOperator` ∈ `ANY`|`ALL`, batch `entities` required with min length 1).
- **GET-only:** raw query length ≤ `FLAGR_EVAL_GET_MAX_URL_BYTES`.
- **Batch (GET and POST):** total evaluations ≤ `FLAGR_EVAL_BATCH_SIZE` when set.
- Invalid JSON syntax → **400** before schema checks.

Remaining differences: POST middleware may return **422**-style composite validation errors from go-swagger before the handler; GET returns **400** with `{ "message": "json is not valid evalContext: …" }` (or batch equivalent). Semantic eval outcomes (unknown flag, no segment match) are still **200** with an empty or partial result, not validation errors.

Because GET still puts the full request in the URL, keep payloads minimal and avoid sensitive fields even when validation passes. Review `FLAGR_BASIC_AUTH_WHITELIST_PATHS` and `FLAGR_JWT_AUTH_WHITELIST_PATHS` — evaluation is whitelisted by default so apps can call without tokens.

### URL length and `FLAGR_EVAL_GET_MAX_URL_BYTES`

Flagr measures the **raw query string** (`json=…` after URL encoding), not the full URL path. Default limit **`FLAGR_EVAL_GET_MAX_URL_BYTES=8192`** (`0` = disabled). Exceeding it returns **400** with a message to use POST.

**Is 8192 enough for normal eval?** Yes. Payloads shaped like integration tests and handler fixtures use only a **small fraction** of the cap (on the order of **~100–250** bytes raw query for single eval; **~200–500** for modest batch bodies). The default is not sized for “typical” traffic — it is a **safety rail** against oversized queries and aligns with many load balancers that allow multi-kilobyte query strings. Treat **~2,048 characters for the entire URL** (path + query + host) as a separate, stricter browser/proxy limit when Flagr sits behind older gateways.

**Typical raw query sizes** (measured from repo fixtures; `go test ./pkg/handler -run TestGetEvalQuerySizesDocumentsTypicalPayloads -v` logs the same):

| Fixture shape | Raw query length | Share of 8192 cap |
|---------------|------------------|-------------------|
| Integration eval (`tier: premium`, `flagID`) | **~155** | **~2%** |
| Handler unit test (`dl_state: CA`, `flagID`) | **~113** | **~1%** |
| Nested `entityContext.user` | **~181** | **~2%** |
| Several context fields + `enableDebug` | **~168–195** | **~2%** |
| Batch: 1 entity, 5 `flagIDs` | **~218** | **~3%** |
| Batch: 3 entities + `flagTags` | **~483** | **~6%** |
| **At Flagr cap** (integration probe) | **8192** | **100%** (`entityContext.blob` **8033** ASCII chars) |

So GET is appropriate for **constraint fields** (`state`, `tier`, `region`, nested paths you already use in segments), not for **large blobs** (serialized carts, feature vectors, JWT claims). Put an **~8 KB** ceiling on custom JSON in `entityContext` if you insist on GET; beyond that, **POST** (or shorten context: pass an id, resolve server-side).

Integration tests pin the exact cap boundary:

| Case | Raw query length | `entityContext.blob` length (chars) | HTTP |
|------|------------------|-------------------------------------|------|
| At limit | **8192** | **8033** | 200 |
| One byte over | **8193** | **8034** | 400 — `GET evaluation query length 8193 exceeds maximum of 8192; use POST` |

```json
{
  "entityID": "get-eval-url-limit-probe",
  "entityType": "user",
  "flagID": 1,
  "entityContext": { "blob": "<8033× 'a'>" }
}
```

### How other stacks enforce URL / header size (and why Flagr uses 8192)

There is **no single HTTP standard** that mandates **2048** or **8192** for every hop. Limits are enforced **per component**, on different units (request line, query string, full request headers, or entire URL). Flagr’s **`FLAGR_EVAL_GET_MAX_URL_BYTES`** applies only to the **raw query string** on GET eval routes; your **first failure** in production is often a **reverse proxy or load balancer**, not the Flagr process.

**Research note:** Automated web search in this environment (SearXNG `it` category) returned **no vendor-specific hits** for these limits; the table below is from **primary documentation** fetched directly (Apache, nginx, AWS, Go `net/http`, MDN).

| Layer | What is measured | Typical default / documented limit | How it is enforced |
|-------|------------------|----------------------------------|--------------------|
| **Flagr GET eval** | Raw `?…` query string only | **`8192`** bytes (`FLAGR_EVAL_GET_MAX_URL_BYTES`; `0` = off) | Handler returns **400** before eval |
| **Flagr HTTP server (Go)** | Entire request headers (request line + all header fields) | **`1 MiB`** — [`net/http.Server.MaxHeaderBytes`](https://pkg.go.dev/net/http#Server), default [`DefaultMaxHeaderBytes`](https://pkg.go.dev/net/http#pkg-constants) = `1 << 20`; Flagr go-swagger **`--max-header-size`** default **`1MiB`** (`swagger_gen/restapi/server.go`) | Connection read stops; **400** “request too large” |
| **Apache httpd** | HTTP request line (method + request-URI + protocol) | **`LimitRequestLine 8190`** ([Apache `core`](https://httpd.apache.org/docs/2.4/mod/core.html#limitrequestline)) | **414** if exceeded |
| **nginx** | Request line + header fields (per-buffer) | **`client_header_buffer_size 1k`**; overflow via **`large_client_header_buffers 4 8k`** ([`ngx_http_core_module`](https://nginx.org/en/docs/http/ngx_http_core_module.html#large_client_header_buffers)) — up to **four 8 KB** buffers for large lines/headers | **414** / **400** / connection close (config-dependent) |
| **AWS Application Load Balancer** | Request line; single header; all request headers | Request line **16 KB**; one header **16 KB**; **entire request header block 64 KB** ([ALB quotas — HTTP headers](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-limits.html#http-headers-quotas)) | ALB rejects request before it reaches targets |
| **HTTP semantics** | URI the server will interpret | No fixed number in RFC; servers may respond [**414 URI Too Long**](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status/414) | Status **414** (or vendor-specific **400** / **431**) |

**Where do “2048” and “8192” come from?**

- **~8192 on the request line / URI** is a **common server default**, not a universal law. **Apache’s default `LimitRequestLine` is 8190** (bytes on the request line, which includes the path and query). That is why many operators treat **~8 KB** as a safe upper bound for **GET URLs** when the app sits behind Apache or similar configs. Flagr’s **8192 on the query string alone** is in the same ballpark but **stricter in meaning**: it does not count `/api/v1/evaluation?` or other headers—only `json=…`.
- **~2048** is often cited for **legacy browsers** (historical IE) and **old proxies** limiting **total URL length** for the bar/bookmark UI, not for modern app servers. It is still worth remembering for **browser-only** GET eval: a URL that passes Flagr and nginx may **fail in the client** or in an ancient corporate proxy. Do not size production GET eval for 2048 unless you truly control every client; do treat **2–4 KB total URL** as a **conservative design target** for public web pages.
- **Go does not cap “URL length” separately.** `MaxHeaderBytes` limits the **whole header block** (request line + headers). A long `?json=` query counts toward that **1 MiB** budget on direct-to-Flagr traffic, which is why Flagr adds an **application-level** cap on the query string so misconfigured clients fail with a clear **“use POST”** message instead of a generic header error.

**Ingress / Kong / Cloudflare / Kubernetes:** There is **no one K8s limit**—`Ingress` controllers delegate to **nginx**, **Envoy** (Istio/Contour), **HAProxy**, etc., each with its own buffer directives. **Kong** (OpenResty/nginx-based data plane) inherits **nginx-style** header buffer behavior unless you tune it. **Cloudflare** and other CDNs publish limits in their own docs (not verified here); assume the **narrowest** hop wins: measure end-to-end with your actual edge + ingress + Flagr chain.

**Practical takeaway:** Flagr’s default **8192** aligns with **Apache’s ~8K request-line culture** and is **far below** Go’s **1 MiB** header ceiling and **ALB’s 16 KB request line**, so for typical deployments **Flagr’s limit bites first** on GET eval—by design. If you raise `FLAGR_EVAL_GET_MAX_URL_BYTES`, re-check **nginx `large_client_header_buffers`**, **Apache `LimitRequestLine`**, and **ALB 16K request line** before relying on longer GET URLs.

CDNs and browsers cache GET by full URL. Use **stable JSON serialization** (consistent key order, no pretty-print) so equivalent requests share a cache key — and only when caching is safe for your data.

```bash
curl -sS -X POST "http://localhost:18000/api/v1/evaluation" \
  -H 'Content-Type: application/json' \
  -d '{"entityID":"user-1","entityType":"user","entityContext":{"tier":"premium"},"flagKey":"my-feature"}'
```

Response field names are camelCase: **`variantKey`**, **`variantAttachment`**, **`evalContext`**.

## Feature flagging

The simplest decision in software is *should this code path run?* A feature flag is that decision, externalized so you can change the answer without redeploying. The most familiar shape is the **kill switch** — a globally visible "off" you can throw the instant a deploy goes wrong, without paging an engineer to roll back. But the same mechanism generalizes immediately to **targeted rollouts**: turn a feature on for internal users first, then a single region, then everyone. The key property is that *deploy* and *release* become independent — you can ship code to production hours before any user sees it.

### The simple boolean flag (a starting template)

A simple boolean flag is the template most teams start with: two variants (`on`/`off`) and a single segment distributing 100% to `on`. It answers the question "is this feature on?" — nothing more.

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

Given an entity (user, request, or cookie), your application evaluates the flag and branches on the assigned variant:

```js
evaluation_result = flagr.postEvaluation(entity)

if (evaluation_result.variantKey == "on") {
    // feature is on for this entity
} else {
    // feature is off (variantKey is empty, or the flag is disabled)
}
```

This is just a convention, not a special flag type. The same flag can grow richer as your needs evolve — without changing your application code:

- **Target a specific audience** — add constraints to the segment (`state == "CA"`, `tier == "beta"`). Now `on` reaches only California users.
- **Roll out gradually** — lower the rollout percent to 10%, then 50%, then 100%. Same flag, same variants, wider audience over time.
- **Run an experiment** — add more variants and split the distribution (`33/33/34`). The boolean flag becomes an A/B test.
- **Serve dynamic config** — attach JSON to each variant (`{"color": "#42b983"}`). The flag now carries configuration, not just on/off.

Every flag also has a top-level `enabled` toggle — separate from the variant pattern above. When `enabled` is `false`, the evaluator returns a blank result immediately, short-circuiting before any segment logic runs (`PUT /api/v1/flags/{id}/enabled` or the UI status switch). Use it as a global kill switch to turn off the entire flag regardless of segments or distributions.

![feature flagging setting demo](/images/demo_ff.png)

To copy an existing flag’s segments, variants, and tags into a new flag (for example to fork an experiment or reuse a rollout pattern), use **`POST /api/v1/flags/{flagID}/duplicate`** or **Duplicate Flag** on the flag detail page in the UI. The clone gets a new key and ` (cloned)` in the description by default; optional `key` and `description` in the API body override those defaults.

## Experimenting — A/B testing

So far the flag answers a binary question: *on or off?* The moment you add a second `on`-like variant and split the audience between them, the question shifts to *which of these is better?* That is the jump from feature flagging to experimentation — and it is the same flag, evaluated by the same `POST /evaluation`, only with more variants and a finer distribution. Flagr's job is the exposure half: deterministic, sticky assignment so the same user always sees the same treatment. The measurement half (conversion rates, significance) is yours; Flagr emits the events but does not compute the verdict.

### Variants in experiments: control and treatment

The names `control` and `treatment` are **convention, not a Flagr contract**. Flagr treats every variant the same — `control` is no more special than `on`, `green`, or `version_b`. The convention comes from experimentation methodology:

- **Control** — the *baseline* experience. Usually the current production behavior. This is what you compare against. If your experiment is "new checkout button," control is the existing button.
- **Treatment** — the *change* you are testing. There can be one or many: `treatment1`, `treatment2`, `treatment3`. Each is a distinct alternative you're measuring against the control.

The distinction matters in your **analytics**, not in Flagr. When you group exposure rows by `variantKey` in your warehouse, the control group is your baseline conversion rate; each treatment's rate is compared against it to measure lift. If you have no control (e.g. testing two brand-new designs), pick one variant as the reference — but the convention of naming a baseline "control" makes the analysis self-documenting.

> **Note:** Flagr does not enforce that a flag has a control variant. You can run a flag with only treatments, or with no variant named "control" at all. The names are for you and your analytics pipeline; Flagr only assigns and records the `variantKey` you configured.

To run an A/B test across several variants with a targeted audience, instrument the code the same way and branch on the assigned variant:

```js
evaluation_result = flagr.postEvaluation(entity)

if (evaluation_result.variantKey == "control") {
    // Control: show the current production checkout (the baseline)
} else if (evaluation_result.variantKey == "treatment1") {
    // Treatment 1: show the new single-page checkout
} else if (evaluation_result.variantKey == "treatment2") {
    // Treatment 2: show the new accordion checkout
} else if (evaluation_result.variantKey == "treatment3") {
    // Treatment 3: show the new one-click checkout
}
```

> **Warning:** Segment order matters. An entity falls into the **first** segment whose constraints **all** match. List targeted segments before broad ones.

A typical A/B test flag in the UI:

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

### Measuring experiment outcomes

Assignment from `POST /evaluation` is **not** enough to measure conversion — you need **impression** events. Use [Exposure logging](flagr_exposure.md) when the user actually sees the treatment, then [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md) (or your own consumer on Kafka/Kinesis/Pub/Sub) to build denominators and join to business metrics. For quick **eval volume** only, see [Datar analytics](flagr_datar.md).

## Dynamic configuration

The progression ends where the flag stops deciding *which experience* and starts carrying *what value* — the threshold for a cache, the copy on a button, the timeout for a retry. You could redeploy for each change, or you can let the variant carry the value. A variant's **Attachment** is an arbitrary JSON object delivered alongside the `variantKey`, so your code reads a configuration value the same way it reads an assignment — no second fetch, no separate config service, no restart. The evaluation call is identical to the kill switch you started with; only the shape of what comes back has widened.

Use a variant's **Attachment** (an arbitrary JSON object) to carry dynamic configuration:

```js
evaluation_result = flagr.postEvaluation(entity)
green_color_hex = evaluation_result.variantAttachment["color_hex"]
```

A typical dynamic-config flag:

```
Variants
  - green
    - attachment: {"color_hex": "#42b983"}   // or {"color_hex": "#008000"}
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

> **Note:** Prior to [v1.1.3](https://github.com/openflagr/flagr/releases/tag/1.1.3), attachments supported only `string:string` key/value pairs. Current Flagr stores attachments as `map[string]any`, so arbitrary JSON values are supported. (The historical boundary is documented in the GitHub release notes; the in-repo `CHANGELOG.md` links there.)