package notification

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailNotifier(t *testing.T) {
	t.Run("returns null notifier when no email URL", func(t *testing.T) {
		en := NewEmailNotifier()
		ctx := context.Background()
		notif := Notification{
			Operation:  "create",
			EntityType: "flag",
			EntityID:   1,
			EntityKey:  "test-flag",
			User:       "test@example.com",
		}

		err := en.Send(ctx, notif)
		assert.NoError(t, err)
	})

	t.Run("formats subject correctly", func(t *testing.T) {
		n := Notification{Operation: "create", EntityType: "flag"}
		subject := formatEmailSubject(n)
		assert.Equal(t, "[Flagr] create flag", subject)
	})

	t.Run("formats body correctly for create", func(t *testing.T) {
		n := Notification{
			Operation:  "create",
			EntityType: "flag",
			EntityKey:  "test-flag",
			EntityID:   1,
			User:       "user@example.com",
		}
		body := formatEmailBody(n)
		assert.Contains(t, body, "🚀")
		assert.Contains(t, body, "create flag")
		assert.Contains(t, body, "Key: test-flag")
		assert.Contains(t, body, "ID: 1")
		assert.Contains(t, body, "User: user@example.com")
	})

	t.Run("formats body correctly for update", func(t *testing.T) {
		n := Notification{
			Operation:  "update",
			EntityType: "flag",
		}
		body := formatEmailBody(n)
		assert.Contains(t, body, "✏️")
		assert.Contains(t, body, "update flag")
	})

	t.Run("formats body correctly for delete", func(t *testing.T) {
		n := Notification{
			Operation:  "delete",
			EntityType: "segment",
		}
		body := formatEmailBody(n)
		assert.Contains(t, body, "🗑️")
		assert.Contains(t, body, "delete segment")
	})

	t.Run("formats body correctly with description and values", func(t *testing.T) {
		n := Notification{
			Operation:   "update",
			EntityType:  "flag",
			EntityKey:   "test-flag",
			EntityID:    1,
			Description: "test description",
			PreValue:    `{"enabled": false}`,
			PostValue:   `{"enabled": true}`,
			User:        "user@example.com",
		}
		body := formatEmailBody(n)
		assert.Contains(t, body, "Description: test description")
		assert.Contains(t, body, "Pre-value:\n{\"enabled\": false}")
		assert.Contains(t, body, "Post-value:\n{\"enabled\": true}")
	})

	t.Run("formats body correctly with diff", func(t *testing.T) {
		n := Notification{
			Operation:  "update",
			EntityType: "flag",
			Diff:       "-old\n+new",
		}
		body := formatEmailBody(n)
		assert.Contains(t, body, "Diff:\n-old\n+new")
	})
}
