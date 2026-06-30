# Duplicate Flag + Transactional Snapshots

Clones an existing flag (variants, segments, constraints, distributions, tags) into a new flag with a new key. Delivers [#724](https://github.com/openflagr/flagr/issues/724) with a native backend endpoint and flag-detail UI. Same PR also refactors **all** CRUD mutation handlers so DB changes and flag snapshots commit in **one** transaction; webhooks fire **after** commit.

**Status:** implemented (2026-06-30). Branch `feat/dup-flag`.

## Motivation

- No first-class way to copy a flag’s full configuration; manual CRUD is slow and error-prone.
- Client-orchestrated multi-request clones can leave partial state on failure.
- Previously `SaveFlagSnapshot` started its own transaction **after** many handlers already committed mutations — snapshot failure left eval cache stale until another change.

## Locked decisions (grill-me, 2026-06-30)

| Topic | Decision |
|-------|----------|
| Delivery | **C** — backend `POST /flags/{flagID}/duplicate` + UI in **one PR** |
| Snapshot atomicity | **B-wide** — every mutation: outer `tx` = mutation + snapshot row + `flag.snapshot_id` update |
| Notifications | **A** — only after successful outer **commit** (never inside open tx) |
| `enabled` on clone | **A** — copy from source |
| New flag `key` | Optional in API; empty → `CreateFlagKey("")` |
| New flag `description` | Optional in API; empty → `{source description} (cloned)` or `(cloned)` if source empty |
| UI entry | **Flag detail only** (no list action) |
| After duplicate | Stay on **source** detail page |
| Confirm UX | **A** — one-step confirm dialog only (no key/description fields in UI) |
| Success toast | **B** — persistent until dismissed; link to `/#/flags/{newId}` |
| Source flag | **A** — must be non-deleted; deleted source → **404** |
| `entityType` | **A** — copy value + `CreateFlagEntityType(tx, et)` in duplicate tx |
| Route | **A** — `POST /flags/{flagID}/duplicate`, `operationId: duplicateFlag` |
| Tree copy | **A** — segment rank, rollout, description; constraints; distributions with remapped variant IDs, same `percent` / DB `bitmap` |
| Response | **A** — `200` + full preloaded `Flag` |
| Webhook operation | **A** — `create` on **new** flag (`component_type: flag`); no separate `duplicate` op |
| PR shape | **A** — single PR |

## API

### `POST /api/v1/flags/{flagID}/duplicate`

| | |
|---|---|
| **Tags** | `flag` |
| **operationId** | `duplicateFlag` |
| **Body** | Optional `duplicateFlagRequest`: `key`, `description` (both optional). `{}` for defaults. |
| **Defaults** | Random `key` via `entity.CreateFlagKey("")`. Description: `util.SafeString(source.Description) + " (cloned)"`, or `"(cloned)"` if source description empty. |
| **Source load** | Scoped preload (`PreloadSegmentsVariantsTags`); not `Unscoped` |
| **Errors** | `404` missing/deleted source; `400` invalid optional `key`; `500` tx/snapshot failures |
| **Success** | `200` + `flag` with segments, variants, tags (nested constraints/distributions) |

**Snapshots:** Exactly **one** new `flag_snapshot` on the **clone**; **zero** on the source flag.

**Copy:** `description` (after default), `notes`, `dataRecordsEnabled`, `entityType`, `enabled`, variants, segments, constraints, distributions (remapped `variant_id`), tags.

**Do not copy:** source `id`, `snapshot_id`, timestamps; `CreatedBy` from request subject.

## As-built: transactional snapshots (B-wide)

### `pkg/handler/crud_snapshot.go`

- **`commitFlagMutation(snapshotFlagID, subject, operation, componentType, mutate)`** — `Begin` → `mutate(tx)` → **`entity.WriteFlagSnapshotTx(tx, flagID, subject)`** → `Commit` → **`entity.NotifyFlagSnapshot`** (post-commit).
- Use `snapshotFlagID == 0` when the new flag ID is assigned inside `mutate` (duplicate, create boolean).

### `pkg/entity/flag_snapshot.go`

- **`WriteFlagSnapshotTx(tx, flagID, updatedBy)`** — load flag on `tx`, preload, marshal, insert snapshot, update flag `SnapshotID` / `UpdatedBy`; no commit, no notify.
- **`NotifyFlagSnapshot`** — webhook + logging after commit.
- **`SaveFlagSnapshot(db, …)`** — thin wrapper for export and legacy paths (own transaction).

### Enforcement

- **`TestAllMutationHandlersCallSaveFlagSnapshot`** — AST guard: handlers must not call snapshot writers except via `crud_snapshot.go`.
- **`TestCommitFlagMutation_RollbackOnMutateFailure`** — rollback leaves no snapshot row.
- **`crud_notification_test.go`** — includes **DuplicateFlag sends notification** (filter by clone flag ID).

## As-built: flag template API

Graph writes for **boolean create** and **duplicate** share one path in `pkg/entity/flag_template.go`:

| Symbol | Role |
|--------|------|
| **`SimpleBooleanFlagTemplate()`** | Starter graph (`on` variant, 100% rollout) for `POST /flags` with `template: simple_boolean_flag` |
| **`SourceFlagTemplate(source *Flag)`** | Clone graph from source (keys preserved; IDs omitted) |
| **`ApplyFlagTemplate(tx, flagID, template)`** | Persist variants, segments, constraints, distributions, tags on `tx` |

**Duplicate handler** (`pkg/handler/crud_duplicate.go`): create flag row + optional `CreateFlagEntityType` → **`ApplyFlagTemplate(tx, created.ID, SourceFlagTemplate(source))`** inside **`commitFlagMutation`**.

**Boolean create** (`pkg/handler/crud_flag_creation.go`): **`ApplyFlagTemplate(..., SimpleBooleanFlagTemplate())`**.

`pkg/entity/flag_clone.go` retains **`AppendTagValueToFlag`** only.

## UI (`browser/flagr-ui`)

| Area | As-built |
|------|----------|
| **Placement** | **Duplicate Flag** on flag detail (`Flag.vue`, `data-testid="duplicate-flag-btn"`) |
| **Flow** | Confirm dialog → `crud.duplicateFlag(flagId)` → stay on source |
| **Toast** | `flagPage.ts`: `duration: 0`, link `a.duplicate-flag-toast-link` → `/#/flags/{id}` |
| **API** | `api/crud.ts` `duplicateFlag`, `DuplicateFlagPayload` in `types.ts` |

## Testing (as-built)

| Layer | Coverage |
|-------|----------|
| **Unit** | `crud_duplicate_test.go`, `flag_template_test.go`, `crud_flag_creation_test.go`, notification duplicate subtest |
| **Integration** | `TestIntegration_DuplicateFlag` — graph parity, custom key/description, **source snapshot count unchanged**, **clone exactly 1 snapshot**, global max snapshot id increases; CRUD tests assert snapshots on segment/variant/tag mutations |
| **E2E** | `flag-detail.spec.ts` — duplicate, toast link, navigate to clone, API/UI parity |
| **Local integration server** | `FLAGR_RECORDER_DATAR_FLUSH_INTERVAL=500ms`, `FLAGR_EVALCACHE_REFRESHINTERVAL=1s`; Datar tests use `entityContext.tier=premium` for seeded flag |

### Verification commands

```bash
make test
make test-integration
make flagr-ui-check
make test-e2e
make ci-swagger   # before push if swagger touched
```

## Files (as-built)

```
swagger/flag_duplicate.yaml
pkg/entity/flag_template.go
pkg/entity/flag_template_test.go
pkg/entity/flag_snapshot.go
pkg/entity/flag_clone.go
pkg/handler/crud_snapshot.go
pkg/handler/crud_duplicate.go
pkg/handler/crud_duplicate_test.go
pkg/handler/crud_flag_creation.go
pkg/handler/crud.go (+ other CRUD handlers via commitFlagMutation)
browser/flagr-ui/src/api/crud.ts
browser/flagr-ui/src/pages/flagPage.ts
browser/flagr-ui/src/components/Flag.vue
browser/flagr-ui/e2e/flag-detail.spec.ts
integration_tests/integration_test.go
integration_tests/integration_server_test.go
docs/flagr_overview.md
docs/flagr_notifications.md
```

## Out of scope (v1)

- Duplicate from flags list or bulk duplicate.
- New notification operation `duplicate`.
- Copying from soft-deleted flags.
- UI fields for optional `key` / `description` (API available for scripts).
- Eval-only mode export behavior unchanged.

## Risks / notes (resolved)

- **B-wide refactor:** All mutation handlers use `commitFlagMutation`; export still uses `SaveFlagSnapshot` in a separate tx.
- **Duplicate snapshots:** One row on new flag only — integration test enforces.
- **Distribution bitmap:** Copied from source entity rows in `ApplyFlagTemplate`.