# Notification Feature

## Overview

Flagr now supports sending notifications for CRUD operations via Slack or other notification providers.

## Configuration

Set these environment variables to enable notifications:

- `FLAGR_NOTIFICATION_ENABLED=true` - Enable notifications (default: false)
- `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true` - Include detailed value diffs in notifications (default: false)
- `FLAGR_NOTIFICATION_TIMEOUT=10s` - Timeout for HTTP requests when sending notifications
- `FLAGR_NOTIFICATION_PROVIDER=slack` - Notification provider (options: `slack`, `email`, `webhook`)

### Slack
- `FLAGR_NOTIFICATION_SLACK_WEBHOOK_URL=...` - Slack webhook URL
- `FLAGR_NOTIFICATION_SLACK_CHANNEL=#channel-name` - Optional Slack channel

### Webhook
- `FLAGR_NOTIFICATION_WEBHOOK_URL=...` - Generic webhook URL to POST JSON payloads to
- `FLAGR_NOTIFICATION_WEBHOOK_HEADERS=...` - Optional comma-separated headers (e.g., `Authorization: Bearer token, X-Custom-Header: value`)

### Email
- `FLAGR_NOTIFICATION_EMAIL_URL=...` - HTTP email API URL
- `FLAGR_NOTIFICATION_EMAIL_TO=...` - Recipient email address
- `FLAGR_NOTIFICATION_EMAIL_FROM=...` - Sender email address
- `FLAGR_NOTIFICATION_EMAIL_API_KEY=...` - Optional API key for email service

## How It Works

1. After successful CRUD operations, `SaveFlagSnapshot()` is called
2. The snapshot is saved to the database
3. A notification is sent asynchronously via `SendFlagNotification()`
4. Notifications are non-blocking - failures are logged but don't affect the operation

## Operations That Trigger Notifications

- Create, Update, Delete, Restore flags
- Enable/Disable flags
- Create tags
- Create, Update, Delete segments
- Create, Update, Delete constraints
- Update distributions
- Create, Update, Delete variants

## Notification Format

### Webhook (JSON)

The generic webhook provider will emit an HTTP POST request with a JSON payload representing the `Notification` object. Depending on whether detailed diffs are enabled, the payload will look like this:

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
  "User": "admin@example.com",
  "Details": {}
}
```

### Slack and Email (Text)

Slack and Email providers format the `Notification` object into a human-readable format. For Slack, it uses Mrkdwn formatting.

```
:rocket: *create flag*
*Key:* my-feature-flag
*ID:* 123
*User:* user@example.com
```

If `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true` is set, updates will include the markdown diff alongside the values:

```
:pencil2: *update flag*
*Description:* Enabled the new login UI
*ID:* 123
*Key:* my-feature-flag
*User:* admin@example.com

*Diff:*
```diff
--- Previous
+++ Current
@@ -1,3 +1,3 @@
 {
-  "enabled": false
+  "enabled": true
 }
```
```

## Testing

```bash
# Run notification tests
go test ./pkg/notification/... -v

# Run tests with mock notifier
go test ./pkg/notification/... -run TestNotification -v
```

## Adding New Providers

1. Implement `Notifier` interface in `pkg/notification/`
2. Update `GetNotifier()` to support the new provider
3. Add configuration options in `pkg/config/env.go`
