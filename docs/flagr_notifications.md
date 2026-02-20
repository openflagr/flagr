# Notifications

Flagr provides an integrated notification system that allows you to monitor changes and updates to your operational resources in real-time. You can configure Flagr to automatically send notifications regarding CRUD (Create, Read, Update, Delete) operations over several distinct channels: **Email**, **Slack**, or generic **Webhooks**.

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

## Global Configuration

You must globally enable the notifications feature via environment variables.

- `FLAGR_NOTIFICATION_ENABLED=true` (Default: `false`) - Globally toggles the notification subsystem.
- `FLAGR_NOTIFICATION_PROVIDER=slack` (Options: `slack`, `email`, `webhook`) - Determines the active transport channel.
- `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true` (Default: `false`) - When enabled, Flagr will embed the precise visual JSON diff of the modified entity within the notification payload.
- `FLAGR_NOTIFICATION_TIMEOUT=10s` (Default: `10s`) - Configures the timeout window for dialing external notification webhooks and email APIs.

### Retry Configuration (HTTP providers only)

- `FLAGR_NOTIFICATION_MAX_RETRIES=3` (Default: `3`) - Maximum number of retry attempts for transient HTTP failures (5xx errors). Set to `0` to disable retries.
- `FLAGR_NOTIFICATION_RETRY_BASE=1s` (Default: `1s`) - Base delay for exponential backoff between retries.
- `FLAGR_NOTIFICATION_RETRY_MAX=10s` (Default: `10s`) - Maximum delay between retries.

### Concurrency & Observability

- Notifications are sent asynchronously with a default concurrency limit of 100 to prevent resource exhaustion under load.
- Metric `notification.sent` is emitted when statsd is enabled, tagged with `provider`, `operation`, `entity_type`, and `status` (`success`/`failure`).

### Important Notes

- **Asynchronous delivery**: Notifications are sent in background goroutines. Failures are logged but **do not affect the API response**.
- **Startup validation**: When `FLAGR_NOTIFICATION_ENABLED=true`, Flagr validates the configuration based on the selected provider and logs warnings if required settings are missing. Notifications will be silently dropped until properly configured.

## Provider Configuration

Depending on the `FLAGR_NOTIFICATION_PROVIDER` selected above, configure the target transport mechanism:

### 1. Slack

When using Slack, the notification is delivered as a formatted `Mrkdwn` message directly to your channel block.

- `FLAGR_NOTIFICATION_SLACK_WEBHOOK_URL=...` - The Incoming Webhook URL provided by your Slack Workspace.
- `FLAGR_NOTIFICATION_SLACK_CHANNEL=#engineering` - (Optional) Overrides the destination Slack channel.

### 2. Email

The Email provider sends beautifully formatted HTML summaries of modifications to a target inbox leveraging the SendGrid REST APIs.

- `FLAGR_NOTIFICATION_EMAIL_URL=https://api.sendgrid.com/v3/mail/send` - HTTP email delivery API endpoint.
- `FLAGR_NOTIFICATION_EMAIL_TO=alerts@your-org.com` - The recipient's email address.
- `FLAGR_NOTIFICATION_EMAIL_FROM=flagr-ops@your-org.com` - The designated sender address.
- `FLAGR_NOTIFICATION_EMAIL_API_KEY=...` - The authorization key for evaluating HTTP API calls.

### 3. Generic Webhook

If you wish to consume these events programmatically, the generic `webhook` provider sends HTTP `POST` requests directly to an arbitrary URL containing a serialized JSON `Notification` object representing the change.

- `FLAGR_NOTIFICATION_WEBHOOK_URL=https://api.your-org.com/webhooks/flagr` - HTTP destination endpoint for generic webhook POST requests.
- `FLAGR_NOTIFICATION_WEBHOOK_HEADERS=Authorization: Bearer secret-token, X-Custom-Header: value` - (Optional) Custom comma-separated HTTP headers, often utilized for securing your webhook receiver with an API token.

---

## The JSON Webhook Payload Format

If `FLAGR_NOTIFICATION_PROVIDER` is set to `webhook`, the target endpoint will receive a structured payload similar to the following:

```json
{
  "Operation": "update",
  "EntityType": "flag",
  "EntityID": 123,
  "EntityKey": "my-feature-flag",
  "Description": "Optional description of the update",
  "PreValue": "{\"key\": \"value\"}",
  "PostValue": "{\"key\": \"new_value\"}",
  "Diff": "--- Previous\n+++ Current\n@@ -1 +1 @@\n-{\"key\": \"value\"}\n+{\"key\": \"new_value\"}",
  "User": "admin@example.com"
}
```

> **Note**: The `Diff` key is visually rendered in Markdown format for rendering natively across internal dashboards or chat systems, but is only populated if `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true` is set on the server.
