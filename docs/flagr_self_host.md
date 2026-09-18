---
title: Self-hosting Flagr
---

# Self-hosting Flagr

Flagr is meant to run on **your** infrastructure. Normal path: official Docker image, env vars for DB (or JSON flag source), port **18000**. No config file required for production.

One process, env-only configuration. Full variable list: [Environment variables](flagr_env.md). Source of truth is `pkg/config/env.go`. API base: **`/api/v1`** (optional [path prefix](#reverse-proxy-and-path-prefix)).

## Container image

Published to **`ghcr.io/openflagr/flagr`**. Tags: `latest` and per-release semver. Architectures: **linux/amd64** and **linux/arm64**.

```bash
docker pull ghcr.io/openflagr/flagr
docker run --rm -p 18000:18000 ghcr.io/openflagr/flagr
```

That boots against the image's SQLite defaults. The image includes the UI bundle and the `./flagr` binary, and runs as non-root **`appuser`**.

Values you almost always override past a demo:

| Variable | Image / process default | Override in prod |
|----------|-------------------------|------------------|
| `HOST` | process default `localhost`; Docker sets `0.0.0.0` | Keep `0.0.0.0` in containers |
| `PORT` | `18000` | As needed |
| `FLAGR_DB_DBDRIVER` | `sqlite3` | `mysql` / `postgres` / `json_http` |
| `FLAGR_DB_DBCONNECTIONSTR` | image: `/data/demo_sqlite3.db`; binary: `flagr.sqlite` | Your DSN / volume path / flags URL |
| `FLAGR_RECORDER_ENABLED` | `false` | When using recorders |

Pin a semver tag in production, not only `latest`.

> **Note:** The Go default for `HOST` is `localhost` (`pkg/config/env.go`). The official `Dockerfile` sets `ENV HOST=0.0.0.0` so the process is reachable outside the container. Do not assume `0.0.0.0` if you run the binary bare.

## Binary from source (optional)

Use a source build when you are **developing Flagr** or need a bare binary without Docker:

```bash
git clone https://github.com/openflagr/flagr.git && cd flagr
make build && ./flagr --port 18000
```

TLS uses `--scheme=https` plus the cert flags from the server bootstrap. Local full stack: `make start` (see [Contributing](CONTRIBUTING.md)).

## Deployment strategy {#deployment-strategy}

Pick **where flags live** and **how you run the process**. Evaluation is served from an in-memory EvalCache (reload is per replica). Extra replicas are only safe when they share a store — SQLite cannot take multiple writers. JSON drivers (`json_file` / `json_http`) force [eval-only mode](flagr_behavioral_contracts.md#eval-only) (read-only UI; writes under `/api/v1/flags` return 403).

| Strategy | Flags live in | How you run it | Scale | Writes |
|----------|---------------|----------------|-------|--------|
| [Docker demo](#docker) | SQLite in the container | `docker run` | 1 | yes |
| [Compose](#compose) | MySQL | Docker Compose | 1 Flagr + DB | yes |
| [Helm / SQLite](#sqlite) | SQLite volume | Helm | **1 writer** | yes |
| [Helm / SQLite HA](#sqlite-ha) | SQLite + `json_http` | Helm `evalReplicas` | 1 writer + N eval pods | primary only |
| [Helm / GitOps](#gitops) | Git (HTTP JSON) | Helm `gitops.enabled` | N identical eval-only pods | no |
| [Helm / MySQL or Postgres](#sql) | Shared SQL | Helm `replicaCount` + `env` | N pods, one DB | yes, every pod |
| [VM / systemd](#vm) | any of the above | `./flagr` or the image | same rules as the store | same |

Headless API (any SQL strategy): `FLAGR_UI_ENABLED=false`. DSN shapes: [Database](#database). Helm values: in-repo [`helm/`](https://github.com/openflagr/flagr/tree/main/helm), OCI `oci://ghcr.io/openflagr/flagr/charts/flagr`.

### Docker (SQLite demo) {#docker}

```bash
docker run --rm -p 18000:18000 ghcr.io/openflagr/flagr
```

Persist flags across restarts:

```bash
docker run --rm -p 18000:18000 \
  -e HOST=0.0.0.0 \
  -e FLAGR_DB_DBDRIVER=sqlite3 \
  -e FLAGR_DB_DBCONNECTIONSTR=/data/flagr.sqlite \
  -v flagr-data:/data \
  ghcr.io/openflagr/flagr
```

Stay at one container. For more eval traffic, use [SQLite HA](#sqlite-ha) or move to SQL / GitOps.

### Docker Compose (MySQL) {#compose}

Starting point, not a hardened blueprint:

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

Swap credentials before any shared environment. The repo CI compose file has more engine examples but is tuned for tests, not production.

### Helm / Kubernetes {#helm}

```bash
helm install flagr oci://ghcr.io/openflagr/flagr/charts/flagr --version 1.0.0 \
  --namespace flagr --create-namespace
kubectl -n flagr port-forward svc/flagr 18000:18000
curl -sS http://127.0.0.1:18000/api/v1/health
```

Pin `--version` to `helm/Chart.yaml` `version`. From a checkout: `helm install flagr ./helm --namespace flagr --create-namespace`. Helm cannot install from a GitHub directory URL (`…/tree/main/helm`); the chart is a subdirectory, not a packaged `.tgz`.

Configure Flagr with `env` / `envFrom` — every knob is in [Environment variables](flagr_env.md). Add your own Ingress, PVC, and HPA.

Default install: **one** replica, SQLite at `/data/flagr.sqlite` on an emptyDir (ephemeral). Chart vs process / Docker defaults: `FLAGR_PPROF_ENABLED=false`, `FLAGR_DB_DBCONNECTION_DEBUG=false`, `FLAGR_LOGRUS_FORMAT=json`.

If you set `FLAGR_WEB_PREFIX`, also override probe `httpGet.path`, `test.path`, and (SQLite HA) `evalReplicas.flagsURL`.

#### Helm / SQLite — one writer {#sqlite}

Keep `replicaCount: 1`. Persist the file with a PVC mounted as volume `data`:

```yaml
extraVolumes:
  - name: data
    persistentVolumeClaim:
      claimName: flagr-data
```

#### Helm / SQLite HA — extra eval capacity, no extra writers {#sqlite-ha}

Do **not** raise `replicaCount`. Set `evalReplicas.replicaCount`. The chart keeps one SQLite primary (UI + CRUD) and adds eval-only pods that poll `GET /api/v1/export/eval_cache/json` over `json_http` (same JSON the GitOps driver reads). Those pods never open the SQLite file.

```yaml
# sqlite-ha.yaml
evalReplicas:
  replicaCount: 3
```

```bash
helm upgrade --install flagr oci://ghcr.io/openflagr/flagr/charts/flagr --version 1.0.0 \
  --namespace flagr -f sqlite-ha.yaml
```

| Traffic | Service |
|---------|---------|
| UI, flag CRUD | `svc/flagr` (primary, SQLite) |
| Evaluation (`/api/v1/evaluation`) | `svc/flagr-eval` (json_http replicas) |

A flag change on the primary is visible on eval replicas within [EvalCache freshness](flagr_behavioral_contracts.md#evalcache-freshness) (default 3s). Eval pods wait for the primary export before starting.

If the primary uses `FLAGR_WEB_PREFIX`, set `evalReplicas.flagsURL` to the prefixed export URL. If you enable JWT/basic on the primary, whitelist `/api/v1/export` (or the prefixed path) so the replicas can fetch.

#### Helm / GitOps — GitHub (or any HTTP JSON) is the source {#gitops}

Every pod is the same: `json_http` eval-only, no SQLite writer. Flags live in git; Flagr polls the raw URL. UI is read-only; writes under `/api/v1/flags` return 403. Spec and PAT setup: [JSON flag source](flagr_json_flag_spec.md).

This is the same `json_http` mechanism as SQLite `evalReplicas`, pointed at GitHub instead of the primary export — so there is only **one** Deployment (`replicaCount` may be > 1). Do not set `evalReplicas` (that fleet exists to wrap a SQLite writer).

```yaml
# gitops.yaml
replicaCount: 3
gitops:
  enabled: true
  flagsURL: https://raw.githubusercontent.com/org/flagr-config/main/flags.json
```

```bash
helm upgrade --install flagr oci://ghcr.io/openflagr/flagr/charts/flagr --version 1.0.0 \
  --namespace flagr -f gitops.yaml
```

Private repo: leave `flagsURL` empty and put the PAT URL in a Secret (PAT as HTTP Basic username, empty password):

```yaml
gitops:
  enabled: true
env:
  - name: FLAGR_DB_DBCONNECTIONSTR
    valueFrom:
      secretKeyRef:
        name: flagr-gitops
        key: url
```

```bash
kubectl create secret generic flagr-gitops \
  --from-literal=url='https://github_pat_xxxx@raw.githubusercontent.com/org/flagr-config/main/flags.json'
```

All traffic (UI + eval) is `svc/flagr`. A merged flags.json is visible within [EvalCache freshness](flagr_behavioral_contracts.md#evalcache-freshness).

#### Helm / MySQL or PostgreSQL — every pod uses the same DB {#sql}

Raise `replicaCount`. Leave `evalReplicas.replicaCount` at **0** and `gitops.enabled` off. There is one shared database, so every replica can serve eval **and** CRUD.

```yaml
# postgres.yaml
replicaCount: 3
env:
  - name: FLAGR_DB_DBDRIVER
    value: postgres
  - name: FLAGR_DB_DBCONNECTIONSTR
    valueFrom:
      secretKeyRef:
        name: flagr-db
        key: FLAGR_DB_DBCONNECTIONSTR
  - name: FLAGR_DB_DBCONNECTION_RETRY_ATTEMPTS
    value: "30"
  - name: FLAGR_DB_DBCONNECTION_RETRY_DELAY
    value: "2s"
```

Flagr fatals after ~900ms if the DB is down (default retries); a startupProbe cannot help. Raise retries as above.

```bash
kubectl create secret generic flagr-db \
  --from-literal=FLAGR_DB_DBCONNECTIONSTR='sslmode=require host=pg.example user=flagr password=… dbname=flagr'
helm upgrade --install flagr oci://ghcr.io/openflagr/flagr/charts/flagr --version 1.0.0 \
  --namespace flagr -f postgres.yaml
```

MySQL is the same overlay with `FLAGR_DB_DBDRIVER=mysql` and a `parseTime=true` DSN ([guide](flagr_env.md)).

### VM / systemd {#vm}

Same image or `make build` binary: inject secrets for DB, JWT, and recorders; bind `0.0.0.0:18000`; probe **`GET /api/v1/health`**. Scaling follows the store in the [matrix](#deployment-strategy) — one SQLite process, or N processes on MySQL/Postgres/json_http.

## Database {#database}

On boot, Flagr retries the connection `FLAGR_DB_DBCONNECTION_RETRY_ATTEMPTS` times (default **9**) with `FLAGR_DB_DBCONNECTION_RETRY_DELAY` (default **100ms**) between attempts. Then GORM **auto-migrate** brings schema up to date for normal upgrades. `json_http` / `json_file` skip GORM and load flags on the EvalCache interval.

**SQLite** — file path in `FLAGR_DB_DBCONNECTIONSTR` (Helm: `/data/flagr.sqlite`).

**MySQL** — DSN like `user:password@tcp(host:3306)/flagr?parseTime=true`. `parseTime` is required for GORM time mapping.

**PostgreSQL** — libpq-style string, e.g. `sslmode=disable host=… user=… password=… dbname=flagr` (prefer `sslmode=require` where you can).

**JSON HTTP** — `FLAGR_DB_DBDRIVER=json_http` and a flag URL. Freshness: [EvalCache freshness](flagr_behavioral_contracts.md#evalcache-freshness). Spec: [JSON flag source](flagr_json_flag_spec.md).

## Reverse proxy and path prefix

```bash
export FLAGR_WEB_PREFIX=/flagr
```

UI at `/flagr/`, API at `/flagr/api/v1/...`. Set `HOST=0.0.0.0` in containers so the process is reachable across the network namespace.

## Production checklist

| Item | Action |
|------|--------|
| Bind | `HOST=0.0.0.0` in containers |
| Logs | `FLAGR_LOGRUS_FORMAT=json` |
| Auth, CORS, recorders, pprof | [Environment variables: guide](flagr_env.md#guide) |
| Recording | [Recording gates](flagr_behavioral_contracts.md#recording-gates) |
| Backups | SQL dump, or `GET /api/v1/export/eval_cache/json`, or Git for JSON mode |

## Verify

```bash
curl -sS http://localhost:18000/api/v1/health
curl -sS -X POST http://localhost:18000/api/v1/evaluation \
  -H 'content-type: application/json' \
  -d '{"entityID":"smoke-1","flagID":1}'
```

Health should be OK; eval returns the assignment (or blank if flag 1 is not configured yet). Client examples: [Integration guide](integration.md).
