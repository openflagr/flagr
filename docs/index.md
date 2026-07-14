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
| A/B tests + trustworthy denominators | [Exposure logging](flagr_exposure.md), [Data recorders](flagr_eval_exposure_pipeline.md) |
| Runtime config on variants | [Use cases: dynamic configuration](flagr_use_cases.md#dynamic-configuration) |
| GitOps / eval-only JSON | [JSON flag source](flagr_json_flag_spec.md) |
| Deploy, DB, auth, recorders | [Self-hosting](flagr_self_host.md), [Environment variables](flagr_env.md) |

To clone an existing flag (segments, variants, tags), use `POST /api/v1/flags/{id}/duplicate` or **Duplicate Flag** in the UI ([#724](https://github.com/openflagr/flagr/issues/724)).

## Deploy

The demo above is local SQLite. Production (MySQL/Postgres, Compose, Kubernetes, TLS) is in **[Self-hosting](flagr_self_host.md)**. Every env knob lives in [Environment variables](flagr_env.md#source-pkgconfigenvgo). The struct in `pkg/config/env.go` is the source of truth.

## Develop Flagr

Clone the repo, then from the root:

```bash
make build
make start   # backend :18000 + UI dev :8080
make test
```

Contributor layout, OpenAPI regen, and test conventions: [Contributing](CONTRIBUTING.md). Docs: `make serve-docs` → http://127.0.0.1:8081/flagr/ ; production: `make build-docs` → `docs/.vitepress/dist`.
