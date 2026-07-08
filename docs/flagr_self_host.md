# Self-hosting Flagr

Flagr is **built to run on your infrastructure** — not as a hosted-only SaaS you must buy. The intended path is simple: run the **official Docker image**, point it at your database (or JSON flag source) with **environment variables**, and expose port **18000**. No config files, no custom build step for normal production use. Everything else on this page is detail for MySQL/Postgres, Compose, K8s, TLS, and hardening.

One process, env-only configuration. Full variable list: [Environment variables](flagr_env.md). API base: **`/api/v1`** (optional [path prefix](#reverse-proxy-and-path-prefix)).

## Container image

The fastest path to a running Flagr is the official image, published to the GitHub Container Registry under **`ghcr.io/openflagr/flagr`**. Tags track both `latest` and per-release semver, and builds cover **linux/amd64** and **linux/arm64** so the same image works on x86 hosts and ARM graviton instances alike.

```bash
docker pull ghcr.io/openflagr/flagr
docker run --rm -p 18000:18000 ghcr.io/openflagr/flagr
```

That single command is enough to boot a working server against an in-memory SQLite database. The image ships the UI bundle alongside the `./flagr` binary and runs as the non-root **`appuser`**, so you do not need to build or bake anything yourself for day-to-day operation.

Out of the box the image is tuned for a frictionless first run, but the values below are the ones you will almost always override for anything beyond a demo:

| Variable | Image default | Override in prod |
|----------|---------------|------------------|
| `HOST` | `0.0.0.0` | Keep for containers |
| `PORT` | `18000` | As needed |
| `FLAGR_DB_DBDRIVER` | `sqlite3` | `mysql` / `postgres` |
| `FLAGR_DB_DBCONNECTIONSTR` | `/data/demo_sqlite3.db` | Your DSN / volume path |
| `FLAGR_RECORDER_ENABLED` | `false` | When using recorders |

Pin a semver tag in production, not only `latest`.

## Binary from source (optional)

The image above is the recommended production artifact. Reach for a source build only when you are **developing Flagr** itself or you need a bare binary without Docker — for example, shipping a static binary into a minimal VM image:

```bash
git clone https://github.com/openflagr/flagr.git && cd flagr
make build && ./flagr --port 18000
```

TLS here uses `--scheme=https` plus the cert flags exposed by the server bootstrap. For local development, `make start` brings the server up with hot-reload conveniences — see [Contributing](CONTRIBUTING.md) for the full workflow.

## Deployment shapes

Before wiring specifics, it helps to pick the shape you are aiming for. The driver you choose for `FLAGR_DB_DBDRIVER` largely determines what Flagr can do, because the JSON drivers put the server into [eval-only mode](contracts.md#eval-only) and drop the management surface entirely.

| Shape | Driver | Notes |
|-------|--------|--------|
| Demo | `sqlite3` (image default) | Ephemeral unless you mount `/data` |
| Prod UI + CRUD | `mysql` or `postgres` | Shared DB; auto-migrate on boot |
| Eval edge | `json_file` / `json_http` | [Eval-only](contracts.md#eval-only) + [JSON flag source](flagr_json_flag_spec.md) |
| Headless API | SQL + `FLAGR_UI_ENABLED=false` | CRUD via API only |

A demo is just the container default with a throwaway SQLite file. The production shape with a UI and CRUD is MySQL or Postgres behind the same image, sharing one database so every replica sees the same flags. The eval-edge shape is for read-only deployments that pull flags from a [JSON source](flagr_json_flag_spec.md) and serve evaluation traffic only — ideal for latency-sensitive or isolated environments. Headless API keeps the SQL backend but turns the UI off, leaving a pure CRUD surface for automation.

## Database

Whatever driver you select, Flagr retries the connection on boot using `FLAGR_DB_DBCONNECTION_RETRY_ATTEMPTS` (default 9) with `FLAGR_DB_DBCONNECTION_RETRY_DELAY` (100ms) between attempts, so a database that starts slightly after the app does not need careful ordering. Once connected, Flagr runs **GORM auto-migrate** on startup, bringing the schema up to date without any hand-run SQL for normal upgrades.

**SQLite** is the zero-dependency option. Mount a volume so the flag file survives restarts:

```bash
docker run --rm -p 18000:18000 \
  -e FLAGR_DB_DBDRIVER=sqlite3 \
  -e FLAGR_DB_DBCONNECTIONSTR=/data/flagr.sqlite \
  -v flagr-data:/data \
  ghcr.io/openflagr/flagr
```

**MySQL** takes a DSN in the form `user:password@tcp(host:3306)/flagr?parseTime=true` — the `parseTime` flag is what lets GORM map Go time types onto MySQL columns correctly.

**PostgreSQL** uses a libpq-style key/value string such as `sslmode=disable host=… user=… password=… dbname=flagr`; switch to `sslmode=require` wherever your cluster supports it.

**JSON HTTP** swaps the database for a URL: set `FLAGR_DB_DBDRIVER=json_http` and point it at your flag endpoint — see [JSON flag source](flagr_json_flag_spec.md). Because the server re-fetches config on its EvalCache cadence, freshness is governed by [EvalCache freshness](contracts.md#evalcache-freshness), not by any database commit.

## Docker Compose (MySQL + Flagr)

For a self-contained stack on a single host, Compose is the most reproducible way to run Flagr with MySQL. The file below brings up a MySQL 8 instance and a Flagr container wired to it, with structured JSON logs ready for log shipping:

```yaml
services:
  mysql:
    image: mysql:8
    environment:
      MYSQL_DATABASE: flagr
      MYSQL_USER: flagr
      MYSQL_PASSWORD: changeme
      MYSQL_ROOT_PASSWORD: changeme
    volumes:
      - mysql-data:/var/lib/mysql

  flagr:
    image: ghcr.io/openflagr/flagr:latest
    ports:
      - "18000:18000"
    environment:
      HOST: "0.0.0.0"
      FLAGR_DB_DBDRIVER: mysql
      FLAGR_DB_DBCONNECTIONSTR: "flagr:changeme@tcp(mysql:3306)/flagr?parseTime=true"
      FLAGR_LOGRUS_FORMAT: json
    depends_on:
      - mysql

volumes:
  mysql-data:
```

Treat this as a starting point, not a hardened blueprint — swap the credentials for real secrets before anything touches a shared environment. The CI compose file shipped with the repo has connection-string examples across more engines, though it is tuned for test coverage rather than production.

## Kubernetes and VMs

There is no in-repo Helm chart, but the deployment story is the same image on any orchestrator. Run the container (or, on a VM, the `make build` binary under systemd), inject secrets for the database, JWT, and recorders, and let the process bind to `0.0.0.0:18000`. Probe **`GET /api/v1/health`** for readiness and liveness, and expose port **18000** through your service or ingress. TLS typically terminates at the ingress; if you prefer to terminate on Flagr itself, use the CLI flags described under binary-from-source above.

To scale horizontally, keep one shared flag store — either the SQL database or a single JSON URL that every replica reads — and let each replica maintain its own EvalCache. Because cache reload is local per process, read the [EvalCache freshness](contracts.md#evalcache-freshness) contract before assuming a flag edit is visible fleet-wide.

## Reverse proxy and path prefix

When Flagr sits behind a reverse proxy under a subpath, set a prefix so the UI and API both generate links relative to it:

```bash
export FLAGR_WEB_PREFIX=/flagr
```

With that, the UI is served at `/flagr/` and the API at `/flagr/api/v1/...`. Keep `HOST=0.0.0.0` so Flagr is reachable across Docker or Kubernetes networking rather than only on localhost.

## Production checklist

Before a stack leaves the staging environment, walk through the items below. Each one maps to a concrete switch documented in the [environment variables guide](flagr_env.md#guide):

| Item | Action |
|------|--------|
| Bind | `HOST=0.0.0.0` in containers |
| Logs | `FLAGR_LOGRUS_FORMAT=json` |
| Auth, CORS, recorders, pprof | [Environment variables — guide](flagr_env.md#guide) |
| Recording | [Recording gates](contracts.md#recording-gates) |
| Backups | SQL backup, or `GET /api/v1/export/eval_cache/json`, or Git for JSON mode |

For backups, a SQL dump covers the management database; the eval-cache export gives a point-in-time snapshot of what evaluators saw; in JSON mode, your Git history is the source of truth.

## Verify

Once the server is up, a two-line smoke test confirms both health and end-to-end evaluation:

```bash
curl -sS http://localhost:18000/api/v1/health
curl -sS -X POST http://localhost:18000/api/v1/evaluation \
  -H 'content-type: application/json' \
  -d '{"entityID":"smoke-1","flagID":1}'
```

The first call should return a healthy status; the second returns the variant assignment for the given entity and flag. For richer client-side examples across languages, move on to the [Integration guide](integration.md).