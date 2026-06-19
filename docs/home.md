# Get Started

Flagr is an open source Go service that delivers the right experience to the right entity and monitors the impact. It provides **feature flags**, **experimentation (A/B testing)**, and **dynamic configuration** — all behind clear swagger REST APIs for flag management and evaluation.

For a deeper introduction, see the [Flagr Overview](flagr_overview).

## What can Flagr do?

| Capability | Highlights |
|------------|------------|
| **Feature flags** | Binary toggles, kill switches, targeted audience rollouts |
| **A/B testing** | Multi-variant experiments with deterministic distribution |
| **Dynamic configuration** | Per-variant JSON attachments for runtime config |
| **GitOps / Flags-as-code** | Load flags from JSON files or HTTP URLs; manage in Git, validate in CI |
| **Datar analytics** | Built-in aggregate analytics — no external pipeline needed |
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

Or try the hosted demo at [https://try-flagr.onrender.com](https://try-flagr.onrender.com) (cold starts may take a moment):

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

- **Go** 1.24+
- **Node** 20+ (for UI development)
- **Make**

### Build and run

```bash
git clone https://github.com/openflagr/flagr.git
cd flagr

# Build the Go server binary
make build

# Start backend (:18000) + frontend dev server (:8080) in parallel
make start

# Or run just the pre-built backend
make run

# Or run just the UI dev server (proxies /api/v1 to :18000)
make run_ui
```

After Go code changes, rebuild and restart in one step:

```bash
make rebuild-run    # build → stop-ui → start
```

Frontend-only development: run `npm run dev` in `browser/flagr-ui/` — Vite proxies `/api/v1` to `:18000` and hot-reloads on save.

## Testing

Flagr has three kinds of tests, each serving a different purpose.

### Unit tests

Run the Go unit tests (no external services required):

```bash
make test
```

Or directly:

```bash
go test ./pkg/...
```

### E2E tests (Flagr UI)

Playwright-based end-to-end tests for the Vue 3 UI. Builds the Go
server, starts the backend and UI servers, runs Playwright, then cleans up:

```bash
make test-e2e
```

### Integration tests (API, multi-DB)

HTTP-level integration tests covering all CRUD and eval endpoints.
Seeds ~50 realistic flags across all 12 constraint operators.

**Local mode** — SQLite `:memory:`, auto-starts server on random port:

```bash
make test-integration
```

**Docker Compose mode** — runs the same test suite against 6 flagr
instances (SQLite, MySQL, MySQL 8, PostgreSQL 9, PostgreSQL 13,
checkr/flagr):

```bash
cd integration_tests && make test
```

To run against a single Docker Compose instance:

```bash
cd integration_tests && make test-instance INSTANCE=flagr_with_mysql
```

**HTTP eval benchmarks** — measures end-to-end eval latency through HTTP:

```bash
make bench-integration
```

## Deploy

We recommend using the `ghcr.io/openflagr/flagr` image directly and configuring everything through environment variables. See [Server Configuration](flagr_env) for the full list.

```bash
# Set env variables. For example,
export HOST=0.0.0.0
export PORT=18000
export FLAGR_DB_DBDRIVER=mysql
export FLAGR_DB_DBCONNECTIONSTR=root:@tcp(127.0.0.1:18100)/flagr?parseTime=true

# Run the docker image. Ideally, the deployment will be handled by Kubernetes or Mesos.
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr
```

For GitOps workflows (flags-as-code, eval-only mode), see the [JSON Flag Source](flagr_json_flag_spec) guide.
