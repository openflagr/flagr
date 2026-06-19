# Debug Console

The Debug Console is a built-in UI tool for testing flag evaluation without leaving the browser. It appears on each flag's detail page and wraps the evaluation API in an interactive JSON editor.

![Debug Console](/images/demo_debugging_console.png)

## What it does

The Debug Console sends evaluation requests to the Flagr API and displays the full response, including debug logs that show exactly how the evaluator arrived at its result. This makes it easy to:

- **Verify segment matching** — see which segments were checked and why they matched or didn't
- **Test constraint logic** — confirm that entity context values trigger the right constraints
- **Debug distribution** — understand which variant was selected and why
- **Test before deploying** — simulate evaluations with different entity contexts before rolling out changes

## Single evaluation

The **Evaluation** panel sends a `POST /api/v1/evaluation` request with a JSON body you can edit inline:

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

- **`variantID` / `variantKey`** — which variant was assigned
- **`evalDebugLog.segmentDebugLogs`** — per-segment breakdown showing:
  - Whether each segment matched
  - Which constraints matched and which didn't
  - The rollout and distribution reasoning
- **`evalContext`** — the full context that was evaluated

The console also renders a summary view showing the matched variant and a segment-by-segment trace.

## Batch evaluation

The **Batch Evaluation** panel sends a `POST /api/v1/evaluation/batch` request, letting you evaluate multiple entities against one or more flags in a single call:

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

This is useful for verifying that different entity contexts route to the correct variants across your segments.

## Tips

- **Always set `enableDebug: true`** — without it, the response omits the debug logs that make the console useful.
- **Use realistic entity contexts** — include the same fields your application sends (e.g. `state`, `age`, `tier`) to test constraint matching accurately.
- **Flag key or flag ID** — the console pre-fills both from the current flag page. You can use either in the request.
