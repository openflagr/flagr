# Get Started

Shipping software is a constant negotiation between **velocity** and
**risk**. You want to move fast, but you also need to decouple *deploy* from
*release* — to turn a feature on for one user, a thousand users, or nobody,
without redeploying. You need to **experiment**: to learn which of two
experiences performs better, with a denominator you can trust. And you need
**dynamic configuration** — change a color, a copy string, a timeout — without
a code change or a restart.

These three problems — feature flagging, A/B testing, and dynamic
configuration — are usually solved by three different tools. Flagr solves all
three with one primitive: the **flag**. A flag is a decision point in your
code. Behind that decision sits an evaluation engine that looks at *who* is
asking and decides *what* they get. The same engine that toggles a kill switch
also splits traffic between variants and serves per-variant JSON attachments.

Flagr is an open-source Go service that delivers the right experience to the
right entity and monitors the impact. It exposes that evaluation engine behind
a Swagger-documented REST API for flag management and evaluation — so your
application code stays thin (`POST /evaluation`, read the `variantKey`),
while operators configure targeting, distribution, and rollout from a UI or
as code.

New to Flagr? Start with the [Overview](flagr_overview) for core concepts.

## Documentation map

| Goal | Read |
|------|-----|
| Learn concepts (flags, segments, rollout, architecture) | [Overview](flagr_overview) |
| See code patterns for flags / A/B / config | [Use Cases](flagr_use_cases) |
| Configure the server (DB, auth, recorders) | [Environment Variables](flagr_env) |
| Run flags from Git (no DB) | [JSON Flag Source](flagr_json_flag_spec) |
| Log impressions after eval (`POST /exposures`) | [Exposure Logging](flagr_exposure) |
| Ship eval + exposure to a recorder and analyze A/B | [Data Recorders & A/B Analysis](flagr_eval_exposure_pipeline) |
| Quick eval counts inside Flagr (no pipeline) | [Datar Analytics](flagr_datar) |
| Test evaluation in the UI | [Debug Console](flagr_debugging) |
| Webhooks on flag changes | [Notifications](flagr_notifications) |
| REST API details | [API Reference](https://openflagr.github.io/flagr/api_docs) |

**Event-recording path:** [Exposure](flagr_exposure) (API) →
[Data Recorders & A/B Analysis](flagr_eval_exposure_pipeline) →
[Datar](flagr_datar) (built-in eval-only aggregates). Recorder env vars:
[Data record destinations](flagr_env?id=data-record-destinations).

## What Flagr can do

| Capability | Highlights |
|------------|------------|
| **Feature flags** | Binary toggles, kill switches, targeted audience rollouts |
| **A/B testing** | Multi-variant experiments with deterministic distribution |
| **Dynamic configuration** | Per-variant JSON attachments for runtime config |
| **GitOps / Flags-as-code** | Load flags from JSON files or HTTP URLs; manage in Git, validate in CI |
| **Exposure logging** | `POST /exposures` after the user sees the variant — [Exposure](flagr_exposure), [Data Recorders & A/B Analysis](flagr_eval_exposure_pipeline) |
| **Webhook notifications** | HTTP POST on every flag change, with retry and backoff |
| **Multi-database** | SQLite (dev), MySQL, PostgreSQL, JSON sources |

See [Use Cases](flagr_use_cases) for practical examples of each pattern.

## Quick demo

Run Flagr with Docker — no dependencies required:

```bash
docker pull ghcr.io/openflagr/flagr
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr

# Open the Flagr UI
open localhost:18000
```

Or try the hosted demo at
[https://try-flagr.onrender.com](https://try-flagr.onrender.com)
(cold starts may take a moment):

```bash
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

## Development

### Prerequisites

- **Go** 1.24+ (CI builds with Go 1.26; see `go.mod`)
- **Node** 20+ (for UI development)
- **Make**

### Build and run

```bash
git clone https://github.com/openflagr/flagr.git
cd flagr

make build        # Build the Go server binary (./flagr)
make start        # Backend (:18000) + frontend dev server (:8080) in parallel

make run          # Run the pre-built backend only
make run_ui       # Run the UI dev server only (proxies /api/v1 to :18000)
```

After Go code changes, rebuild and restart in one step:

```bash
make rebuild-run    # build → stop-ui → start
```

> **Note:** `make stop-ui` kills processes bound to `:18000` and `:8080` (via
> `lsof -ti:<port>`), so it never touches processes from other projects.

Frontend-only development: run `npm run dev` in `browser/flagr-ui/` — Vite
proxies `/api/v1` to `:18000` and hot-reloads on save.

## Testing

Flagr has three kinds of tests, each serving a different purpose.

### Unit tests

Go unit tests for `pkg/` — no external services required:

```bash
make test
```

Or directly:

```bash
go test ./pkg/...
```

> **Note:** `make test` also runs `verifiers` first (golangci-lint + Swagger
> validation). `go test ./pkg/...` skips that prerequisite.

### E2E tests (UI)

Playwright end-to-end tests for the Vue 3 UI. Builds the Go server, starts the
backend and UI servers, runs Playwright, then cleans up:

```bash
make test-e2e
```

### Integration tests (API, multi-DB)

HTTP-level integration tests covering all CRUD and eval endpoints. Seeds ~48
realistic flags across all 12 constraint operators.

**Local mode** — SQLite `:memory:`, auto-starts a server on a random port:

```bash
make test-integration
```

**Docker Compose mode** — runs the same suite against 6 Flagr instances
(SQLite, MySQL, MySQL 8, PostgreSQL 9, PostgreSQL 13, checkr/flagr):

```bash
cd integration_tests && make test
```

> **Note:** CI runs `make test-and-bench` (tests + benchmarks against the same
> instances). Other compose targets: `make bench` (benchmarks only),
> `make retest` (tear down then re-run).

**HTTP eval benchmarks** — measures end-to-end eval latency through HTTP:

```bash
make bench-integration
```

## Deploy

Use the `ghcr.io/openflagr/flagr` image directly and configure everything
through environment variables. See
[Server Configuration](flagr_env) for the full list.

```bash
export HOST=0.0.0.0
export PORT=18000
export FLAGR_DB_DBDRIVER=mysql
export FLAGR_DB_DBCONNECTIONSTR=root:@tcp(127.0.0.1:18100)/flagr?parseTime=true

docker run -it -p 18000:18000 ghcr.io/openflagr/flagr
```

For GitOps workflows (flags-as-code, eval-only mode), see the
[JSON Flag Source](flagr_json_flag_spec) guide.