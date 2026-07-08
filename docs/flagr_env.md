# Server configuration

Flagr ships with **no config file**. Every knob the server understands is an environment variable read at startup, bound to a single struct in the codebase. That struct is the source of truth: when this page and the code disagree, the code wins. So this page shows the full source first — every `env` tag, default, and comment — and then offers a short guide for the variables you actually touch in common deployments. If a variable isn't mentioned in the guide, it's still in the source above, usually because it's a niche tuning knob or a default that rarely needs changing.

For deployment recipes — Docker, Compose, Kubernetes — see [Self-hosting](flagr_self_host.md).

---

## Source (`pkg/config/env.go`) :id=source-pkgconfigenvgo

The block below is the live source from the `main` branch, embedded directly. On feature branches, compare against your own checkout until it merges. You can also open it on GitHub.

[Open on GitHub](https://github.com/openflagr/flagr/blob/main/pkg/config/env.go)

[env.go](https://raw.githubusercontent.com/openflagr/flagr/main/pkg/config/env.go ':include :type=code')

---

## Quick start

The fastest path to a running server is the [Self-hosting](flagr_self_host.md) guide, which covers Docker, Compose, and Kubernetes. If you just want the minimal environment for a MySQL-backed server, copy these four variables and go:

```sh
export HOST=0.0.0.0
export PORT=18000
export FLAGR_DB_DBDRIVER=mysql
export FLAGR_DB_DBCONNECTIONSTR='user:pass@tcp(127.0.0.1:3306)/flagr?parseTime=true'
```

If you'd rather serve flags from a static JSON file or URL with no database at all, set `FLAGR_DB_DBDRIVER` to `json_file` or `json_http`. That puts the server into eval-only mode automatically — see [contracts — eval-only](contracts.md#eval-only) and the [JSON flag source](flagr_json_flag_spec.md) spec.

---

## Guide

### Server & HTTP

The first group of variables governs how the process binds, what it serves, and what it logs. Most of these defaults are sane for development; the only ones operators commonly touch in production are the log format (switch to `json` for structured ingestion) and the UI toggle (turn it off for a pure API backend).

| Variable | Default | Notes |
|----------|---------|--------|
| `HOST` / `PORT` | `localhost` / `18000` | Bind address |
| `FLAGR_WEB_PREFIX` | *(empty)* | UI + API base path |
| `FLAGR_UI_ENABLED` | `true` | `false` = API-only |
| `FLAGR_LOGRUS_LEVEL` / `FORMAT` | `info` / `text` | Use `json` in production |
| `FLAGR_PPROF_ENABLED` | `true` | pprof endpoints |
| `FLAGR_MIDDLEWARE_VERBOSE_LOGGER_*` | on | Exclude hot paths via `…_EXCLUDE_URLS` |
| `FLAGR_MIDDLEWARE_GZIP_ENABLED` | `true` | |

CORS is its own family of variables under `FLAGR_CORS_*` — enabled by default with permissive origins. The full list (allowed headers, methods, origins, credentials, max age) lives in the source above; tune it only if you're locking down a browser-facing deployment.

### Evaluation & cache

Evaluation requests never hit the database directly. They read from an in-memory cache that the server rebuilds on a fixed interval, so the two variables that matter most here are the reload cadence and the per-fetch timeout. A third, debug, only takes effect when an evaluation request also sets `enableDebug: true` — it's a global gate on per-request segment logging, documented in the [Debug console](flagr_debugging.md) page. The batch-size caps are guardrails against oversized requests rather than tuning knobs.

| Variable | Default | Notes |
|----------|---------|--------|
| `FLAGR_EVALCACHE_REFRESHINTERVAL` | `3s` | EvalCache reload period |
| `FLAGR_EVALCACHE_REFRESHTIMEOUT` | `59s` | Single fetch timeout |
| `FLAGR_EVAL_DEBUG_ENABLED` | `true` | + `enableDebug` on request → segment logs ([Debug console](flagr_debugging.md)) |
| `FLAGR_EVAL_BATCH_SIZE` | `0` | `0` = unlimited batch eval (POST and GET batch) |
| `FLAGR_EVAL_GET_MAX_URL_BYTES` | `8192` | GET `json=` query cap; `0` = off — [use cases](flagr_use_cases.md#get-evaluation-browser-friendly) |
| `FLAGR_EXPOSURE_BATCH_SIZE` | `100` | Max rows per `POST /exposures` |

Because the cache is eventually consistent, a flag change won't be visible to evaluators until the next reload lands — and until it does, **`variantKey`** comes back blank for affected entities. That staleness window is a contract, not a bug; see [EvalCache freshness](contracts.md#evalcache-freshness) for why automated tests should wait at least one interval before asserting on new configuration.

There's also an explicit `FLAGR_EVAL_ONLY_MODE` flag, though in practice eval-only is usually implied: whenever the database driver is `json_file` or `json_http`, the server enters eval-only mode on its own.

### Database

Two variables decide where flags live: the driver and the connection string. The defaults give you a local SQLite file, which is enough for development and tests. Production wants MySQL or Postgres, and the JSON drivers exist for read-only evaluation from a file or remote URL.

`FLAGR_EVAL_GET_MAX_URL_BYTES` (default **8192**, **0** = disabled) caps the **raw query string** on `GET /api/v1/evaluation` and batch GET. POST-vs-GET security, proxy limits, and examples: [Use cases — GET evaluation](flagr_use_cases.md#get-evaluation-browser-friendly).

#### Built-in context injection

| Variable | Default | Notes |
|----------|---------|--------|
| `FLAGR_INJECTED_CONTEXT_ENABLED` | `false` | Merge `@ts*` and `@http_*` into `entityContext` before eval |
| `FLAGR_INJECTED_CONTEXT_HTTP_HEADERS` | `""` | Comma-separated headers → `@http_*` keys |
| `FLAGR_INJECTED_CONTEXT_HTTP_HEADER_PREFIXES` | `""` | Prefix match (e.g. `CF-` for Cloudflare) |

Full guide: [Built-in context injection](flagr_injected_context.md).



| Variable | Default |
|----------|---------|
| `FLAGR_DB_DBDRIVER` | `sqlite3` |
| `FLAGR_DB_DBCONNECTIONSTR` | `flagr.sqlite` |

| Driver | Role |
|--------|------|
| `sqlite3` | Local dev (default) |
| `mysql` / `postgres` | Production |
| `json_file` / `json_http` | Flags from file or URL ([JSON spec](flagr_json_flag_spec.md)) |

#### Eval cache export :id=eval-cache-export

A running server can dump its in-memory cache as JSON, which is useful for debugging, seeding another instance, or snapshotting what evaluators actually see. The endpoint is `GET /api/v1/export/eval_cache/json`, and it accepts optional `enabled`, `ids`, `keys`, `tags`, and `tagsOperator` (`ANY` / `ALL`) query parameters to narrow the dump.

### Authentication

Authentication is off by default, so a freshly started server is open until you turn something on. Flagr supports two layers that can be used independently: basic auth, which guards the UI, and JWT auth, which guards the API. Both layers let you whitelist paths so hot evaluation traffic doesn't have to authenticate — and both ship with defaults that leave `/api/v1/evaluation` and `/api/v1/exposures` open, so turning auth on won't break your integration.

A minimal basic-auth setup is three variables:

```sh
FLAGR_BASIC_AUTH_ENABLED=true
FLAGR_BASIC_AUTH_USERNAME=admin
FLAGR_BASIC_AUTH_PASSWORD=password
```

JWT is richer. The variables cover enabling it (`FLAGR_JWT_AUTH_ENABLED`), the shared secret or PEM key (`FLAGR_JWT_AUTH_SECRET`), the signing method (`HS256` / `HS512` / `RS256`), and a set of prefix and exact whitelist paths. All of them are in the source above. JWT tokens can arrive by cookie or by `Authorization: Bearer` header; when both are present, the header wins.

Separately, Flagr can identify *who* made a mutation for audit logging without doing full authentication. `FLAGR_HEADER_AUTH_*` reads a user identifier from a header (handy behind a corporate proxy), and `FLAGR_COOKIE_AUTH_*` reads one from a cookie (handy behind something like Cloudflare Zero Trust). These stamp `created_by` / `updated_by` on changes; they don't gate access.

One thing worth calling out: the default JWT whitelist allows unauthenticated exposure logging. If the integrity of your impression stream matters, narrow the whitelist to lock down `/api/v1/exposures` and rate-limit it at the edge. The [Exposure logging](flagr_exposure.md) page walks through the tradeoffs.

### Data recorders :id=data-record-destinations

Recording is opt-in by design — it always costs something, so Flagr makes you ask for each layer explicitly. The master switch is `FLAGR_RECORDER_ENABLED`, which defaults to `false`. Even with it on, nothing streams until each flag also sets **`dataRecordsEnabled: true`** in its own configuration. Those two gates cascade the same way for every recorder type.

`FLAGR_RECORDER_TYPE` selects which recorder(s) run, and it's a comma-separated list so you can combine them.

| `FLAGR_RECORDER_TYPE` | Doc |
|------------------------|-----|
| `kafka`, `kinesis`, `pubsub` | Eval + exposure stream — [Recorders & A/B](flagr_eval_exposure_pipeline.md) |
| `datar` | In-process eval counts only — [Datar](flagr_datar.md) (no exposures) |

The streaming recorders ship eval and exposure rows to a broker; Datar keeps in-process evaluation counts and flushes them to the database on an interval. Combining `kafka,datar` is a common pattern — a live stream for analytics plus cheap counts for dashboards. The `FLAGR_RECORDER_FRAME_OUTPUT_MODE` variable controls how each row is framed: `payload_string` stringifies the payload (and respects encryption), while `payload_raw_json` emits the raw object (and ignores encryption).

The minimal Kafka setup is four variables:

```sh
FLAGR_RECORDER_ENABLED=true
FLAGR_RECORDER_TYPE=kafka
FLAGR_RECORDER_KAFKA_BROKERS=kafka1:9092
FLAGR_RECORDER_KAFKA_TOPIC=flagr-records
```

Everything else under `FLAGR_RECORDER_*` — broker TLS and SASL, compression, Kinesis batch tuning, Pub/Sub credentials — is in the source above. Those knobs exist for production hardening; the defaults are meant to get a row onto a topic, not to survive a misconfigured cluster.

### Webhooks

Flagr can fire a webhook whenever a flag changes, which is how teams wire approvals, audit trails, or cache invalidation downstream. The webhook provider is one part of the notification system: you enable it, point it at a URL, give it headers, and it retries with exponential backoff. There's also a toggle for detailed diffs, so the payload can include exactly which fields changed before and after. The full variable set and the retry semantics are on the [Notifications](flagr_notifications.md) page.

### Observability

The last group is how you watch the server once it's running. Flagr exports metrics in three shapes — Prometheus scrape, Statsd push, and two hosted APMs — and you typically pick one rather than stacking them.

| Area | Switch |
|------|--------|
| Prometheus | `FLAGR_PROMETHEUS_ENABLED`, `FLAGR_PROMETHEUS_PATH` |
| Statsd | `FLAGR_STATSD_ENABLED`, host/port/prefix |
| Sentry / New Relic | `FLAGR_SENTRY_ENABLED`, `FLAGR_NEWRELIC_ENABLED` |

Prometheus is the default choice for Kubernetes; Statsd suits traditional infrastructure; Sentry and New Relic are for error tracking and distributed tracing respectively. Each family has its own tuning variables in the source above — latency histograms for Prometheus, APM ports for Statsd, DSNs and app names for the hosted services.

---

## Maintaining this page

When you add or change variables in `pkg/config/env.go`, update the **guide** tables here only if operators need a one-line summary; the embedded **source** updates automatically on merge to `main`.


