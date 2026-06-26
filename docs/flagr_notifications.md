# Notifications

Flag configuration changes are high-leverage events: a rollout percentage
creep, a disabled flag, a swapped distribution can change the experience for
every user in seconds. In a team, you want those changes to *announce
themselves* — to trigger an audit trail, a cache invalidation, a Slack alert,
or a sync to an external config store. Flagr's notification system does this
by sending an HTTP `POST` webhook whenever a flag is created, updated,
deleted, or restored. The webhook carries what changed, on which flag, by
whom, and (optionally) a JSON diff — so the receiver can decide whether to
act, log, or alarm. Delivery is asynchronous and never blocks the API
response that made the change.

## Tracked operations

Notifications fire on mutations to a flag and its related entities (segments,
variants, constraints, distributions, tags). The `operation` field identifies
what happened:

| Operation | Fires when |
|-----------|------------|
| `create` | A flag is created, **or** a segment, variant, constraint, or tag is created on a flag |
| `update` | A flag's metadata, enabled state, segments, variants, constraints, distributions, or tags are modified |
| `delete` | A flag is soft-deleted, **or** a segment, variant, constraint, or tag is deleted from a flag |
| `restore` | A soft-deleted flag is restored |

The `component_type` field identifies **what** changed (`flag`, `segment`,
`variant`, `constraint`, `distribution`, or `tag`).

> **Note:** Enabling or disabling a flag is an `update` with
> `component_type: "flag"`. Reordering segments is an `update` with
> `component_type: "segment"`. Changing distributions is an `update` with
> `component_type: "distribution"`.

## Configuration

Notifications are off by default — a Flagr instance with no webhook
destination stays silent. Turning it on is a single env var; the rest tunes
the delivery. Retries exist because webhook endpoints *do* go down, and a
flag change you lose is one you can't audit.

| Variable | Default | Description |
|----------|---------|-------------|
| `FLAGR_NOTIFICATION_WEBHOOK_ENABLED` | `false` | Enable webhook notifications |
| `FLAGR_NOTIFICATION_WEBHOOK_URL` | — | HTTP destination endpoint for POST requests |
| `FLAGR_NOTIFICATION_WEBHOOK_HEADERS` | — | Comma-separated custom HTTP headers (e.g. `Authorization: Bearer token, X-Custom: value`) |
| `FLAGR_NOTIFICATION_TIMEOUT` | `10s` | Timeout for dialing the webhook endpoint |
| `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED` | `false` | Embed the JSON diff of the modified flag in the payload |
| `FLAGR_NOTIFICATION_MAX_RETRIES` | `3` | Max retry attempts for transient HTTP failures (5xx). Set `0` to disable retries |
| `FLAGR_NOTIFICATION_RETRY_BASE` | `1s` | Base delay for exponential backoff between retries |
| `FLAGR_NOTIFICATION_RETRY_MAX` | `10s` | Maximum delay between retries |

### Delivery behavior

Flagr treats notification delivery as best-effort: the API call that changed
the flag returns immediately, and the webhook fires in the background. This is
a deliberate trade-off — you don't want a flaky webhook receiver to stall a
flag update — but it means your receiver must be idempotent and observably
reliable on its own.

- **Asynchronous** — notifications are sent in background goroutines. Failures
  are logged but **do not affect the API response**.
- **Concurrency** — a hardcoded semaphore limits delivery to **100** concurrent
  notifications to prevent resource exhaustion under load.
- **Retry** — on HTTP `5xx`, the webhook retries up to `MAX_RETRIES` times with
  exponential backoff (delay doubling, capped at `RETRY_MAX`, with jitter).
  `4xx` responses are treated as final (no retry).
- **Silent fallback** — if webhooks are enabled but `FLAGR_NOTIFICATION_WEBHOOK_URL`
  is missing, notifications are dropped. Flagr logs a warning at startup.

### Observability

The Statsd counter `notification.sent` is emitted for every delivery attempt,
tagged with:

- `provider` — the notifier (e.g. `webhook`)
- `operation` — `create`, `update`, `delete`, or `restore`
- `status` — `success` or `failure`

> **Note:** Flagr validates the notification configuration at startup and logs
> a warning if `FLAGR_NOTIFICATION_WEBHOOK_URL` is not set while webhooks are
> enabled.

## Webhook payload

The payload is intentionally flat and self-describing: one JSON object per
event, with the `operation` and `component_type` telling the receiver what
happened and to what, and the `flag_id`/`flag_key` anchoring it to a flag your
system already knows. The optional diff fields (`pre_value`, `post_value`,
`diff`) carry the full before/after snapshot when you want rich change
records — off by default because they roughly double the payload size.

The target endpoint receives a JSON payload:

```json
{
  "operation": "update",
  "flag_id": 123,
  "flag_key": "my-feature-flag",
  "component_type": "segment",
  "component_id": 7,
  "component_key": "power-users",
  "pre_value": "...",
  "post_value": "...",
  "diff": "--- Previous\n+++ Current\n@@ ...",
  "user": "admin@example.com",
  "timestamp": "2026-04-26T18:51:03Z"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `operation` | string | `create`, `update`, `delete`, or `restore` |
| `flag_id` | uint | Database ID of the parent flag |
| `flag_key` | string | Unique key of the parent flag |
| `component_type` | string | What changed: `flag`, `segment`, `variant`, `constraint`, `distribution`, or `tag` |
| `component_id` | uint | Database ID of the changed component |
| `component_key` | string | Key/name of the changed component (e.g. variant key, tag value) |
| `pre_value` | string | Previous flag snapshot JSON (only if `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true`) |
| `post_value` | string | Current flag snapshot JSON (only if `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true`) |
| `diff` | string | Unified diff between previous and current (only if `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true`) |
| `user` | string | Identity of the user who made the change |
| `timestamp` | string | UTC timestamp of the change in RFC 3339 format |