<p align="center">
    <a href="https://github.com/openflagr/flagr/actions/workflows/ci.yml?query=branch%3Amain+" target="_blank">
        <img src="https://github.com/openflagr/flagr/actions/workflows/ci.yml/badge.svg?branch=main">
    </a>
    <a href="https://goreportcard.com/report/github.com/openflagr/flagr" target="_blank">
        <img src="https://goreportcard.com/badge/github.com/openflagr/flagr">
    </a>
    <a href="https://godoc.org/github.com/openflagr/flagr" target="_blank">
        <img src="https://img.shields.io/badge/godoc-reference-green.svg">
    </a>
    <a href="https://github.com/openflagr/flagr/releases" target="_blank">
        <img src="https://img.shields.io/github/release/openflagr/flagr.svg?style=flat&color=green">
    </a>
    <a href="https://codecov.io/gh/openflagr/flagr">
        <img src="https://codecov.io/gh/openflagr/flagr/branch/main/graph/badge.svg?token=iwjv26grrN">
    </a>
    <a href="https://deepwiki.com/openflagr/flagr">
        <img src="https://deepwiki.com/badge.svg?color=green" alt="Ask DeepWiki">
    </a>
</p>

## What is Flagr?

Flagr is an open-source Go service for **feature flags**, **A/B testing**, and
**dynamic configuration**. One primitive — the *flag* — covers all three: a
decision point in your code that the evaluation engine resolves at runtime
based on *who* is asking.

It exists to decouple *deploy* from *release* — turn a feature on for one user,
a thousand, or nobody, without redeploying. Run experiments and trust the
numbers. Change configuration without a code change or a restart.

`openflagr/flagr` is the community-driven home of Flagr, advancing development
beyond the original [`checkr/flagr`](https://github.com/checkr/flagr).

---

## 📖 Documentation

**[https://openflagr.github.io/flagr](https://openflagr.github.io/flagr)**

**Developers:** clone the repo and run **`make help`** for build, test, and UI commands (single entrypoint for CI and local work).

| Page | Content |
|------|---------|
| [Home](https://openflagr.github.io/flagr/) | Quick start, dev, testing, deploy |
| [Overview](https://openflagr.github.io/flagr/#/flagr_overview) | Concepts, running example, architecture |
| [Use Cases](https://openflagr.github.io/flagr/#/flagr_use_cases) | Feature flags, A/B testing, dynamic configuration |
| [Debug Console](https://openflagr.github.io/flagr/#/flagr_debugging) | UI evaluation testing |
| [Server Configuration](https://openflagr.github.io/flagr/#/flagr_env) | Environment variables, DB, auth, recorders |
| [JSON Flag Source](https://openflagr.github.io/flagr/#/flagr_json_flag_spec) | GitOps, JSON format, validator |
| [Notifications](https://openflagr.github.io/flagr/#/flagr_notifications) | Webhooks on flag changes |
| [Exposure Logging](https://openflagr.github.io/flagr/#/flagr_exposure) | `POST /exposures` API |
| [Data Recorders & A/B Analysis](https://openflagr.github.io/flagr/#/flagr_eval_exposure_pipeline) | Kafka, Kinesis, Pub/Sub; sample consumer; A/B analytics |
| [Datar Analytics](https://openflagr.github.io/flagr/#/flagr_datar) | In-process eval aggregates |
| [API Reference](https://openflagr.github.io/flagr/api_docs) | Swagger/OpenAPI spec |

---

## Features

- **Feature flags** — binary on/off, kill switches, targeted rollouts by audience
- **Duplicate flag** — clone variants, segments, constraints, distributions, and tags to a new flag (`POST /flags/{id}/duplicate` or **Flag Management** on flag detail)
- **A/B testing** — multi-variant experiments with deterministic, sticky assignment
- **Dynamic configuration** — per-variant JSON attachments, no redeploy or restart
- **GitOps / Flags-as-code** — load flags from JSON or HTTP; manage in Git, validate in CI, rollback with `git revert`
- **Exposure logging** — `POST /exposures` for client-reported impressions, the trustworthy A/B denominator
- **Webhook notifications** — HTTP POST on every flag change, with retry and backoff
- **Multi-database** — SQLite (dev), MySQL, PostgreSQL, and JSON sources
- **Vue 3 UI** — TypeScript management UI (Vite, typed REST via `ApiResult`); see `make help` for `build-ui` / `test-e2e`

## Quick start

```sh
docker pull ghcr.io/openflagr/flagr
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr

# Open the Flagr UI
open localhost:18000
```

Or try the hosted demo at
[https://try-flagr.onrender.com](https://try-flagr.onrender.com) (cold starts
may take a moment):

```sh
curl --request POST \
     --url https://try-flagr.onrender.com/api/v1/evaluation \
     --header 'content-type: application/json' \
     --data '{
       "entityID": "127",
       "entityType": "user",
       "entityContext": { "state": "NY" },
       "flagID": 1,
       "enableDebug": true
     }'
```

## Flagr UI

<p align="center">
    <img src="./docs/images/demo_readme.png" width="900">
</p>

## Architecture

Flagr has three components:

- **Evaluator** — serves evaluation from an in-memory cache of all flags. The
  cache refreshes periodically (default 3s) and short-circuits when nothing
  changed, so evaluation never touches the database on the request path.
- **Manager** — the CRUD gateway; all flag mutations flow through here.
- **Metrics** — fans evaluation and exposure events out to your data pipeline
  (Kafka, Kinesis, Pub/Sub) or the built-in Datar aggregates. Recording is
  asynchronous, so a slow backend never stalls an evaluation.

See the [architecture overview](https://openflagr.github.io/flagr/#/flagr_overview?id=architecture)
for diagrams, request flows, and the deterministic bucketing algorithm.

## Performance

Tested with [`vegeta`](./benchmark) — 2,000 req/s sustained, sub-millisecond
median latency:

```
Requests      [total, rate]            56521, 2000.04
Duration      [total, attack, wait]    28.26s, 28.26s, 365.53µs
Latencies     [mean, 50, 95, 99, max]  371.63µs, 327.99µs, 614.92µs, 1.39ms, 12.50ms
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:56521
```

## Client Libraries

| Language | Client |
| -------- | ------ |
| Go | [goflagr](https://github.com/openflagr/goflagr) |
| JavaScript | [jsflagr](https://github.com/openflagr/jsflagr) |
| Python | [pyflagr](https://github.com/openflagr/pyflagr) |
| Ruby | [rbflagr](https://github.com/openflagr/rbflagr) |

## License and Credit

- [`openflagr/flagr`](https://github.com/openflagr/flagr) — Apache 2.0
- [`checkr/flagr`](https://github.com/checkr/flagr) — Apache 2.0 (original project)