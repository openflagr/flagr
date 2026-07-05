# Testing

Three test layers — all via **`make`** from the repo root (`make help` → **Test**).

## Unit tests

```bash
make test           # golangci-lint + swagger validate + go test ./pkg/...
```

## E2E tests (UI)

```bash
make test-e2e       # build server + UI lint/typecheck + Playwright
```

Flagr UI is **TypeScript** (`browser/flagr-ui`); architecture and patterns: [`docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md`](plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md) (As-built).

## Integration tests (API, multi-DB)

**Local** — SQLite `:memory:`, auto-started server:

```bash
make test-integration
```

**Docker Compose** — same suite against six instances (SQLite, MySQL, PostgreSQL, …):

```bash
make test-integration-compose
```

CI runs **`make ci-integration`** (tests + benchmarks on Compose). Local benchmarks:

```bash
make bench-integration
```

## Writing parallel-safe tests

CI runs `go test ./pkg/...` which already parallelizes across packages. To go
faster, tests *within* a package should call `t.Parallel()` so the test runner
can execute them concurrently on multi-core CI runners.

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

The runner scans the whole package first, collects all `t.Parallel()` tests,
then resumes them concurrently once serial tests finish.

### Safe — add `t.Parallel()`

Pure functions with no shared state:

```go
func TestMyPureFunction(t *testing.T) {
    t.Parallel()
    result := MyPureFunction("input")
    assert.Equal(t, "expected", result)
}
```

Tests that create their own isolated resources (in-memory DB, local HTTP
server, temp directory):

```go
func TestWithIsolatedDB(t *testing.T) {
    t.Parallel()
    db := entity.NewTestDB() // fresh :memory: SQLite per test
    // ... use db freely
}
```

### Unsafe — skip `t.Parallel()`

Tests that mutate **global state** must not run in parallel:

| Pattern | Why it's unsafe | Example |
|---------|----------------|---------|
| Direct `config.Config.X = ...` | Races with other tests reading/writing the same field | `config.Config.RecorderEnabled = true` |
| `gostub.Stub(&config.Config.X, val)` | Same as above, via reflection | `gostub.Stub(&config.Config.RecorderType, ...)` |
| Package-level singletons | Shared mutable state | `singletonDataRecorderOnce = sync.Once{}` |
| `os.Setenv` / `t.Setenv` | Process-wide env mutation | OK inside `t.Run` subtest with `t.Setenv` (auto-restores) |

### The safe subtest pattern

When a parent test needs global setup but subtests are independent, parallelize
the **subtests** — the parent serializes setup:

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

```bash
go test -race ./pkg/...     # requires CGO_ENABLED=1
go test -count=5 ./pkg/...  # repeat to catch flaky ordering
```

## CI performance

### Cache warming

The CI workflow caches Go modules, npm dependencies, and Docker layers between
runs. The **first push** to a new branch populates the cache; subsequent pushes
hit the cache and skip dependency downloads. To warm the cache before your PR
runs its checks, simply push an empty commit or amend-and-force-push — the
first run saves the cache, and the second run benefits from it.

### Integration test Docker caching

The integration test Docker image (`integration_tests/Dockerfile-Integration-Test`)
uses two caching strategies:

1. **Dependency layer** — `go.mod`/`go.sum` are copied and `go mod download`
   runs before source code. Changing application code doesn't re-download
   dependencies.

2. **BuildKit cache mounts** — `--mount=type=cache,target=/root/.cache/go-build`
   and `--mount=type=cache,target=/go/pkg/mod` persist Go's build and module
   caches across builds. In CI, `docker/setup-buildx-action` + GHA cache
   backend persists these layers between runs.

Locally, Docker BuildKit (default on modern Docker) activates these
automatically — no extra setup needed.
