# Debug Console

Evaluations are deterministic, but the *reasoning* behind them — which segment
matched, which constraints held, where the hash landed — is invisible from a
raw `POST /evaluation` call. When a user reports "I'm in California but I'm
not getting the green button," you need to see the trace, not just the answer.
The Debug Console is that trace, built into each flag's detail page. It wraps
the evaluation API in an interactive JSON editor so you can replay an
evaluation with a specific entity context and inspect exactly how Flagr
arrived at its verdict — without writing a throwaway script or hitting the
API with curl.

The console exercises **`POST /evaluation`** and
**`POST /evaluation/batch`** only (not GET). For browser-friendly **`GET /evaluation?json=…`**, see [Use Cases — GET evaluation](flagr_use_cases.md#get-evaluation-browser-friendly). For production **impression** logging after
render, see [Exposure Logging](flagr_exposure.md). For segment, constraint, and
rollout concepts, see the [Overview](flagr_overview.md).

![Debug Console](/images/demo_debugging_console.png)

## What it does

The Debug Console sends evaluation requests to the Flagr API and displays the
full response, including debug logs that show how the evaluator arrived at its
result. Each log entry is keyed to a `segmentID` and carries a free-text `msg`
that narrates the decision: "matched all constraints" or "constraint not
match," followed by the bucket and distribution reasoning. Read the trace top
to bottom to see the segments Flagr considered, in order, and where it
stopped. Use it to:

- **Verify segment matching** — see which segments were checked and whether
  each matched.
- **Test constraint logic** — confirm entity-context values trigger the right
  constraints.
- **Debug distribution** — understand which variant was selected and why.
- **Test before deploying** — simulate evaluations with different entity
  contexts before rolling out changes.

## Single evaluation

The **Evaluation** panel sends a `POST /api/v1/evaluation` request with a JSON
body you can edit inline:

```json
{
  "entityID": "a1234",
  "entityType": "report",
  "entityContext": { "state": "CA" },
  "enableDebug": true,
  "flagID": 1
}
```

The response includes:

- **`variantID` / `variantKey`** — which variant was assigned.
- **`evalDebugLog.segmentDebugLogs`** — a per-segment trace. Each entry has:
  - `segmentID` — the segment evaluated.
  - `msg` — free-text reasoning: whether all constraints matched, and the
    rollout/distribution bucket reasoning (bucket number, distribution array,
    rollout percent).
- **`evalContext`** — the full context that was evaluated.

> **Note:** `segmentDebugLogs` is a flat list of `{ segmentID, msg }` entries,
> not a structured breakdown. Constraint results are reported as one boolean
> expression per segment — on mismatch the whole expression is dumped; there is
> no per-constraint matched/didn't split.

The console also renders a summary view showing the matched variant and a
segment-by-segment trace.

## Batch evaluation

The **Batch Evaluation** panel sends a `POST /api/v1/evaluation/batch` request,
letting you evaluate multiple entities against one or more flags in a single
call:

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

This is useful for verifying that different entity contexts route to the
correct variants across your segments.

## Tips

- **Always set `enableDebug: true`** — without it, the response omits the debug
  logs that make the console useful. (The server must also have
  `FLAGR_EVAL_DEBUG_ENABLED` on, which defaults to `true`.)
- **Use realistic entity contexts** — include the same fields your app sends
  (e.g. `state`, `age`, `tier`) to test constraint matching accurately.
- **Flag key or flag ID** — the console pre-fills both from the current flag
  page. You can use either `flagID` or `flagKey` in the request; both resolve to
  the same flag.