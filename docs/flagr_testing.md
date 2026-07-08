# Testing

Flagr's test suite is layered so you can start small and widen only when you need to. The fastest feedback loop is a plain unit test against a pure function; the widest is the same integration suite run across six databases in Docker Compose. Every layer is reachable through **`make`** from the repo root (`make help` → **Test**), so you never need to remember the underlying toolchain to get feedback.

CI gates and PR conventions live in [Contributing](CONTRIBUTING.md) and [AGENTS.md](https://github.com/openflagr/flagr/blob/main/AGENTS.md); this guide is about running and writing the tests themselves.

## Unit tests

The first stop is always the unit suite. One command runs the linter, validates the OpenAPI spec, and exercises every package under `pkg/`:

```bash
make test           # golangci-lint + swagger validate + go test ./pkg/...
```

Unit tests are where most feature work begins and ends. They compile fast, run in seconds, and — because `go test ./pkg/...` already parallelizes across packages — they make good use of CI cores without any extra effort from you. When you do need to reach for parallelism *within* a package, see [Writing parallel-safe tests](#writing-parallel-safe-tests) below.

## E2E tests (UI)

The Flagr UI lives in `browser/flagr-ui` and is **TypeScript**. End-to-end coverage builds the server, typechecks and lints the UI, then drives it with Playwright:

```bash
make test-e2e       # build server + UI lint/typecheck + Playwright
```

The UI's architecture and the patterns it follows are documented in [`docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md`](plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md) (As-built). Treat that plan as the source of truth for how the frontend is structured.

## Integration tests (API, multi-DB)

Unit tests prove the pieces; integration tests prove the pieces talk to each other. Flagr's integration suite is a single `go:build integration` package in `integration_tests/` that hits a *live* server end to end — health checks, CRUD, eval, batch eval, export, exposures (when routed), and Datar (when the recorder is on).

The quickest way to run it is against a local binary backed by SQLite in memory, which the make target starts for you:

```bash
make test-integration
```

That same suite scales up to a full matrix. Docker Compose spins up six instances — SQLite, MySQL, PostgreSQL, and the rest — and runs the tests against each one:

```bash
make test-integration-compose
```

In CI, the canonical invocation is **`make ci-integration`**, which runs the suite *and* its benchmarks on the Compose matrix. To run just the benchmarks locally:

```bash
make bench-integration
```

The two local targets differ only in topology. **`make test-integration`** starts one binary with SQLite; **`make test-integration-compose`** runs the same suite against six URLs — the current image on ports **18001–18005** and legacy **`checkr/flagr:1.1.12`** on port **18006** for backward-compatibility skips. Newer routes (duplicate flag, snapshot `max_id`) **fail** on current images when the route is missing, while the legacy instance on **18006** skips cleanly when the router reports the path was never registered. For the full breakdown, see [`integration_tests/README.md`](https://github.com/openflagr/flagr/blob/main/integration_tests/README.md).

## Writing parallel-safe tests

CI already parallelizes *across* packages when it runs `go test ./pkg/...`. The remaining speedup comes from parallelizing *within* a package: tests that call `t.Parallel()` hand themselves back to the runner, which schedules them concurrently across the available cores on multi-core CI runners.

**Rule of thumb:** every new test function should start with `t.Parallel()`
unless it has a specific reason not to.

### How `t.Parallel()` works

```go
// go test ./pkg/foo/...
//
// Runner starts:
//   TestA  (t.Parallel) → paused, added to parallel set
//   TestB  (serial)     → runs immediately, blocks runner
//   TestC  (t.Parallel) → paused, added to parallel set
//
//   [runner finishes all serial tests]
//
//   Resumes ALL parallel tests concurrently:
//     TestA ────────────┐
//     TestC ─────────┐  │
//                    ▼  ▼
//               GOMAXPROCS goroutines
```

The runner's behavior is worth internalizing. It scans the whole package first, collecting every `t.Parallel()` test into a set, then resumes them concurrently once the serial tests have finished. The practical consequence is that the *order* tests appear in the file does not determine the order they run in — parallel tests all start together after the serial ones are done.

### Safe — add `t.Parallel()`

The safe cases are the easy ones. A pure function with no shared state is the clearest example:

```go
func TestMyPureFunction(t *testing.T) {
    t.Parallel()
    result := MyPureFunction("input")
    assert.Equal(t, "expected", result)
}
```

So is any test that creates its own isolated resources — an in-memory database, a local HTTP server, a temp directory. Each test owns its fixture, so there is nothing to race over:

```go
func TestWithIsolatedDB(t *testing.T) {
    t.Parallel()
    db := entity.NewTestDB() // fresh :memory: SQLite per test
    // ... use db freely
}
```

### Unsafe — skip `t.Parallel()`

Things get dangerous when a test mutates **global state**. Global config, package-level singletons, process-wide environment variables — any of these can be read or written by other tests running at the same time, and the resulting races are notoriously hard to debug. The table below collects the patterns to watch for:

| Pattern | Why it's unsafe | Example |
|---------|----------------|---------|
| Direct `config.Config.X = ...` | Races with other tests reading/writing the same field | `config.Config.RecorderEnabled = true` |
| `gostub.Stub(&config.Config.X, val)` | Same as above, via reflection | `gostub.Stub(&config.Config.RecorderType, ...)` |
| Package-level singletons | Shared mutable state | `singletonDataRecorderOnce = sync.Once{}` |
| `os.Setenv` / `t.Setenv` | Process-wide env mutation | OK inside `t.Run` subtest with `t.Setenv` (auto-restores) |

### The safe subtest pattern

Global setup and parallel execution are not mutually exclusive — you just have to split them across two scopes. Let the parent test do the global mutation *serially*, then parallelize the independent subtests it spawns. The parent serializes setup; the subtests run concurrently once setup is done:

```go
func TestWithGlobalSetup(t *testing.T) {
    // NO t.Parallel() here — this test mutates globals
    config.Config.FeatureEnabled = true
    defer func() { config.Config.FeatureEnabled = false }()

    t.Run("case A", func(t *testing.T) {
        t.Parallel() // safe — parent already finished setup
        // reads config, doesn't write
    })
    t.Run("case B", func(t *testing.T) {
        t.Parallel()
        // ...
    })
}
```

The flow is the same as at the package level, just nested. The parent runs serially, collects its parallel subtests, and only resumes them after setup is complete — so each subtest sees a stable global state:

```
TestWithGlobalSetup starts (serial)
├── sets config.Config.FeatureEnabled = true
├── t.Run("case A")  → paused (parallel)
├── t.Run("case B")  → paused (parallel)
│
│   [all subtests collected]
│
│   case A ─────┐
│   case B ───┐ │
│             ▼ ▼
│        run concurrently
│
├── defer restores config.Config
└── done
```

### Decision tree for new tests

When you're unsure, walk the decision tree. It encodes the reasoning above into a quick check you can run before writing the first line of a test:

```
New test function
│
├── Does it write to globals (config.Config, singletons, env vars)?
│   ├── YES → NO t.Parallel()
│   └── NO ↓
│
├── Does it read globals that other tests might write?
│   ├── YES → NO t.Parallel() (or isolate with subtest pattern)
│   └── NO ↓
│
├── Does it use shared resources (same DB, same port)?
│   ├── YES → NO t.Parallel()
│   └── NO ↓
│
└── ✅ Add t.Parallel()
```

### Checking for races locally

The race detector is the final safety net. Run it before you push, and repeat the suite a few times to shake out ordering-dependent flakes:

```bash
go test -race ./pkg/...     # requires CGO_ENABLED=1
go test -count=5 ./pkg/...  # repeat to catch flaky ordering
```

## CI performance

Fast CI is a feature. Two caching strategies keep the integration pipeline from re-doing work it has already done, and both are worth understanding whether you're debugging a slow pipeline or just trying to land a PR quickly.

### Cache warming

The CI workflow caches Go modules, npm dependencies, and Docker layers between runs. The **first push** to a new branch populates the cache; subsequent pushes hit the cache and skip dependency downloads. To warm the cache before your PR runs its checks, simply push an empty commit or amend-and-force-push — the first run saves the cache, and the second run benefits from it.

### Integration test Docker caching

The integration test Docker image (`integration_tests/Dockerfile-Integration-Test`) uses two caching strategies:

1. **Dependency layer** — `go.mod`/`go.sum` are copied and `go mod download`
   runs before source code. Changing application code doesn't re-download
   dependencies.

2. **BuildKit cache mounts** — `--mount=type=cache,target=/root/.cache/go-build`
   and `--mount=type=cache,target=/go/pkg/mod` persist Go's build and module
   caches across builds. In CI, `docker/setup-buildx-action` + GHA cache
   backend persists these layers between runs.

Locally, Docker BuildKit (default on modern Docker) activates these
automatically — no extra setup needed.