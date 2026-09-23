# Context enrichers

A **context enricher** contributes properties to the evaluation context *before* constraints run. It is the one mechanism behind every server-injected property in Flagr: time (`@ts*`), request headers (`@http_*`), and model-derived answers (`@jev_*`).

Constraints stay exactly what they always were — `{property, operator, value}` — and match enriched properties like any client-provided field. There is no "enricher constraint type".

## Namespaces

An enricher is identified by a **namespace**. The namespace is hardcoded in Flagr (adding an integration is a code change), and each namespace owns a set of **flat, underscore-joined** property names:

| Namespace | Scope | Properties | Enabled by |
|-----------|-------|------------|------------|
| `ts` | global (built-in) | `@ts`, `@ts_hour`, `@ts_weekday`, `@ts_month` | `FLAGR_INJECTED_CONTEXT_ENABLED` |
| `http` | global (built-in) | `@http_<header>` | `FLAGR_INJECTED_CONTEXT_ENABLED` + header config |
| `jev` | flag-scoped | `@jev_<question>` | non-empty `FLAGR_INJECTED_CONTEXT_JEV_BASE_URL` |

**Global** enrichers are supplied by the server and apply to every flag. **Flag-scoped** enrichers are declared on a flag and travel with it through snapshots, templates, duplicate, the JSON flag source, and the eval-cache export.

## How evaluation uses them

```
request entityContext
      ↓  global enrichers (ts, http)
      ↓  flag-scoped enrichers (jev), in declaration order
enriched entityContext  ← constraints evaluate against this
      ↓
segment match → distribution → rollout
```

- **Order**: global first, then flag-scoped. An enricher sees the accumulated context, so a `jev` question can use `@ts_hour` or `@http_*`.
- **Own properties are masked**: an enricher never sees its own prior output, and `@jev_*` answers are never fed back to the model.
- **Enriched properties overwrite client keys** of the same name (a client cannot spoof `@ts`).
- **Declaration = execution**: every declared enricher runs on every evaluation of that flag. There is no reference-scanning or lazy pruning, so an enricher can also be declared purely to decorate the logged evaluation context.
- **Fail-closed**: a disabled, timed-out, or erroring enricher contributes no properties. A constraint referencing a missing property errors and the segment falls through. There is no fail-open default.
- **Public context**: enriched properties are part of the evaluation context and appear in `EvalResult.evalContext` and data records. A `jev` enricher writes **only its answers**, never the request state it sent to the model.

## Built-in enrichers

`ts` and `http` are documented in full on [Built-in context injection](flagr_injected_context.md). Enable them with `FLAGR_INJECTED_CONTEXT_ENABLED=true`; headers are exposed with `FLAGR_INJECTED_CONTEXT_HTTP_HEADERS` / `..._HTTP_HEADER_PREFIXES`.

## Jev / System One

The `jev` enricher asks a [System One](https://docs.typesafe.ai/) model typed questions and exposes each answer as `@jev_<question>`. It is a boolean gate: variant selection and rollout stay with Flagr.

### Configure an endpoint

Setting a non-empty base URL enables the enricher.

```sh
FLAGR_INJECTED_CONTEXT_JEV_BASE_URL=http://127.0.0.1:8009
FLAGR_INJECTED_CONTEXT_JEV_MODEL=kev-latest
# FLAGR_INJECTED_CONTEXT_JEV_API_KEY=...        # optional
# FLAGR_INJECTED_CONTEXT_JEV_TIMEOUT=1s         # per-call budget, includes retries
# FLAGR_INJECTED_CONTEXT_JEV_MAX_RETRIES=2
# FLAGR_INJECTED_CONTEXT_JEV_RETRY_BASE=100ms
# FLAGR_INJECTED_CONTEXT_JEV_RETRY_MAX=500ms
```

The same `POST /v1/systemone` contract is implemented by the hosted TypeSafe API and by open-source drop-in servers — [`jaredpalmer/kev`](https://github.com/jaredpalmer/kev), [`oido-systemone`](https://github.com/Djancyp/oido-systemone), and [`jeff`](https://github.com/logan-markewich/jeff).

### Author questions

In the UI, a flag has a **Context enrichers** card. Add a Jev enricher and author one or more questions: a name, a type, instructions, type-specific criteria, and — for `choice` / `score` — an optional confidence threshold. The editor shows the property each question will produce (`@jev_<name>`) and a one-line reminder of its value shape as you type.

Each question name becomes the constraint property `@jev_<name>`. Ask **atomic** questions: the model answers one question at a time, so keep arithmetic, dates, and multi-factor logic in Flagr constraints.

### How answers become context

The endpoint returns a typed answer per question, and Flagr maps it to exactly one property value:

| Type | Response field | Property value | Confidence gate | Operators |
|------|----------------|----------------|-----------------|-----------|
| `noul` (true/false) | `noul` | a number in `[0,1]` — **P(true)** | none | `GTE` `GT` `LTE` `LT` |
| `choice` | `choice` | the chosen **option label** (string) | yes | `EQ` `NEQ` `IN` `NOT IN` |
| `score` | `score` | the **level number** | yes | `GTE` `GT` `LTE` `LT` |

`noul` is **not** a boolean. It is the probability that the answer is true, so you pick the cut-off yourself with a comparison (`@jev_billing GTE 0.7`). That is also why `noul` has **no confidence threshold** — the comparison *is* the threshold. Setting one is rejected rather than silently ignored.

You can define what true and false mean for the model with optional `criteria`, which sharpens the boundary:

```json
{
  "type": "noul",
  "instructions": "Is this message about billing?",
  "criteria": {
    "true": "Payments, invoices, refunds, or charges",
    "false": "Anything else"
  }
}
```

Either key may be omitted; `criteria` itself is optional. The UI shows these as the question's **True means** / **False means** fields.

For `choice` and `score`, the model also reports a self-assessed `confidence`. If it is below the question's threshold (or missing), Flagr **drops the property** — it is not defaulted. A constraint referencing it then errors and the segment falls through (fail-closed). A question with no threshold has no gate.

A `choice` answer is a **single** option label with a single confidence. The per-option `probabilities` an endpoint may return are not exposed: one question yields one comparable value. To branch on a second option, ask a second `noul` question (or use a `score`).

Limits: `choice` accepts up to 255 options; `score` needs 2 to 10 levels; `noul` needs no criteria.

#### Worked example

The endpoint answers four questions:

```json
{
  "answers": {
    "billing":   {"type": "noul",   "noul": 0.82},
    "plan_tier": {"type": "choice", "choice": "pro", "confidence": 0.91,
                  "probabilities": {"pro": 0.91, "free": 0.06, "enterprise": 0.03}},
    "unsure":    {"type": "choice", "choice": "pro", "confidence": 0.31},
    "sentiment": {"type": "score",  "score": 4, "confidence": 0.77}
  }
}
```

With thresholds `plan_tier: 0.5`, `unsure: 0.5`, `sentiment: 0.5`, the evaluation context becomes:

```json
{
  "message": "charged twice",
  "@ts_hour": 17,
  "@jev_billing": 0.82,
  "@jev_plan_tier": "pro",
  "@jev_sentiment": 4
}
```

`@jev_unsure` is **absent** (confidence `0.31 < 0.5`), while `@jev_billing` is present because `noul` has no gate.

Constraints and outcomes:

```
Segment "enterprise":
  {@jev_plan_tier} EQ "pro"     → true   (value is the label "pro")
  {@jev_billing} GTE 0.7        → true   (0.82 ≥ 0.7)
  {@jev_sentiment} GTE 3        → true   (4 ≥ 3)

Segment "unsure":
  {@jev_unsure} EQ "pro"        → error: argument @jev_unsure not found
                                        → segment falls through
```

Value syntax in constraints:

- `choice` values are strings, so quote them: `EQ "pro"`, `IN "pro,enterprise"`.
- `noul` / `score` values are numbers, so do not quote them: `GTE 0.7`, `GTE 3`.
- The property is `@jev_<question name>` — flat and underscore-joined, like `@ts_hour` / `@http_x_env`.

### Match

Add a normal constraint on `@jev_<name>`; the operator family must fit the type (see the table above). A mismatched operator is reported by `flagr-validate` and fails closed.

```
Segment "likely enterprise":
  Constraint: {@jev_plan_tier} EQ "enterprise"
  Constraint: {@jev_churn_risk} LT 0.3
```

Because a `choice` / `score` answer below its threshold is dropped, a segment built only on such a question can be *silently* unreachable. The Debug Console reports `argument: @jev_<name> not found` for a dropped answer.

## API

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/v1/flags/{flagID}` | `enrichers` is the **effective catalog**: flag-scoped plus enabled built-ins |
| `POST` | `/api/v1/flags/{flagID}/enrichers` | create `{namespace, config}` |
| `PUT` | `/api/v1/flags/{flagID}/enrichers/{namespace}` | replace `{config}` |
| `DELETE` | `/api/v1/flags/{flagID}/enrichers/{namespace}` | remove |

`config` for `jev` is `{"questions": {"<name>": {"type": "noul|choice|score", "instructions": ..., "criteria": ..., "confidenceThreshold": ...}}}`.

The flag's `enrichers` read model also carries read-only `scope`, `enabled`, and `properties`, which is what powers the constraint property picker in the UI.

## JSON flag source (GitOps)

Flag-scoped enrichers are part of the flag in the [JSON flag source](flagr_json_flag_spec.md), so `GET /api/v1/export/eval_cache/json` round-trips them. Built-ins are server configuration and are re-derived on load.

## Validation

- Enricher definitions are validated strictly: an unknown namespace, a global namespace written to a flag, or an invalid `jev` config is rejected (400 on the API, an error from `flagr-validate`).
- A constraint referencing an enriched property no enricher provides is a **warning**, not an error. It is accepted and fails closed at evaluation. `flagr-validate` reports it; check the Debug Console when a segment unexpectedly never matches.

## Related

- [Built-in context injection](flagr_injected_context.md) — `@ts*` / `@http_*`
- [Behavioral contracts](flagr_behavioral_contracts.md#context-enrichers) — the invariants above
- [Environment variables](flagr_env.md#context-enrichers) — the full list
