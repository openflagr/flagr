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

## Introduction

Flagr is an open source Go service that delivers the right experience to the right entity and monitors the impact. It provides feature flags, experimentation (A/B testing), and dynamic configuration — all behind clear swagger REST APIs for flag management and evaluation.

`openflagr/flagr` is the community-driven home of Flagr, advancing development beyond the original [`checkr/flagr`](https://github.com/checkr/flagr).

---

## 📖 Documentation

**[https://openflagr.github.io/flagr](https://openflagr.github.io/flagr)**

| Page | Content |
|------|---------|
| [Overview](https://openflagr.github.io/flagr/#/flagr_overview) | Concepts, running example, architecture |
| [Use Cases](https://openflagr.github.io/flagr/#/flagr_use_cases) | Feature flagging, A/B testing, dynamic configuration patterns |
| [Server Configuration](https://openflagr.github.io/flagr/#/flagr_env) | All environment variables, database drivers, auth, data recorders |
| [JSON Flag Source](https://openflagr.github.io/flagr/#/flagr_json_flag_spec) | GitOps workflows, JSON format spec, validator, CI integration |
| [Datar Analytics](https://openflagr.github.io/flagr/#/flagr_datar) | In-memory aggregate analytics engine |
| [Exposure Logging](https://openflagr.github.io/flagr/#/flagr_exposure) | Client-reported impressions for A/B testing |
| [Notifications](https://openflagr.github.io/flagr/#/flagr_notifications) | Webhook configuration and payload format |
| [API Reference](https://openflagr.github.io/flagr/api_docs) | Swagger/OpenAPI spec |

---

## Features

| Capability | Description |
|------------|-------------|
| **Feature flags** | Binary on/off toggles, kill switches, targeted rollouts by audience segment |
| **A/B testing** | Multi-variant experiments with deterministic distribution and rollout control |
| **Dynamic configuration** | Per-variant JSON attachments for runtime config without redeploy |
| **GitOps / Flags-as-code** | Load flags from JSON files or HTTP URLs. Manage flags in Git, validate in CI, rollback with `git revert` |
| **Datar analytics** | Built-in in-memory aggregate analytics — evaluation counts by variant, segment, and day. No external pipeline required |
| **Exposure logging** | `POST /exposures` for client-reported impressions; same **data recorders** as eval (`AsyncRecord`), with `recordSource: exposure` |
| **Webhook notifications** | HTTP POST webhooks on every flag create/update/delete/restore with retry and exponential backoff |
| **Multi-database** | SQLite (dev), MySQL, PostgreSQL, and JSON sources |
| **Eval cache** | In-memory cache with short-circuit reload — only refreshes when flag snapshots change |
| **Vue 3 UI** | Modern management UI built with Vite, Vue 3, and Element Plus |

## Quick start

```sh
docker pull ghcr.io/openflagr/flagr
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr

# Open the Flagr UI
open localhost:18000
```

Or try the hosted demo at [https://try-flagr.onrender.com](https://try-flagr.onrender.com) (cold starts may take a moment; every push to `main` triggers a redeploy):

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

Flagr has three core components:

- **Evaluator** — evaluates incoming requests against an in-memory `EvalCache` of all flags, segments, variants, constraints, and distributions. The cache refreshes periodically (default 3s) and short-circuits when no new snapshots exist.
- **Manager** — CRUD gateway for all flag mutations.
- **Metrics** — **Data recorders** for evaluation and exposure rows (`GetDataRecorder().AsyncRecord`). Kafka, AWS Kinesis, Google Pub/Sub, and built-in Datar. Exposures skip eval metrics and Datar aggregation.

See the [architecture overview](https://openflagr.github.io/flagr/#/flagr_overview) for the full diagram and evaluation algorithm.

## Performance

Tested with [`vegeta`](./benchmark) — 2,000 req/s sustained:

```
Requests      [total, rate]            56521, 2000.04
Duration      [total, attack, wait]    28.26s, 28.26s, 365.53µs
Latencies     [mean, 50, 95, 99, max]  371.63µs, 327.99µs, 614.92µs, 1.39ms, 12.50ms
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:56521
```

## Client Libraries

| Language   | Client                                          |
| ---------- | ----------------------------------------------- |
| Go         | [goflagr](https://github.com/openflagr/goflagr) |
| JavaScript | [jsflagr](https://github.com/openflagr/jsflagr) |
| Python     | [pyflagr](https://github.com/openflagr/pyflagr) |
| Ruby       | [rbflagr](https://github.com/openflagr/rbflagr) |

## License and Credit

- [`openflagr/flagr`](https://github.com/openflagr/flagr) — Apache 2.0
- [`checkr/flagr`](https://github.com/checkr/flagr) — Apache 2.0 (original project)
