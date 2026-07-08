# Behavioral contracts

Flagr has a few rules that every integrator needs to internalize once and then trust forever. These are the invariants that keep evaluations correct, data clean, and experiments trustworthy — the page other docs link to instead of repeating themselves. When something here seems to contradict another page, this one wins; when in doubt, the runtime behavior is the final authority.

---

## Eval vs exposure

The single most common source of bad experiment data is treating evaluation and exposure as the same event. They are not. Evaluation is the server's *decision* — which variant this entity gets right now. Exposure is the client's *observation* — the user actually saw the experiment surface. A user who is assigned a variant but never renders it should never land in the denominator of an A/B test. Keeping these two streams separate is what makes a flag's analytics mean anything at all.

| | **Evaluation** | **Exposure** |
|---|----------------|--------------|
| API | `POST /api/v1/evaluation` (and batch) | `POST /api/v1/exposures` |
| Meaning | Server **assignment** — which variant (if any) for this entity now | Client **impression** — user **saw** the experiment surface |
| Typical use | Branch in app code; cache `variantKey`, `flagSnapshotID` | A/B **denominator** after render / in-viewport |
| Stream tag | `recordSource: "evaluation"` | `recordSource: "exposure"` |

Never count evaluation volume as experiment participants. Evaluations can fire before render, on every navigation, and even for no-match blanks — none of those tell you anything about what a user *saw*.

The healthy client flow is **eval → render → exposure**, batching exposures on unload if needed. For the mechanics of each step, see [Exposure logging](flagr_exposure.md) and [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md).

---

## Recording gates

Recording is opt-in by design. Streaming flags, warehouse rows, and in-process counters all have a cost — network hops, storage, operational surface area — so Flagr makes you ask for each layer explicitly. The gates cascade: a global switch, a recorder-type allowlist, and a per-flag toggle. All three must agree before a single row leaves the process. This three-level design lets you turn recording on for one critical flag during an incident without flooding your stream with noise from every other flag in the system.

Streaming recorders (Kafka, Kinesis, Pub/Sub) and the in-process **Datar** engine both honor the same three gates:

1. `FLAGR_RECORDER_ENABLED=true`
2. Recorder listed in `FLAGR_RECORDER_TYPE` (e.g. `kafka`, `kinesis`, `pubsub`, `datar`)
3. Per-flag **`dataRecordsEnabled: true`** (UI or `PUT /api/v1/flags/{id}`)

**Datar** counts **evaluations only** — it deliberately skips any row tagged `recordSource: "exposure"`. If you need impressions in a warehouse for A/B analysis, pair a streaming recorder with the exposure API; Datar alone will not give you a denominator.

When the gates are off, valid exposure requests still return **200** with `loggedCount: 0`. The request succeeded; nothing was recorded. This keeps client code simple — it never has to distinguish "recording disabled" from "recording broken."

For wire format and consumer setup, see [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md). For the full environment variable reference, see [Environment variables](flagr_env.md#guide).

---

## Eval-only mode :id=eval-only

Not every deployment needs the full Flagr surface. Edge nodes serving evaluations at high volume, read-only replicas, and CI environments don't need the CRUD UI, the exposure endpoint, or even a database — they just need to answer "which variant?" fast. Eval-only mode strips the runtime down to exactly that. When the database driver is `json_file` or `json_http`, the server registers a minimal set of routes and skips everything else, keeping the footprint small and the attack surface narrow.

In this mode the server exposes only:

- `GET /api/v1/health`
- Evaluation APIs (`POST /evaluation`, batch, tag eval)
- `GET /api/v1/eval_cache/json` (export)

Everything else is gone: the CRUD UI, `POST /exposures`, Datar APIs, SQLite export, and the `flag_snapshot` short-circuit that normally avoids redundant reloads. In eval-only mode there is no database to snapshot, so the cache is re-fetched from the JSON source on every poll interval.

For the JSON flag source workflow, see [JSON flag source](flagr_json_flag_spec.md).

---

## EvalCache freshness :id=evalcache-freshness

Evaluation requests are the hottest path in Flagr, and they read from an in-memory cache, not the database. Every reload rebuilds the lookup maps from whatever the configured fetcher returns — a SQL query in database mode, a JSON file or HTTP endpoint in eval-only mode. The tradeoff for this speed is a small staleness window: after you change a flag, the cache doesn't know until the next reload. Understanding that window is the difference between a test that flakes and one that trusts what it sees.

The reload runs on a fixed interval set by `FLAGR_EVALCACHE_REFRESHINTERVAL` (default **3s**). In database-backed deployments there is a second, faster trigger: every API mutation that affects evaluation data creates a `flag_snapshot` row, and the cache checks whether that snapshot's max ID has advanced. If it hasn't, the reload is skipped — no point rebuilding identical maps. This short-circuit is what makes frequent polling cheap.

After a flag change, **`variantKey` may be blank** until the next reload lands. Wait at least one refresh interval before automated tests assert on new configuration. The integration test suite handles this with a **`waitForEvalReady`** helper that polls the evaluation endpoint directly — not the export endpoint — because only a real eval confirms the cache has been rebuilt.

A blank result (empty variant, no streaming record) means one of three things: the flag is missing, the flag has `enabled: false`, or the flag has no matching segments. All three are normal evaluation outcomes, not errors.

---

## Where to read more

| Topic | Page |
|-------|------|
| HTTP examples (eval, batch, exposures) | [Integration guide](integration.md) |
| Concepts, bucketing, diagram | [Overview](flagr_overview.md) |
| Exposure API & validation | [Exposure logging](flagr_exposure.md) |
| Recorders, frame, A/B SQL | [Data recorders & A/B analysis](flagr_eval_exposure_pipeline.md) |
| Deploy | [Self-hosting](flagr_self_host.md) |
| Env vars | [Environment variables](flagr_env.md) |