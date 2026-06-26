# Exposure Logging

There is a quiet gap between *assigning* a variant and *showing* it. When you
call `POST /evaluation`, Flagr decides which treatment a user *should* get —
but the user never sees that decision until your UI renders. The eval call
happens on the server, often before render; it runs on every navigation; it
fires even when the result is "no match." If you count eval calls as
experiment participants, you'll inflate denominators with users who prefetched,
scrolled past, or never saw the surface at all.

Exposure logging closes that gap. It records when a user **actually saw** a
flag or experiment surface — a client-reported impression, fired on render or
visibility, separate from the server-side assignment. The typical flow: call
`POST /evaluation` to get a variant, cache it client-side, then call
`POST /exposures` when the UI actually shows the treatment. That impression
is your trustworthy experiment denominator.

For data recorders (Kafka, Kinesis, Pub/Sub), wire format, a sample Kafka
consumer, and A/B analysis, see
[Data Recorders & A/B Analysis](flagr_eval_exposure_pipeline.md). For where
exposures sit in the server architecture (Evaluator vs Metrics), see
[Overview — Architecture](flagr_overview.md#architecture).

## Endpoint

`POST /api/v1/exposures`

Request body:

```json
{
  "exposures": [
    {
      "flagKey": "my-feature",
      "variantKey": "treatment",
      "entityID": "user-123",
      "entityType": "user",
      "entityContext": { "country": "US", "page": "/checkout" },
      "flagSnapshotID": 42,
      "timestamp": "2026-06-25T12:00:00Z"
    }
  ]
}
```

Each row also accepts `flagID` and `variantID` as alternatives to the key-based
fields.

### `entityContext`

Use the same **`entityContext`** object as `POST /evaluation`: any optional
JSON object, stored on the recorded row as `evalContext.entityContext` (for
your warehouse or analytics). Put segment-style attributes (`country`, `tier`)
and impression-specific attributes (`page`, `component`) in one map — there is
no separate `metadata` field.

> **Warning:** Exposure ingest does **not** re-run segment constraints. Context
> is only copied onto the data record — it is not evaluated.

Single impressions use an array of one item. Maximum batch size:
`FLAGR_EXPOSURE_BATCH_SIZE` (default **100**).

### Typical client flow

1. `POST /evaluation` — get the variant assignment.
2. Render the experiment surface in the UI.
3. `POST /exposures` — report that the user saw it (batch on page unload if
   needed).

## Response

- **200** with `loggedCount`, `message`, and optional `errors[]` (`index`,
  `message`) for per-row validation failures.
- **400** for an empty body or a batch over the limit.

`loggedCount` is the number of rows written to the data recorder. If the
recorder is off or the flag's `dataRecordsEnabled` is false, valid rows still
return **200** with `loggedCount: 0`.

## Recording

Exposure rows reuse the exact same recording pipeline as evaluations — the
same gate, the same fan-out, the same wire format. This is deliberate: your
downstream consumer doesn't need a second topic or a second parser. The only
difference is the `recordSource` tag, which lets one consumer route both
event types. The gate is the same three conditions as evaluation data records:

1. `FLAGR_RECORDER_ENABLED=true`
2. At least one recorder in `FLAGR_RECORDER_TYPE` (e.g. `kafka`, `kinesis`,
   `pubsub` — not `datar` alone if you need a stream)
3. Per-flag `dataRecordsEnabled: true`

Recorded rows use the same `evalResult` JSON shape as evaluations on
**Kafka, Kinesis, and Pub/Sub**, with `recordSource: "exposure"`. **Datar does
not count exposures.** Exposure ingest uses separate Statsd metrics
(`exposure.ingest`, `exposure.recorded`).

## How this connects to evaluation and `AsyncRecord`

Understanding the split means understanding the two code paths. Evaluation is
the *server's* record of a decision; exposure is the *client's* record of an
impression. They converge at `AsyncRecord` but arrive there differently.

**Evaluation** runs bucketing, then `logEvalResult` (eval Statsd/Prometheus).
Results use `recordSource: "evaluation"`. When `dataRecordsEnabled` and
`FLAGR_RECORDER_ENABLED`, it calls `GetDataRecorder().AsyncRecord(evalResult)`.

**Exposure** skips `logEvalResult`. For valid rows it builds an exposure
`evalResult` with `recordSource: "exposure"` and, when the same gates pass,
calls `GetDataRecorder().AsyncRecord` directly.

`AsyncRecord` fans out to configured **data recorders** (Kafka, Kinesis,
Pub/Sub, Datar). **Datar** ignores rows with `recordSource == "exposure"`.
Exposure uses Statsd `exposure.ingest` / `exposure.recorded`.

### Eval-only mode

`json_file` / `json_http` deployments run in eval-only mode and register
**evaluation** only; **`POST /exposures` is not available** on those
deployments.

### Downstream consumers (Kafka, Kinesis, Pub/Sub)

Filter or branch on `recordSource`:

- `evaluation` — rows from the evaluation API path, including the no-segment-
  match blank (no variant assigned). Flag-not-found and flag-disabled paths
  early-return **before** recording, so they do not appear in the stream.
- `exposure` — client-reported impression.

Do not treat exposure rows as assignments for experiment analysis.

### Statsd (exposure ingest)

Separate from the eval `evaluation` metric:

- `exposure.ingest` — tags: `status=accepted|rejected|recorded`, `FlagID`,
  `FlagKey`
- `exposure.recorded` — emitted when a row is passed to `AsyncRecord`

## Validation

| Field | Rule |
|-------|------|
| `entityID` | Required |
| `flagID` / `flagKey` | At least one; if both present, must match the same flag |
| `variantID` / `variantKey` | Optional; if set, must exist on the flag (and match each other if both set) |
| `flagSnapshotID` | Optional; if set, recorded as sent (downstream can join to snapshots); if omitted, uses the current eval-cache snapshot |
| Disabled flags | Allowed |

### Authentication

`/exposures` is on the default auth whitelist alongside `/evaluation`, so
both are open by default when JWT or Basic auth is enabled. To require
credentials for impressions, remove `/api/v1/exposures` from
`FLAGR_JWT_AUTH_WHITELIST_PATHS` / `FLAGR_BASIC_AUTH_WHITELIST_PATHS`.

> **Warning:** `/exposures` is a write endpoint — callers supply the variant
> and entity, and Flagr records it as an impression. If you leave it open,
> an unauthenticated caller can submit fabricated impressions and inflate
> your experiment denominators. Flagr does not ship an in-process rate limiter
> for this endpoint; protect it at your edge (Cloudflare, AWS API Gateway, Kong,
> or a similar reverse proxy with rate limiting) to bound volume from
> untrusted sources. For deployments where impression integrity matters, also
> remove `/api/v1/exposures` from `FLAGR_JWT_AUTH_WHITELIST_PATHS` /
> `FLAGR_BASIC_AUTH_WHITELIST_PATHS` so impressions require credentials.

Design notes are in the repo under
`docs/plans/2026-06-25-001-exposure-logging-plan.md`
([view on GitHub](https://github.com/openflagr/flagr/blob/master/docs/plans/2026-06-25-001-exposure-logging-plan.md)).