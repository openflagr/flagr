# GET Evaluation API (`?json=`)

Functionally equivalent **GET** routes for flag evaluation, motivated by [#613](https://github.com/openflagr/flagr/issues/613): CORS-simple requests (no preflight), HTML preload, and HTTP caching. Delivers on top of community work in [#631](https://github.com/openflagr/flagr/pull/631) by [@iamafanasyev](https://github.com/iamafanasyev) with a revised wire contract.

**Status:** implemented (2026-07-05). Branch `feat/get-eval-api`.

**Credit:** `Co-authored-by:` on commits retaining #631 logic; PR description **Based on #631**.

## Locked decisions (grill-me, 2026-07-05)

| Topic | Decision |
|-------|----------|
| Query contract | Single param **`json`** on both routes; value = **percent-encoded JSON** matching POST body schema for that path |
| `GET /evaluation` | `json` unmarshals to **`evalContext`** (same as POST body) |
| `GET /evaluation/batch` | `json` unmarshals to **`evaluationBatchRequest`** (same as POST body) |
| POST parity | Same field names and models as POST; no abbreviated query keys (`dbg`, `id`, …) |
| Batch encoding | Whole batch request in one `json` param (not repeated `entities`) |
| URL limits | Document ~2K conservative / ~8K with proxy tuning; **POST** remains escape hatch |
| Caching | Clients should use **stable JSON serialization** (e.g. sorted keys) for cache keys |
| Batch DoS | **`FLAGR_EVAL_BATCH_SIZE`** applies to GET batch same as POST |
| Optional guard | **`FLAGR_EVAL_GET_MAX_URL_BYTES`** (default `8192`, `0` = disabled) on raw query length |
| v1 scope | **API + user docs**; Debug Console GET later |
| Credit | **`Co-authored-by:`** + #631 in PR |

## API

### `GET /api/v1/evaluation`

| | |
|---|---|
| **operationId** | `getEvaluation` |
| **Query** | `json` (required, string): URL-decoded value is JSON for `#/definitions/evalContext` |
| **Success** | `200` + `evalResult` (same as POST) |
| **Errors** | `400` missing/invalid JSON; optional URL too long |

### `GET /api/v1/evaluation/batch`

| | |
|---|---|
| **operationId** | `getEvaluationBatch` |
| **Query** | `json` (required, string): URL-decoded value is JSON for `#/definitions/evaluationBatchRequest` |
| **Success** | `200` + `evaluationBatchResponse` (same as POST) |
| **Errors** | `400` invalid JSON / batch size exceeded; optional URL too long |

### Migration (POST → GET)

```javascript
// Single
const ctx = { entityID: 'u1', flagTags: ['web'], entityContext: { tier: 'premium' } };
const url = `/api/v1/evaluation?json=${encodeURIComponent(JSON.stringify(ctx))}`;

// Batch — same object as POST body
const req = { entities: [{ entityID: 'u1', entityContext: { region: 'us' } }], flagTags: ['integ'] };
const url = `/api/v1/evaluation/batch?json=${encodeURIComponent(JSON.stringify(req))}`;
```

## Implementation

1. **Swagger** — add `get:` to `swagger/evaluation.yaml`, `swagger/evaluation_batch.yaml` with `json` query param.
2. **`make gen`** — `api_docs` + `swagger` (do not hand-edit `swagger_gen/`).
3. **`pkg/handler/eval.go`** — `GetEvaluation` / `GetEvaluationBatch`; extract **`EvaluateBatch`** shared by POST/GET batch; URL length check helper.
4. **`pkg/handler/handler.go`** — register GET handlers in `setupEvaluation`.
5. **`pkg/config/env.go`** — `EvalGetMaxURLBytes` if not present.
6. **Tests** — unit in `eval_test.go`; integration `TestIntegration_GetEvaluation` (and extend evaluation tests).
7. **Docs** — `docs/flagr_get_evaluation.md`, sidebar link; touch `flagr_overview.md` architecture table; `flagr_env.md` for new env.

## Out of scope (v1)

- Debug Console GET URL builder / buttons.
- Andrei’s abbreviated query param aliases.
- `entity` repeated-param compatibility.

## Verification

```bash
make test
make test-integration
make ci-swagger   # before push
```

## References

- Issue [#613](https://github.com/openflagr/flagr/issues/613)
- PR [#631](https://github.com/openflagr/flagr/pull/631) (closed, `ama/get-evaluation-api`)