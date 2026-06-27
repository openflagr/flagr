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

CI: `make ci`, `make ci-swagger`, `make ci-integration`, `make build-ui` — see `Makefile` **CI** section.

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
- UI architecture: **`docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md`** (As-built); reviewer guide **`docs/review/feat-flagr-ui-typescript-effect.md`**

## Constraints

- **Don't edit `swagger_gen/`** — `make swagger`
- Dev mode uses SQLite, no external deps needed
- Process management uses `lsof -ti:<port>` not `pkill -f` — never touches other projects' processes
- See [deepwiki.com/openflagr/flagr](https://deepwiki.com/openflagr/flagr) and `docs/`
- When creating a PR, follow `PULL_REQUEST_TEMPLATE.md`