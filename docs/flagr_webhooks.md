# Webhooks

Flagr supports webhooks to notify external systems about changes to flags, segments, variants, and constraints. This allows you to integrate Flagr with your existing systems and workflows.

## Overview

Webhooks in Flagr are HTTP callbacks that are triggered when specific events occur. Each webhook can be configured to listen for specific events and will be triggered for all flags.

## Events

The following events are supported:

### Flag Events
- `flag.created` - When a new flag is created
- `flag.updated` - When a flag's properties are updated
- `flag.deleted` - When a flag is deleted
- `flag.enabled` - When a flag is enabled
- `flag.disabled` - When a flag is disabled

### Segment Events
- `segment.created` - When a new segment is created
- `segment.updated` - When a segment's properties are updated
- `segment.deleted` - When a segment is deleted
- `segment.rollout_percent.updated` - When a segment's rollout percentage is updated

### Variant Events
- `variant.created` - When a new variant is created
- `variant.updated` - When a variant's properties are updated
- `variant.deleted` - When a variant is deleted
- `variant.attachment.updated` - When a variant's attachment is updated

### Constraint Events
- `constraint.created` - When a new constraint is created
- `constraint.updated` - When a constraint's properties are updated
- `constraint.deleted` - When a constraint is deleted

### Distribution Events
- `distribution.updated` - When distributions are updated

## Webhook Payload

Each webhook payload includes the following information:

```json
{
  "event": "flag.updated",
  "flag_id": 1,
  "flag_key": "my_feature",
  "timestamp": "2024-03-21T12:00:00Z",
  "data": {
    // Complete flag data including segments, variants, and constraints
  }
}
```

The `data` field contains the complete flag data at the time of the event, including:
- Flag properties (key, description, enabled state, etc.)
- All segments with their constraints
- All variants
- All distributions

## Security

Webhooks can be secured using a secret key. When a secret is configured, Flagr will include an `X-Flagr-Signature` header in the webhook request. The signature is an HMAC SHA-256 hash of the payload, using the secret as the key.

To verify the webhook:
1. Get the signature from the `X-Flagr-Signature` header
2. Calculate the HMAC SHA-256 hash of the raw request body using your secret
3. Compare the calculated hash with the signature

## Configuration

Webhooks can be configured through the Flagr UI or API in the "Webhooks" section. For each webhook, you can configure:
- URL: The endpoint that will receive the webhook
- Events: The list of events to subscribe to
- Secret: (Optional) A secret key for webhook verification
- Description: A description of the webhook's purpose

## Best Practices

1. **Always verify webhook signatures** when a secret is configured
2. **Handle webhook failures gracefully** - Flagr will retry failed webhooks
3. **Keep webhook processing fast** - Flagr has a 10-second timeout for webhook delivery
4. **Monitor webhook delivery** - Check the webhook events in the UI to ensure successful delivery
5. **Use HTTPS** for webhook endpoints to ensure secure delivery

## Example

Here's an example of a webhook payload for a flag update:

```json
{
  "event": "flag.updated",
  "flag_id": 1,
  "flag_key": "new_feature",
  "timestamp": "2024-03-21T12:00:00Z",
  "data": {
    "id": 1,
    "key": "new_feature",
    "description": "A new feature flag",
    "enabled": true,
    "segments": [
      {
        "id": 1,
        "description": "Beta users",
        "constraints": [...],
        "distributions": [...]
      }
    ],
    "variants": [
      {
        "id": 1,
        "key": "control",
        "attachment": null
      },
      {
        "id": 2,
        "key": "treatment",
        "attachment": {
          "message": "New feature enabled!"
        }
      }
    ],
    "tags": [
      {
        "id": 1,
        "value": "beta"
      }
    ]
  }
}
``` 