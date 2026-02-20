# Notifications

Flagr provides an integrated notification system that allows you to monitor changes and updates to your operational resources in real-time. You can configure Flagr to automatically send notifications regarding CRUD (Create, Read, Update, Delete) operations over several distinct channels: **Email**, **Slack**, or generic **Webhooks**.

## Tracked Operations

Flagr monitors the following operations across your entities and immediately broadcasts them:
- **Flags**: Create, Update, Delete, Restore, Enable, Disable
- **Tags**: Create
- **Segments**: Create, Update, Delete
- **Constraints**: Create, Update, Delete
- **Distributions**: Update
- **Variants**: Create, Update, Delete

## Global Configuration

You must globally enable the notifications feature via environment variables and define the timeout for HTTP providers.

- `FLAGR_NOTIFICATION_ENABLED=true` (Default: `false`) - Globally toggles the notification subsystem.
- `FLAGR_NOTIFICATION_PROVIDER=slack` (Options: `slack`, `email`, `webhook`) - Determines the active transport channel.
- `FLAGR_NOTIFICATION_DETAILED_DIFF_ENABLED=true` (Default: `false`) - When enabled, Flagr will embed the precise visual JSON diff of the modified entity within the notification payload.
- `FLAGR_NOTIFICATION_TIMEOUT=10s` (Default: `10s`) - Configures the timeout window for dialing external notification webhooks and email APIs.

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
