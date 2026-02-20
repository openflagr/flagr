# Notifications

Flagr supports sending notifications for CRUD operations via Slack, Webhooks, or Email.

## Configuration

Set these environment variables to enable notifications:

- `FLAGR_NOTIFICATION_ENABLED=true` - Enable notifications (default: false)
- `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true` - Include detailed value diffs in notifications (default: false)
- `FLAGR_NOTIFICATION_TIMEOUT=10s` - Timeout for HTTP requests when sending notifications
- `FLAGR_NOTIFICATION_PROVIDER=slack` - Notification provider (options: `slack`, `email`, `webhook`)

### Webhook
- `FLAGR_NOTIFICATION_WEBHOOK_URL=...` - Generic webhook URL to POST JSON payloads to
- `FLAGR_NOTIFICATION_WEBHOOK_HEADERS=...` - Optional comma-separated headers (e.g., `Authorization: Bearer token, X-Custom-Header: value`)

### Slack
- `FLAGR_NOTIFICATION_SLACK_WEBHOOK_URL=...` - Slack webhook URL
- `FLAGR_NOTIFICATION_SLACK_CHANNEL=#channel-name` - Optional Slack channel

### Email
- `FLAGR_NOTIFICATION_EMAIL_URL=...` - HTTP email API URL
- `FLAGR_NOTIFICATION_EMAIL_TO=...` - Recipient email address
- `FLAGR_NOTIFICATION_EMAIL_FROM=...` - Sender email address
- `FLAGR_NOTIFICATION_EMAIL_API_KEY=...` - Optional API key for email service

## Operations That Trigger Notifications

- Create, Update, Delete, Restore flags
- Enable/Disable flags
- Create tags
- Create, Update, Delete segments
- Create, Update, Delete constraints
- Update distributions
- Create, Update, Delete variants

## Notification Format

Internal to Flagr, every CRUD notification is structured as a generic `Notification` object.

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

Depending on your configured `FLAGR_NOTIFICATION_PROVIDER`, this generic object is formatted and delivered:
- **`webhook`**: The JSON representation above is serialized and HTTP `POST`ed directly to the target URL.
- **`slack`** & **`email`**: The object is parsed into a human-readable text document (with Mrkdwn formatting for Slack) and delivered to the channel or inbox. If `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true`, the `Diff` property is also included visually in the message.

## How It Works

1. After successful CRUD operations, `SaveFlagSnapshot()` is called
2. The snapshot is saved to the database
3. A notification is sent asynchronously via `SendFlagNotification()`
4. Notifications are non-blocking - failures are logged but don't affect the operation

## Testing

```bash
# Run notification tests
go test ./pkg/notification/... -v

# Run tests with mock notifier
go test ./pkg/notification/... -run TestNotification -v
```

## Adding New Providers

1. Implement the `Notifier` interface in `pkg/notification/`
2. Update `GetNotifier()` in `pkg/notification/notifier.go` to support the new provider
3. Add configuration options in `pkg/config/env.go`
