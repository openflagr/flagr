package notification

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestCalculateDiff(t *testing.T) {
	t.Run("empty cases", func(t *testing.T) {
		assert.Empty(t, CalculateDiff("", ""))
		assert.Empty(t, CalculateDiff("a", ""))
		assert.Empty(t, CalculateDiff("", "b"))
	})

	t.Run("simple diff", func(t *testing.T) {
		pre := "line1\nline2\n"
		post := "line1\nline3\n"
		diff := CalculateDiff(pre, post)
		assert.NotEmpty(t, diff)
		assert.Contains(t, diff, "-line2")
		assert.Contains(t, diff, "+line3")
	})

	t.Run("JSON diff visibility", func(t *testing.T) {
		pre := `{"id":1,"key":"flag1","enabled":false}`
		post := `{"id":1,"key":"flag1","enabled":true}`
		diff := CalculateDiff(pre, post)
		t.Logf("Pretty JSON Diff:\n%s", diff)
		// Pretty JSON diff shows individual field changes
		assert.Contains(t, diff, "-  \"enabled\": false")
		assert.Contains(t, diff, "+  \"enabled\": true")
	})
}

func TestSendNotification(t *testing.T) {
	t.Run("sends to multiple notifiers concurrently", func(t *testing.T) {
		mock1 := NewMockNotifier()
		mock2 := NewMockNotifier()
		mock3 := NewMockNotifier()

		// First reset to nil, then stub to desired value
		Notifiers = nil
		stubs := gostub.Stub(&Notifiers, []Notifier{mock1, mock2, mock3})
		defer stubs.Reset()

		SendNotification(Notification{
			Operation:  OperationCreate,
			FlagID:   1,
			FlagKey:  "test-flag",
			User:       "user",
		})

		// Wait for goroutine to complete
		assert.Eventually(t, func() bool {
			return len(mock1.GetSentNotifications()) == 1 &&
				len(mock2.GetSentNotifications()) == 1 &&
				len(mock3.GetSentNotifications()) == 1
		}, 1*time.Second, 10*time.Millisecond)

		// Verify each notifier received the same notification
		for _, mock := range []*MockNotifier{mock1, mock2, mock3} {
			sent := mock.GetSentNotifications()
			assert.Len(t, sent, 1)
			assert.Equal(t, OperationCreate, sent[0].Operation)
			assert.Equal(t, uint(1), sent[0].FlagID)
			assert.Equal(t, "test-flag", sent[0].FlagKey)
		}
	})

	t.Run("handles errors from some notifiers", func(t *testing.T) {
		mock1 := NewMockNotifier()
		mock1.SetSendError(errors.New("error from mock1"))

		mock2 := NewMockNotifier()
		// mock2 succeeds

		mock3 := NewMockNotifier()
		mock3.SetSendError(errors.New("error from mock3"))

		Notifiers = nil
		stubs := gostub.Stub(&Notifiers, []Notifier{mock1, mock2, mock3})
		defer stubs.Reset()

		SendNotification(Notification{
			Operation:  OperationUpdate,
			FlagID:   2,
			FlagKey:  "test-flag-2",
		})

		// Wait for goroutine to complete
		assert.Eventually(t, func() bool {
			return len(mock1.GetSentNotifications()) == 1 &&
				len(mock2.GetSentNotifications()) == 1 &&
				len(mock3.GetSentNotifications()) == 1
		}, 1*time.Second, 10*time.Millisecond)

		// All notifiers should still have been called (fire all)
		assert.Len(t, mock1.GetSentNotifications(), 1)
		assert.Len(t, mock2.GetSentNotifications(), 1)
		assert.Len(t, mock3.GetSentNotifications(), 1)
	})

	t.Run("does nothing when notifiers is empty", func(t *testing.T) {
		Notifiers = nil
		stubs := gostub.Stub(&Notifiers, []Notifier(nil))
		defer stubs.Reset()

		// Should not panic
		SendNotification(Notification{
			Operation:  OperationCreate,
			FlagID:   1,
		})
	})

	t.Run("sends notification with correct entity type and fields", func(t *testing.T) {
		mock := NewMockNotifier()
		Notifiers = nil
		stubs := gostub.Stub(&Notifiers, []Notifier{mock})
		defer stubs.Reset()

		SendNotification(Notification{
			Operation:   OperationCreate,
			FlagID:    42,
			FlagKey:   "my-flag",
			User:        "creator",
		})

		assert.Eventually(t, func() bool {
			return len(mock.GetSentNotifications()) >= 1
		}, 1*time.Second, 10*time.Millisecond)

		sent := mock.GetSentNotifications()
		assert.Len(t, sent, 1)
		assert.Equal(t, OperationCreate, sent[0].Operation)
		assert.Equal(t, uint(42), sent[0].FlagID)
		assert.Equal(t, "my-flag", sent[0].FlagKey)
		assert.Equal(t, "creator", sent[0].User)
	})

	t.Run("sets timestamp when not provided", func(t *testing.T) {
		mock := NewMockNotifier()
		Notifiers = nil
		stubs := gostub.Stub(&Notifiers, []Notifier{mock})
		defer stubs.Reset()

		before := time.Now()
		SendNotification(Notification{
			Operation:  OperationCreate,
			FlagID:   1,
		})

		assert.Eventually(t, func() bool {
			return len(mock.GetSentNotifications()) >= 1
		}, 1*time.Second, 10*time.Millisecond)

		sent := mock.GetSentNotifications()
		assert.Len(t, sent, 1)
		assert.False(t, sent[0].Timestamp.IsZero())
		assert.True(t, sent[0].Timestamp.After(before) || sent[0].Timestamp.Equal(before))
	})
}

func TestSendNotificationConcurrency(t *testing.T) {
	t.Run("concurrent sends are handled safely", func(t *testing.T) {
		mock := NewMockNotifier()
		stubs := gostub.Stub(&Notifiers, []Notifier{mock})
		defer stubs.Reset()

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id uint) {
				defer wg.Done()
				SendNotification(Notification{
					Operation:  OperationCreate,
					FlagID:   id,
					FlagKey:  "flag",
				})
			}(uint(i))
		}

		wg.Wait()

		// All notifications should eventually be delivered
		assert.Eventually(t, func() bool {
			return len(mock.GetSentNotifications()) == 50
		}, 2*time.Second, 50*time.Millisecond)
	})
}

func TestNotifierDirectSend(t *testing.T) {
	t.Run("can send to notifier directly with context", func(t *testing.T) {
		mock := NewMockNotifier()

		ctx := context.Background()
		notif := Notification{
			Operation:   OperationCreate,
			FlagID:    1,
			FlagKey:   "direct-test",
			User:        "tester",
		}

		err := mock.Send(ctx, notif)
		assert.NoError(t, err)

		sent := mock.GetSentNotifications()
		assert.Len(t, sent, 1)
		assert.Equal(t, "direct-test", sent[0].FlagKey)
	})
}
