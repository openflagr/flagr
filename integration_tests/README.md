# Integration tests

HTTP tests against a running Flagr server (`go:build integration`).

## How compose CI runs

`make ci-integration` builds the **current branch** into image `flagr_integration_tests`, starts `docker-compose.yml`, then runs one test binary pass **per server URL** (`FLAGR_SERVER_URLS`).

| Port | Compose service | Image | Role |
|------|-----------------|-------|------|
| 18001 | `flagr_with_sqlite` | `flagr_integration_tests` | Current Flagr + Datar recorder |
| 18002 | `flagr_with_mysql` | `flagr_integration_tests` | Same |
| 18003 | `flagr_with_mysql8` | `flagr_integration_tests` | Same |
| 18004 | `flagr_with_postgres9` | `flagr_integration_tests` | Same |
| 18005 | `flagr_with_postgres13` | `flagr_integration_tests` | Same |
| **18006** | **`checkr_flagr_with_sqlite`** | **`checkr/flagr:1.1.12`** | **Legacy baseline** (pinned upstream release) |

Local `make test-integration` uses a single auto-started binary (always “current”); it does not hit the legacy image unless you point `FLAGR_SERVER_URL` at it.

## Test coverage vs backend

**Pass or skip on every backend (including legacy 1.1.12)** — no capability gate at test start:

| Test | Notes |
|------|--------|
| `TestIntegration_Health` | |
| `TestIntegration_SegmentCRUD` | Per-flag snapshot list only |
| `TestIntegration_ConstraintCRUD` | |
| `TestIntegration_VariantCRUD` | |
| `TestIntegration_DistributionCRUD` | |
| `TestIntegration_Evaluation` | |
| `TestIntegration_Preload` | |
| `TestIntegration_Export` | |
| `TestIntegration_TagCRUD` | |
| `TestIntegration_BatchEval` | |
| `TestIntegration_BatchEvalOperator` | |

**Current Flagr API only** — `requireCurrentFlagrAPI` at start; **skipped** on legacy when route returns 404 (see `integration_compat.go`):

| Test | Required route |
|------|----------------|
| `TestIntegration_FlagCRUD` | `GET /api/v1/flags/snapshots/max_id` (and snapshot assertions) |
| `TestIntegration_DuplicateFlag` | `POST /api/v1/flags/{flagID}/duplicate` (+ max_id) |
| `TestIntegration_DuplicateFlag_Errors` | `POST /api/v1/flags/{flagID}/duplicate` |

**Optional / recorder-specific** — inline probe; **skipped** when endpoint or Datar is unavailable:

| Test | Skip when |
|------|-----------|
| `TestIntegration_Exposures` | `POST /api/v1/exposures` → 404 |
| `TestIntegration_DatarSummary` | Datar summary not OK |
| `TestIntegration_DatarFlagSummary` | Datar per-flag summary not OK |

Capability gates live in `integration_compat.go` (`requireFlagSnapshotMaxIDAPI`, `requireDuplicateFlagAPI`, `requireOptionalAPI`).

Skipped tests are **not** failures; the job still passes if all non-skipped assertions succeed on that URL.

## Adding a new test

1. If it only uses routes present in **checkr/flagr:1.1.12**, add the test with no gate (legacy compatibility).
2. If it needs **new** routes or snapshot max_id, call the appropriate `requireCurrentFlagrAPI(...)` helper first and document the route in this README.
3. If it needs **Datar / exposures / other optional** stacks, follow `TestIntegration_Exposures` or the Datar tests (probe + `t.Skip`).