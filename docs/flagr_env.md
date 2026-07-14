# Server configuration

Flagr has **no config file**. Every knob is an environment variable bound at startup to one struct: `pkg/config/env.go`. That struct is the source of truth. When this page and the code disagree, the code wins.

This page embeds `pkg/config/env.go` from the repo tree at docs build time (every `env` tag and default), then a short operator guide for the variables you actually touch. Niche knobs may appear only in the source block.

Deploy recipes: [Self-hosting](flagr_self_host.md).

## Source (`pkg/config/env.go`) {#source-pkgconfigenvgo}

The block below is the checked-in source at the commit used to build the docs site. You can also open it on GitHub.

[Open on GitHub](https://github.com/openflagr/flagr/blob/main/pkg/config/env.go)

<<< @/snippets/env.go{go}

## Quick start

The fastest path to a running server is the [Self-hosting](flagr_self_host.md) guide, which covers Docker, Compose, and Kubernetes. If you just want the minimal environment for a MySQL-backed server, copy these four variables and go:

```sh
export HOST=0.0.0.0
export PORT=18000
export FLAGR_DB_DBDRIVER=mysql
export FLAGR_DB_DBCONNECTIONSTR='user:pass@tcp(127.0.0.1:3306)/flagr?parseTime=true'
```

If you'd rather serve flags from a static JSON file or URL with no database at all, set `FLAGR_DB_DBDRIVER` to `json_file` or `json_http`. That puts the server into eval-only mode automatically - see [behavioral contracts - eval-only](flagr_behavioral_contracts.md#eval-only) and the [JSON flag source](flagr_json_flag_spec.md) spec.

## Guide

### Server & HTTP

How the process binds, what it serves, and what it logs. Defaults are fine for local `./flagr`; containers should set `HOST=0.0.0.0` (the official Dockerfile already does).

| Variable | Default | Notes |
|----------|---------|--------|
| `HOST` / `PORT` | `localhost` / `18000` | Bind address (`env.go`); Docker image sets `HOST=0.0.0.0` |
| `FLAGR_WEB_PREFIX` | *(empty)* | UI + API base path |
| `FLAGR_UI_ENABLED` | `true` | `false` = API-only |
| `FLAGR_LOGRUS_LEVEL` / `FORMAT` | `info` / `text` | Use `json` in production |
| `FLAGR_PPROF_ENABLED` | `true` | pprof endpoints |
| `FLAGR_MIDDLEWARE_VERBOSE_LOGGER_*` | on | Exclude hot paths via `…_EXCLUDE_URLS` |
| `FLAGR_MIDDLEWARE_GZIP_ENABLED` | `true` | |

CORS lives under `FLAGR_CORS_*`, enabled by default with permissive origins. Full list is in the source block; tighten it only for browser-facing lockdowns.

### Evaluation & cache

Evaluation never hits the database on the hot path. It reads an in-memory EvalCache rebuilt on a fixed interval.

| Variable | Default | Notes |
|----------|---------|--------|
| `FLAGR_EVALCACHE_REFRESHINTERVAL` | `3s` | EvalCache reload period |
| `FLAGR_EVALCACHE_REFRESHTIMEOUT` | `59s` | Single fetch timeout |
| `FLAGR_EVAL_DEBUG_ENABLED` | `true` | + `enableDebug` on request → segment logs ([Debug console](flagr_debugging.md)) |
| `FLAGR_EVAL_BATCH_SIZE` | `0` | `0` = unlimited batch eval (POST and GET batch) |
| `FLAGR_EVAL_GET_MAX_URL_BYTES` | `8192` | GET `json=` raw query cap; `0` = off - [use cases](flagr_use_cases.md#get-evaluation-browser-friendly) |
| `FLAGR_EXPOSURE_BATCH_SIZE` | `100` | Max rows per `POST /exposures` |

After a flag change, **`variantKey`** can stay blank or stale until the next reload. That lag is a contract, not a bug. See [EvalCache freshness](flagr_behavioral_contracts.md#evalcache-freshness). Automated tests should wait at least one interval (this repo uses **`waitForEvalReady`**).

Eval-only is the usual product path when `FLAGR_DB_DBDRIVER` is `json_file` or `json_http` (`setupEvalOnlyMode` in `pkg/config/config.go`). `FLAGR_EVAL_ONLY_MODE=true` can be set on other drivers as an edge case; prefer JSON drivers for eval-edge deploys. Surface: [behavioral contracts: eval-only](flagr_behavioral_contracts.md#eval-only).

#### Built-in context injection

| Variable | Default | Notes |
|----------|---------|--------|
| `FLAGR_INJECTED_CONTEXT_ENABLED` | `false` | Merge `@ts*` and `@http_*` into `entityContext` before eval |
| `FLAGR_INJECTED_CONTEXT_HTTP_HEADERS` | `""` | Comma-separated headers → `@http_*` keys |
| `FLAGR_INJECTED_CONTEXT_HTTP_HEADER_PREFIXES` | `""` | Prefix match (e.g. `CF-` for Cloudflare) |

Full guide: [Built-in context injection](flagr_injected_context.md).

#### Eval cache export {#eval-cache-export}

A running server can dump its in-memory cache as JSON via `GET /api/v1/export/eval_cache/json`, with optional `enabled`, `ids`, `keys`, `tags`, and `tagsOperator` (`ANY` / `ALL`) query parameters.

### Database

Two variables decide where flags live: the driver and the connection string. Defaults are local SQLite; production typically uses MySQL or Postgres. JSON drivers load flags from a file or URL for read-only eval.

| Variable | Default |
|----------|---------|
| `FLAGR_DB_DBDRIVER` | `sqlite3` |
| `FLAGR_DB_DBCONNECTIONSTR` | `flagr.sqlite` |

| Driver | Role |
|--------|------|
| `sqlite3` | Local dev (default) |
| `mysql` / `postgres` | Production |
| `json_file` / `json_http` | Flags from file or URL ([JSON spec](flagr_json_flag_spec.md)) |


### Authentication

Authentication is off by default, so a freshly started server is open until you turn something on. Flagr supports two layers that can be used independently: basic auth, which guards the UI, and JWT auth, which guards the API. Both layers let you whitelist paths so hot evaluation traffic doesn't have to authenticate - and both ship with defaults that leave `/api/v1/evaluation` and `/api/v1/exposures` open, so turning auth on won't break your integration.

A minimal basic-auth setup is three variables:

```sh
FLAGR_BASIC_AUTH_ENABLED=true
FLAGR_BASIC_AUTH_USERNAME=admin
FLAGR_BASIC_AUTH_PASSWORD=password
```

JWT is richer. The variables cover enabling it (`FLAGR_JWT_AUTH_ENABLED`), the shared secret or PEM key (`FLAGR_JWT_AUTH_SECRET`), the signing method (`HS256` / `HS512` / `RS256`), and a set of prefix and exact whitelist paths. All of them are in the source above. JWT tokens can arrive by cookie or by `Authorization: Bearer` header; when both are present, the header wins.

Separately, Flagr can identify *who* made a mutation for audit logging without doing full authentication. `FLAGR_HEADER_AUTH_*` reads a user identifier from a header (handy behind a corporate proxy), and `FLAGR_COOKIE_AUTH_*` reads one from a cookie (handy behind something like Cloudflare Zero Trust). These stamp `created_by` / `updated_by` on changes; they don't gate access.

One thing worth calling out: the default JWT whitelist allows unauthenticated exposure logging. If the integrity of your impression stream matters, narrow the whitelist to lock down `/api/v1/exposures` and rate-limit it at the edge. The [Exposure logging](flagr_exposure.md) page walks through the tradeoffs.

### Data recorders {#data-record-destinations}

Recording gates (master switch, recorder type, per-flag `dataRecordsEnabled`): [behavioral contracts: recording gates](flagr_behavioral_contracts.md#recording-gates). Blank assignment vs whether a row is written: [blank vs stream](flagr_behavioral_contracts.md#blank-vs-stream).

`FLAGR_RECORDER_TYPE` is a comma-separated list so you can combine recorders.

| `FLAGR_RECORDER_TYPE` | Doc |
|------------------------|-----|
| `kafka`, `kinesis`, `pubsub` | Eval + exposure stream - [Recorders & A/B](flagr_eval_exposure_pipeline.md) |
| `datar` | In-process eval counts only - [Datar](flagr_datar.md) (no exposures) |

Streaming recorders ship eval and exposure rows to a broker; Datar keeps in-process evaluation counts and flushes them to the DB. Combining `kafka,datar` is common: live stream plus cheap dashboards. `FLAGR_RECORDER_FRAME_OUTPUT_MODE`: `payload_string` stringifies the payload (and respects encryption); `payload_raw_json` embeds the object (and ignores encryption).

The minimal Kafka setup is four variables:

```sh
FLAGR_RECORDER_ENABLED=true
FLAGR_RECORDER_TYPE=kafka
FLAGR_RECORDER_KAFKA_BROKERS=kafka1:9092
FLAGR_RECORDER_KAFKA_TOPIC=flagr-records
```

Everything else under `FLAGR_RECORDER_*` - broker TLS and SASL, compression, Kinesis batch tuning, Pub/Sub credentials - is in the source above. Those knobs exist for production hardening; the defaults are meant to get a row onto a topic, not to survive a misconfigured cluster.

### Webhooks

Flagr can fire a webhook whenever a flag changes, which is how teams wire approvals, audit trails, or cache invalidation downstream. The webhook provider is one part of the notification system: you enable it, point it at a URL, give it headers, and it retries with exponential backoff. There's also a toggle for detailed diffs, so the payload can include exactly which fields changed before and after. The full variable set and the retry semantics are on the [Notifications](flagr_notifications.md) page.

### Observability

The last group is how you watch the server once it's running. Flagr exports metrics in three shapes - Prometheus scrape, Statsd push, and two hosted APMs - and you typically pick one rather than stacking them.

| Area | Switch |
|------|--------|
| Prometheus | `FLAGR_PROMETHEUS_ENABLED`, `FLAGR_PROMETHEUS_PATH` |
| Statsd | `FLAGR_STATSD_ENABLED`, host/port/prefix |
| Sentry / New Relic | `FLAGR_SENTRY_ENABLED`, `FLAGR_NEWRELIC_ENABLED` |

Prometheus is the default choice for Kubernetes; Statsd suits traditional infrastructure; Sentry and New Relic are for error tracking and distributed tracing respectively. Each family has its own tuning variables in the source above - latency histograms for Prometheus, APM ports for Statsd, DSNs and app names for the hosted services.

## Maintaining this page

When you add or change variables in `pkg/config/env.go`, update the **guide** tables here only if operators need a one-line summary. The embedded **source** is copied from `pkg/config/env.go` by `make build-docs` / `make serve-docs` into `docs/snippets/env.go` at build time.


