# flagr-ui

Vue 3 + Vite + TypeScript. **Commands live in the repo root `Makefile`** — run `make help` there.

| Goal | From repo root |
|------|----------------|
| Dev server (`:8080`) | `make run-ui` (or `make start` with backend) |
| Production build | `make build-ui` |
| Playwright e2e | `make test-e2e` |

Layout: `src/api/` · `src/pages/` · `src/components/` · `src/helpers/`

Config: `VITE_API_URL` (default `/api/v1`) in `src/helpers/constants.ts`.

**Architecture (As-built):** [`docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md`](../../docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md) · **PR review:** [`docs/review/feat-flagr-ui-typescript-effect.md`](../../docs/review/feat-flagr-ui-typescript-effect.md)