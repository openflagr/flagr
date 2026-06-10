# refactor: Migrate integration tests to Go with HTTP benchmarks

**Date:** 2026-06-09
**Status:** implemented

## Summary

Replace the shell/shakedown integration tests (`integration_tests/test.sh`) with a Go test suite that runs via `go test`, seeds ~50 realistic flags covering all 12 constraint types, and provides native Go `BenchmarkXxx` functions over HTTP eval endpoints. Supports three execution modes: local auto-start (SQLite), Docker Compose multi-instance (6 flagr instances with different DB backends), and CI.

## Problem Frame

- Shell tests are slow, fragile, invisible to `go test`/`go test -bench`, and produce no timing metrics
- In-process benchmarks in `pkg/handler/eval_test.go` don't measure HTTP overhead, serialization, or middleware
- Need a clean way to create ~50 flags with all constraint operators for realistic eval scenarios

**Success criteria:**
1. `go test -tags=integration ./integration_tests/` auto-starts a server and runs the suite
2. `cd integration_tests && make test` runs the Go suite against all 6 Docker Compose flagr instances
3. `go test -tags=integration -bench=. ./integration_tests/` produces benchmark results for HTTP eval endpoints
4. ~50 flags seeded covering all 12 constraint operators with diverse entity contexts
5. CI `integration_test` job uses the Go suite

## Key Technical Decisions

1. **Build tag `integration`** — Prevents `go test ./...` from running integration tests. Explicit `-tags=integration` required.

2. **Three-mode TestMain** — `TestMain` reads `FLAGR_SERVER_URL`:
   - **Unset** → Auto-start server (SQLite `:memory:`, random port), run tests, kill server
   - **Set** → Connect to that URL, run tests. Used for Docker Compose targeting
   - The compiled binary (`go test -c`) always requires `FLAGR_SERVER_URL` — local auto-start only works when `go test` is invoked from source

3. **Pre-compiled binary for Docker** — `go test -tags=integration -c -o integration.test ./integration_tests/` produces a fully static binary. Volume-mounted into an `alpine:3` runner service in Docker Compose — no shakedown, no dependencies.

4. **Seeding via HTTP API** — Creates flags/segments/constraints/variants/distributions/tags through the same REST API a client would use. Validates the CRUD pipeline end-to-end. Called once from `TestMain` before any test runs.

5. **Polling replaces sleep** — After seeding, poll the eval cache or run a test eval until the data is available (15s max). Eliminates the shell script's `sleep 5` hack.

6. **No new Go dependencies** — Stdlib HTTP + testify (already in go.mod).

## Scope

**In scope:** Go test file, TestMain, seed generator, 12 test functions ported from shell, 12 HTTP benchmarks, Docker Compose Makefile + alpine runner, root Makefile targets, CI workflow update, deletion of `test.sh`.

**Out of scope:** UI e2e tests, in-process benchmarks.
**Deferred:** `-flagr.url` test flag, stress/soak benchmarks, parallel per-instance execution.

## Architecture

```
                         ┌──────────────────────────┐
                         │       TestMain            │
                         │  FLAGR_SERVER_URL?        │
                         └──────┬───────────────────┘
                                │
              ┌─────────────────┼──────────────────┐
              ▼                 ▼                    ▼
     ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐
     │ Local dev    │  │ BYO server  │  │ Compiled binary  │
     │ auto-start   │  │ single URL  │  │ (Docker mode)    │
     └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘
            └─────────────────┼───────────────────┘
                              ▼
                    ┌─────────────────────┐
                    │ waitForServer(/health)│
                    └──────────┬──────────┘
                               ▼
                    ┌─────────────────────┐
                    │ seedFlags(~52 flags) │
                    └──────────┬──────────┘
                               ▼
                    ┌─────────────────────┐
                    │ waitForEvalCache()  │
                    └──────────┬──────────┘
                               ▼
                    ┌─────────────────────┐
                    │ m.Run()             │
                    │ 12 tests + benches  │
                    └─────────────────────┘
```

Docker mode wraps this in a Makefile loop:
```
for instance in [sqlite, mysql, mysql8, postgres9, postgres13, checkr]:
    docker compose exec -T integration-runner \
        env FLAGR_SERVER_URL=http://$instance:18000 \
        /app/integration.test -test.v
```

## File Layout

```
integration_tests/
├── integration_test.go            # Build tag, test functions
├── integration_server.go          # TestMain, server lifecycle, seed runner
├── integration_client.go          # HTTP client helpers
├── integration_benchmark_test.go  # HTTP benchmarks
├── seed.go                        # Flag spec data (52 flag definitions)
├── integration.test               # Pre-compiled binary (gitignored)
├── docker-compose.yml             # Replaces shakedown with alpine integration-runner
├── Makefile                       # build, test, bench targets
├── Dockerfile-Integration-Test    # Unchanged
└── ...
```

## Implementation Units

### U1. Scaffold — test file + TestMain

**Files:** `integration_tests/integration_test.go` (create)

Create the file with `//go:build integration` tag, `package flagr_integration`.

**TestMain:**
- Read `FLAGR_SERVER_URL` env var
- If empty: build server binary (`go build ./swagger_gen/cmd/flagr-server/`), start on random port with SQLite `:memory:` + `?cache=shared`, register cleanup
- Set `baseURL` package var
- Call `waitForServer(baseURL, 30s)` — poll `/api/v1/health` until 200
- Call `seedFlags(baseURL)` (defined in U2)
- Call `m.Run()`

Package vars: `baseURL`, `seedFlagIDs`, `seedFlagKeys`, `httpClient`.

**Verification:** `go test -tags=integration ./integration_tests/` compiles and runs. `FLAGR_SERVER_URL=http://localhost:18000 go test -tags=integration ./integration_tests/` works against running server.

---

### U2. Seed data + seed runner

**Files:** `integration_tests/seed.go` (create), `integration_tests/integration_test.go` (modify)

**Seed data** — Declarative flag specs covering:
- **48 flags** (4 per operator × 12 operators): EQ, NEQ, LT, LTE, GT, GTE, EREG, NEREG, IN, NOTIN, CONTAINS, NOTCONTAINS
- **4+ extra flags**: multi-segment (3 segments, varying rollout%), complex AND constraints, entity-type-override, disabled flag
- Each flag has 1 segment, 2 variants, 1 distribution (100% to variant_1), 1-2 tags
- Property values use realistic domain names (`region`, `age`, `email`, `tags`, etc.)
- Diverse entity contexts across flags: nested objects, arrays, mixed types, variable references
- Shared tags for batch-by-tag eval testing

**Seed runner** — `seedFlags(baseURL)` iterates specs, calling HTTP API:
- `POST /api/v1/flags` → create flag, get flagID
- `POST /api/v1/flags/{flagID}/segments` → create segment(s)
- `POST /api/v1/flags/{flagID}/segments/{segmentID}/constraints` → create constraints
- `POST /api/v1/flags/{flagID}/variants` → create variants
- `PUT /api/v1/flags/{flagID}/segments/{segmentID}/distributions` → set distributions
- `POST /api/v1/flags/{flagID}/tags` → create tags
- Record flagID/key in package vars

**Verification:** After seeding, `GET /api/v1/flags?preload=true` returns 52+ fully populated flags.

---

### U3. HTTP client helpers

**Files:** `integration_tests/integration_test.go` (modify)

Helpers: `getJSON`, `postJSON`, `putJSON`, `deleteResource`, `requireOK`, `requireStatus`. All use `net/http`, JSON marshal/unmarshal, `Content-Type: application/json`. Test helpers call `t.Fatal` on error; seed helper calls `log.Fatal`. A `httpPostRaw` variant for benchmarks returns the raw response without decoding.

---

### U4. Test functions — port of 12 shell steps

**Files:** `integration_tests/integration_test.go` (modify)

Port shell `step_1` through `step_12` to `TestIntegration_*` functions:

| Test | Shell step | Coverage |
|------|-----------|----------|
| `TestIntegration_Health` | step_1 | GET /health → 200 |
| `TestIntegration_FlagCRUD` | step_2 | Create/get/put/enable/delete/restore flag, set entity type, query with preload |
| `TestIntegration_SegmentCRUD` | step_3 | Create/update/reorder segments |
| `TestIntegration_ConstraintCRUD` | step_4 | Create/update constraints |
| `TestIntegration_VariantCRUD` | step_5 | Create/update variants |
| `TestIntegration_DistributionCRUD` | step_6 | Set/get distributions |
| `TestIntegration_Evaluation` | step_7 | POST /evaluation — match, no-match, disabled, entity type override |
| `TestIntegration_Preload` | step_8 | GET flags with/without preload |
| `TestIntegration_Export` | step_9 | GET /export/sqlite, /export/eval_cache/json |
| `TestIntegration_TagCRUD` | step_10 | Create/list tags |
| `TestIntegration_BatchEval` | step_11 | POST /evaluation/batch with flagTags |
| `TestIntegration_BatchEvalOperator` | step_12 | Batch eval with ANY/ALL operators |

Tests create their own data for mutation scenarios; seeded flags are read-only references. Eval cache warmup uses polling with backoff instead of `sleep 5`. All assertions use `testify/require`.

**Verification:** `make test-integration` passes all 12 tests.

---

### U5. HTTP benchmarks

**Files:** `integration_tests/integration_test.go` (modify)

Add `BenchmarkXxx` functions that measure end-to-end HTTP eval latency. Each pre-builds the request body in `b.StopTimer()`, then calls `http.Post` in the `for i := 0; i < b.N; i++` loop, reading and discarding the response body.

| Benchmark | Endpoint | Scenario |
|-----------|----------|----------|
| `BenchmarkEvalByFlagID` | POST /evaluation | Single eval, flagID |
| `BenchmarkEvalByFlagKey` | POST /evaluation | Single eval, flagKey |
| `BenchmarkEvalNoMatch` | POST /evaluation | No matching constraints |
| `BenchmarkEvalDisabledFlag` | POST /evaluation | Disabled flag |
| `BenchmarkEvalBatchByIDs` | POST /evaluation/batch | 10 flagIDs, 1 entity |
| `BenchmarkEvalBatchByTags` | POST /evaluation/batch | 1 tag, 10 entities |
| `BenchmarkEvalBatchLarge` | POST /evaluation/batch | 50 flagIDs × 5 entities |
| `BenchmarkEvalEQ` | POST /evaluation | Single EQ constraint |
| `BenchmarkEvalIN` | POST /evaluation | Single IN constraint |
| `BenchmarkEvalRegex` | POST /evaluation | EREG constraint |
| `BenchmarkEvalMultiConstraint` | POST /evaluation | 3+ AND'd constraints |
| `BenchmarkEvalNestedContext` | POST /evaluation | Nested entityContext |

Benchmarks only run in local/byo mode (not in Docker Compose loop, which omits `-bench`).

**Verification:** `go test -tags=integration -bench=. -benchmem ./integration_tests/` produces ns/op and B/op.

---

### U6. Docker Compose — alpine runner + Makefile

**Files:** `integration_tests/docker-compose.yml` (modify), `integration_tests/Makefile` (modify)

**docker-compose.yml:** Remove the `shakedown` service. Add:
```yaml
integration-runner:
    image: alpine:3
    container_name: flagr-integration-runner
    command: tail -f /dev/null
    volumes:
      - ./integration.test:/app/integration.test:ro
    depends_on:
      - flagr_with_sqlite
      - flagr_with_mysql
      - flagr_with_mysql8
      - flagr_with_postgres9
      - flagr_with_postgres13
      - checkr_flagr_with_sqlite
```
**Makefile targets:**
- `build-integration-test` — `go test -tags=integration -c -o integration.test ./integration_tests/`
- `test` (default) — Build image + binary, `up`, loop over 6 instances, `docker compose exec -T integration-runner` with `FLAGR_SERVER_URL` per instance. Track pass/total, exit non-zero for any failure.
- `test-instance INSTANCE=x` / `bench-instance INSTANCE=x` — Single-instance helpers
- `retest` — `down test`

**Verification:** `cd integration_tests && make test` passes all 6 instances.

---

### U7. Root Makefile targets

**Files:** `Makefile` (modify)

```makefile
test-integration: build
    go test -tags=integration -count=1 -v ./integration_tests/

bench-integration: build
    go test -tags=integration -bench=. -benchmem -count=1 -run=^$$ ./integration_tests/
```

---

### U8. CI workflow

**Files:** `.github/workflows/ci.yml` (modify)

Update the `integration_test` job to:
1. Add `actions/setup-go@v5` (needed for `go test -c`)
2. `make build-image` + `make down && make up`
3. `make test` (now runs the Go suite against all 6 instances)
4. `if: always()` log collection step

---

### U9. Documentation

**Files:** `AGENTS.md` (modify)

Add entries:
- `make test-integration` — Run Go integration tests (local auto-start)
- `make bench-integration` — Run HTTP eval benchmarks
- `cd integration_tests && make test` — Run against all 6 Docker Compose instances

## Risks

- SQLite `:memory:` requires `?cache=shared` for multi-connection support from subprocess
- `checkr/flagr:1.1.12` may lack newer endpoints — tests report 404 but don't skip
- Benchmark numbers are host-load-sensitive; use `benchstat` for CI comparison
- `CGO_ENABLED=0` is default in root Makefile; compiled binary is fully static

## Sources

- Shell tests: `integration_tests/test.sh` — 12-step CRUD+eval flow, 6-instance orchestration
- In-process benchmarks: `pkg/handler/eval_test.go` — `BenchmarkEvalFlag`, `genBenchmarkEvalCache`
- Entity fixtures: `pkg/entity/fixture.go` — `GenFixtureFlag`
- Constraint operators: `swagger_gen/models/constraint.go` — 12 enum values
- CI workflow: `.github/workflows/ci.yml` — `integration_test` job

## Delta from Plan

Changes made during implementation that differ from the draft:

- **File split:** `integration_test.go` was split into `integration_server.go`, `integration_client.go`, `integration_benchmark_test.go`, and the test-only `integration_test.go` — each under 260 lines vs the planned 929.
- **Seed data flattened:** Replaced `opGroup` nested loop machinery with a flat `[]entry` literal — every flag spec visible at a glance.
- **No test-shell fallback:** `test.sh` was deleted entirely; the `test-shell` Makefile target and docs were removed. The Go suite fully replaces the shell tests.
- **Stale-binary guard removed:** `startLocalServer` always builds to a temp directory instead of checking for a pre-built `./flagr`.
- **`init()` removed:** Seed data construction moved from `init()` side effect to an explicit `initFlagDefs()` call from `TestMain`.