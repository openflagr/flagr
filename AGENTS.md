# AGENTS.md

Flagr — Go feature flag service with Vue 3 UI.

## Commands

| Command | What it does |
|---|---|
| `make build` | Build Go server binary (`./flagr`) |
| `make build_ui` | Build UI for production (`browser/flagr-ui/dist/`) |
| `make start` | Backend (`:18000`) + frontend dev server (`:8080`) in parallel |
| `make stop-ui` | Kill processes on `:18000` or `:8080` (port-based via `lsof`, not `pkill`) |
| `make rebuild-run` | `build` → `stop-ui` → `start` — one step after Go changes |
| `make test` | Go unit tests |
| `make test-e2e` | Build Go binary → start servers → Playwright → cleanup |
| `make swagger` | Regenerate `swagger_gen/` from OpenAPI spec |
| `make test-integration` | Go integration tests (auto-starts local server, SQLite `:memory:`) |
| `make bench-integration` | HTTP eval benchmarks against local server |
| `go build -o flagr-validate ./cmd/flagr-validate/` | Build standalone JSON flag validator |
| `go build -o flagr ./cmd/flagr-server/` | Build server binary directly (same as `make build`) |

**UI-only** (`browser/flagr-ui/`): `npm run dev` (Vite), `npm run build`, `npm run typecheck` (`vue-tsc --noEmit`), `npm run lint`, `npm run test:e2e` (needs servers; repo root: `make test-e2e`).

## Key Code

**Backend (`pkg/`):**
- `handler/eval.go` — evaluation engine; `handler/crud.go` — CRUD API handlers
- `handler/exposure.go` — exposure (impression) logging; `handler/data_recorder*.go` — recorders (Kafka, Kinesis, Pub/Sub, Datar)
- `entity/` — domain models (flag, segment, constraint, variant, distribution)
- `config/env.go` — all environment variables (single source of truth)

**Frontend (`browser/flagr-ui/src/`):** `api/`, `pages/`, `components/`, `helpers/` — Vite compiles TS; `npm run typecheck` = `vue-tsc --noEmit`.
- New REST: add functions in `api/flags.ts` or `api/evaluation.ts`; call from UI via `helpers/runApi` (not `fetch` in components).
- Effect usage and conventions: **`docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md`** (§ Effect) and **`browser/flagr-ui/docs/EFFECT.md`**.

## Constraints

- **Don't edit `swagger_gen/`** — regenerate with `make swagger`
- Dev mode uses SQLite, no external deps needed
- Process management uses `lsof -ti:<port>` not `pkill -f` — never touches other projects' processes
- See [deepwiki.com/openflagr/flagr](https://deepwiki.com/openflagr/flagr) and `docs/`
- When creating a PR, follow `PULL_REQUEST_TEMPLATE.md`