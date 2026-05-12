package notification

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestNotification(t *testing.T) {
	t.Run("null notifier should not fail", func(t *testing.T) {
		n := &nullNotifier{}
		ctx := context.Background()
		notif := Notification{
			Operation:  OperationCreate,
			FlagID:   1,
			FlagKey:  "test-flag",
			User:       "test@example.com",
		}

		err := n.Send(ctx, notif)
		assert.NoError(t, err)
	})

	t.Run("null notifier name", func(t *testing.T) {
		n := &nullNotifier{}
		assert.Equal(t, "null", n.Name())
	})

	t.Run("mock notifier records sent notifications", func(t *testing.T) {
		m := NewMockNotifier()
		ctx := context.Background()

		notif1 := Notification{
			Operation:  OperationCreate,
			FlagID:   1,
			FlagKey:  "test-flag-1",
			User:       "user1@example.com",
		}

		notif2 := Notification{
			Operation:  OperationUpdate,
			FlagID:   2,
			FlagKey:  "test-flag-2",
			User:       "user2@example.com",
		}

		err1 := m.Send(ctx, notif1)
		err2 := m.Send(ctx, notif2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		sent := m.GetSentNotifications()
		assert.Len(t, sent, 2)
		assert.Equal(t, OperationCreate, sent[0].Operation)
		assert.Equal(t, uint(1), sent[0].FlagID)
		assert.Equal(t, "test-flag-1", sent[0].FlagKey)

		assert.Equal(t, OperationUpdate, sent[1].Operation)
		assert.Equal(t, uint(2), sent[1].FlagID)
		assert.Equal(t, "test-flag-2", sent[1].FlagKey)
	})

	t.Run("mock notifier can return errors", func(t *testing.T) {
		m := NewMockNotifier()
		m.SetSendError(errors.New("test error"))

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

func TestGetNotifiers(t *testing.T) {
	t.Run("GetNotifiers returns empty when disabled", func(t *testing.T) {
		stubs := gostub.Stub(&Notifiers, []Notifier(nil))
		stubs.Stub(&once, sync.Once{})
		defer stubs.Reset()

		n := GetNotifiers()
		assert.Empty(t, n)
	})

	t.Run("GetNotifiers returns pre-set notifiers for testing", func(t *testing.T) {
		mock := NewMockNotifier()
		stubs := gostub.Stub(&Notifiers, []Notifier{mock})
		stubs.Stub(&once, sync.Once{})
		defer stubs.Reset()

		n := GetNotifiers()
		assert.Len(t, n, 1)
		assert.Equal(t, "mock", n[0].Name())
	})
}

func TestNotifierConcurrency(t *testing.T) {
	t.Run("MockNotifier is safe for concurrent use", func(t *testing.T) {
		mock := NewMockNotifier()

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				mock.Send(context.Background(), Notification{Operation: OperationCreate})
			}()
		}

		wg.Wait()

		// All notifications should have been sent
		assert.Len(t, mock.GetSentNotifications(), 100)
	})
}