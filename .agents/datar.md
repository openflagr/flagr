# Datar — Embedded Aggregate Analytics for Flagr

Datar is an optional in-memory analytics engine for Flagr. It counts evaluation events by `(flag_id, variant_id, segment_id, hour)` and periodically flushes the counts to a shared `datar_hourly_events` table — no external pipeline, no Kafka consumer, no separate analytics stack required.

## Motivation

A simple, zero-dependency solution for trivial aggregate analytics. Built directly into Flagr, Datar fills the gap between Prometheus metrics (rate-based, no segment index) and a full external analytics pipeline.

## How to enable

```bash
FLAGR_DATAR_ENABLED=true
FLAGR_DATAR_FLUSH_INTERVAL=60s   # default; use shorter values in dev
```

The table is always created by AutoMigrate. The kill switch only controls whether the in-memory aggregator runs.

## What you get

- **`GET /api/v1/datar/summary`** — flags with traffic, sorted by count desc, top 100
- **`GET /api/v1/datar/flags/{flagID}/summary`** — variant split, segment breakdown, daily time-series

Only flags with actual evaluation events appear in the summary (INNER JOIN, no COALESCE zeros).

## Cost

- ~87ns per evaluation on the hot path (existing key), ~98ns for new keys; zero allocations
- ~210 bytes per active (flag, variant, segment) tuple; ~2.1 MB for 10K keys
- One DB batch transaction every flush interval
- Table grows ~2.4K rows/month per 100 flags (no retention)

## Architecture

```
logEvalResult → Engine.Record() → in-memory map[FlushKey]int32
                                       ↓ every FLAGR_DATAR_FLUSH_INTERVAL
                                   GORM Clauses.Create → datar_hourly_events
                                       ↓
                                   GET /api/v1/datar/*
```

- **In-memory only.** No WAL, no files, no recovery. On crash, ≤1 flush interval of data is lost.
- **Multi-instance.** Each instance flushes independently. Additive UPSERT sums correctly.
- **No coordination.** No sequence numbers, no flush-log table.
- **Eval cache dependency.** Flags must be loaded by the eval cache (~3s refresh) before evaluations reach Datar.

## Files

```
pkg/datar/
├── engine.go              — Engine: buffer, flush loop, queries, nil-safe methods
├── engine_test.go         — 27 tests, 95%+ coverage

pkg/handler/
├── datar.go               — GetDatar singleton, ResetDatar (test helper)
├── datar_handler.go       — go-swagger handler implementations
├── datar_test.go          — 6 endpoint tests

pkg/entity/datar.go        — HourlyEvent GORM model
pkg/entity/db.go           — AutoMigrateTables includes HourlyEvent
pkg/config/env.go          — FLAGR_DATAR_ENABLED, FLAGR_DATAR_FLUSH_INTERVAL
pkg/handler/eval.go        — logEvalResult: GetDatar().Record(r)
pkg/handler/handler.go     — setupDatar: handler assignment + shutdown hook
swagger/                   — datar_summary.yaml, datar_flag_summary.yaml

docs/flagr_datar.md        — user-facing docs
integration_tests/
├── docker-compose.yml     — Datar enabled on all DB services
├── test.sh                — step_13_test_datar (real-data assertions)
```
