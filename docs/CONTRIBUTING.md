# Contributing

Welcome. This guide walks you through the full contributor journey: clone the repo, build it, find your way around the code, run the tests, regenerate Swagger, and touch the docs site. The commands and conventions below are the ones CI enforces, so following them from your first commit keeps PRs green and reviews short.

Before you open a PR, run the checks that match what you changed. The canonical **pre-commit / CI** matrix and Makefile mapping live in the repo root **[AGENTS.md](https://github.com/openflagr/flagr/blob/main/AGENTS.md)** (same rules CI enforces). Also see [PULL_REQUEST_TEMPLATE.md](https://github.com/openflagr/flagr/blob/main/PULL_REQUEST_TEMPLATE.md).

```bash
make help    # full command catalog
```

## Clone, build, run

**Prerequisites:** Go 1.24+ (`go.mod`), Node 20+ for UI, Make.

Start from a fresh clone and let the Makefile drive the build. `make build` produces the `./flagr` binary; `make start` brings up both the API on `:18000` and the UI dev server on `:8080`. After editing Go code, `make rebuild-run` rebuilds and restarts the server in one step, and `make stop-ui` frees the ports again via `lsof`.

```bash
git clone https://github.com/openflagr/flagr.git
cd flagr
make build          # ./flagr
make start          # API :18000 + UI dev :8080
make rebuild-run    # after Go changes
make stop-ui        # free :18000 / :8080 via lsof
```

The default database is SQLite (`flagr.sqlite`), so there's nothing extra to provision for local development. The UI dev server proxies `/api/v1` to the Go server on `:18000`, so frontend and backend stay in sync without a separate API URL.

## Where the code lives

Once the server is running, the next step is knowing where to make your change. The table below maps each area of interest to the package that owns it. Evaluation and its cache refresh live in the eval handlers; exposures, CRUD, and the data recorders each have their own files; the entity package holds the domain models; and the UI sources sit under `browser/flagr-ui/src/`.

| Path | Responsibility |
|------|----------------|
| `pkg/handler/eval.go`, `eval_cache.go` | Evaluation, cache refresh |
| `pkg/handler/exposure.go` | `POST /exposures` |
| `pkg/handler/crud*.go` | CRUD, snapshots, duplicate flag |
| `pkg/handler/data_recorder*.go` | Kafka, Kinesis, Pub/Sub, Datar |
| `pkg/entity/` | Domain models |
| `pkg/config/env.go` | Environment variables (documented in [flagr_env.md](flagr_env.md)) |
| `browser/flagr-ui/src/` | UI — `api/crud.ts`, `api/eval.ts`, `pages/flagPage.ts` |
| `swagger/` → `make swagger` → `swagger_gen/` | OpenAPI; do not hand-edit `swagger_gen/` |
| `cmd/flagr-server/` | Server entry |
| `cmd/flagr-validate/` | JSON flag file validator for CI |

One rule overrides everything else here: when docs and code disagree, **code wins**. If you spot a doc that no longer matches the implementation, trust the code and fix the doc.

## HTTP handlers (`/api/v1`)

Most API surface lives under `/api/v1`, and each area maps to a handler package. The evaluation endpoints cover both single and batch evaluation; exposures have their own handler; flag CRUD and duplication share a set of files; the eval-cache export and datar summaries round out the list.

| Area | Package |
|------|---------|
| `POST /evaluation`, batch | `pkg/handler/eval.go` |
| `POST /exposures` | `pkg/handler/exposure.go` |
| Flags CRUD, duplicate | `pkg/handler/crud.go`, `crud_duplicate.go` |
| `GET /export/eval_cache/json` | `pkg/handler/export.go`, `eval_cache_fetcher.go` |
| Datar summaries | datar handlers in `pkg/handler` |

Before changing handler behavior, skim the contract tests in `pkg/handler/*_test.go`. They pin down the invariants you need to preserve — eval-cache short-circuiting, snapshots taken on mutate, and exposure recording — and they're the fastest way to understand what a handler promises.

## Testing

Flagr runs a layered test suite: Go unit tests, Playwright browser E2E, and API integration tests. The [testing guide](flagr_testing.md) covers each in depth — unit via `make test`, Playwright E2E via `make test-e2e`, API integration via `make test-integration` and `make test-integration-compose`, plus the `t.Parallel()` rules every suite follows.

```bash
make test
make test-e2e
make test-integration
go test -race ./pkg/...    # when debugging flakes
```

Reach for the race detector when a test flakes intermittently; it catches the data races that sequential runs hide.

## OpenAPI / Swagger workflow

The API contract starts in `swagger/index.yaml` and the files it references under `swagger/`. The flow is always the same three steps: edit the Swagger source, regenerate the docs bundle, then regenerate the Go server bindings.

1. Edit `swagger/index.yaml` and referenced files under `swagger/`.
2. `make api_docs` → `docs/api_docs/bundle.yaml`
3. `make swagger` → `swagger_gen/`

If you'd rather not remember the order, `make gen` runs all three in sequence. In CI, `make ci-swagger` fails if the generated output is dirty, so always commit regenerated files alongside your Swagger edits.

## Documentation site

The docs site is a Docsify app whose sources live right here in `docs/`, with `home.md` as the homepage. Preview it locally with `make serve-docs`; on push to `main` it publishes to GitHub Pages via the [`pages.yml`](https://github.com/openflagr/flagr/blob/main/.github/workflows/pages.yml) workflow.

Several pages serve specific audiences and are worth knowing as a contributor: the [integration guide](integration.md) is the hub for integrators, the [behavioral contracts](contracts.md) hold the canonical descriptions of cross-cutting behavior, and this page is the repo guide. Design and as-built notes live in `docs/plans/` — they aren't linked in the sidebar, but they're valuable internal history.

**Docs conventions**

- User-facing filenames: `flagr_*.md`, `integration.md`, `contracts.md`, `home.md`. Link text: **sentence case** (e.g. "Exposure logging", "Data recorders & A/B analysis").
- Cross-cutting behavior (eval vs exposure, recording gates, eval-only, EvalCache lag): edit **[contracts.md](contracts.md)** first; other pages link there instead of copying paragraphs.
- **Deploy / topology:** edit **[flagr_self_host.md](flagr_self_host.md)**; **`flagr_env.md`** = embedded `env.go` + variable guide (no duplicate Compose/mysql runbooks).
- **`flagr_env.md`:** embedded `pkg/config/env.go` first, then the short guide; embed tracks **`main`** on GitHub.
- Off-site sidebar links use a subtle **↗** via CSS (`a[href^="http"]`); client SDKs are listed in [integration.md](integration.md) and README, not in `_sidebar.md`.
- URL compatibility when renaming pages or anchors: [link-compatibility.md](link-compatibility.md) (not in sidebar).

Whenever you change the API docs, refresh `docs/api_docs/bundle.yaml` so the hosted [API reference](https://openflagr.github.io/flagr/api_docs) stays current.

## UI architecture

The flagr-ui TypeScript migration laid down the patterns the UI follows today — component layout, API client structure, and the as-built decisions behind them. Read it before a non-trivial frontend change: [`docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md`](plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md).

## Performance

Load numbers and the `benchmark/` tooling live with the README rather than the docs site: [README — Performance](https://github.com/openflagr/flagr/blob/main/README.md#performance). Treat that as the source of truth for throughput and latency claims.