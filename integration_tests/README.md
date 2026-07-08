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

**Current Flagr API only** — skipped **only** on legacy `:18006` when the router reports `path … was not found`; on `18001`–`18005` and local auto-start, missing routes **fail** the run:

| Test | Required route |
|------|----------------|
| `TestIntegration_FlagCRUD` | `GET /api/v1/flags/snapshots/max_id` (and snapshot assertions) |
| `TestIntegration_DuplicateFlag` | `POST /api/v1/flags/{flagID}/duplicate` (+ max_id) |
| `TestIntegration_DuplicateFlag_Errors` | `POST /api/v1/flags/{flagID}/duplicate` |
| `TestIntegration_BuiltInContext` | `isLegacyIntegrationBaseline` skip (route exists on legacy but `@ts` injection does not) |
| `TestIntegration_BuiltInContextHTTPHeader` | `isLegacyIntegrationBaseline` skip (route exists on legacy but `@http_*` injection does not) |
**Optional / recorder-specific** — `requireOptionalAPI` (router 404 → skip on legacy only); Datar uses `requireRecorderEndpointOK` (non-200 → skip on legacy, **fail** on current images with Datar enabled):

| Test | Gate |
|------|------|
| `TestIntegration_Exposures` | `requireOptionalAPI` → `POST /api/v1/exposures` |
| `TestIntegration_GetEvaluation` | `requireOptionalAPI` → `GET /api/v1/evaluation` (schema **400**, malformed JSON **400**, missing `json` **422**) |
| `TestIntegration_GetEvaluation_QueryURLBytesLimit` | `requireOptionalAPI` → GET eval; local auto-start only (8192-byte query cap) |
| `TestIntegration_DatarSummary` | optional route probe + `requireRecorderEndpointOK` |
| `TestIntegration_DatarFlagSummary` | same |

Capability gates: `integration_compat.go` (`responseIndicatesRouteNotRegistered`, `requireFlagSnapshotMaxIDAPI`, `requireDuplicateFlagAPI`, `requireOptionalAPI`, `requireRecorderEndpointOK`).

### Skip vs fail (reliability)

- **Legacy only (`http://…:18006`)**: skip when the swagger router returns `path … was not found`, when the method is not allowed (**405**, e.g. GET `/evaluation` on checkr/flagr:1.1.12), or when optional recorder/Datar returns non-200.
- **Current images (`18001`–`18005`, local auto-start)**: gated routes must be registered — otherwise **`Fatal`**. Datar summary probes must return **200** — otherwise **`Fatal`**.
- Probes distinguish **router 404** from **application 404** (duplicate gate uses `POST /flags/999999999/duplicate` so current servers prove the handler exists without cloning a flag).

Skipped tests are **not** failures; the job still passes if all non-skipped assertions succeed on that URL.

## Adding a new test

1. If it only uses routes present in **checkr/flagr:1.1.12**, add the test with no gate (legacy compatibility).
2. If it needs **new** routes or snapshot max_id, call the appropriate helper first and document the route in this README.
3. If it needs **Datar / exposures / other optional** stacks, use `requireOptionalAPI` and, when the feature must be enabled on current images, `requireRecorderEndpointOK`.
4. If the route exists on legacy but the **behavior** is new (e.g. built-in context injection), use a direct `isLegacyIntegrationBaseline()` skip at the top of the test — the compat probe approach cannot distinguish behavior differences, only missing routes.