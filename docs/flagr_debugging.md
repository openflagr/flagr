# Debug console

You evaluated an entity and got back the wrong variant — or no variant at all. The evaluation API tells you *what* happened; it does not tell you *why*. The **Debug Console**, available on every flag page, fills that gap. It sends an evaluation request with **`enableDebug: true`** and surfaces `evalDebugLog.segmentDebugLogs`, so you can replay an entity's context and walk through, segment by segment, exactly where evaluation stopped and what it decided.

The console uses **`POST /evaluation`** and **`POST /evaluation/batch`** only (not GET). For **`GET /evaluation?json=…`**, see [Use cases — GET evaluation](flagr_use_cases.md#get-evaluation-browser-friendly). It works when **`FLAGR_EVAL_DEBUG_ENABLED=true`** (default) and the request sets `enableDebug`; turn that env off in production if you do not want segment logs on eval traffic. For impressions after render, [Exposure logging](flagr_exposure.md); for concepts, [Overview](flagr_overview.md). Gating: [Behavioral contracts](contracts.md#evalcache-freshness).

![Debug Console](/images/demo_debugging_console.png)

## What you get

Each entry in `segmentDebugLogs` carries a `segmentID` and a free-text `msg` describing what the evaluator saw: which constraints matched, which did not, and — when a segment *did* match — the bucket and distribution reasoning that followed. Read the list in **segment rank order**. Evaluation walks segments from first to last and stops at the first match, so the first entry whose `msg` reports a match is your culprit; everything after it was never evaluated.

## Single evaluation

Start with a single entity when you are chasing one wrong assignment. The **Evaluation** panel sends `POST /api/v1/evaluation` with an editable body, so you can paste in the exact context your application sends and see what the server actually does with it:

```json
{
  "entityID": "a1234",
  "entityType": "report",
  "entityContext": { "state": "CA" },
  "enableDebug": true,
  "flagID": 1
}
```

The response carries three things you care about when debugging:

- **`variantID` / `variantKey`** — the variant that was assigned, or nothing if no segment matched.
- **`evalDebugLog.segmentDebugLogs`** — a per-segment trace. Each entry pairs a `segmentID` with a `msg` that reports whether all constraints matched and, when they did, the rollout and distribution reasoning: the bucket number, the distribution array, and the rollout percent.
- **`evalContext`** — the full context the server actually evaluated, echoed back so you can confirm it received what you intended to send.

> **Note:** `segmentDebugLogs` is a flat list of `{ segmentID, msg }` entries, not a structured breakdown. Constraint results are reported as one boolean expression per segment — on mismatch the whole compiled expression is dumped; there is no per-constraint matched/didn't split.

That compiled expression is the fastest path to the most common bug. When a segment's constraints fail, the debug text includes the expression the server compiled and the **`entityContext`** map it evaluated against. Compare the two: a property name typo, a missing field, or a JSON type mismatch (string `"30"` versus number `30`) jumps out immediately, because the expression names the field it expected and the map shows what you actually sent.

The console renders a summary view alongside the raw response — the matched variant up top, and a segment-by-segment trace below — so you can scan the outcome before drilling into the JSON.

## Batch evaluation

When the question is not "why did this one entity fail?" but "do my segments route the way I think they do across the board?", switch to the **Batch Evaluation** panel. It sends `POST /api/v1/evaluation/batch`, letting you evaluate multiple entities against one or more flags in a single call:

```json
{
  "entities": [
    { "entityID": "a1234", "entityType": "report", "entityContext": { "state": "CA" } },
    { "entityID": "a5678", "entityType": "report", "entityContext": { "state": "NY" } }
  ],
  "enableDebug": true,
  "flagIDs": [1]
}
```

This is the right tool for regression-checking a constraint change: line up the entity contexts that should land in each variant, run them together, and confirm every one routed where you expected. If one did not, its `segmentDebugLogs` tells you exactly which segment rejected it and why.

## Tips

- **Always set `enableDebug: true`** — without it, the response omits the debug logs that make the console useful. The server must also have `FLAGR_EVAL_DEBUG_ENABLED` on, which defaults to `true`.
- **Use realistic entity contexts** — include the same fields your app sends (e.g. `state`, `age`, `tier`) to test constraint matching accurately.
- **Flag key or flag ID** — the console pre-fills both from the current flag page. You can use either `flagID` or `flagKey` in the request; both resolve to the same flag.