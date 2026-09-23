# feat: Context Enrichers — Unified Evaluation-Context Enrichment

**Date:** 2026-09-23
**Status:** draft (→ as-built on completion)
**Branch:** `zz/ctx-enrich`
**Supersedes:** PR [#790](https://github.com/openflagr/flagr/pull/790) *Jev / System One calibrated constraints* (open, unreleased) — redesigned before merge into a general concept.

## Summary

Introduce a first-class **context enricher**: a named component that contributes
properties to the evaluation context *before* constraints run. The existing
built-in injections (`@ts*`, `@http_*`) and the Jev / System One integration
become instances of the same abstraction. Constraints stay exactly
`{Property, Operator, Value}` and match enriched properties such as
`@jev_<question>` with the existing operators — the constraints UI is reused
unchanged apart from a property picker.

The core reframe versus PR #790: Jev is **not a new constraint type**. It is a
**flag-scoped enricher** (`namespace: "jev"`, `config.questions`) that feeds the
same evaluation context as the ambient enrichers. Enriched properties are part
of the evaluation context and are returned/persisted (uniform public
visibility).

User guide target: `docs/flagr_context_enrichers.md` (replacing the Jev-only
guide). Built-in behaviour folds into `docs/flagr_injected_context.md`.

## Problem Frame

PR #790 modelled Jev as a constraint attribute: the System One question lived
inline on the `constraint` row (`Constraint.JevJSON`) and was exposed as
`@jev.<name>`. That works for one question matched by one segment, but it is:

- **not reusable** — the same question cannot feed two segments, or two flags;
- **not composable** — a question cannot be declared for its own sake (e.g. to
  log a calibrated confidence with the evaluation) without a matching segment;
- **inconsistent** — `@ts`/`@http` already enrich the context through a
  completely different code path (`InjectBuiltInContext`), with different
  visibility and different debug plumbing.

Introducing a single **context enricher** concept removes that split, makes the
feature scalable to future sources, and keeps the constraint engine a pure
deterministic comparison over an enriched context map.

## Key Technical Decisions

1. **Enricher = namespace populator.** An enricher is `{namespace, config}`; it
   owns a set of property names and contributes `map[property]value` to the
   evaluation context. It is not a context transform and not a single-key
   producer. `kind` and `scope` are derived in code, never stored.
2. **Flag-scoped definitions.** Flag-scoped enrichers live on the flag
   (`Flag.Enrichers`), so they ride snapshots, templates, duplicate-flag, the
   JSON flag spec, and the eval-only export. There is no user-defined global
   registry.
3. **Unified pipeline.** One enrichment step consolidates ambient (built-in,
   from the request/middleware layer) and flag-scoped (from the flag)
   enrichers into the evaluation context. Constraints read the enriched map via
   the unchanged `conditions.Evaluate` path.
4. **Uniform public visibility.** Enriched properties are merged into
   `entityContext` and therefore appear in `EvalResult.EvalContext` and data
   records. Jev writes **only its answers** (`@jev_*`) — never the request
   `state`, which may contain sensitive client context. The debug log may
   include `state` under `enableDebug`.
5. **Namespace identity + prefix linkage.** One enricher per namespace per flag.
   A constraint property resolves to an enricher by its registered property
   set; no explicit FK on the constraint.
6. **Effective catalog on `GET flag`.** The flag read model returns the merged
   effective enrichers (flag-scoped + global) with derived `scope`/`properties`
   for the UI picker. The write model / snapshot / JSON spec / export carry
   **flag-scoped entries only**; writes to global namespaces are ignored.
7. **Exact-namespace registry, flat underscore keys.** Code registry maps
   namespace → `{scope, property set, factory}`. v1 namespaces: `ts`, `http`
   (global), `jev` (flag-scoped). Properties are flat and underscore-joined:
   `@ts`, `@ts_hour`, `@ts_weekday`, `@ts_month`, `@http_<header>`,
   `@jev_<question>`. Each enricher declares explicit property names (so the
   bare `@ts` needs no special handling). The PR's nested `@jev.<name>` is
   dropped (never released).
8. **Declaration = execution.** Every declared enricher runs on every flag
   evaluation, with its full config. No constraint scanning, no lazy pruning.
   This preserves the "declare `@jev_*` purely as logged decision-confidence
   data" use case enabled by decision 4.
9. **Ordering.** Global enrichers first (cheap/synchronous), then flag-scoped
   in declaration order. Each enricher sees the accumulated context with its
   **own properties masked** (no self-feedback; Jev answers are never fed back
   to the model). Enricher properties always overwrite client-provided keys.
10. **Isolated, fail-closed.** A failing enricher contributes no properties,
    records its error in the debug section, and evaluates nothing for the
    referencing constraints (segment falls through). No whole-eval abort, no
    fail-open default. Per-enricher deadline; Jev keeps `FLAGR_INJECTED_CONTEXT_JEV_TIMEOUT`
    with bounded retries inside it. Pipeline runs per `(entity, flag)`.
11. **Warn-only reference validation.** Constraint create/update and
    `flagr-validate`/JSON load accept unknown enriched properties and emit a
    warning; evaluation fails closed. Enricher **config** validation stays
    strict (`JevQuestion.Validate()` moves from constraint to enricher config).
12. **Debug provenance.** A per-eval `enrichers` debug section lists each
    enricher that ran: `{namespace, propertiesAdded, latencyMs, status,
    error}`, plus the vendor payload for flag-scoped enrichers (Jev: `state`,
    `questions`, `answers`, `usage`, `retries`). Hoisted off
    `segmentDebugLog.jev` because enrichers run once per flag, not per segment.
13. **API = sub-resource endpoints.** `POST /flags/{flagID}/enrichers`,
    `PUT /flags/{flagID}/enrichers/{namespace}`,
    `DELETE /flags/{flagID}/enrichers/{namespace}`. `PUT /flags/{id}` stays
    scalar-only.
14. **Naming.** `entity.Enricher`, `Flag.Enrichers`, JSON/API field
    `enrichers`, docs term "context enricher(s)".

## Enricher Model

```go
// pkg/entity/enricher.go
type Enricher struct {
    gorm.Model

    FlagID    uint   `gorm:"index:idx_enricher_flagid"`
    Namespace string `gorm:"type:varchar(64);index:idx_enricher_flagid_namespace"`

    // Kind-specific JSON. For namespace "jev": {"questions": {<name>: JevQuestion}}.
    ConfigJSON string `gorm:"type:text"`
}

type JevQuestion struct { /* moved from constraint_jev.go, unchanged fields */ }
```

- `namespace` is the identity; uniqueness per flag is enforced at create/update
  and at JSON load.
- `ConfigJSON` is validated by the namespace's registered factory before
  persistence (strict).
- `Flag.PrepareEvaluation()` preloads enrichers and resolves the effective
  catalog (flag-scoped + global) once for the hot path.

### Property resolution

```go
// pkg/handler/enricher_registry.go
type EnricherKind interface {
    Scope() Scope                 // Global | FlagScoped
    Properties(cfg any) []string  // exact property names, e.g. @ts, @ts_hour, @jev_plan_tier
    Enrich(ctx EnrichInput) (map[string]any, error)
}
```

Registry (code, exact namespace match):

| Namespace | Scope | Properties | Config (flag) |
|---|---|---|---|
| `ts` | global | `@ts`, `@ts_hour`, `@ts_weekday`, `@ts_month` | none (always on when injection enabled) |
| `http` | global | `@http_<header>` (configured set/prefixes) | none (env: `FLAGR_INJECTED_CONTEXT_HTTP_*`) |
| `jev` | flag-scoped | `@jev_<question>` | `{questions: {<name>: JevQuestion}}` |

Ambient enrichers are registered unconditionally; their behaviour is gated by
existing env flags (`FLAGR_INJECTED_CONTEXT_ENABLED`; the jev enricher is
enabled by a non-empty `FLAGR_INJECTED_CONTEXT_JEV_BASE_URL`).

## Runtime Flow

```
PostEvaluation / GetEvaluation / batch
  └─ for each (entity, flag)
       ├─ base = caller entityContext
       ├─ effective = PrepareEvaluation(flag).EffectiveEnrichers  // flag + globals
       ├─ enriched = base
       │    for enricher in effective (globals first, then flag order):
       │        out, err := enricher.Enrich(input{enriched minus own props, request, identity})
       │        enriched = merge(enriched, out)   // enricher wins
       │        debug.enrichers += provenance
       ├─ evalContext.EntityContext = enriched     // public
       └─ evalSegment loop (unchanged; reads enriched map)
```

- `InjectBuiltInContext` becomes the `ts` + `http` global enrichers; the eager
  mutation in `eval.go` collapses into the single enrichment step.
- `resolveJevForFlag` becomes the `jev` enricher; the private answers map and
  the `evalSegment` merge disappear — constraints read `@jev_*` directly.
- Fail-closed falls out of `conditions.Evaluate` erroring on a missing
  property (existing behaviour).

**As-built note (two-stage invocation).** The enrichment *logic* is
centralised in one registry + pipeline (`EnrichContext`), but it is invoked in
two stages: global enrichers at the request boundary
(`EnrichGlobalContext`, where the `*http.Request` exists) and flag-scoped
enrichers inside `EvalFlagWithContext` (`EnrichFlagContext`, where the flag is
known). This preserves the `EvalFlag` / `EvalFlagWithContext` package-var test
seams, which the handler tests stub heavily; threading the request through those
seams would churn a large test surface for no behavioural gain. Ordering still
holds (globals run first, so `jev` sees `@ts*`/`@http_*`), and both stages share
the same merge rule and fail-closed behaviour.

## Persistence & Plumbing

- `Flag.Enrichers []Enricher` gorm relation; added to
  `PreloadSegmentsVariantsTags` so snapshots, templates, duplicate, eval-only
  export, and `SourceFlagTemplate` carry flag-scoped enrichers automatically.
- JSON flag spec: `Flags[].enrichers` = array of `{namespace, config}`
  (flag-scoped only). The eval-only export dump mirrors it.
- `eval_cache_validate.go`: validate enricher config (strict) and warn on
  dangling `@jev_*` constraint references.
- Eval cache stores the resolved effective catalog per flag, so eval does not
  re-parse config JSON on the hot path.

## API (swagger)

- `enricher` object: `{namespace: string, config: object}`; `config` typed as
  `object` at the array level (per-namespace schema validated in Go). `jev`
  config documented as `{questions: {<name>: JevQuestion}}`.
- `flag` model gains `enrichers: [enricher]` (effective, read-only union).
- New paths under `swagger/`: enrichers sub-resource (POST/PUT/DELETE) or the
  existing `flags.yaml` split convention.
- `createFlagRequest`/`putFlagRequest` unchanged (scalar-only).
- `segmentDebugLog.jev` removed; `evalDebugLog.enrichers` added.
- `make gen` (api_docs + swagger) and commit all three artifacts.

## UI

- Flag detail page: **"Context enrichers"** section (sibling of
  Segments/Variants). Lists effective enrichers grouped by scope; `@ts`/`@http`
  read-only with their properties, `@jev` editable via the copied
  `JevQuestionEditor.vue`.
- `ConstraintAddRow.vue`: property input becomes a combobox seeded from
  `GET flag.enrichers` properties (grouped, free text allowed).
- `api/types.ts`: `Enricher` DTO; `jevQuestion.ts` reused.
- Remove the Jev toggle from the constraint row (`JevToggleSwitch.vue`).

## File Reuse Map (from `zz-jev-test`)

| Copy, adapt | Source |
|---|---|
| System One client + tests | `pkg/handler/jev_client.go`, `jev_client_test.go` |
| `JevQuestion` type/validation + tests | `pkg/entity/constraint_jev.go`, `constraint_jev_test.go` (relocate to `enricher_jev.go`) |
| `FLAGR_INJECTED_CONTEXT_JEV_*` config | `pkg/config/env.go` (hunk) |
| Jev question editor + helper + tests | `browser/flagr-ui/src/components/JevQuestionEditor.vue`, `JevCriteriaEditor.vue`, `helpers/jevQuestion.ts(.test.ts)` |
| Jev docs content | `docs/flagr_jev.md` (rewrite as `flagr_context_enrichers.md`) |
| Debug payload shapes | `jev_eval.go` (adapt to generic provenance) |

**Do not copy:** `Constraint.JevJSON`, `crud_jev.go` name-uniqueness logic,
the constraint-embedded swagger hunks, `segmentDebugLog.jev`.

## Execution Plan

1. **Entity + registry** — `entity/enricher.go`, `enricher_jev.go`, gorm
   migration, `Flag.Enrichers`, `PreloadSegmentsVariantsTags`,
   `PrepareEvaluation` effective catalog; unit tests.
2. **Eval pipeline** — `handler/enricher*.go`, port `ts`/`http` into global
   enrichers, remove `InjectBuiltInContext` call sites, port Jev client,
   per-eval debug section; unit + integration tests.
3. **API** — swagger model + sub-resource paths, mapper (r2e/e2r), CRUD
   handlers, effective-catalog read; `make gen`; handler tests.
4. **UI** — enrichers section, editor wiring, constraint property combobox,
   types/API; `make flagr-ui-check`.
5. **Docs** — context enrichers guide, injected-context rewrite, env table,
   behavioral contracts, llms.txt, sidebar.
6. **Validation** — `make test`, `make flagr-ui-check`, `make test-integration`.
7. Push branch, open PR (`.github/PULL_REQUEST_TEMPLATE.md`).

## Testing

- Go: property resolution/validation per namespace; effective-catalog merge;
  ordering + self-masking; ambient+flag pipeline; fail-closed isolation; public
  visibility (answers in result, state never); debug provenance; CRUD
  sub-resources; snapshot/template/duplicate round-trip; warn-only dangling
  reference; JSON spec load/export.
- Mock System One contract test reused from `jev_client_test.go`.
- UI: combobox catalog rendering, enricher section, editor re-use.
- `require.Eventually` for async notification/cache side effects (no sleeps).

## Risks

- **Latency**: sequential flag-scoped enrichers add to the eval path; only
  `jev` in v1, documented as low-QPS/high-value.
- **Visibility change**: enriched keys now include `@jev_*` in result/data
  records — intended, but model output becomes part of the analysis stream.
- **Public headers**: `@http_*` can carry sensitive headers already today;
  exposure remains opt-in via `FLAGR_INJECTED_CONTEXT_HTTP_*`.
- **Snapshot size**: enricher config is small, but a question library can grow;
  first-segment-rank collision handling is unnecessary (namespace-unique).
- **No answer cache** in v1 (unchanged from PR #790 rationale).

## As-built refinements (post-review)

- **`entity.Enricher` is generic.** The entity layer stores only
  `{namespace, configJSON}` and checks storage invariants (namespace present,
  well-formed JSON). Namespace scope and config schema are validated by the
  handler registry, which owns the namespace implementations. This removes the
  "jev is the only flag-scoped namespace" assumption that previously sat in
  `Enricher.Validate`/`DecodeJevConfig`.
- **All Jev domain and behaviour consolidated in `pkg/handler`**:
  `jev_config.go` (question/config types, validation, `@jev_` property naming),
  `jev_client.go` (System One transport + answer interpretation/confidence
  gate), `enricher_jev.go` (builder + state). `pkg/entity` no longer imports or
  models Jev.
- **One resolved `enricher` type.** The registry is
  `map[string]func(configJSON string) (*enricher, error)` — no separate
  `enricherKind` wrapper, no ignored builder parameter.
- **One env family for enrichers: `FLAGR_INJECTED_CONTEXT_*`.** "Injected context"
  is the env- and config-facing name (`Config.InjectedContextEnabled`,
  `Config.InjectedContextJev*`, ...); the runtime concept in `pkg/handler` is
  `enricher`. The three released tags are unchanged, and the new (unreleased)
  Jev switches join the same family. No rename and no compatibility shim.
- **Env switches respected when listing/evaluating.** A resolved `enricher`
  carries `enabled`; `enabledEnrichers` drives the read model/catalog and the
  eval pipeline, while `effectiveEnrichers` (all namespaces) is used for
  warn-only reference validation so a disabled namespace does not produce false
  "unknown property" warnings.
