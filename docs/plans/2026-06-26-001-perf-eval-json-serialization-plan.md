# perf: Eval JSON serialization improvements

Reduce CPU and allocations on the **HTTP JSON layer** for flag evaluation. Engine-side wins landed in #719; this plan targets **request decode**, **response encode**, and **data-recorder JSON** without changing evaluation semantics (same buckets, same constraint results).

**Branch:** `perf/eval-json-serialization` (from `origin/main`, includes #719).

## Current behavior (baseline)

| Stage | Where | What happens |
|-------|--------|----------------|
| Request decode | `swagger_gen/restapi/configure_flagr.go` | `json.Decoder` + `UseNumber()` → `models.EvalContext` / batch models |
| Dynamic fields | `swagger_gen/models/*.go` | `entityContext`, `variantAttachment` as `any` → reflection + nested maps |
| Eval logic | `pkg/handler/eval.go` | `EntityContext` must be `map[string]any` for constraints |
| Response | `post_evaluation_responses.go` | Full `*models.EvalResult` JSON, including **echoed** `evalContext` (same size as request context) |
| Recorders | `pkg/handler/data_record_frame.go` | `evalResult.MarshalBinary()` (JSON via jsonutils) then **`json.Marshal(drf)`** again in `Output()` |

**Not in scope for v1:** Replacing all go-swagger models with easyjson; changing OpenAPI major version; vendor-only `replace` hacks.

### go-swagger / `make gen` (2026-06-26)

go-swagger **has no** `--json-lib` / sonic / easyjson switch. Model codegen only sets `--struct-tags=json`. HTTP codec is **not** generated into models — it lives in **`swagger/templates/server/configureapi.gotmpl`** (override via `make swagger --template-dir swagger/templates --allow-template-override`). Runtime: **`FLAGR_EVAL_JSON_CODEC`** → `pkg/jsoncodec`. See `swagger/README.md`.

## Goals

1. Measurable before/after on **JSON-heavy** paths (large `entityContext`, batch eval).
2. **Simple, incremental** commits; each option can ship alone.
3. Preserve default API contract unless an option explicitly opts into behavior change.

## Non-goals (deferred)

- CRUD list/get flag JSON (low QPS vs eval).
- `jsoniter` / `sonic` for entire API surface.
- Skipping go-swagger `Validate` on eval without explicit safety review.

---

## Options (pick one track or combine A + B2)

### Option A — Recorder single-pass JSON (recommended first)

**Idea:** `DataRecordFrame.Output()` should not marshal the frame twice. Today `MarshalJSON` builds inner payload bytes, then `Output()` calls `json.Marshal(drf)` on the whole struct again.

**Approach (simple):**
- Add `OutputBytes() ([]byte, error)` (or fix `Output()`) that returns the **final** wire bytes from one composition step.
- For `payload_raw_json`: one `json.Marshal` of `rawPayload{Payload: raw}` where `raw` comes from **one** `json.Marshal(evalResult)` (or `MarshalBinary` once).
- For encrypted/string modes: one outer marshal after inner payload is built.

**API / contract:** None (Kafka/PubSub payload shape unchanged).

**Risk:** Low. Covered by existing `data_record_frame` / recorder tests.

**Files:** `pkg/handler/data_record_frame.go`, `pkg/handler/data_record_frame_test.go`, callers of `Output()` if any.

---

### Option B — Slimmer eval **HTTP** response

**Idea:** Responses repeat the full request `evalContext` (including large `entityContext`). Many clients only need variant + flag ids; echo is redundant and expensive to encode.

| Variant | Behavior | API impact |
|---------|----------|------------|
| **B1** | Env `FLAGR_EVAL_RESPONSE_MINIMAL=true` (default **false**): omit `evalContext.entityContext` on response; keep other `evalContext` fields | Opt-in; safe default |
| **B2** | When `enableDebug == false`, omit `entityContext` on response (keep rest of `evalContext`) | **Behavior change** for clients that read echoed context without debug |
| **B3** | OpenAPI: document that response never includes `entityContext`; strip always | Breaking; needs release note |

**Recommended for “simple + measurable”:** **B1** first (operator choice), document **B2** as a follow-up RFC if product agrees.

**Approach (implementation sketch, no new swagger types in v1):**
- Helper in `pkg/handler` e.g. `evalResultForHTTPResponse(r *models.EvalResult, minimal bool) *models.EvalResult` — shallow copy, nil or clear `EntityContext` on nested `EvalContext` when minimal.
- Wire in `PostEvaluation` / batch result assembly **before** returning responder (not in `BlankResult`, so recorders still get full context if needed today).

**Risk:** Medium for B2/B3 (client compatibility). Low for B1 default-off.

**Files:** `pkg/handler/eval.go`, `pkg/config/env.go` (B1), tests, `docs/flagr_env.md`.

---

### Option C — Decode: normalize `entityContext` once

**Idea:** After bind, convert `EntityContext` `any` → `map[string]any` once (and optionally normalize `json.Number` → `float64` for leaf numbers) so eval + debug paths don’t re-walk ambiguous types.

**Approach:**
- `normalizeEvalContext(ec *models.EvalContext) error` in `pkg/handler` called at start of `PostEvaluation` / batch entity loop.
- Reject or clear non-object context with same semantics as today (constraints skipped if not a map).

**API:** None if semantics unchanged.

**Win:** Smaller than B on encode; helps consistency and may shave reflection in edge debug paths.

**Risk:** Low if normalization matches current `UseNumber` behavior (tests with nested maps + json.Number).

**Files:** `pkg/handler/eval.go`, `eval_test.go`.

---

### Option D — Faster serializer at HTTP boundary only

**Idea:** Swap `JSONProducer` / `JSONConsumer` in `configure_flagr.go` for eval routes only, or globally for `EvalResult` encoding.

**Verdict for v1:** **Defer.** go-openapi routes share one consumer/producer; per-route serializers need middleware or custom responders. Higher complexity; bench Option A+B first.

---

## Recommended phasing

| Phase | Units | Ships |
|-------|--------|--------|
| **P0** | Benchmark harness (below) | Baseline numbers in PR description |
| **P1** | **Option A** | Recorder alloc/latency ↓, no API change |
| **P2** | **Option B1** (env-gated minimal response) | HTTP encode ↓ when enabled |
| **P3** | **Option C** (optional) | Decode/normalize |
| **P4** | Re-bench + integration notes | Evidence for merge |

Do **not** block P1 on OpenAPI changes.

---

## Benchmark strategy

Keep three layers so we know **where** time goes (compare to existing `BenchmarkEvalFlag*` from #719).

### Layer 1 — In-process JSON only (new, fast, CI-friendly)

**File:** `pkg/handler/eval_json_bench_test.go` (new)

| Benchmark | Measures |
|-----------|----------|
| `BenchmarkJSONDecodeEvalContext_Small` | Decoder + `UseNumber`, tiny map (1–2 keys) |
| `BenchmarkJSONDecodeEvalContext_Large` | Nested map ~50 keys + array (fixture `[]byte` in testdata) |
| `BenchmarkJSONEncodeEvalResult_Full` | Full `EvalResult` with large echoed `EvalContext` |
| `BenchmarkJSONEncodeEvalResult_Minimal` | Same result after strip `entityContext` (Option B) |
| `BenchmarkDataRecordFrame_Output` | `Output()` / `OutputBytes()` before vs after Option A |

**Setup:** Reuse the same `JSONConsumer` / `JSONProducer` funcs as `configure_flagr.go` (copy into test helper or export `testJSONCodec()` in `pkg/handler` test file) so numbers match production.

**Compare options:** Run with `benchstat` on `main` vs branch; report ns/op, B/op, allocs/op.

### Layer 2 — Handler without network (existing + one addition)

| Benchmark | Notes |
|-----------|--------|
| `BenchmarkEvalFlag_*` | Engine only (already on main) — **hold constant** to prove JSON work doesn’t regress eval |
| `BenchmarkEvalFlag_LargeEntityContext` (new) | Same as `AllMatch` but fixture context 10× larger — isolates **in-process** eval cost of big maps (not HTTP) |

### Layer 3 — HTTP integration (manual / CI optional)

| Command | Measures |
|---------|----------|
| `make bench-integration` with env | Full stack: marshal client body + server decode + eval + encode |
| Extend `integration_benchmark_test.go` | Add `BenchmarkEvalLargeContext` POST body with ~4–16 KB `entityContext` |

**Protocol for comparing options:**
1. Record baseline on `origin/main` (or first commit on branch): Layer 1 + one Layer 3 bench.
2. After each option (A, B1): re-run Layer 1 relevant benches + `go test ./pkg/...`.
3. PR table: **before → after** for encode/decode/recorder only; note Layer 3 if run.

**Success criteria (indicative, arm64 dev machine):**
- Option A: `BenchmarkDataRecordFrame_Output` **−1 alloc/op** or **≥15%** ns/op (target; validate empirically).
- Option B1 (minimal on): `BenchmarkJSONEncodeEvalResult_Minimal` **≥25%** smaller B/op vs `Full` on large-context fixture.

---

## Test scenarios (by unit)

### U1 — Benchmark harness (P0)

- Layer 1 benchmarks compile and run with `-count=5`.
- Large context fixture checked into `pkg/handler/testdata/eval_context_large.json` (generated once, stable).

### U2 — Option A recorder (P1)

- Happy: raw JSON mode output byte-equal to previous (golden or structural compare).
- Happy: encrypted mode still valid base64 wrapper.
- Regression: `exposure_test` / kafka recorder tests that unmarshal payload still pass.

### U3 — Option B1 minimal response (P2)

- `FLAGR_EVAL_RESPONSE_MINIMAL=false` (default): response JSON unchanged vs today.
- `true`: `entityContext` absent or null on response; `variantKey` / `flagID` unchanged.
- Debug: with `enableDebug: true`, define policy explicitly (recommend: still omit context unless debug needs it — document).

### U4 — Option C normalize (P3)

- Nested `entityContext` + `json.Number` in JSON decodes to map usable by `evalSegment`.
- Invalid context types behave as today (no match / debug msg).

---

## Dependencies

- #719 merged on `main` (eval hot path) — **done** on branch base.
- `conditions` v0.2.6 — unchanged; no `go.mod` `replace`.

## Open decisions (defaults if no reply)

| Question | Default |
|----------|---------|
| First ship A only or A+B1? | **A then B1** |
| B1 env name | `FLAGR_EVAL_RESPONSE_MINIMAL` |
| Strip context when `enableDebug`? | **No** (only when minimal flag true); debug keeps full echo for console |

## PR checklist

- [ ] Layer 1 bench numbers in PR description
- [ ] No `replace` in `go.mod`
- [ ] `make test` green
- [ ] If B1: `docs/flagr_env.md` + release note snippet
- [ ] Swagger regen only if OpenAPI changed (B1/B2 without spec change = no regen)

## References

- JSON boundary: `swagger_gen/restapi/configure_flagr.go`
- Echo context: `BlankResult` in `pkg/handler/eval.go`
- Prior engine perf: PR #719