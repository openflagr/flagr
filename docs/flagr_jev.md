# Jev / System One constraints

A **Jev constraint** is an audience-targeting predicate whose value comes from a
[System One](https://docs.typesafe.ai/) model instead of a client-provided
`entityContext` field. You author a typed question (`noul`, `choice`, or `score`)
in the Flagr UI and the model's answer is compared with ordinary operators. It
is, literally, a "fancy `if`".

Flagr talks to any endpoint that implements `POST /v1/systemone` with the same
`{state, model, questions}` contract — the hosted TypeSafe API, or a local
open-source server such as [Kev](https://github.com/jaredpalmer/kev),
[`oido-systemone`](https://github.com/Djancyp/oido-systemone), or
[`jeff`](https://github.com/logan-markewich/jeff).

## 1. Configure an endpoint (do this first)

Flagr reads env vars at startup, so export them before starting the server. Pick
one of the endpoints below, then verify it answers (last subsection) before
authoring constraints.

### Hosted TypeSafe

Sign in at [typesafe.ai](https://typesafe.ai), create an API key, and use the
hosted endpoint:

```bash
export FLAGR_JEV_ENABLED=true
export FLAGR_JEV_BASE_URL=https://api.typesafe.ai
export FLAGR_JEV_API_KEY="YOUR_TYPESAFE_API_KEY"
export FLAGR_JEV_MODEL=jev-latest
export FLAGR_JEV_TIMEOUT=15s   # model inference is slower than the 1s default
```

### Local: Kev

[Kev](https://github.com/jaredpalmer/kev) runs small Jev-style models (0.8B / 4B
/ 9B) locally on CUDA, ROCm, or Apple Silicon. It needs Python 3.12+ and
[`uv`](https://docs.astral.sh/uv/), and downloads the base model on first run:

```bash
git clone https://github.com/jaredpalmer/kev.git && cd kev
uv sync --extra serve
KEV_DTYPE=bf16 uv run --extra serve python -m kev.serve --run jaredpalmer/kev-4b --port 8009
```

Then point Flagr at it (Kev does not require an API key):

```bash
export FLAGR_JEV_ENABLED=true
export FLAGR_JEV_BASE_URL=http://127.0.0.1:8009
export FLAGR_JEV_MODEL=kev-latest
export FLAGR_JEV_TIMEOUT=30s
```

### Local: oido-systemone

[oido-systemone](https://github.com/Djancyp/oido-systemone) is a single Go binary
that runs a GGUF model (MiniCPM5-2B or Qwen3.5-4B) in-process. It needs Go
1.27+, a C toolchain, and internet on first run to download the model:

```bash
git clone https://github.com/Djancyp/oido-systemone.git && cd oido-systemone
go build -o oido-systemone . && ./oido-systemone
```

It listens on `:8080` (Swagger UI at `/docs`). Point Flagr at it:

```bash
export FLAGR_JEV_ENABLED=true
export FLAGR_JEV_BASE_URL=http://127.0.0.1:8080
export FLAGR_JEV_MODEL=jev-latest   # or oido-rlhf-minicpm5-2b / oido-rlhf-qwen3.5-4b
export FLAGR_JEV_TIMEOUT=30s
```

If you set `API_KEY` on the server, set the same value as `FLAGR_JEV_API_KEY`.

### Local: jeff

[jeff](https://github.com/logan-markewich/jeff) is a lightweight Python server
(GLiFormer, 400M). It needs [`uv`](https://docs.astral.sh/uv/) and downloads
weights on first run:

```bash
uv sync --extra dev
uv run hf download knowledgator/gliformer-large-v1 --local-dir models/gliformer-large-v1
JEFF_API_KEYS=devkey uv run jeff
```

It serves on `:8000`. Point Flagr at it:

```bash
export FLAGR_JEV_ENABLED=true
export FLAGR_JEV_BASE_URL=http://127.0.0.1:8000
export FLAGR_JEV_API_KEY=devkey
export FLAGR_JEV_MODEL=jev-latest
export FLAGR_JEV_TIMEOUT=15s
```

### Verify the endpoint

Before wiring up constraints, confirm the endpoint answers. This is the exact
request Flagr sends:

```bash
curl -sS "$FLAGR_JEV_BASE_URL/v1/systemone" \
  -H "Authorization: Bearer $FLAGR_JEV_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "'"$FLAGR_JEV_MODEL"'",
    "state": {"message": "I was charged twice for my subscription"},
    "questions": {
      "billing": {"type": "noul", "instructions": "Is this message about billing?"}
    }
  }'
```

Expect `answers.billing.noul` in the response. If this fails, Flagr will too —
Jev is **fail-closed**, so a bad key/model/timeout silently evaluates the
constraint to `false`. Drop the `Authorization` header for keyless local servers
(Kev, default oido-systemone).

Then start Flagr with those variables still exported:

```bash
make rebuild-run      # build + backend :18000 + UI :8080
# or: make build && ./flagr --port 18000
```

Full list: [Environment variables](flagr_env.md#jev).

## 2. Add a Jev constraint

1. In a segment, add a constraint and flip the **JEV** switch on.
2. Give the question a name — the answer is used as `@jev.<name>`. Names must
   start with a letter or underscore and contain only letters, digits, and
   underscores (the UI slugs them for you).
3. Pick a type and write the instructions:
   - **Noul** (true / false) returns `P(true)`; match with `≥` / `<`.
   - **Choice** (pick one) — match **any of** / **none of** the choices.
   - **Scale** (rate) — match **at least** / **below** a level.
4. Set the **confidence** gate (choice / scale; defaults to 0.5).
5. Save.

Operators are validated against the question type: `noul` / `scale` accept
`≥` / `>` / `≤` / `<`, and `choice` accepts `=` / `≠` / `in` / `not in`. Each
`@jev.<name>` may be defined by only one constraint per flag — a second
constraint reusing the name is rejected.

The **state** sent to the model is the full `entityContext` plus `entityID` and
`entityType`. Enable built-in context injection
(`FLAGR_INJECTED_CONTEXT_ENABLED=true`) to also include `@ts*` / `@http_*`.
Answers are used for constraint evaluation only and are never written back into
the result context.

## 3. How it works

### The state sent to the model

Flagr builds the System One `state` from the evaluation's `entityContext`, then
adds the entity identity and drops its own answer namespace:

- starts from `entityContext`,
- adds `entityID` and `entityType` (canonical values win over same-named
  `entityContext` keys),
- removes `@jev` (answers are never fed back to the model),
- keeps server-injected `@ts*` / `@http_*` keys when
  [`FLAGR_INJECTED_CONTEXT_ENABLED`](flagr_injected_context.md) is on — injection
  runs before Jev.

Given this evaluation request:

```json
{
  "flagID": 42,
  "entityID": "user-4711",
  "entityType": "account",
  "entityContext": {
    "message": "I was charged twice for my subscription",
    "plan": "pro"
  }
}
```

the model receives:

```json
{
  "message": "I was charged twice for my subscription",
  "plan": "pro",
  "entityID": "user-4711",
  "entityType": "account"
}
```

With built-in context injection enabled it also carries `@ts`, `@ts_hour`,
`@ts_weekday`, `@ts_month`, and any configured `@http_*` header keys. Because the
state is an object, questions can point at nested values by name in backticks,
e.g. <code>`account.plan`</code>.

### One batched call per flag

Every `@jev.<name>` question in the flag is collected once at flag-load time and
sent in a single `POST /v1/systemone` (questions are answered in parallel):

```json
{
  "model": "jev-latest",
  "state": { "message": "...", "plan": "pro", "entityID": "user-4711", "entityType": "account" },
  "questions": {
    "billing":  { "type": "choice", "instructions": "Which team should handle this message?", "criteria": { "billing": "Charges, invoices, payments", "shipping": "Deliveries", "returns": "Exchanges and refunds" } },
    "escalate": { "type": "noul", "instructions": "Does this need urgent human attention?" },
    "urgency":  { "type": "score", "instructions": "How urgent is this message?", "criteria": ["Not urgent", "Somewhat urgent", "Very urgent"] }
  }
}
```

Answers come back under the same keys:

```json
{
  "model": "jev-latest",
  "answers": {
    "billing":  { "type": "choice", "choice": "billing", "confidence": 0.91 },
    "escalate": { "type": "noul", "noul": 0.87 },
    "urgency":  { "type": "score", "score": 2.1, "confidence": 0.78 }
  },
  "usage": { "input_tokens": 96, "output_tokens": 12 },
  "latency_ms": 210
}
```

### How the constraint is evaluated

Flagr turns the answers into a synthetic `@jev` object and merges it into a
**private copy** of the context, used only for that segment's constraint
expression:

```json
{ "@jev": { "billing": "billing", "escalate": 0.87, "urgency": 2.1 } }
```

Each constraint's `property` is `@jev.<name>`, so the existing conditions engine
resolves it like any dotted path. A segment with these constraints:

| Property | Operator | Value |
|---|---|---|
| `@jev.billing` | `=` | `"billing"` |
| `@jev.escalate` | `≥` | `0.8` |
| `@jev.urgency` | `≥` | `2` |

compiles to one AND expression:

```text
({@jev.billing} == "billing") AND ({@jev.escalate} >= 0.8) AND ({@jev.urgency} >= 2)
```

which evaluates to `true AND true AND true` → the segment matches, and Flagr runs
its rollout and distribution to pick a variant. A miss (or an error) falls
through to the next segment, exactly like a normal constraint. The answers are
**never written back** into `EvalResult.evalContext` or data records — they exist
only for the constraint comparison.

### Confidence, errors, and fail-closed

- **Confidence gate** — for `choice` / `score`, an answer whose confidence is
  below the constraint's `confidenceThreshold` is dropped. The property is then
  missing, the expression errors, and the segment falls through. `noul` has no
  separate confidence: its `P(true)` is the value you compare.
- **Fail-closed** — if `FLAGR_JEV_ENABLED` is off, the endpoint errors, or the
  call times out, there are no `@jev` answers, so every Jev constraint is
  `false`. There is no fail-open mode.
- **Bounded retries** — transient failures (network errors, 5xx, 429) are
  retried with backoff inside `FLAGR_JEV_TIMEOUT`; a failure after that still
  fails closed.

Turn on the [Debug Console](flagr_debugging.md) (`enableDebug: true`) to see the
`state`, `questions`, `answers`, `latencyMs`, `retries`, and any `error` for each
segment.

## Cost & limits

Each flag evaluation with Jev constraints makes one batched System One call. It
is slower than a normal constraint and adds model usage cost, and Jev is
rate-limited, so keep Jev constraints off high-QPS request paths. Transient
failures (network errors, 5xx, 429) are retried with exponential backoff and
jitter within the `FLAGR_JEV_TIMEOUT` budget; once it is exhausted the constraint
falls through. The eval debug log (`enableDebug: true`) includes the request,
response, latency, and retry count.

Design notes: [Jev constraints plan](https://github.com/openflagr/flagr/blob/main/docs/plans/2026-09-22-001-jev-calibrated-constraints-plan.md).
