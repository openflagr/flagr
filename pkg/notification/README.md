# Notification Feature

## Overview

Flagr now supports sending notifications for CRUD operations via Slack or other notification providers.

## Configuration

Set these environment variables to enable notifications:

- `FLAGR_NOTIFICATION_ENABLED=true` - Enable notifications (default: false)
- `FLAGR_NOTIFICATION_PROVIDER=slack` - Notification provider (default: slack)
- `FLAGR_NOTIFICATION_SLACK_WEBHOOK_URL=...` - Slack webhook URL
- `FLAGR_NOTIFICATION_SLACK_CHANNEL=#channel-name` - Optional Slack channel

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

```
:rocket: *create flag*
*Key:* my-feature-flag
*ID:* 123
*User:* user@example.com
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
