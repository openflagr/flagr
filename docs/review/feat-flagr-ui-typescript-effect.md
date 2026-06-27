# Reviewer guide: `feat/flagr-ui-typescript-effect`

Use this page to gain **high confidence** without reading the full diff (~3k LOC). Proof is layered: fast gates → unit → browser e2e → what did *not* change.

## 30-second verdict

| Question | Answer |
|----------|--------|
| Does the Go API change? | **No** (UI + Makefile/CI/docs only; backend binary unchanged for this feature) |
| Does user-visible behavior intentionally change? | **No** — same routes, toasts, 401 redirect, CRUD flows |
| What could still break? | UI-only regressions (wiring, state refresh). Mitigated by **typecheck + Playwright** |
| Where is the risk concentrated? | `pages/flagPage.ts` orchestration (~430 lines), `api/http.ts` + `runApi` boundary |

## Proof ladder (run in order)

Copy-paste from repo root:

```bash
# 1) UI static + unit (no server) — ~10s
make flagr-ui-check

# 2) Full browser regression — build + above + Playwright (~2–5 min)
make test-e2e
```

**CI alignment:** `make ci` (Go) is unchanged for this PR’s core; UI gates are `make build-ui` / `make test-e2e` (see `.github/workflows/ci.yml`).

### What each gate proves

| Gate | Proves |
|------|--------|
| `npm run lint` | No obvious TS/Vue footguns in `src/` + `e2e/` |
| `npm run typecheck` (`vue-tsc`) | Props, page VMs, API types consistent — catches #720-class `ReferenceError`/shape bugs at compile time |
| `npm run test` (Vitest) | `requestJson` (204, 401, decode, 5xx) and `listFlagsIfStale` cache logic |
| `make test-e2e` (Playwright) | Real backend `:18000` + Vite UI `:8080`: list, create, flag detail CRUD, segments reorder, distributions, etc. |

**Verified on branch (2026-06-27):** `make flagr-ui-check` + `make test-e2e` → **30/30** Playwright tests passed (`e2e/*.spec.ts` only; `playwright.config.mjs` `testDir: 'e2e'`).

## Playwright coverage map

Specs live in `browser/flagr-ui/e2e/`:

| Spec | Confidence for |
|------|----------------|
| `smoke.spec.ts` | App shell, nav, title |
| `flags-list.spec.ts` | List load, search, create flag / boolean flag |
| `flag-detail.spec.ts` | Flag config, variants, segments (incl. reorder), constraints, distributions, **debug console eval** (`.dc-summary` after POST; disabled flags show variant `—`), delete guards |

E2e uses the same stack as dev: `scripts/e2e-server.sh` → `make build` (if needed) + `flagr` + `make run-ui`.

## Architecture (what reviewers should expect)

```
Vue template `flagPage.*(page)`  →  pages/*  →  runApi(promise)  →  api/*
```

- **Stack** — plan **As-built**: `docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md` (`ApiResult`, no Effect).
- **Types** — single module `src/api/types.ts`; e2e helpers import from `../src/api/types`.
- **`page`** — computed `castFlagPage(this)` / `castFlagsList(this)`; templates use `flagPage.*(page, …)` (one `vm`-first contract; tab click is the only thin `methods` wrapper).
- **API composition** — multi-step REST in `api/flags.ts` (`loadFlagPageContext`, `createTagAndRefreshAllTags`, `deleteTagAndReload`, …), not nested `runApi` in pages.
- **Presentational** `FlagHistory` / `DebugConsole` — I/O in `flagPage.ts`; `evalSummaryFromResult` shows rendered result whenever eval identifies a flag (not only when `segmentDebugLogs` is non-empty).

## Suggested review order (60 min → 15 min)

1. Read this file + PR “Verification” section (paste `make test-e2e` output).
2. Skim plan **As-built** (rules for new code).
3. Spot-check `api/http.ts`, `helpers/runApi.ts`, `pages/flagPage.ts` (mount + one mutation).
4. Trust gates for the rest unless you care about Makefile/CI wording.

## Out of scope / known non-goals

- No Vue `<script setup>` rewrite
- No OpenAPI codegen into the frontend
- `flagPage.ts` **not** split into submodules (explicit product choice — single orchestration file)

## Post-merge maintainer commands

```bash
make help          # catalog
make start         # dev backend + UI
make rebuild-run   # after Go/UI changes
make test-e2e      # before release UI changes
```