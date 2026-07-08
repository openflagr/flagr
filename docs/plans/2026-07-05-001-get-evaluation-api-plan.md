# GET Evaluation API (`?json=`)

**Status:** shipped on `feat/get-eval-api` (2026-07-05). Credit [#631](https://github.com/openflagr/flagr/pull/631) / [@iamafanasyev](https://github.com/iamafanasyev) in commits and PR.

## Contract

| Route | Query `json` decodes to | Same as POST |
|-------|-------------------------|--------------|
| `GET /api/v1/evaluation` | `evalContext` | body of `POST /evaluation` |
| `GET /api/v1/evaluation/batch` | `evaluationBatchRequest` | body of `POST /evaluation/batch` |

- **POST is primary;** GET is for CORS-simple / cacheable reads ([#613](https://github.com/openflagr/flagr/issues/613)).
- **`FLAGR_EVAL_GET_MAX_URL_BYTES`** (default `8192`, `0` = off) on raw query length; **`FLAGR_EVAL_BATCH_SIZE`** applies to GET batch like POST.
- User docs: [Use Cases — GET evaluation](../flagr_use_cases.md#get-evaluation-browser-friendly). Env: [flagr_env.md](../flagr_env.md).

## Verify

```bash
make test && make test-integration && make ci-swagger
```

## Review (2026-07-05)

**Approve.** Handler in `pkg/handler/eval.go`; GET tests in `eval_get_test.go`; `decodeFromGetQuery` shared pipeline.