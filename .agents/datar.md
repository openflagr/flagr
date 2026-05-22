# Datar — Embedded Aggregate Analytics for Flagr

Datar is an optional in-memory analytics engine for Flagr. It counts evaluation events by `(flag_id, variant_id, segment_id, hour)` and periodically flushes the counts to a shared `datar_hourly_events` table — no external pipeline, no Kafka consumer, no separate analytics stack required.

## Motivation

A simple, zero-dependency solution for trivial aggregate analytics. Built directly into Flagr, Datar fills the gap between Prometheus metrics (rate-based, no segment index) and a full external analytics pipeline.

## How to enable

```bash
FLAGR_DATAR_ENABLED=true
```

The table is always created by AutoMigrate. The kill switch only controls whether the in-memory aggregator runs.

## What you get

- **`/api/v1/datar/summary`** — all flags with 7-day traffic totals
- **`/api/v1/datar/flags/{flagID}/summary`** — variant split, segment breakdown, daily time-series

## Cost

- ~87ns per evaluation on the hot path (existing key), ~98ns for new keys; zero allocations
- ~210 bytes per active (flag, variant, segment) tuple; ~2.1 MB for 10K keys
- One DB batch transaction every flush interval
- Table grows ~2.4K rows/month per 100 flags (no retention)


## Architecture

```
logEvalResult → Aggregator.Record() → in-memory map[FlushKey]int32
                                          ↓ every 60s
                                      UPSERT batch → datar_hourly_events
                                          ↓
                                      GET /api/v1/datar/*
```

- **In-memory only.** No WAL, no files, no recovery. On crash, ≤60s of aggregate data is lost.
- **Multi-instance.** Each instance flushes independently. Additive UPSERT sums correctly.
- **No coordination.** No sequence numbers, no flush-log table.

## Files

```
pkg/datar/
├── aggregator.go       — Aggregator (Record, SnapshotAndReset)
├── models.go           — FlushKey DTO
├── store.go            — FlushAggregates (raw SQL upsert), query helpers
├── aggregator_test.go  — 10 tests
├── store_test.go       — 10 tests

pkg/handler/
├── datar.go            — Datar singleton lifecycle (GetDatar, Shutdown, flush loop)
├── datar_handler.go    — go-swagger handler implementations
├── datar_test.go       — 4 endpoint tests

pkg/entity/datar.go     — HourlyEvent GORM model
pkg/entity/db.go        — AutoMigrateTables includes HourlyEvent
pkg/config/env.go       — FLAGR_DATAR_ENABLED, FLAGR_DATAR_FLUSH_INTERVAL
pkg/handler/eval.go     — logEvalResult: GetDatar().Record(r)
pkg/handler/handler.go  — setupDatar: handler assignment + shutdown hook
swagger/                — datar_summary.yaml, datar_flag_summary.yaml

integration_tests/
├── docker-compose.yml  — Datar enabled on flagr_with_sqlite
├── test.sh             — step_13_test_datar (real-data assertions)
```
