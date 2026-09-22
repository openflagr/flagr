# Jev / System One constraints

A **Jev constraint** is an audience-targeting predicate whose value comes from a
[System One](https://docs.typesafe.ai/) model instead of a client-provided
`entityContext` field. You author a typed question (`noul`, `choice`, or `score`)
in the Flagr UI and the model's answer is compared with ordinary operators. It
is, literally, a "fancy `if`".

Flagr talks to any endpoint that implements `POST /v1/systemone` — the hosted
API or a self-hosted open-source server such as
[`oido-systemone`](https://github.com/Djancyp/oido-systemone) or
[`jeff`](https://github.com/logan-markewich/jeff).

## 1. Configure an endpoint (do this first)

Set at minimum `FLAGR_JEV_ENABLED=true` and point `FLAGR_JEV_BASE_URL` at a
System One endpoint. Add `FLAGR_JEV_API_KEY` if the endpoint requires a bearer
token, and `FLAGR_JEV_MODEL` to pin a model.

```bash
FLAGR_JEV_ENABLED=true \
FLAGR_JEV_BASE_URL=http://127.0.0.1:8009 \
FLAGR_JEV_MODEL=kev-latest \
FLAGR_JEV_TIMEOUT=15s \
./flagr
```

Full list (including `FLAGR_JEV_CONFIDENCE_THRESHOLD`): [Environment variables](flagr_env.md#jev).

## 2. Add a Jev constraint

1. In a segment, add a constraint and flip the **JEV** switch on.
2. Give the question a name — the answer is used as `@jev.<name>`.
3. Pick a type and write the instructions:
   - **Noul** (true / false) returns `P(true)`; match with `≥` / `<`.
   - **Choice** (pick one) — match **any of** / **none of** the choices.
   - **Scale** (rate) — match **at least** / **below** a level.
4. Optionally set the **confidence** gate (choice / scale).
5. Save.

Operators are validated against the question type: `noul` / `scale` accept
`≥` / `>` / `≤` / `<`, and `choice` accepts `=` / `≠` / `in` / `not in`.

The **state** sent to the model is the full `entityContext` plus `entityID` and
`entityType`. Enable built-in context injection
(`FLAGR_INJECTED_CONTEXT_ENABLED=true`) to also include `@ts*` / `@http_*`.
Answers are used for constraint evaluation only and are never written back into
the result context.

## Cost & limits

Each flag evaluation with Jev constraints makes one batched System One call. It
is slower than a normal constraint and adds model usage cost, and Jev is
rate-limited, so keep Jev constraints off high-QPS request paths. The eval debug
log (`enableDebug: true`) includes the request, response, and latency.

Design notes: [Jev constraints plan](https://github.com/openflagr/flagr/blob/main/docs/plans/2026-09-22-001-jev-calibrated-constraints-plan.md).
