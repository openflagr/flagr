# Datar Aggregate Analytics

Datar is an optional in-memory aggregate analytics engine built into Flagr. It tallies evaluation counts by flag, variant, segment, and hour, then exposes the results through two REST endpoints — no external pipeline, no Kafka consumer, no separate analytics stack required.

## When to use

A simple, zero-dependency solution for trivial aggregate analytics. Use it when you need basic evaluation counts broken down by variant and segment without setting up an external pipeline.

Prometheus covers rate-based metrics and variant-level time-series well, but it cannot index by `segment_id` due to high cardinality. Use Datar when you need:

- **Segment breakdowns** — how many evaluations each segment received
- **Historic totals** — cumulative counts (not just rates) over days or weeks
- **Per-flag dashboards** — a simple summary view across all flags without setting up a separate analytics stack

## Enabling

```bash
export FLAGR_DATAR_ENABLED=true
export FLAGR_DATAR_FLUSH_INTERVAL=60s    # default
```

The `datar_hourly_events` table is created automatically by AutoMigrate on startup. No schema migration is needed.

## Recording

Datar recording is gated on two conditions:

1. The server-level flag `FLAGR_DATAR_ENABLED=true`
2. The per-flag toggle `dataRecordsEnabled: true` (configurable via PUT /api/v1/flags/{id})

This means you can selectively enable recording per flag, even when Datar is globally enabled.

## Endpoints

### GET /api/v1/datar/summary

Returns all flags with aggregate totals over a time window.

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

Flags with no evaluation data in the time window appear with `totalEvalCount: 0`.

### GET /api/v1/datar/flags/{flagID}/summary

Detailed breakdown for a single flag.

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `from` | RFC 3339 | 7 days ago | Start of time window |
| `to` | RFC 3339 | now | End of time window |

Response:

```json
{
  "flagID": 1,
  "trafficByVariant": {
    "1": 30188,
    "2": 15095
  },
  "trafficBySegment": [
    { "segmentID": 10, "description": "US users", "evalCount": 30188 },
    { "segmentID": 20, "description": "EU users", "evalCount": 15095 }
  ],
  "trafficByDay": [
    { "date": "2026-05-21", "count": 22100 },
    { "date": "2026-05-22", "count": 23183 }
  ]
}
```

## Data model

Counts are bucketed by hour using `time.Now().Truncate(time.Hour)`. Each row in the `datar_hourly_events` table represents one unique combination of:

- `flag_id` — the evaluated flag
- `variant_id` — the matched variant
- `segment_id` — the matched segment (0 if no segment matched)
- `bucket_hour` — the truncated hour timestamp

A unique index on `(flag_id, variant_id, segment_id, bucket_hour)` ensures additive UPSERTs work correctly across concurrent instances.

## Resource usage

- **CPU**: ~87ns per evaluation on the hot path (existing key), ~98ns for new keys; zero allocations
- **RAM**: ~210 bytes per active (flag, variant, segment) tuple; ~2.1 MB for 10K keys

- **DB writes**: One batch transaction every flush interval
- **Table growth**: ~2.4K rows/month per 100 flags (hourly buckets, no retention)

## Limitations

- Data is in-memory until the periodic flush. If the process crashes, up to one flush interval of aggregate data is lost (acceptable for dashboard analytics).
- No data retention policy is built in — the table grows unbounded. Deploy a cron job or retention policy if needed.
- No unique entity counting (HyperLogLog or similar). Each evaluation is counted once regardless of the entity identity.
