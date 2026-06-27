# AGENTS.md

Flagr — Go feature flag service with Vue 3 UI.

## Commands

Run **`make help`** from the repo root for the full catalog. Common targets:

| Command | What it does |
|---|---|
| `make build` | Go server → `./flagr` |
| `make build-ui` | UI: lint, typecheck, Vite → `browser/flagr-ui/dist/` |
| `make start` | Backend `:18000` + UI dev `:8080` |
| `make stop-ui` | Free ports `:18000` / `:8080` (`lsof`, not `pkill`) |
| `make rebuild-run` | `build` → `stop-ui` → `start` |
| `make test` | Lint + swagger validate + Go unit tests |
| `make test-e2e` | `build` + UI lint/typecheck + Playwright |
| `make test-integration` | API integration tests (SQLite, local server) |
| `make test-integration-compose` | Same suite vs Docker Compose (6 DBs) |
| `make bench-integration` | HTTP eval benchmarks (local) |
| `make swagger` | Regenerate `swagger_gen/` |

## Before commit / push (PR)

Run from **repo root**. Match what [`.github/workflows/ci.yml`](.github/workflows/ci.yml) enforces so PR checks stay green.

| You changed | Run before commit | Run before push (recommended) |
|-------------|-------------------|-------------------------------|
| **`browser/flagr-ui/`** only | `make flagr-ui-check` | `make test-e2e` |
| **`pkg/`** or Go tests | `make test` | `make test` (+ `make test-integration` if handler/API behavior) |
| **Swagger** (`swagger/`, handlers → OpenAPI) | `make swagger` then commit `swagger_gen/` + `cmd/flagr-server/main.go` | `make ci-swagger` (regen + `git diff --exit-code`) |
| **UI + Go** or unsure | `make test` **and** `make flagr-ui-check` | `make test` + `make test-e2e` |

**CI mapping (same commands):**

| GitHub Actions job | Makefile |
|--------------------|----------|
| `unit_test` | `make ci-swagger` then `make ci` (= `make test`: **golangci-lint** + swagger validate + `go test ./pkg/...`) |
| `ui_lint` | `make build-ui` (= `flagr-ui-check` + Vite production build) |
| `e2e_test` | `make test-e2e` (= `make build` + `flagr-ui-check` + Playwright) |
| `integration_test` | `make ci-integration` (Docker Compose; usually not every UI PR) |

**Fast UI loop:** `make flagr-ui-check` ≈ ESLint + `vue-tsc` + Vitest (~10s). **Do not** rely on `make run-ui` alone — it does not lint.

**PR hygiene:** Follow [`PULL_REQUEST_TEMPLATE.md`](PULL_REQUEST_TEMPLATE.md). For UI work, use plan **As-built** in `docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md`.

## Key Code

**Backend (`pkg/`):**
- `handler/eval.go` — evaluation engine; `handler/crud.go` — CRUD API handlers
- `handler/exposure.go` — exposure (impression) logging; `handler/data_recorder*.go` — recorders (Kafka, Kinesis, Pub/Sub, Datar)
- `entity/` — domain models (flag, segment, constraint, variant, distribution)
- `config/env.go` — all environment variables (single source of truth)

**Frontend (`browser/flagr-ui/src/`):**
- `api/types.ts` — UI DTOs aligned with `docs/api_docs/bundle.yaml`; `api/flags.ts`, `api/evaluation.ts`, `http.ts`
- `pages/flagPage.ts`, `pages/flagsListPage.ts` — orchestration; templates call `flagPage.*(page)` with computed `page` = `castFlagPage(this)` / `castFlagsList(this)`
- New REST: extend `api/*`; multi-step calls in `api/flags.ts`; UI via `helpers/runApi`
- UI architecture: **`docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md`** (As-built)

## Constraints

- **Don't edit `swagger_gen/`** — `make swagger`
- Dev mode uses SQLite, no external deps needed
- Process management uses `lsof -ti:<port>` not `pkill -f` — never touches other projects' processes
- See [deepwiki.com/openflagr/flagr](https://deepwiki.com/openflagr/flagr) and `docs/`
- When creating a PR, follow `PULL_REQUEST_TEMPLATE.md`