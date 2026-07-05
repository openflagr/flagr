# GET Evaluation API

Flagr supports **GET** requests that mirror **POST** evaluation bodies, for use cases described in [issue #613](https://github.com/openflagr/flagr/issues/613): CORS-simple browser requests (no preflight), HTML `<link rel="preload">`, and HTTP caching.

Implementation builds on community work in [PR #631](https://github.com/openflagr/flagr/pull/631) by [@iamafanasyev](https://github.com/iamafanasyev); the shipped wire format uses a single query parameter **`json`** with the same JSON as the POST body.

## Endpoints

| Method | Path | `json` decodes to |
|--------|------|-------------------|
| `GET` | `/api/v1/evaluation` | `evalContext` (same as POST body) |
| `GET` | `/api/v1/evaluation/batch` | `evaluationBatchRequest` (same as POST body) |

Responses are identical to POST (`evalResult` / `evaluationBatchResponse`).

## Migration from POST

Use the **same object** you would send as the POST JSON body; percent-encode it as the `json` query value.

```javascript
const ctx = {
  entityID: 'user-1',
  entityType: 'user',
  entityContext: { tier: 'premium' },
  flagID: 42,
};

// POST
await fetch('/api/v1/evaluation', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(ctx),
});

// GET — same ctx
const url = `/api/v1/evaluation?json=${encodeURIComponent(JSON.stringify(ctx))}`;
await fetch(url);
```

Batch:

```javascript
const req = {
  entities: [{ entityID: 'user-1', entityContext: { region: 'us' } }],
  flagTags: ['web'],
  flagTagsOperator: 'ANY',
};
const url = `/api/v1/evaluation/batch?json=${encodeURIComponent(JSON.stringify(req))}`;
```

## URL length and limits

- Many proxies and clients treat **~2,048** characters (full URL) as a safe default; **~8,192** bytes often works with tuned load balancers. Large `entityContext` or many batch entities may exceed GET limits — **use POST** for those cases.
- **`FLAGR_EVAL_GET_MAX_URL_BYTES`** (default `8192`, `0` = disabled) rejects oversized query strings with **400** and a message to use POST. See [Environment Variables](flagr_env.md).
- **`FLAGR_EVAL_BATCH_SIZE`** applies to GET batch the same as POST batch (total evaluations per request).

## Caching

CDNs and browsers cache GET URLs by full URL string. Use **stable JSON serialization** on the client (consistent key order, no pretty-print) so equivalent requests hit the same cache key.

## Errors

| Condition | Status |
|-----------|--------|
| Missing `json` | 400 |
| Invalid JSON / wrong schema | 400 |
| Batch size over `FLAGR_EVAL_BATCH_SIZE` | 400 |
| Query over `FLAGR_EVAL_GET_MAX_URL_BYTES` | 400 |

## Related

- [Overview](flagr_overview.md) — evaluation architecture
- [API Reference](https://openflagr.github.io/flagr/api_docs)
- Plan: [2026-07-05-001-get-evaluation-api-plan.md](plans/2026-07-05-001-get-evaluation-api-plan.md)