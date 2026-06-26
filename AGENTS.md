# AGENTS.md

Flagr — Go feature flag service with Vue 3 UI.

## Commands

| Command | What it does |
|---|---|
| `make build` | Go server → `./flagr` |
| `make build-ui` | UI: lint, typecheck, Vite → `browser/flagr-ui/dist/` |
| `make start` | Backend `:18000` + UI dev `:8080` |
| `make stop-ui` | Free ports `:18000` / `:8080` (`lsof`, not `pkill`) |
| `make rebuild-run` | `build` → `stop-ui` → `start` |
| `make test` | Go unit tests |
| `make test-e2e` | `build` + UI lint/typecheck + Playwright |
| `make swagger` | Regenerate `swagger_gen/` |
| `make test-integration` | Go API integration tests (SQLite `:memory:`) |
| `make bench-integration` | HTTP eval benchmarks |
| `go build -o flagr ./cmd/flagr-server/` | Same as `make build` |

**UI** (`browser/flagr-ui/`): `npm run dev` / `build` / `lint` / `typecheck` — or `make build-ui` / `make run-ui` from repo root.

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