# Datar Aggregate Analytics

Not every team that runs feature flags has a data pipeline. Sometimes you
just want to know: *how many evaluations did this flag get this week, split
by variant and segment?* Standing up Kafka, a consumer, and a warehouse to
answer that question is overkill. Datar is the answer for that case — an
optional in-memory aggregate analytics engine built into Flagr. It tallies
evaluation counts by flag, variant, segment, and hour, then exposes the
results through two REST endpoints. No external pipeline, no Kafka consumer,
no separate analytics stack — just one more entry in `FLAGR_RECORDER_TYPE`.

Datar is an optional in-memory aggregate analytics engine built into Flagr.
It tallies evaluation counts by flag, variant, segment, and hour, then
exposes the results through two REST endpoints — no external pipeline, no
Kafka consumer, no separate analytics stack required.

## When to use

Datar exists in the gap between "no analytics" and "full pipeline." If you
already run Kafka and a warehouse, you don't need it — your streaming
recorders give you richer data. Datar is for the team that wants a quick
dashboard without standing up infrastructure: basic evaluation counts broken
down by variant and segment, queryable through a REST endpoint, persisted in
a single table.

Prometheus covers rate-based metrics and variant-level time-series well, but it
cannot index by `segment_id` due to high cardinality. Use Datar when you need:

- **Segment breakdowns** — how many evaluations each segment received
- **Historic totals** — cumulative counts (not just rates) over days or weeks
- **Per-flag dashboards** — a simple summary view across all flags without a
  separate analytics stack

For **client-reported impressions** and warehouse-style A/B analysis, use
[Exposure Logging](flagr_exposure.md) and
[Data Recorders & A/B Analysis](flagr_eval_exposure_pipeline.md) (Kafka,
Kinesis, or Pub/Sub). Datar does not ingest exposure rows.

## Enabling

Requires `FLAGR_RECORDER_ENABLED=true`. With the master switch on, list
`datar` in `FLAGR_RECORDER_TYPE` to activate the in-memory aggregator:

```bash
export FLAGR_RECORDER_ENABLED=true
export FLAGR_RECORDER_TYPE=kafka,datar
export FLAGR_RECORDER_DATAR_FLUSH_INTERVAL=60s    # default
```

The `datar_hourly_events` table is created automatically by AutoMigrate on
startup. No schema migration is needed.

## Recording

Datar recording is gated on three conditions:

1. `FLAGR_RECORDER_ENABLED=true` (master switch)
2. `datar` listed in `FLAGR_RECORDER_TYPE`
3. The per-flag toggle `dataRecordsEnabled: true`
   (configurable via `PUT /api/v1/flags/{id}`)

This means you can selectively enable recording per flag, even when Datar is
globally enabled.

> **Note:** After creating or updating a flag, wait at least one eval cache
> refresh cycle (~3s by default) before sending evaluations. The eval cache
> needs to pick up the new flag's configuration, or evaluations will return
> "not found" and won't reach the Datar recorder.

## Endpoints

### GET /api/v1/datar/summary

Returns flags with aggregate totals over a time window. **Only flags that have
actual evaluation traffic in the window appear** — zero-traffic flags are
excluded.

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `from` | RFC 3339 | 7 days ago | Start of time window |
| `to` | RFC 3339 | now | End of time window |
| `limit` | int | 100 | Max results |
| `offset` | int | 0 | Result offset |

Response:

```json
{
  "flags": [
    {
      "flagID": 1,
      "flagKey": "my-feature",
      "enabled": true,
      "description": "Controls feature X",
      "totalEvalCount": 45283,
      "lastEvaluatedAt": "2026-05-22T14:30:00Z"
    }
  ]
}
```

### GET /api/v1/datar/flags/{flagID}/summary

Detailed breakdown for a single flag. Returns traffic grouped by variant,
segment, and day — all three arrays sorted (descending by count for
variant/segment, ascending by date for day).

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `from` | RFC 3339 | 7 days ago | Start of time window |
| `to` | RFC 3339 | now | End of time window |

Response:

```json
{
  "flagID": 1,
  "trafficByVariant": [
    { "variantID": 1, "count": 30188 },
    { "variantID": 2, "count": 15095 }
  ],
  "trafficBySegment": [
    { "segmentID": 10, "count": 30188 },
    { "segmentID": 20, "count": 15095 }
  ],
  "trafficByDay": [
    { "date": "2026-05-21", "count": 22100 },
    { "date": "2026-05-22", "count": 23183 }
  ]
}
```

> **Note:** `trafficBySegment` **excludes** rows with `segment_id = 0` (the
> no-segment-match bucket). Evaluations that matched no segment are still
> counted in `trafficByVariant` and `trafficByDay`, but do not appear in the
> per-segment breakdown.

## Data model

Datar trades precision for simplicity: it counts evaluations, nothing more.
There are no per-user rows, no unique-entity counting, no event payloads. Each
evaluation bumps an in-memory counter keyed by `(flag, variant, segment, hour)`,
and a background goroutine flushes those counters to one table periodically.
The trade-off is that you lose entity-level detail (you can't ask "which users
saw this variant") but you gain a tiny, fast, zero-dependency store that can
run alongside evaluation without measurable cost.

Counts are bucketed by hour using `time.Now().Truncate(time.Hour)`. Each row
in the `datar_hourly_events` table represents one unique combination of:

- `flag_id` — the evaluated flag
- `bucket_hour` — the truncated hour timestamp
- `variant_id` — the matched variant
- `segment_id` — the matched segment (`0` if no segment matched)

A unique composite index on `(flag_id, bucket_hour, variant_id, segment_id)`
ensures additive UPSERTs work correctly across concurrent instances.

## Resource usage

The hot path uses `sync.Map` with atomic increments (zero allocations on the
existing-key path) and one batch transaction per flush interval.

> **Note:** The numbers below are indicative, not benchmark-verified — the
> repo currently has no Datar-specific benchmark. Measure on your own hardware
> before sizing capacity.

- **CPU**: ~87ns per evaluation on the hot path (existing key), ~98ns for new
  keys; zero allocations.
- **RAM**: ~210 bytes per active (flag, variant, segment) tuple; ~2.1 MB for
  10K keys.
- **DB writes**: One batch transaction every flush interval (configurable,
  default 60s).
- **Table growth**: ~2.4K rows/month per 100 flags (hourly buckets, no
  retention).

## Limitations

Datar's simplicity comes from what it *doesn't* do. Knowing the boundaries
tells you when to reach for a streaming recorder instead:

- **Crash loss** — data is in-memory until the periodic flush. If the process
  crashes, up to one flush interval of aggregate data is lost (acceptable for
  dashboard analytics).
- **No retention policy** — the table grows unbounded. Deploy a cron job or
  retention policy if needed.
- **No unique entity counting** — each evaluation is counted once regardless
  of entity identity (no HyperLogLog or similar).
- **Evaluations only** — rows with `recordSource: exposure` from
  `POST /exposures` are not counted. Use a streaming recorder (Kafka, Kinesis,
  or Pub/Sub) and
  [Data Recorders & A/B Analysis](flagr_eval_exposure_pipeline.md) for
  impression-based experiments.