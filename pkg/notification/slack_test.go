package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatNotification(t *testing.T) {
	t.Run("basic format create", func(t *testing.T) {
		n := Notification{
			Operation:  OperationCreate,
			EntityType: EntityTypeFlag,
			EntityID:   123,
			EntityKey:  "my-flag",
			User:       "testuser",
		}
		msg := formatNotification(n)
		assert.Contains(t, msg, ":rocket: *create flag*")
		assert.Contains(t, msg, "*Key:* my-flag")
		assert.Contains(t, msg, "*ID:* 123")
		assert.Contains(t, msg, "*User:* testuser")
	})

	t.Run("basic format with description, diff and values", func(t *testing.T) {
		n := Notification{
			Operation:   OperationUpdate,
			EntityType:  EntityTypeFlag,
			EntityID:    123,
			EntityKey:   "my-flag",
			Description: "updated description",
			PreValue:    "{\"enabled\": false}",
			PostValue:   "{\"enabled\": true}",
			Diff:        "-false\n+true",
		}
		msg := formatNotification(n)
		assert.Contains(t, msg, ":pencil2: *update flag*")
		assert.Contains(t, msg, "*Description:* updated description")
		assert.Contains(t, msg, "*Diff:*\n```diff\n-false\n+true\n```")
		assert.Contains(t, msg, "*Pre-value:*\n```json\n{\"enabled\": false}\n```")
		assert.Contains(t, msg, "*Post-value:*\n```json\n{\"enabled\": true}\n```")
		assert.Contains(t, msg, "*User:* anonymous") // Default user
	})

	t.Run("basic format delete", func(t *testing.T) {
		n := Notification{
			Operation:  OperationDelete,
			EntityType: EntityTypeFlag,
		}
		msg := formatNotification(n)
		assert.Contains(t, msg, ":wastebasket: *delete flag*")
	})

	t.Run("basic format other", func(t *testing.T) {
		n := Notification{
			Operation:  OperationRestore,
			EntityType: EntityTypeFlag,
		}
		msg := formatNotification(n)
		assert.Contains(t, msg, ":information_source: *restore flag*")
	})
}
