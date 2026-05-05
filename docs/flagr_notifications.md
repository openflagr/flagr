# Notifications

Flagr provides an integrated notification system that allows you to monitor changes and updates to your operational resources in real-time. You can configure Flagr to send HTTP `POST` webhooks whenever a flag is created, updated, deleted, or restored.

## Tracked Operations

Flagr monitors changes to **flags** and their related configuration. All notifications have `EntityType: "flag"` in the payload.

The following operations trigger notifications:

| Operation | Description |
|-----------|-------------|
| `create` | A new flag is created |
| `update` | Any change to a flag's metadata, enabled state, or any of its associated entities (segments, variants, constraints, distributions, tags) |
| `delete` | A flag is soft-deleted |
| `restore` | A soft-deleted flag is restored |

**Note**: Operations such as adding/removing tags, updating segment rollout percentages, modifying constraints, or changing variant attachments all trigger an `update` notification for the parent flag. Enabling or disabling a flag is also considered an update.

## Configuration

To enable notifications, set the following environment variables:

- `FLAGR_NOTIFICATION_WEBHOOK_ENABLED=true` (Default: `false`) — Enable webhook notifications.
- `FLAGR_NOTIFICATION_WEBHOOK_URL=https://api.your-org.com/webhooks/flagr` — HTTP destination endpoint for POST requests.
- `FLAGR_NOTIFICATION_WEBHOOK_HEADERS=Authorization: Bearer secret-token, X-Custom-Header: value` — (Optional) Custom comma-separated HTTP headers, often utilized for securing your webhook receiver with an API token.
- `FLAGR_NOTIFICATION_TIMEOUT=10s` (Default: `10s`) — Configures the timeout window for dialing the webhook endpoint.
- `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true` (Default: `false`) — When enabled, Flagr will embed the precise visual JSON diff of the modified flag within the notification payload.
- `FLAGR_NOTIFICATION_MAX_RETRIES=3` (Default: `3`) — Maximum number of retry attempts for transient HTTP failures (5xx errors). Set to `0` to disable retries.
- `FLAGR_NOTIFICATION_RETRY_BASE=1s` (Default: `1s`) — Base delay for exponential backoff between retries.
- `FLAGR_NOTIFICATION_RETRY_MAX=10s` (Default: `10s`) — Maximum delay between retries.

### Concurrency & Observability

- Notifications are sent asynchronously with a default concurrency limit of 100 to prevent resource exhaustion under load.
- Metric `notification.sent` is emitted when statsd is enabled, tagged with `provider`, `operation`, `entity_type`, and `status` (`success`/`failure`).

### Important Notes

- **Asynchronous delivery**: Notifications are sent in background goroutines. Failures are logged but **do not affect the API response**.
- **Startup validation**: Flagr validates the notification configuration at startup and logs a warning if `FLAGR_NOTIFICATION_WEBHOOK_URL` is not set while webhooks are enabled.
- **Silent fallback**: If webhooks are enabled but the URL is missing, notifications will be silently dropped. A warning is logged at startup to help diagnose misconfiguration.

## Webhook Payload Format

The target endpoint receives a structured JSON payload:

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
| `component_type` | string | What part of the flag changed: `flag`, `segment`, `variant`, `constraint`, `distribution`, or `tag` |
| `component_id` | uint | Database ID of the changed component |
| `component_key` | string | Key/name of the changed component (e.g. variant key, tag value) |
| `pre_value` | string | Previous flag snapshot JSON (only if `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true`)
| `post_value` | string | Current flag snapshot JSON (only if `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true`)
| `diff` | string | Unified diff between previous and current (only if `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true`)
| `user` | string | Identity of the user who made the change |
| `timestamp` | string | UTC timestamp of the change in RFC 3339 format |
