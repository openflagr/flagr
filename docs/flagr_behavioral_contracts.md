---
title: Behavioral contracts
---

# Behavioral contracts

These are the invariants integrators can rely on. Other pages link here instead of restating them. If another doc disagrees with this page, **this page wins**. If this page and the runtime disagree, **runtime wins** - file a docs bug.

---

## Eval vs exposure {#eval-vs-exposure}

Evaluation is the server's **assignment**. Exposure is the client's **impression** that the user actually saw the surface.

They answer different questions. Prefetches, disabled flags, and no-match blanks all produce evaluations; those rows are not "someone saw the treatment."

| | **Evaluation** | **Exposure** |
|---|----------------|--------------|
| API | `POST /api/v1/evaluation` (and batch; GET variants exist) | `POST /api/v1/exposures` |
| Meaning | Which variant (if any) for this entity **now** | User **saw** the experiment surface |
| Typical use | Branch in code; cache `variantKey`, `flagSnapshotID` | Impression after render / in-viewport |
| Stream tag | `recordSource: "evaluation"` | `recordSource: "exposure"` |

**When is eval volume enough?** Many cases only need assignment counts: kill switches, gradual rollouts, server-side config, ops dashboards, and rough traffic splits. Eval (or [Datar](flagr_datar.md)) is fine there.

**When do you need exposure?** For **rigid A/B experiments** where the denominator must be people who actually saw a surface (not prefetches, not never-rendered assignments), send explicit exposure after render. Mixing eval volume into that denominator is how lift gets diluted.

Healthy UI experiment flow when you need impressions: **eval → render → exposure** (batch exposures on unload if you defer). Mechanics: [Exposure logging](flagr_exposure.md), [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md).

---

## Segment evaluation {#segment-evaluation}

Segments run in **rank order** (lower rank first). For each segment:

1. If constraints miss (or there are constraints but no valid `entityContext` map), try the **next** segment.
2. If constraints match (or the segment has no constraints), run distribution + rollout for that segment, then **stop**. Later segments never run.

**Match means constraints matched**, not "got a non-empty `variantKey`."

Low rollout on a matched segment can still yield an empty `variantKey`. That is a holdout / partial rollout, not a fallthrough to the next segment. Put narrow audiences above catch-alls; do not rely on later segments to catch rollout misses.

Stickiness: send a stable **`entityID`**. If the client omits it, the evaluator injects a random id **before** bucketing (`randomly_generated_*` in `pkg/handler/eval.go`), so that request is non-sticky by design.

Bucketing algorithm (CRC32, 1000 buckets, in-range rollout): [Overview](flagr_overview.md#rollout-and-deterministic-bucketing). Source: `pkg/handler/eval.go` (`evalSegment`), `pkg/entity/distribution.go`.

---

## Recording gates {#recording-gates}

Recording is opt-in. Three gates must all pass before a row leaves the process:

1. `FLAGR_RECORDER_ENABLED=true`
2. Recorder listed in `FLAGR_RECORDER_TYPE` (e.g. `kafka`, `kinesis`, `pubsub`, `datar`)
3. Per-flag **`dataRecordsEnabled: true`** (UI or `PUT /api/v1/flags/{id}`)

Streaming recorders and **Datar** share those gates.

**Datar** counts **evaluations only**. It ignores `recordSource: "exposure"`. For warehouse denominators, pair a streaming recorder with the exposure API.

When gates are off, valid exposure requests still return **200** with `loggedCount: 0`. Success means "request accepted," not "row written."

Wire format and consumer setup: [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md). Env reference: [Environment variables](flagr_env.md#guide). Defaults live in `pkg/config/env.go`.

---

## Blank assignment vs stream rows {#blank-vs-stream}

An empty `variantKey` is a normal evaluation outcome (HTTP 200), not a transport failure. Whether a **stream row** is enqueued is a separate question. With [recording gates](#recording-gates) open:

| Outcome | `variantKey` | Stream row (`recordSource: evaluation`)? |
|---------|--------------|------------------------------------------|
| Flag missing / not found | blank | **No** (early return; no `logEvalResult`) |
| Flag `enabled: false` | blank | **No** |
| Flag has no segments | blank | **No** |
| Constraints never match any segment | blank | **Yes** (evaluator ran; useful for "assigned but never exposed" analysis) |
| Constraints match, rollout / distribution yields no variant | blank | **Yes** (`segmentID` stays `0` when no variant was chosen) |
| Variant assigned | set | **Yes** |

Exposure rows are independent: see [recording gates](#recording-gates) and [Exposure logging](flagr_exposure.md). Full matrix and frame shape: [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md).

---

## Eval-only mode {#eval-only}

**Usual path:** drivers `json_file` and `json_http` force **eval-only** mode (`setupEvalOnlyMode` in `pkg/config/config.go`). That is the supported GitOps / eval-edge shape.

`FLAGR_EVAL_ONLY_MODE=true` can also be set explicitly on other drivers; that is an edge case, not the normal product path. Prefer JSON drivers when you want eval-only.

Registered surface:

- `GET /api/v1/health`
- Evaluation APIs (`POST` / `GET /evaluation`, batch, tag eval)
- `GET /api/v1/export/eval_cache/json` (export)

Absent: CRUD UI, `POST /exposures`, Datar APIs, SQLite export, and the `flag_snapshot` short-circuit. There is no DB to snapshot, so EvalCache re-fetches the JSON source every poll interval.

JSON workflow: [JSON flag source](flagr_json_flag_spec.md). Route wiring: `pkg/handler/handler.go`.

---

## EvalCache freshness {#evalcache-freshness}

Hot-path evaluation reads **EvalCache** only. Reloads rebuild lookup maps from the configured fetcher (SQL, file, or HTTP).

- Interval: `FLAGR_EVALCACHE_REFRESHINTERVAL` (default **3s**)
- Fetch timeout: `FLAGR_EVALCACHE_REFRESHTIMEOUT` (default **59s**)

In **database** mode, each mutating API write creates a `flag_snapshot` row. The cache polls `MAX(flag_snapshot.id)` and skips rebuild when the max is unchanged. External consumers can also poll **`GET /api/v1/flags/snapshots/max_id`**. In **eval-only** mode there is no snapshot table, so every poll refetches.

After you change a flag, **`variantKey` may stay blank or stale** until the next reload. Automated tests should wait at least one refresh interval. This repo's integration suite uses **`waitForEvalReady`**, which polls a real evaluation (not the export endpoint) until the new config is live.

Blank assignment vs whether a stream row is written: [blank vs stream](#blank-vs-stream).

Source: `pkg/handler/eval_cache.go`, `pkg/config/env.go`.

---

## Where to read more

| Topic | Page |
|-------|------|
| HTTP examples (eval, batch, exposures) | [Integration guide](integration.md) |
| Concepts, bucketing, architecture | [Overview](flagr_overview.md) |
| Exposure API & validation | [Exposure logging](flagr_exposure.md) |
| Recorders, frame, A/B SQL | [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md) |
| Deploy | [Self-hosting](flagr_self_host.md) |
| Env vars | [Environment variables](flagr_env.md) |
