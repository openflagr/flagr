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

`segmentDebugLog` gains an optional free-form `jev` object exposing the answers,
confidence, and probabilities used while evaluating that segment (debug only).

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

| Env var | Default | Purpose |
|---|---|---|
| `FLAGR_JEV_ENABLED` | `false` | Master switch |
| `FLAGR_JEV_BASE_URL` | `https://api.typesafe.ai` | Hosted API or self-hosted System One |
| `FLAGR_JEV_API_KEY` | `""` | `Authorization: Bearer` (optional for some self-hosted) |
| `FLAGR_JEV_MODEL` | `jev-latest` | Model / alias |
| `FLAGR_JEV_TIMEOUT` | `1s` | Per-request timeout |
| `FLAGR_JEV_CONFIDENCE_THRESHOLD` | `0.5` | Default when a constraint omits one |

## UI

- `ConstraintAddRow.vue` / `ConstraintExistingRow.vue`: a "Jev question" source
  toggle on the property cell.
- New `JevQuestionEditor.vue`: type selector, `instructions`, type-specific
  criteria (noul yes/no descriptions; choice option rows; score ordered levels),
  confidence slider, and a per-field **Edit as JSON** escape hatch.
- `ConstraintValueCell.vue`: when the constraint is a Jev question, show the
  match widget (probability / option select / level select) and the question
  summary.
- `api/types.ts` + `api/crud.ts`: `JevQuestion` DTO and wiring.

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
  conversions and readiness validation.
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

## Risks

- **Hot-path latency**: 70–500 ms per Jev call. Mitigated by one
  batched call per flag, and fail-closed. Documented as suitable for
  low-QPS / high-value targeting, not every request.
- **Self-hosted parity**: `oido-systemone` is an independent reimplementation;
  the client tolerates missing `usage`/extra fields and only relies on the
  documented answer fields.
- **Question-name collisions** across segments in one flag: first segment-rank
  wins; conflicting definitions are a known follow-up validation.
- **Model jaggedness** (Jev docs): keep arithmetic, dates, and multi-factor
  reasoning in Flagr constraints; ask atomic questions.
