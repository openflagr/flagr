# Documentation URL compatibility

Shared links use GitHub Pages + Docsify hash routes: `https://openflagr.github.io/flagr/#/<page>` and `#/<page>?id=<anchor>`.

## Page routes (stable)

| Route | File | Notes |
|-------|------|--------|
| `#/home` | `home.md` | Default landing / get started |
| `#/integration` | `integration.md` | Was the target of `#/USING_FLAGR` |
| `#/CONTRIBUTING` | `CONTRIBUTING.md` | Was the target of `#/DEVELOPER_GUIDE` |
| `#/flagr_behavioral_contracts` | `flagr_behavioral_contracts.md` | Canonical behavioral contracts (invariants) |
| `#/contracts` | `contracts.md` | **Stub** - redirects readers to `flagr_behavioral_contracts.md` |
| `#/flagr_self_host` | `flagr_self_host.md` | Dedicated deploy guide |
| `#/flagr_injected_context` | `flagr_injected_context.md` | Built-in `@ts` / `@http_*` context |

| `#/flagr_*` | `flagr_*.md` | Unchanged filenames |

## Former bookmark routes (no stub pages)

Old Docsify routes used removed filenames. Open the replacement route directly - there is no `USING_FLAGR.md` or `DEVELOPER_GUIDE.md`.

| Old route | Use instead |
|-----------|-------------|
| `#/USING_FLAGR` | `#/integration` - [Integration guide](integration.md) |
| `#/DEVELOPER_GUIDE` | `#/CONTRIBUTING` - [Contributing](CONTRIBUTING.md); client API → `#/integration` |
| `#/contracts` (content) | `#/flagr_behavioral_contracts` - stub remains for the hash; prefer the new route in links |

## Deep anchors - preserved on purpose

| Old / common link | Current target |
|-------------------|----------------|
| `#/flagr_env?id=source-pkgconfigenvgo` | Same id (explicit `:id=source-pkgconfigenvgo`) |
| `#/flagr_env?id=data-record-destinations` | Alias on **Data recorders** (`:id=data-record-destinations`) |
| `#/flagr_env?id=eval-cache-export` | Alias `:id=eval-cache-export` under Database |
| `#/flagr_env?id=guide` | `## Guide` |
| `#/flagr_env?id=database` | `### Database` |
| `#/integration?id=eval-vs-exposure` | Same id on integration (links to [behavioral contracts](flagr_behavioral_contracts.md#eval-vs-exposure)) |
| `#/flagr_overview?id=architecture` | `## Architecture` |
| `#/flagr_overview?id=rollout-and-deterministic-bucketing` | Unchanged heading |
| `#/flagr_overview?id=constraint-property-access` | Unchanged heading |
| `#/flagr_use_cases?id=get-evaluation-browser-friendly` | `## GET evaluation (browser-friendly)` |

## Deep anchors - changed (update external links)

| Old anchor | Use instead |
|------------|-------------|
| `#/flagr_eval_exposure_pipeline?id=…` (env link to `data-record-destinations`) | `#/flagr_env?id=guide` or `#/flagr_behavioral_contracts` |
| Eval-only narrative only on `integration` | `#/flagr_behavioral_contracts?id=eval-only` |
| Recording gates copy on many pages | `#/flagr_behavioral_contracts?id=recording-gates` |
| `#/contracts?id=…` | `#/flagr_behavioral_contracts?id=…` (same anchors; stub has no section ids) |

## Canonical anchors (`flagr_behavioral_contracts.md`)

| Anchor | Topic |
|--------|--------|
| `#eval-vs-exposure` | Assignment vs impression; when eval volume is enough vs exposure |
| `#segment-evaluation` | Rank order, constraint match stops eval, no rollout fallthrough |
| `#recording-gates` | Recorder + `dataRecordsEnabled` |
| `#blank-vs-stream` | Empty `variantKey` vs whether a stream row is written |
| `#eval-only` | `json_file` / `json_http` (usual eval-edge path) |
| `#evalcache-freshness` | Reload lag; links blank vs stream |

When you change cross-cutting behavior, update **`flagr_behavioral_contracts.md`** and extend this table if you add or rename ids. After renaming a page or `:id=` anchor, grep the repo for the old hash (`#/old-page`, `?id=old-anchor`) and update this file.
