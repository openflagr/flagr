# Get started

Flagr is an open-source Go service for **feature flags**, **A/B tests**, and **dynamic configuration**. The idea is simple: your application calls **`POST /api/v1/evaluation`** with an `entityID` and, optionally, some `entityContext`. Flagr decides which variant applies — and hands back a `variantKey` plus an optional `variantAttachment` JSON payload. That's it. One round trip, deterministic answer, no per-user state.

You can run it in front of a database, or as a stateless eval-only sidecar fed from a JSON file in your repo. You can target a rollout to a single state, split a checkout flow fifty-fifty, or serve a different timeout value per customer tier — all without redeploying your app. The same flag can be a kill switch today, an experiment tomorrow, and a piece of runtime config the day after.

If you want the rules of the road before you start, the canonical reference is [Behavioral contracts](contracts.md). For hands-on HTTP examples, see the [Integration guide](integration.md). And if you'd rather just build something, the quick demo below takes about thirty seconds.

## Quick demo

The fastest way to see Flagr in action is to pull the image and point your browser at it:

```bash
docker pull ghcr.io/openflagr/flagr
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr

open http://localhost:18000
```

Prefer not to install anything? There's a hosted demo at [try-flagr.onrender.com](https://try-flagr.onrender.com) — you can evaluate a flag against it right now:

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

The debug log in the response walks you through the segment evaluation, so you can watch how a `state == "NY"` constraint turns into a variant decision.

## What Flagr does

Flagr wears a few hats — feature-flag engine, experiment platform, runtime configuration store — but they all share one mechanism: evaluate an entity, pick a variant, attach some JSON. The table below is less a feature list and more a map: each row is a thing you might want to do, with a link to the page that explains it properly.

| Capability | Where to read more |
|------------|-------------------|
| Feature flags, rollouts, kill switches | [Overview](flagr_overview.md), [Use cases](flagr_use_cases.md) |
| Browser-friendly eval (`GET ?json=`) | [Use cases — GET evaluation](flagr_use_cases.md#get-evaluation-browser-friendly) |
| Time/header targeting (`@ts`, `@http_*`) | [Built-in context injection](flagr_injected_context.md) |
| A/B tests + trustworthy denominators | [Exposure logging](flagr_exposure.md), [Data recorders](flagr_eval_exposure_pipeline.md) |
| Runtime config on variants | [Use cases — dynamic configuration](flagr_use_cases.md) |
| GitOps / eval-only JSON | [JSON flag source](flagr_json_flag_spec.md) |
| Deploy, DB, auth, recorders | [Self-hosting](flagr_self_host.md), [Environment variables](flagr_env.md) |

One small convenience worth knowing about: you can duplicate an existing flag with `POST /flags/{id}/duplicate`, or use the **Duplicate Flag** button in the UI ([#724](https://github.com/openflagr/flagr/issues/724)) — handy when you want to stage a new experiment from a known-good baseline.

## Deploy

The [quick demo](#quick-demo) above is everything you need to kick the tires locally. When you're ready for production — MySQL or Postgres, Docker Compose, Kubernetes, TLS termination — the walk-through lives in **[Self-hosting](flagr_self_host.md)**. Every tunable, from database drivers to recorder types, is documented in [Environment variables](flagr_env.md#source-pkgconfigenvgo).

## Develop Flagr

Flagr is plain Go: clone the repo, run `make build`, then `make start` to bring it up and `make test` to run the suite. The full contributor guide — repo layout, testing conventions, how to add a feature — is in [Contributing](CONTRIBUTING.md). If you're working on these docs, `make serve-docs` previews them locally.