---
title: Get started
description: Flagr is an open-source Go service for feature flags, A/B tests, and dynamic configuration. Self-hosted evaluation API with sticky variants.
---

# Get started

Flagr is an open-source **Go** service for feature flags, A/B tests, and dynamic configuration.

Your app calls **`POST /api/v1/evaluation`** with an `entityID` and optional `entityContext`. Flagr returns a `variantKey` and optional `variantAttachment` JSON. One round trip. Sticky for a stable entity. No per-user state store.

You can run it against SQLite (local demo), MySQL, or Postgres, or as an eval-only sidecar fed from a JSON file in Git. Same flag can be a kill switch today, an experiment tomorrow, and a runtime config knob the day after, without redeploying the app.

Hard rules (eval vs exposure, segment stop, blank vs stream, recording gates, cache lag): [Behavioral contracts](flagr_behavioral_contracts.md). HTTP copy-paste: [Integration guide](integration.md). Just want something running? Use the demo below.

## Quick demo

```bash
docker pull ghcr.io/openflagr/flagr
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr

open http://localhost:18000
```

No install? Hit the hosted demo at [try-flagr.onrender.com](https://try-flagr.onrender.com) (cold start possible):

```bash
curl -sS -X POST https://try-flagr.onrender.com/api/v1/evaluation \
  -H 'content-type: application/json' \
  -d '{
    "entityID": "127",
    "entityType": "user",
    "entityContext": { "state": "NY" },
    "flagID": 1,
    "enableDebug": true
  }'
```

With `enableDebug: true`, the response includes a segment walk so you can see how a constraint like `state == "NY"` becomes a variant.

## What Flagr does

One evaluation primitive covers several jobs. Use the map below as a routing table, not a marketing feature list.

| Capability | Where to read more |
|------------|-------------------|
| Feature flags, rollouts, kill switches | [Overview](flagr_overview.md), [Use cases](flagr_use_cases.md) |
| Browser-friendly eval (`GET ?json=`) | [Use cases: GET evaluation](flagr_use_cases.md#get-evaluation-browser-friendly) |
| Time / header targeting (`@ts`, `@http_*`) | [Built-in context injection](flagr_injected_context.md) |
| Model-backed targeting (Jev / System One) | [Jev constraints](flagr_jev.md) |
| A/B tests + trustworthy denominators | [Exposure logging](flagr_exposure.md), [Data recorders](flagr_eval_exposure_pipeline.md) |
| Runtime config on variants | [Use cases: dynamic configuration](flagr_use_cases.md#dynamic-configuration) |
| GitOps / eval-only JSON | [JSON flag source](flagr_json_flag_spec.md) |
| Deploy, DB, auth, recorders | [Self-hosting](flagr_self_host.md), [Environment variables](flagr_env.md) |

To clone an existing flag (segments, variants, tags), use `POST /api/v1/flags/{id}/duplicate` or **Duplicate Flag** in the UI ([#724](https://github.com/openflagr/flagr/issues/724)).

## Deploy

The demo above is local SQLite. Production (MySQL/Postgres, Compose, Kubernetes, TLS) is in **[Self-hosting](flagr_self_host.md)**. Every env knob lives in [Environment variables](flagr_env.md#source-pkgconfigenvgo). The struct in `pkg/config/env.go` is the source of truth.

## Develop Flagr

Same `make` targets on Linux, macOS, and Windows. Default DB is SQLite (`flagr.sqlite` in the repo root), so a clone is enough to run the stack.

**Prerequisites:** Go 1.26+ (`go.mod`), Node 20+ for the UI, GNU Make. Then:

```bash
git clone https://github.com/openflagr/flagr.git
cd flagr
make deps            # swagger + golangci-lint into $(go env GOPATH)/bin
make build
make start           # API :18000 + UI dev :8080
make test
```

`make deps` tools must be on `PATH` (`$(go env GOPATH)/bin`). Full command catalog: `make help`. Tests: [Testing](flagr_testing.md). How to open a PR: [Contributing](CONTRIBUTING.md). Code layout and CI: [AGENTS.md](https://github.com/openflagr/flagr/blob/main/AGENTS.md). Docs site: `make serve-docs` → http://127.0.0.1:8081/flagr/.

### Windows {#develop-windows}

Native Windows uses the same Makefile through Git Bash (`sh.exe`). Install:

1. **[Git for Windows](https://git-scm.com/download/win)** — includes Git Bash.
2. **GNU Make 4.x**, **Go 1.26+**, **Node 20+** (winget):

```powershell
winget install --id ezwinports.make -e
winget install --id GoLang.Go -e
winget install --id OpenJS.NodeJS.LTS -e
```

3. Put **`C:\Program Files\Git\bin`** on your User `PATH` (ahead of `WindowsApps`, so `sh` / `bash` are Git’s). The Go installer usually adds `%USERPROFILE%\go\bin` as well; keep it so `make deps` tools resolve.
4. Open a **new** terminal and check:

```powershell
go version          # go1.26 or newer
node -v             # v20 or newer
make --version      # GNU Make 4.x
where.exe sh        # ...\Git\bin\sh.exe
```

Then the same `make deps`, `make build`, `make start`, `make test` as above. The server binary is `.\flagr.exe`. Drive the UI through `make` (`make run-ui`, `make flagr-ui-check`); PowerShell’s `npm` shim can be blocked by execution policy (`npm.cmd` works if you call npm directly).

Docker Desktop is only needed for `make test-integration-compose`. `make build-docs` uses `python` (Windows) / `python3` (Unix).
