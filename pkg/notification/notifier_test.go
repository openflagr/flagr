package notification

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotification(t *testing.T) {
	t.Run("null notifier should not fail", func(t *testing.T) {
		n := &nullNotifier{}
		ctx := context.Background()
		notif := Notification{
			Operation:  OperationCreate,
			EntityType: EntityTypeFlag,
			EntityID:   1,
			EntityKey:  "test-flag",
			User:       "test@example.com",
		}

		err := n.Send(ctx, notif)
		assert.NoError(t, err)
	})

	t.Run("mock notifier records sent notifications", func(t *testing.T) {
		m := NewMockNotifier()
		ctx := context.Background()

		notif1 := Notification{
			Operation:  OperationCreate,
			EntityType: EntityTypeFlag,
			EntityID:   1,
			EntityKey:  "test-flag-1",
			User:       "user1@example.com",
		}

		notif2 := Notification{
			Operation:  OperationUpdate,
			EntityType: EntityTypeFlag,
			EntityID:   2,
			EntityKey:  "test-flag-2",
			User:       "user2@example.com",
		}

		err1 := m.Send(ctx, notif1)
		err2 := m.Send(ctx, notif2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		sent := m.GetSentNotifications()
		assert.Len(t, sent, 2)
		assert.Equal(t, OperationCreate, sent[0].Operation)
		assert.Equal(t, EntityTypeFlag, sent[0].EntityType)
		assert.Equal(t, uint(1), sent[0].EntityID)
		assert.Equal(t, "test-flag-1", sent[0].EntityKey)

		assert.Equal(t, OperationUpdate, sent[1].Operation)
		assert.Equal(t, EntityTypeFlag, sent[1].EntityType)
		assert.Equal(t, uint(2), sent[1].EntityID)
		assert.Equal(t, "test-flag-2", sent[1].EntityKey)
	})

	t.Run("mock notifier can return errors", func(t *testing.T) {
		m := NewMockNotifier()
		m.SetSendError(assert.AnError)

		ctx := context.Background()
		notif := Notification{Operation: Operation("test")}

		err := m.Send(ctx, notif)
		assert.Error(t, err)
	})

	t.Run("mock notifier clear works", func(t *testing.T) {
		m := NewMockNotifier()
		ctx := context.Background()

		m.Send(ctx, Notification{Operation: "test"})
		m.Send(ctx, Notification{Operation: "test"})

		assert.Len(t, m.GetSentNotifications(), 2)

		m.ClearSent()
		assert.Len(t, m.GetSentNotifications(), 0)
	})
}
