# AGENTS.md

Flagr — Go feature flag service with Vue 3 UI.

## Commands

| Command | What it does |
|---|---|
| `make build` | Build Go server binary (`./flagr`) |
| `make build_ui` | Build UI for production (`browser/flagr-ui/dist/`) |
| `make start` | Run backend (`:18000`) + frontend dev server (`:8080`) in parallel |
| `make stop-ui` | Kill processes on `:18000` or `:8080` (port-based, not `pkill`) |
| `make rebuild-run` | `build` → `stop-ui` → `start` — one step after Go changes |
| `make test` | Go unit tests |
| `make test-e2e` | Build Go binary → start servers via `scripts/e2e-server.sh` → Playwright → cleanup |
| `make swagger` | Regenerate `swagger_gen/` from OpenAPI spec |
| `go build -o flagr-validate ./cmd/flagr-validate/` | Build standalone JSON flag validator |

**UI-only** (`browser/flagr-ui/`): `npm run dev` (Vite), `npm run build`, `npm run test:e2e` (needs servers running).

## Key Code

### Backend (`pkg/`)
- `handler/crud.go` — CRUD API handlers, `handler/eval.go` — evaluation engine
- `entity/` — domain models (flag, segment, constraint, variant, distribution)
- `mapper/entity_restapi/` — conversions between entities and API models

### Frontend (`browser/flagr-ui/src/`)
- `components/Flag.vue` — flag detail page (orchestrates sub-components)
- `components/Flags.vue` — flag list page with search/filter
- `components/FlagConfigCard.vue` — flag key/description/tags/notes editor
- `components/VariantsSection.vue` — variant CRUD
- `components/SegmentsSection.vue` — segment + constraint + distribution display
- `components/DistributionDialog.vue` — distribution editing modal
- `components/DebugConsole.vue` — inline eval request/response tool
- `components/FlagHistory.vue` — snapshot diff viewer
- `components/MarkdownEditor.vue` — flag notes with markdown + KaTeX
- `helpers/helpers.js` — utility functions (`pluck`, `sum`, `handleErr`)
- `constants.js` — env-var-backed config (`VITE_API_URL`, entity types)

## Workflows

**Frontend-only dev:** `npm run dev` in `browser/flagr-ui/` (Vite proxies `/api/v1` to `:18000`).

**Backend changes:** Frontend auto-reloads via Vite HMR. Backend needs rebuild: `make rebuild-run`.

**E2E tests:** `make test-e2e` — single command. Uses `scripts/e2e-server.sh` (idempotent, port-safe) and Playwright's `webServer` lifecycle. Always works regardless of leftover processes.

**Process management** uses `lsof -ti:<port>` not `pkill -f`, so it never touches processes from other projects.

## Constraints

- **Don't edit `swagger_gen/`** — regenerate with `make swagger`
- Dev mode uses SQLite, no external deps needed
- See [deepwiki.com/openflagr/flagr](https://deepwiki.com/openflagr/flagr) and `docs/`
