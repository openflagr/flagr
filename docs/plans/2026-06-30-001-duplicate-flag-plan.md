# Duplicate Flag + Transactional Snapshots

Clones an existing flag (variants, segments, constraints, distributions, tags) into a new flag with a new key. Delivers [#724](https://github.com/openflagr/flagr/issues/724) with a native backend endpoint and flag-detail UI. Same PR also refactors **all** CRUD mutation handlers so DB changes and flag snapshots commit in **one** transaction; webhooks fire **after** commit.

## Motivation

- No first-class way to copy a flag’s full configuration; manual CRUD is slow and error-prone.
- Client-orchestrated multi-request clones can leave partial state on failure.
- Today `SaveFlagSnapshot` starts its own transaction **after** many handlers already committed mutations — snapshot failure leaves eval cache stale until another change.

## Locked decisions (grill-me, 2026-06-30)

| Topic | Decision |
|-------|----------|
| Delivery | **C** — backend `POST /flags/{flagID}/duplicate` + UI in **one PR** |
| Snapshot atomicity | **B-wide** — every mutation: outer `tx` = mutation + snapshot row + `flag.snapshot_id` update |
| Notifications | **A** — only after successful outer **commit** (never inside open tx) |
| `enabled` on clone | **A** — copy from source (new key not in SDKs yet; safe for debug) |
| New flag `key` | Optional in API; empty → `CreateFlagKey("")` / `NewSecureRandomKey()` (same as create) |
| New flag `description` | Optional in API; empty → `{source description} (cloned)` |
| UI entry | **Flag detail only** (no list action) |
| After duplicate | Stay on **source** detail page |
| Confirm UX | **A** — one-step confirm dialog only (no key/description fields in UI) |
| Success toast | **B** — persistent until dismissed (`duration: 0`); link to `/#/flags/{newId}`; copy: cloned successfully, open clone to update details |
| Source flag | **A** — must be non-deleted; deleted source → **404** (same as `GetFlag`) |
| `entityType` | **A** — copy value + `CreateFlagEntityType(tx, et)` in duplicate tx |
| Route | **A** — `POST /flags/{flagID}/duplicate`, `operationId: duplicateFlag` |
| Tree copy | **A** — mirror segment `rank`, rollout, description; constraints in preload order; distributions with remapped variant IDs, same `percent` / DB `bitmap` |
| Response | **A** — `200` + full preloaded `Flag` (same shape as `GET /flags/{id}`) |
| Webhook operation | **A** — `OperationCreate`, `ComponentFlag`, new flag id/key (no new `duplicate` op) |
| PR shape | **A** — single PR; rely on tests for large refactor |

## API

### `POST /api/v1/flags/{flagID}/duplicate`

| | |
|---|---|
| **Tags** | `flag` |
| **operationId** | `duplicateFlag` |
| **Body** | Optional `duplicateFlagRequest`: `key` (string, optional), `description` (string, optional). Omit body or send `{}` for defaults. |
| **Defaults** | `key` empty/missing → random key via `entity.CreateFlagKey("")`. `description` empty/missing → `util.SafeString(source.Description) + " (cloned)"` (define trim-only source → `" (cloned)"` or `"Untitled (cloned)"` in impl if needed). |
| **Source load** | Scoped `GetFlag`-equivalent preload (`PreloadSegmentsVariantsTags`); not `Unscoped` |
| **Errors** | `404` missing/deleted source; `400` invalid optional `key`; `500` tx/snapshot/map failures |
| **Success** | `200` + `flag` model with segments, variants, tags (and nested constraints/distributions on segments) |

**Do not copy:** source `id`, `snapshot_id`, timestamps, `created_by`/`updated_by` from source (set `CreatedBy` from request subject like `CreateFlag`).

**Copy:** `description` (after default rule), `notes`, `dataRecordsEnabled`, `entityType`, `enabled`, all variants (key + attachment), all segments (rank, rollout, description), constraints, distributions (remap `variant_id`), tags (reuse global tag rows via `flags_tags` association).

## Transactional snapshot refactor (B-wide)

### Goal

Handlers that today: (1) mutate DB, (2) call `entity.SaveFlagSnapshot(getDB(), …)` in a **separate** transaction — instead: **one** `tx := getDB().Begin()`, perform mutation(s), write snapshot **on `tx`**, `Commit`, then notify.

### `pkg/entity/flag_snapshot.go`

1. Extract **`writeFlagSnapshotTx(tx, flagID, updatedBy, operation, componentType, componentID, componentKey) error`** (name TBD):
   - `Unscoped` load flag by `flagID` on **`tx`**
   - `Preload` on **`tx`**
   - Marshal, `Create` `FlagSnapshot`, update flag `UpdatedBy` + `SnapshotID` on **`tx`**
   - **No** `Begin`/`Commit` inside; **no** notification inside
   - Return error on failure (caller rolls back)

2. **`SaveFlagSnapshot(db, …)`** becomes thin wrapper for existing call sites during migration:
   - `tx := db.Begin()` → `writeFlagSnapshotTx` → `Commit` → existing notify/diff/`logFlagSnapshotUpdate` logic (unchanged semantics post-commit)

3. Mutation handlers use pattern:

```text
tx := getDB().Begin()
// ... mutations on tx ...
if err := entity.writeFlagSnapshotTx(tx, flagID, subject, op, component, componentID, componentKey); err != nil {
  tx.Rollback()
  return 5xx
}
if err := tx.Commit().Error; err != nil { ... }
notifyAfterSnapshot(...) // same payload as today
```

4. **Notifications** stay **after** `Commit` only (grill Q10 **A**).

### Handler scope

Refactor all mutation paths in `pkg/handler/crud.go` and `pkg/handler/crud_flag_creation.go` that call `SaveFlagSnapshot`, including:

- `CreateFlag`, `PutFlag`, `DeleteFlag`, `RestoreFlag`, `SetFlagEnabledState`
- Tag create/delete
- Segment create/put/delete/reorder
- Constraint create/put/delete
- Distribution put
- Variant create/put/delete
- **New** `DuplicateFlag`

Handlers that already use a tx for part of the work (`PutDistributions`, `PutSegmentsReorder`, `CreateFlag` + template) should extend that tx to include snapshot write before commit, then notify once.

### Enforcement

- Keep / extend **`TestAllMutationHandlersCallSaveFlagSnapshot`** — may need to accept `writeFlagSnapshotTx` as well as `SaveFlagSnapshot`.
- **`crud_notification_test.go`** — still eventually assert notifications after mutations.
- **`eval_cache_test.go`** — snapshot still bumps `max(snapshot.id)` after commit.

## Duplicate implementation

### Handler: `DuplicateFlag`

1. Validate source exists (scoped `First` + preload); else `404`.
2. Parse optional body; resolve `key` and `description` defaults.
3. `tx := getDB().Begin()`.
4. Create new `entity.Flag` row (metadata + `CreatedBy`).
5. If `entityType` non-empty: `CreateFlagEntityType(tx, et)`.
6. Create variants; build `map[oldVariantID]newVariantID` (and key map if needed).
7. For each source segment (stable order by `rank` then id):
   - Create segment on new flag (same rank, rollout, description).
   - Create constraints.
   - Create distributions with remapped `VariantID`, same `Percent`, copy `Bitmap` from source row.
8. Associate tags (`Association("Tags").Append`) on **`tx`** — same pattern as `CreateTag` but batched inside duplicate tx.
9. `writeFlagSnapshotTx(tx, newFlagID, subject, OperationCreate, ComponentFlag, newFlagID, newKey)`.
10. `Commit`.
11. Post-commit notification (create on **new** flag).
12. Preload new flag on `getDB()` for response mapping (or preload on tx before commit if already in memory).
13. Return `duplicateFlagOK` with `e2rMapFlag`.

Consider extracting **`cloneFlagGraph(tx, source *entity.Flag, newFlag *entity.Flag, …) error`** in `pkg/handler/` or `pkg/entity/` for unit tests without HTTP.

### Swagger

- New `swagger/flag_duplicate.yaml` (or section in `swagger/flag.yaml`).
- Register in `swagger/index.yaml`.
- Run `make swagger`; commit `swagger_gen/` + `cmd/flagr-server/main.go` per AGENTS.md.
- Wire `api.FlagDuplicateFlagHandler` in `pkg/handler/handler.go`.

## UI (`browser/flagr-ui`)

| Area | Change |
|------|--------|
| **Placement** | Duplicate control on flag detail only (e.g. near Delete / config header in `Flag.vue` / `FlagConfigCard`) |
| **Flow** | Confirm dialog → `duplicateFlag(flagId)` → no navigation |
| **API** | `api/crud.ts`: `duplicateFlag(flagId, body?)` → `ApiResult<Flag>`; types for optional payload |
| **Page** | `flagPage.ts`: `duplicateFlag(vm)` using `confirmAndRunApi` or dialog + `runApi` |
| **Toast** | `onSuccess`: custom `ElMessage` with `duration: 0`, `showClose: true`, action/link to clone via `router.push` or `<a href="#/flags/{id}">` |
| **List** | No duplicate button in `Flags.vue` |

## Testing

| Layer | What |
|-------|------|
| **Unit** | `DuplicateFlag` happy path: full graph, variant ID remap, default key/description, optional body overrides; `404` deleted/missing source; duplicate `key` → `400` |
| **Unit** | `writeFlagSnapshotTx` rollback leaves no snapshot row when mutation fails mid-tx |
| **Unit** | Existing `crud_*_test.go` + notification tests still pass after B-wide refactor |
| **Integration** | `integration_tests`: duplicate seeded flag; `GET` clone matches structure; eval optional smoke |
| **E2E** | Playwright on flag detail: confirm duplicate → persistent toast with link → navigate to clone via link; verify clone description suffix and segments |

### Verification commands (repo root)

```bash
make test                    # lint + swagger validate + go unit tests
make test-integration        # if handler/API behavior touched
make flagr-ui-check          # ESLint + vue-tsc + Vitest
make test-e2e                # UI + duplicate flow
make ci-swagger              # before push if swagger changed
```

## Files (expected touch)

```
swagger/flag_duplicate.yaml
swagger/index.yaml
swagger_gen/...               # make swagger
cmd/flagr-server/main.go      # if regenerated
pkg/entity/flag_snapshot.go
pkg/entity/flag_snapshot_test.go
pkg/handler/crud.go
pkg/handler/crud_flag_creation.go
pkg/handler/crud_duplicate.go          # or crud_flag_duplicate.go
pkg/handler/crud_duplicate_test.go
pkg/handler/crud_test.go
pkg/handler/crud_notification_test.go
pkg/handler/handler.go
browser/flagr-ui/src/api/crud.ts
browser/flagr-ui/src/api/types.ts
browser/flagr-ui/src/pages/flagPage.ts
browser/flagr-ui/src/components/Flag.vue
browser/flagr-ui/e2e/flag-detail.spec.ts
integration_tests/integration_test.go    # optional new case
docs/flagr_notifications.md              # only if wording needs “duplicate = create”
```

## Implementation order (single PR)

1. **Snapshot refactor** — `writeFlagSnapshotTx` + migrate all mutation handlers + tests green.
2. **Swagger + `DuplicateFlag` handler** + unit tests.
3. **UI** + e2e.
4. **Integration test** (if time).
5. **Docs** — short note in API docs / home if exposed publicly.

## Out of scope (v1)

- Duplicate from flags list or bulk duplicate.
- New notification operation `duplicate`.
- Copying from soft-deleted flags.
- UI fields for optional `key` / `description` (API remains available for scripts/forks).
- Changing eval-only mode export behavior.

## Risks / notes

- **Large diff:** B-wide touches every mutation; treat `make test` + integration as gate, not optional.
- **Tag association inside tx:** Must use `tx` session for `Model(flag).Association("Tags")` — verify GORM behavior in tests.
- **Distribution bitmap:** Copy stored `Bitmap` + `Percent` from source entity rows; do not rely on API-exposed bitmap (json omitted).
- **GORM version:** `go.mod` pins gorm; avoid upgrades in this PR.

## Decisions log

Grill-me session (2026-06-30): choices in table above; user confirmed B-wide transactional snapshots, single PR, API optional fields with simple UI confirm, persistent toast with link to clone.