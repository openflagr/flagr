# feat: Read-only UI for eval-only mode (`json_file` / `json_http`)

**Date:** 2026-07-30 (reworked 2026-08-17 after review)
**Status:** implemented

## Summary

Eval-only mode (`json_file` / `json_http` drivers, or `FLAGR_EVAL_ONLY_MODE=true`)
registers only health, evaluation, and eval-cache export handlers. The UI static
assets are still served (`FLAGR_UI_ENABLED` defaults to true), but every CRUD
call the UI makes returns 501, so the UI is effectively broken on eval-only
deployments.

The flag data is already in memory (EvalCache) and already exposed via
`GET /api/v1/export/eval_cache/json`. This plan makes the UI a **read-only
client of that export endpoint** so operators can browse flags, inspect
segments/variants, and use the Debug Console on eval-only (GitOps) nodes —
while all writes stay rejected.

> **Review rework (2026-08-17):** the first implementation registered a second
> 32-method `CRUD` implementation (`crud_readonly.go`) serving reads from the
> EvalCache and overloading `GET /flags/snapshots/max_id` with a content
> fingerprint. Review (PR #764) asked to drop the dual CRUD implementation —
> a permanent maintenance tax whose filter semantics would drift from GORM —
> and serve the read-only UI from the existing export dump plus a frontend
> mapper. This document describes the as-built (reworked) design.

## Design

### Backend (small)

1. **Eval-only registered surface is unchanged from main**: health, evaluation,
   eval-cache export. CRUD routes stay unregistered (reads hit the generated
   501s; the UI never calls them in this mode).

2. **Write denial: 403 middleware** (`evalOnlyDeny` in
   `pkg/config/middleware.go`). In eval-only mode, any non-GET/HEAD/OPTIONS
   request under `<WebPrefix>/api/v1/flags` returns **403** with a message
   pointing at the JSON source. Every CRUD write endpoint lives under
   `/api/v1/flags`, so one method+prefix check covers all 19 write operations —
   no `CRUD` implementation needed. Evaluation POSTs (`/api/v1/evaluation`)
   pass through. Writes were previously unregistered (501 "not implemented");
   an explicit 403 with a pointer to the JSON source is the real contract.

3. **Mode discovery: `evalOnlyMode` on `GET /health`**

   The UI needs to know it should render read-only. `health` is the one
   endpoint every mode serves, so the response gains an `evalOnlyMode` boolean
   (`swagger/index.yaml` → `make gen`). Response stays backward-compatible.

### Frontend (`browser/flagr-ui/`)

1. **Server mode state** — `src/api/health.ts` (`getHealth`) +
   `src/helpers/serverMode.ts`: module-level reactive `evalOnlyMode` ref,
   fetched once at app start. Fail-open: if health can't be read, assume
   writable (a broken health check shouldn't lock the UI).
2. **Export adapter** — `src/api/evalCache.ts`: fetches
   `GET /export/eval_cache/json` (the GitOps `entity.Flag` shape: PascalCase,
   `{ Flags: [...] }`) and maps it once at the API boundary to the swagger
   camelCase `Flag` the rest of the UI speaks. GORM-only fields (`DeletedAt`,
   `SnapshotID`, `FlagID`, `SegmentID`, timestamps we don't display) are
   ignored. A module-level cache of the mapped dump makes list → detail → back
   a single fetch; concurrent readers share one in-flight request. Flags are
   sorted by ID (the export iterates a Go map — order is random per fetch);
   segment order is kept as exported because in json mode source order **is**
   evaluation order.
3. **Read branch in `crud.ts`** — the read functions check `evalOnlyMode`:
   - `listFlagsIfStale` → always refetch the export on list mount (no change
     token; the dump is small), reverse for newest-first display.
   - `getFlag` / `listAllTags` → from the mapped dump (tags deduped by value —
     in the JSON source the same value on two flags is two entities).
   - `listEntityTypes`, `listFlagSnapshots`, `listDeletedFlags` → resolve `[]`
     locally without network (entity types are DB-only; history lives in Git;
     JSON sources have no soft-deletes).
   Components and pages are unchanged — the branch lives in one module.
4. **Read-only rendering** — components import the ref directly (no prop
   drilling):
   - Global banner: "Read-only (GitOps) mode — flags are managed via the JSON source".
   - Flags list: hide New Flag form and deleted-flags view.
   - Flag page: hide save/delete buttons, enabled toggle, tag add/remove,
     variant/segment/constraint/distribution editors and reorder; hide the
     History tab (no snapshots).
   - **Debug Console stays** — evaluation works in eval-only mode and is the
     main reason to open the UI on an eval edge node.
5. UI hiding is UX, not enforcement — the backend 403 is the gate.

### Docs

- `docs/flagr_behavioral_contracts.md` — eval-only surface stays
  "evaluation + health + export"; the UI is a client of export; writes 403.
- `docs/flagr_env.md`, `docs/flagr_json_flag_spec.md`, `docs/integration.md`,
  `docs/flagr_overview.md` — mention the read-only UI on eval-only deployments.

## Decisions

- **No new env var.** The UI reads the existing eval-cache export endpoint, so
  there is no new exposure surface; auth middlewares (JWT/Basic) apply
  unchanged. An opt-out flag can be added later if a use case appears.
- **Export + mapper, not a second CRUD implementation.** The export payload is
  `entity.Flag` JSON and the UI's types are the swagger models — that gap is a
  mapper, not a new API. The export JSON shape is the GitOps source of truth
  and is not changed to suit the UI; all adaptation happens in
  `evalCache.ts`. (Superseded first cut: `crud_readonly.go`, see rework note.)
- **No server-side change token.** The first cut faked
  `GET /flags/snapshots/max_id` with a content fingerprint; swagger documents
  that endpoint as a monotonic snapshot ID, and external pollers may rely on
  that. The list page simply refetches the export on mount instead.
- **403 via middleware, not handlers.** All write endpoints share one deny
  path; there is nothing per-endpoint about the denial.

## Files changed (as-built)

| File | Change |
|------|--------|
| `swagger/index.yaml` | `health` definition gains `evalOnlyMode` boolean |
| `docs/api_docs/bundle.yaml`, `swagger_gen/` | regenerated (`make gen`) |
| `pkg/handler/handler.go` | health returns `evalOnlyMode` |
| `pkg/config/middleware.go` (+ test) | `evalOnlyDeny`: 403 for non-GET under `/api/v1/flags` in eval-only mode |
| `browser/flagr-ui/src/api/evalCache.ts` (+ test) | export fetch + PascalCase→camelCase mapper + mapped-dump cache |
| `browser/flagr-ui/src/api/crud.ts` (+ test) | read functions branch to the export adapter in eval-only mode |
| `browser/flagr-ui/src/api/health.ts`, `api/types.ts` | `getHealth` + `Health` DTO |
| `browser/flagr-ui/src/helpers/serverMode.ts` (+ test) | reactive `evalOnlyMode` ref, `initServerMode()` (fail-open) |
| `browser/flagr-ui/src/main.ts` | mode resolved before first paint (1.5s bound, fail-open) |
| `browser/flagr-ui/src/App.vue` | read-only banner |
| `browser/flagr-ui/src/components/Flags.vue` | hide Create Flag + Deleted Flags in read-only |
| `browser/flagr-ui/src/components/Flag.vue` | hide Flag Management + History tab; pass `readonly` to sections; snap to Config if mode turns read-only |
| `browser/flagr-ui/src/pages/flagPage.ts` (+ test) | `applyDeepLink` routes history deep links to Config in read-only mode |
| `browser/flagr-ui/src/components/FlagConfigCard.vue` | `readonly` prop: disable inputs/switches, hide save/tag/notes-edit controls |
| `browser/flagr-ui/src/components/VariantsSection.vue` | `readonly` prop: disable key input, read-only attachment editor, hide actions/add row |
| `browser/flagr-ui/src/components/SegmentsSection.vue` | `readonly` prop: disable inputs, hide reorder/new/save/delete/edit-distribution |
| `browser/flagr-ui/src/components/ConstraintExistingRow.vue`, `ConstraintValueCell.vue` | `readonly`/`disabled` props threaded to constraint cells |
| `browser/flagr-ui/e2e/readonly.spec.ts` | Playwright: banner, hidden write affordances, disabled inputs, Debug Console present, history deep link → Config |
| `docs/flagr_behavioral_contracts.md`, `flagr_env.md`, `flagr_json_flag_spec.md`, `integration.md`, `flagr_overview.md` | eval-only contract update |

## Screenshots (json_file source, 3 sample flags)

Flags list — banner, no Create Flag, no Deleted Flags:

![read-only flags list](./assets/readonly-flags-list.png)

Flag page — inputs disabled, no save/delete/reorder/add controls, no History
tab, Debug Console available:

![read-only flag detail](./assets/readonly-flag-detail.png)

`?tab=history` deep link lands on Config:

![history deep link lands on Config](./assets/readonly-history-deeplink.png)

## As-built notes

- Deep-linking `?tab=history` on a read-only instance routes to the Config tab
  (`applyDeepLink` guards on `evalOnlyMode`; a `Flag.vue` watcher covers the
  race where `/health` resolves after the deep link already opened History).
  Change history for JSON-sourced flags lives in Git.
- The app resolves the server mode **before first paint**: `main.ts` awaits
  `initServerMode()` (bounded by a 1.5s timeout) before `app.mount`, so a
  read-only deployment never flashes editable controls. If `/health` exceeds
  the timeout, the app mounts fail-open (editable UI, backend 403 backstop)
  and the first fetch goes through the CRUD path, which 501s on a real
  eval-only server. When the late health response flips `evalOnlyMode`,
  watchers on the list and detail pages refetch through the export path, so
  the UI self-heals instead of spinning forever (covered by the
  "late /health" Playwright tests). `refreshFlags` also ends its loading
  state on failure — error toast + empty state, never an endless spinner.
- In eval-only mode the CRUD read routes return the generated 501s (same as
  main). Tooling that needs flag data from an eval edge node should read
  `GET /api/v1/export/eval_cache/json`, exactly like the UI does.
