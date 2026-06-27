## Description

Migrate **`browser/flagr-ui`** from JavaScript to **TypeScript** (Vite transpile; **`vue-tsc --noEmit`** type gate). Replace **axios** + `handleErr` with **`fetch`**, typed **`ApiResult<T>`**, and a single UI interpreter (**`helpers/runApi.ts`**). Options API and user-visible behavior are unchanged; orchestration lives in **`pages/flagPage.ts`** / **`pages/flagsListPage.ts`**.

**Reviewer shortcut:** [`docs/review/feat-flagr-ui-typescript-effect.md`](docs/review/feat-flagr-ui-typescript-effect.md) — proof ladder, e2e map, review order.

**As-built layout:**

| Layer | Location |
|--------|-----------|
| HTTP + errors | `src/api/http.ts`, `errors.ts`, `result.ts`, `flags.ts`, `evaluation.ts`, `types.ts` |
| Page orchestration | `pages/flagPage.ts`, `pages/flagsListPage.ts`, list cache `pages/flagsList.ts` |
| Vue edge | `Flag.vue` / `Flags.vue` — templates call `flagPage.*(page)` / `flagsListPage.*(page)`; `page` = `castFlagPage(this)` / `castFlagsList(this)` |
| Composed API flows | `listFlagsIfStale`, `loadFlagPageContext`, `createTagAndRefreshAllTags`, `deleteTagAndReload` in `api/flags.ts` |
| Docs | Canonical: [`docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md`](docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md) (**As-built**). No Effect runtime (removed). |

**Repo ergonomics:** Root **`Makefile`** (`make help`): `flagr-ui-check`, `build-ui`, `run-ui`, `start`, `test-e2e`. CI uses `build-ui` and `test-e2e`. E2e: `scripts/e2e-server.sh` → `make build` + `make run-ui`.

## Motivation and Context

- Catch UI regressions at **typecheck** time, not only in e2e/runtime ([#720](https://github.com/openflagr/flagr/pull/720)-class bugs: missing symbols, wrong shapes on flag detail).
- One typed error channel (`ApiError` + `runApi`) instead of scattered axios callbacks.
- Align local dev and CI on **`make`** targets for the UI.

**Canonical context:** migration plan **As-built** + reviewer guide (links above).

## How Has This Been Tested?

From repo root:

```bash
make flagr-ui-check   # lint + vue-tsc + Vitest (6 tests)
make test-e2e         # above + go build + Playwright (e2e/*.spec.ts)
```

| Layer | Command | What it guards |
|-------|---------|----------------|
| Static | `vue-tsc --noEmit` | SFCs, pages, `api/*` |
| Unit | `npm run test` | HTTP decode/errors, `listFlagsIfStale` |
| Browser | Playwright | Smoke, flags list, flag detail (CRUD, segments, distributions) |

**Verification (re-run before merge):**

```text
make flagr-ui-check  # lint + vue-tsc + Vitest 6/6
make test-e2e        # Playwright 30 passed
```

**E2e environment:** Flagr `:18000`, Vite UI `:8080` (same as `make start`).

### Not changed

- Go **`pkg/`** REST contract (UI calls existing paths).
- Playwright **intent** (flows unchanged; specs renamed `.ts`).

## Types of changes

- [ ] Bug fix
- [x] New feature (TS migration + typed API layer, Makefile/CI/docs)
- [ ] Breaking change (API/server)

## Checklist

- [x] Code style (`npm run lint`)
- [x] Documentation (`AGENTS.md`, root `README.md`, `browser/flagr-ui/README.md`, migration plan As-built, reviewer guide)
- [x] Tests added (Vitest `src/api/*.test.ts`; Playwright e2e)
- [x] All new and existing tests passed — **`make flagr-ui-check` + `make test-e2e`**

## Review focus (optional)

1. `api/http.ts` + `api/errors.ts` — failure mapping, 401 `WWW-Authenticate`
2. `helpers/runApi.ts` — `ApiResult`, toasts, redirect
3. `pages/flagPage.ts` — single orchestration file (by design); uniform `flagPage.*(page, …)` handlers
4. `playwright.config.mjs` — `testDir: 'e2e'`