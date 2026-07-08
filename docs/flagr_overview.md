# Flagr overview

Every evaluation answers: **given this entity, which variant right now?** Segments, constraints, distribution, and rollout implement that answer. Evaluation is deterministic (same `entityID` → same bucket) and served from memory, not per-request SQL.

REST field names: [API reference](https://openflagr.github.io/flagr/api_docs). Code layout: [Contributing](CONTRIBUTING.md#where-the-code-lives).

## Concepts

For the authoritative field definitions, see the
[API doc](https://openflagr.github.io/flagr/api_docs).

The simplest thing Flagr knows about is a **flag**: a decision point in your
application. Behind it lives the question you want to answer at runtime — *who
gets what?* A single flag can act as a feature flag (on/off), an experiment
(control vs. treatment), or a piece of dynamic configuration (which JSON to
serve). It is the top-level unit you evaluate against, and everything else in
Flagr exists to refine the answer that flag returns. Flags are identified by a
unique `key` and can be turned off wholesale with `enabled: false`, which makes
evaluation return a blank result regardless of how the rest of the flag is
configured. You can attach **tags** to a flag — descriptive labels like
"frontend", "experiment", or "ops" — purely for lookup and filtering, so you
never have to encode team ownership or lifecycle stage into the flag key
itself.

A flag by itself does not know what it can return. That is the job of a
**variant**: one possible outcome of the flag, such as `control`/`treatment` or
`green`/`blue`/`pink`. Each variant carries an **Attachment** — an arbitrary
JSON object — for dynamic configuration. The same variant mechanism that picks a
button color can also carry the hex code, a copy string, or a timeout value, so
the client never has to branch on business logic: it branches on the
`variantKey` and reads whatever it needs from the attachment. The variant is
what turns an abstract decision point into a concrete thing your application can
act on.

> **Note:** "Variant Attachment" in these docs refers to the `Attachment` field
> on the `Variant` entity (type `map[string]any` — arbitrary JSON). There is no
> separate `VariantAttachment` type.

To decide which variant an entity receives, Flagr first needs to know *who* is
asking. That is the **entity**: the context you evaluate against, made up of
`entityID`, `entityType`, and `entityContext`. The `entityID` is what makes
evaluation **deterministic** — the same ID always hashes to the same bucket, so
a user sees the same variant on every request, across restarts, redeploys, and
load-balanced instances. The `entityContext` is a free-form JSON object that
constraints read from; this is where targeting attributes like `state`, `age`,
or `device` live. If no `entityID` is supplied, Flagr generates a random one,
which makes that particular evaluation non-sticky by design.

With outcomes (variants) and a subject (entity) in place, the next question is
*which entities should this flag apply to at all?* That is a **segment**: the
audience you want to target, and the smallest unit you can analyze in Flagr
Metrics. A segment is defined by a set of **constraints** connected by logical
`AND`, and segments are evaluated in rank order — an entity falls into the
**first** segment that matches, then evaluation stops. That ordering is how you
express "specific audiences first, everyone else last." A segment with no
constraints matches every entity, which is how you build a catch-all fallback
or a global rollout.

Constraints are the rules that make a segment *targeted* rather than global.
A **constraint** is a single comparison against a field in `entityContext`, such
as `state == "CA"` or `age >= 21`. All constraints in a segment must match for
the segment to match; a single miss means the entity falls through to the next
segment. Constraints let you say "only users in California, on mobile, over 21"
without writing that logic in your app — the targeting lives with the flag, not
scattered across call sites.

Once a segment matches, the last question is *which variant does this matched
entity actually get?* That is **distribution**: how matched entities are split
across variants within a segment. A `50/50` split of `control`/`treatment` is
the bread and butter of A/B testing; a `100/0` split is a rollout. Distribution
turns a segment from a *who* into a *who gets which*, and it works hand in hand
with **rollout** — the percentage of a matched segment that receives any
variant at all, letting you expose a flag to a sliver of traffic before going
wide.

### Constraint property access

Constraint `Property` supports nested field access via dotted and bracketed
syntax against `entityContext`:

| Syntax | Resolves to |
|--------|-------------|
| `state` | `entityContext["state"]` (flat, as before) |
| `user.name` | `entityContext["user"]["name"]` |
| `users[0]` | `entityContext["users"][0]` |
| `users[0].role` | `entityContext["users"][0]["role"]` |

Missing keys, out-of-bounds indices, and type mismatches evaluate to
**no match** (the segment does not match) rather than erroring. Negative
indices are supported (`users[-1]` = last element).

## Running example

The screenshots below walk one flag from a simple rollout to geo-targeted
segments and a full launch. Each UI image maps to the concepts above: variants
define outcomes, segments define *who*, constraints narrow the audience,
distribution splits variants, and rollout limits what fraction of a matched
segment actually receives a variant.

**Setup** — three button-color variants (`green`, `blue`, `pink`) for a new
checkout experience:

![Flag variants for the running example](images/flagr_running_example_1.png)
![Flag detail with segments and distributions](images/flagr_running_example_4.png)

**Step 1 — targeted rollout** — expose the flag to California users only
(`state == "CA"`), with a partial rollout before going wide:

![California segment with constraints](images/flagr_running_example_2.png)

**Step 2 — geo-specific preferences** — add segments per state (and combine
constraints, e.g. `state == "NY" AND age >= 21`):

![Multiple state-based segments](images/flagr_running_example_3.png)
![Per-segment distribution and rollout](images/flagr_running_example_5.png)

**Step 3 — experiment then launch** — A/B within a segment (`50/50` green/blue
at `20%` rollout), then raise rollout and distribution to ship `100%` green
globally:

![Raised rollout on a segment](images/flagr_running_example_7.png)
![Full launch distribution](images/flagr_running_example_6.png)

## Rollout and deterministic bucketing

The defining property of a good feature-flag engine is **stickiness**: the
same user must see the same variant on every request, even across restarts,
redeploys, and load-balanced instances. Flagr achieves this without storing
any per-user state. Instead it derives the variant **deterministically** from
the entity ID, so the answer is reproducible anywhere the same flag
configuration is loaded.

Given an entity context, evaluation works as follows:

1. **Hash** — `crc32.ChecksumIEEE(flagIDString + entityID)`.
2. **Modulo 1000** — the hash mod `1000` (the total bucket count) gives a
   bucket in `[0, 999]`.
3. **Distribution** — variants occupy contiguous bucket ranges. A `50/50`
   split of control/treatment means buckets `0–499` → control, `500–999` →
   treatment.
4. **Rollout** — the rollout percentage truncates the variant's bucket range.
   Within the variant's range, only the first `rolloutPercent` of buckets
   receive the variant; the rest fall through to the next segment (or no match).

> **Note:** The rollout percentage is applied **within whichever variant's
> bucket sub-range the entity hashed into**, not to a specific variant's full
> range. This means a low rollout can cause a matched segment to still not
> assign a variant, letting evaluation continue to later segments.

## Architecture

Flagr separates three concerns that scale differently:

| Concern | Path | Consistency | Latency goal |
|---------|------|-------------|--------------|
| **Read** (evaluation) | `POST` or `GET /evaluation` | Stale-tolerant (cache refresh) | Sub-ms hot path |
| **Write** (configuration) | CRUD APIs + UI | Strong (database) | Rare, not on eval path |
| **Record** (analytics) | `AsyncRecord` fan-out | Best-effort async | Must not block eval |

Diagrams below: database-backed (default) and eval-only JSON. When diagram and code disagree, **code wins**. Shared rules: [Behavioral contracts](contracts.md). Handler layout: [Contributing](CONTRIBUTING.md#where-the-code-lives).

```mermaid
flowchart TB
  subgraph clients [Clients]
    App[App / SDK]
    UI[Flagr UI]
  end

  subgraph evaluator [Evaluator — hot path]
    Eval["POST /evaluation"]
    EvalGet["GET /evaluation"]
    Exp["POST /exposures"]
    EC[(EvalCache)]
    Eval --> EC
    EvalGet --> EC
    Exp --> EC
  end

  subgraph manager [Manager — cold path]
    CRUD[CRUD APIs]
    DB[(Database)]
    FS[flag_snapshots]
    WH[Webhooks]
    UI --> CRUD
    CRUD --> DB
    CRUD --> FS
    CRUD -. webhooks after snapshot commit .-> WH
  end

  subgraph metrics [Metrics — async]
    AR[AsyncRecord fan-out]
    Stream[Kafka / Kinesis / Pub/Sub]
    DT[Datar eval only]
    AR --> Stream
    AR --> DT
  end

  subgraph gitops [Eval-only optional]
    JF[json_file]
    JH[json_http]
  end

  App --> Eval
  App --> EvalGet
  App --> Exp
  Eval --> AR
  Exp --> AR
  FS -. MAX id poll .-> EC
  DB -. reload on change .-> EC
  JF -. every poll .-> EC
  JH -. every poll .-> EC
```

### Paths

### How to read the diagram

- **Clients** call `POST` or `GET /evaluation` for assignment and optionally `POST /exposures` for impressions — [Exposure logging](flagr_exposure.md), [contracts](contracts.md#eval-vs-exposure).
- **Evaluator** uses **EvalCache** only on the hot path (no per-request SQL).
- **Manager** — CRUD + `flag_snapshot`; webhooks after commit.
- **Metrics** — async `AsyncRecord` to stream recorders or Datar.

### Request flows

**Evaluation** — `POST` or `GET /api/v1/evaluation` ([GET evaluation](flagr_use_cases.md#get-evaluation-browser-friendly)) → segments in rank order → variant → optional `AsyncRecord` (`recordSource: evaluation`). No stream when flag missing, disabled, or has no segments ([contracts](contracts.md#evalcache-freshness)).

**Configuration** — CRUD + snapshot → EvalCache reload on `MAX(flag_snapshot.id)` or poll interval.

**Exposure** — `POST /api/v1/exposures` → cache validation → `AsyncRecord` (`recordSource: exposure`).

**Eval-only** — [contracts — eval-only](contracts.md#eval-only).

### Components

**Evaluator** — batch/tag eval; `FLAGR_EVALCACHE_REFRESHINTERVAL` (default 3s); `GET /api/v1/flags/snapshots/max_id`.



**Manager** — CRUD, **`POST /flags/{flagID}/duplicate`**, **`commitFlagMutation`** (transaction + snapshot). UI: **Duplicate Flag**, **Delete Flag**.

**Metrics** — [contracts — recording](contracts.md#recording-gates); [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md).

## Related documentation

- [Behavioral contracts](contracts.md)
- [Use cases](flagr_use_cases.md) — flags, experiments, dynamic config
- [Exposure logging](flagr_exposure.md) and [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md)
- [Datar analytics](flagr_datar.md)
- [Environment variables](flagr_env.md)