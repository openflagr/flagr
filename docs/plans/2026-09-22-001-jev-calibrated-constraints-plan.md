# feat: Jev / System One Constraints — Calibrated Audience Targeting

**Date:** 2026-09-22
**Status:** as-built
**Branch:** `zz-jev-test`

## Summary

Add a first-class **Jev constraint** to Flagr: an audience-targeting predicate whose
value is produced by a [TypeSafe System One model](https://docs.typesafe.ai/)
(`noul` / `choice` / `score`) instead of a client-provided `entityContext` field.
A Jev constraint is authored entirely in the Flagr UI, stored inline on the
`constraint` row, and used as a synthetic property
`@jev.<name>` so the existing conditions engine performs the comparison
unchanged. It is, literally, a "fancy `if`".

The same `POST /v1/systemone` contract is implemented by the hosted TypeSafe API
and by open-source drop-in servers — [`oido-systemone`](https://github.com/Djancyp/oido-systemone)
(Go, local GGUF) and [`jeff`](https://github.com/logan-markewich/jeff) (Python,
GLiFormer). Flagr talks to whichever endpoint `FLAGR_JEV_BASE_URL` points at.

User guide: [`docs/flagr_jev.md`](../flagr_jev.md).

## Problem Frame

Flagr constraints are deterministic (`EQ`, `LT`, `IN`, regex, …) evaluated against
`entityContext`. This is exact but brittle: expressing "is this visitor a likely
enterprise buyer?" requires the caller to precompute the judgment and send a
boolean. There is no way to make a fuzzy, calibrated audience decision inside the
flag itself.

Jev exposes exactly that as a typed primitive: state in, calibrated decision out.
Because Jev answers are typed values, they slot into Flagr's existing
`conditions.Evaluate(expr, map)` path as ordinary variables — no evaluator change.

## Key Technical Decisions

1. **Boolean-only semantics** — every Jev question compiles to a `true`/`false`
   gate that participates in the segment's `AND`. Variant selection and rollout
   stay entirely with Flagr. No value-returning mode in v1.
2. **Synthetic property `@jev.<name>`** — the answer is exposed under a nested
   `@jev` map for constraint evaluation only, so `{@jev.plan_tier} == "pro"` resolves via the conditions
   library's existing dotted-path traversal. **Verified** against
   `conditions@v0.2.6` (see `pkg/handler/jev_eval_test.go`). No parser or
   operator changes.
3. **Inline authoring** — the question lives on the `constraint` row, not in a
   separate registry. Structured fields + JSON escape hatch (`instructions` /
   `criteria` are free-form JSON, matching Jev's `EntryType`).
4. **State = `entityContext` + entity identity** — the whole evaluation context is
   the Jev `state`, including server-injected `@ts*`/`@http_*` keys, plus
   `entityID` and `entityType` (which are not part of `entityContext`). Only
   Flagr's own `@jev` answer namespace is excluded, so answers are never fed back
   to the model.
5. **One batched call per flag evaluation** — all `@jev.*` questions in a flag are
   fanned out in a single `POST /v1/systemone`, leveraging Jev's parallel
   question evaluation.
6. **Confidence gate** — `choice` / `score` carry a per-constraint
   `confidenceThreshold`. Below it, the answer is omitted, the expression
   errors, and `evalSegment` already treats an error as "no match" → the segment
   falls through. `noul` has no separate confidence; its probability comparison
   *is* the gate.
7. **Fail-closed** — Jev disabled, timed out, or erroring ⇒ no `@jev` answers ⇒
   every Jev constraint evaluates false. There is no fail-open knob.
8. **Env-var config** — `FLAGR_JEV_*`. No secrets in the DB, no UI settings page.
9. **API-first** — the swagger contract is defined before the implementation.
10. **No answer cache (v1)** — the state (entityContext + entity identity) changes
    per request, so caching answers is low value. Revisit later if a stable
    subset can be keyed.

## API Contract (swagger)

Both `constraint` and `createConstraintRequest` gain an optional `jev` object:

```yaml
jev:
  type: object
  description: >
    Optional Jev / System One question backing this constraint. When present,
    `property` must be `@jev.<name>`. The model answer is used as
    `@jev.<name>` for constraint evaluation and compared with `operator`/`value`;
    it is not written into the result context.
  properties:
    type:
      type: string
      enum: ["noul", "choice", "score"]
    instructions:
      type: object            # string | object | array (go-swagger: any)
    criteria:
      type: object            # choice: {name: description}; score: [level, ...]
    confidenceThreshold:
      type: number
      format: double
      minimum: 0
      maximum: 1
```

`segmentDebugLog` gains an optional free-form `jev` object carrying the System
One request and response for that segment (debug only):
`{model, state, questions, answers, usage, latencyMs, serverLatencyMs, error}`.
It is present only when the segment has Jev constraints, and `questions` /
`answers` are scoped to that segment.

## Data Model (`pkg/entity/constraint.go`)

```go
const (
    JevTypeNoul   = "noul"
    JevTypeChoice = "choice"
    JevTypeScore  = "score"

    JevPropertyPrefix = "@jev."
    JevContextKey     = "@jev"
)

type Constraint struct {
    gorm.Model
    SegmentID uint
    Property  string
    Operator  string
    Value     string `gorm:"type:text"`

    JevType                string   `gorm:"type:varchar(16)"`
    JevInstructions        string   `gorm:"type:text"` // JSON EntryType
    JevCriteria            string   `gorm:"type:text"` // JSON criteria
    JevConfidenceThreshold *float64
}

type JevQuestion struct {
    Type                string   `json:"type"`
    Instructions        any      `json:"instructions,omitempty"`
    Criteria            any      `json:"criteria,omitempty"`
    ConfidenceThreshold *float64 `json:"confidenceThreshold,omitempty"`
}
```

`Constraint.Validate()` rejects a `jev.type` that is not `noul|choice|score`,
requires `property` to start with `@jev.`, requires choice criteria to be a
JSON object with ≥1 option, score criteria to be a JSON array of 2–10 levels,
and `confidenceThreshold` to be within `[0, 1]`.

`Flag.PrepareEvaluation()` collects all Jev constraints into
`FlagEvaluation.JevQuestions` (deduped by name, first segment-rank wins) so the
hot path never parses JSON.

## Runtime Flow (`pkg/handler/jev_*.go`, `eval.go`)

```
EvalFlagWithContext
  ├─ resolve flag + entityContext
  ├─ resolveJevForFlag(evalContext, flag)        # new, does not mutate evalContext
  │    ├─ skip if !FLAGR_JEV_ENABLED or no Jev questions
  │    ├─ state = entityContext + entityID/entityType (minus `@jev`)
  │    ├─ POST {base}/v1/systemone (Bearer key, timeout)
  │    ├─ build answers map; omit answers below confidenceThreshold
  │    └─ return (answers, debug); the result context stays clean
  └─ evalSegment loop
       ├─ merge answers under `@jev` into a private eval map
       └─ conditions.Evaluate parses {@jev.<name>} via path traversal
```

## Configuration

Env vars live in `pkg/config/env.go` and are documented in
[Environment variables](../flagr_env.md#jev) and the
[Jev guide](../flagr_jev.md). Summary: `FLAGR_JEV_ENABLED`, `FLAGR_JEV_BASE_URL`,
`FLAGR_JEV_API_KEY`, `FLAGR_JEV_MODEL`, `FLAGR_JEV_TIMEOUT`,
`FLAGR_JEV_CONFIDENCE_THRESHOLD`. There is no answer cache in v1.

## UI

- A subtle **JEV switch** sits in the constraint row's action area (right side).
  The `@jev.` prefix appears on the property input when on, and toggling off
  clears the property.
- `JevQuestionEditor.vue` owns the whole match: type selector, `instructions`,
  type-aware criteria (noul `true`/`false`; choice option rows; scale ordered
  levels), and a confidence control.
  - noul: `P(true)` threshold slider (`≥` / `<`)
  - choice: `any of` / `none of` option multi-select + confidence
  - scale: `at least` / `below` level + confidence
  - `Edit as JSON` escape hatch for structured `instructions` / `criteria`
- `helpers/jevQuestion.ts`: pure `reduceJevMatch` reducer + value/direction
  conversions, unit-tested in `jevQuestion.test.ts`.
- `api/types.ts`: `JevQuestion` DTO.

## Non-Goals (deferred)

Value-returning answers; a reusable `Audience` registry; secrets in a UI settings
page; fail-open; calibration ledger; persisting probabilities into data records;
freezing Jev answers in eval-only snapshots; cross-flag batch optimization;
in-editor "try this question" preview.

## Testing

- `pkg/entity/constraint_test.go`: validation matrix (types, criteria, threshold).
- `pkg/handler/jev_client_test.go`: client against an `httptest` mock that
  implements the `oido-systemone` contract (request shape, Bearer auth, error
  mapping, timeout).
- `pkg/handler/jev_eval_test.go`: end-to-end segment match for all three types,
  confidence fall-through, fail-closed, mixed plain+Jev constraints, per-segment
  debug scoping, and that the result context is never mutated with `@jev`.
- `pkg/handler/jev_integration_test.go`: full path (real HTTP client → mock
  System One server → on-the-fly answers → conditions → variant) with a single
  batched call carrying all three question types, plus a fail-closed case.
- `pkg/handler/crud_jev_test.go`: create/find/update through the REST CRUD
  handlers, including rejection of invalid questions and non-`@jev.` properties.
- `browser/flagr-ui/src/helpers/jevQuestion.test.ts`: criteria <-> form
  conversions, readiness validation, the `reduceJevMatch` reducer (incl. choice
  IN/NOTIN and direction preservation), and the any-of/none-of direction mapping.
- `TestJevConstraintQuestionTypes` (table-driven): noul/choice/scale across
  `GTE`/`LT`, `EQ`/`NEQ`, `IN`/`NOTIN` with match and no-match cases, plus the
  confidence gate.
- `make test`, `make flagr-ui-check`.

### Local testing against a mock or self-hosted endpoint

Run the mock-backed suite:

```bash
go test ./pkg/entity/ ./pkg/handler/ -run Jev -count=1
```

The `jev_integration_test.go` mock speaks the exact `POST /v1/systemone` request
and response shapes (including `noul` / `choice` / `score` answers and Bearer
auth), so it doubles as the contract fixture.

To exercise a real self-hosted open-source endpoint, run one of the drop-in
servers and point Flagr at it:

```bash
# Python (GLiFormer), default http://localhost:8000
JEFF_API_KEYS=devkey uv run jeff

FLAGR_JEV_ENABLED=true \
FLAGR_JEV_BASE_URL=http://localhost:8000 \
FLAGR_JEV_API_KEY=devkey \
./flagr
```

```bash
# Go (local GGUF), default http://localhost:8080
./oido-systemone

FLAGR_JEV_ENABLED=true FLAGR_JEV_BASE_URL=http://localhost:8080 ./flagr
```

Then create a segment constraint with property `@jev.<name>`, operator/value
comparison, and the question in the UI editor.

## Refinements after first implementation

- **Dropped the answer cache.** The state (entityContext + entity identity)
  changes per request, so a `(model, state, questions)` cache rarely hit. Removed
  `FLAGR_JEV_CACHE_*`. Revisit only if a genuinely stable key exists.
- **No `@jev` in the evaluation context.** `resolveJevForFlag` returns the answers
  separately; `evalSegment` merges them into a private map for
  `conditions.Evaluate`. The result context (and data records) stay clean.
- **Entity identity in the state.** `entityID` / `entityType` are added to the
  System One state (canonical values win over same-named context keys).
- **Debug payload.** `segmentDebugLog.jev` carries the request/response, usage,
  latency, and errors; it is scoped to the segment and absent for non-Jev
  segments.
- **`true` / `false` copy.** Noul criteria and match use Jev's `{true, false}`
  keys (`P(true) ≥ …`).
- **UI polish.** JEV toggle moved to the row actions as a subtle switch; the
  editor shows a notice that Jev constraints are slower and add model cost, and
  points at the [Jev guide](../flagr_jev.md) to set up an endpoint first.
- **One question type everywhere.** `entity.JevQuestion` backs the flag's
  evaluation map, the client request, and the debug payload; the transient
  `JevConstraintSpec` / `JevDebugQuestion` wrappers were dropped.
- **Jev-free `evalSegment`.** `EvalFlagWithContext` merges answers into a private
  constraint context; `evalSegment` keeps its original signature and knows
  nothing about Jev.
- **Operator/type validation.** `noul` / `scale` accept `GTE`/`GT`/`LTE`/`LT`,
  `choice` accepts `EQ`/`NEQ`/`IN`/`NOTIN`, so an incompatible match is rejected
  instead of silently never matching.

## Risks

- **Hot-path latency**: 70–500 ms per Jev call. Mitigated by one
  batched call per flag, and fail-closed. Documented as suitable for
  low-QPS / high-value targeting, not every request.
- **Cost & rate limits**: Jev is billed per input token and rate-limits requests,
  and Flagr makes one batched call per flag evaluation, so cost scales with
  `entities × flags-with-Jev × evaluations`. A high-QPS path can hit the request
  limit; there is no retry/backoff yet, so Jev constraints stay fail-closed.
- **Self-hosted parity**: `oido-systemone` is an independent reimplementation;
  the client tolerates missing `usage`/extra fields and only relies on the
  documented answer fields.
- **Question-name collisions** across segments in one flag: first segment-rank
  wins; conflicting definitions are a known follow-up validation.
- **Model jaggedness** (Jev docs): keep arithmetic, dates, and multi-factor
  reasoning in Flagr constraints; ask atomic questions.
