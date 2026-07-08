# Documentation URL compatibility

Shared links use GitHub Pages + Docsify hash routes: `https://openflagr.github.io/flagr/#/<page>` and `#/<page>?id=<anchor>`.

## Page routes (stable)

| Route | File | Notes |
|-------|------|--------|
| `#/home` | `home.md` | Default landing / get started |
| `#/integration` | `integration.md` | Was the target of `#/USING_FLAGR` |
| `#/CONTRIBUTING` | `CONTRIBUTING.md` | Was the target of `#/DEVELOPER_GUIDE` |
| `#/contracts` | `contracts.md` | **New** — no old URL |
| `#/flagr_self_host` | `flagr_self_host.md` | **New** — dedicated deploy guide |
| `#/flagr_injected_context` | `flagr_injected_context.md` | Built-in `@ts` / `@http_*` context |

| `#/flagr_*` | `flagr_*.md` | Unchanged filenames |

## Former bookmark routes

| Old route | Use instead |
|-----------|-------------|
| `#/USING_FLAGR` | [integration.md](integration.md) (one-line redirect page kept) |
| `#/DEVELOPER_GUIDE` | [CONTRIBUTING.md](CONTRIBUTING.md) (repo work) or [integration.md](integration.md) (client API) |

## Deep anchors — preserved on purpose

| Old / common link | Current target |
|-------------------|----------------|
| `#/flagr_env?id=source-pkgconfigenvgo` | Same id (explicit `:id=source-pkgconfigenvgo`) |
| `#/flagr_env?id=data-record-destinations` | Alias on **Data recorders** (`:id=data-record-destinations`) |
| `#/flagr_env?id=eval-cache-export` | Alias `:id=eval-cache-export` under Database |
| `#/flagr_env?id=guide` | `## Guide` |
| `#/flagr_env?id=database` | `### Database` |
| `#/integration?id=eval-vs-exposure` | Same id on integration (links to [contracts](contracts.md#eval-vs-exposure)) |
| `#/flagr_overview?id=architecture` | `## Architecture` |
| `#/flagr_overview?id=rollout-and-deterministic-bucketing` | Unchanged heading |
| `#/flagr_overview?id=constraint-property-access` | Unchanged heading |
| `#/flagr_use_cases?id=get-evaluation-browser-friendly` | `## GET evaluation (browser-friendly)` |

## Deep anchors — changed (update external links)

| Old anchor (pre–contracts refactor) | Use instead |
|-------------------------------------|-------------|
| `#/flagr_eval_exposure_pipeline?id=…` (env link to `data-record-destinations`) | `#/flagr_env?id=guide` or `#/contracts` |
| Eval-only narrative only on `integration` | `#/contracts?id=eval-only` |
| Recording gates copy on many pages | `#/contracts?id=recording-gates` |

## New canonical anchors (`contracts.md`)

| Anchor | Topic |
|--------|--------|
| `#eval-vs-exposure` | Assignment vs impression |
| `#recording-gates` | Recorder + `dataRecordsEnabled` |
| `#eval-only` | `json_file` / `json_http` |
| `#evalcache-freshness` | Reload lag, blank `variantKey` |

When you change cross-cutting behavior, update **`contracts.md`** and extend this table if you add or rename ids.