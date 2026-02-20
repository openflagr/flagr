package notification

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestWebhookNotifier(t *testing.T) {
	t.Run("returns null notifier when no webhook URL", func(t *testing.T) {
		wn := NewWebhookNotifier()
		ctx := context.Background()
		notif := Notification{
			Operation:  "create",
			EntityType: "flag",
			EntityID:   1,
			EntityKey:  "test-flag",
			User:       "test@example.com",
		}

		err := wn.Send(ctx, notif)
		assert.NoError(t, err)
	})

	t.Run("sends custom headers", func(t *testing.T) {

		var receivedAuth string
		var receivedCustomHeader string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuth = r.Header.Get("Authorization")
			receivedCustomHeader = r.Header.Get("X-Custom-Header")
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		stubs := gostub.Stub(&config.Config.NotificationWebhookURL, ts.URL)
		defer stubs.Reset()

		stubs.Stub(&config.Config.NotificationWebhookHeaders, "Authorization: Bearer secret-token, X-Custom-Header: custom-value ")

		wn := NewWebhookNotifier()
		ctx := context.Background()
		notif := Notification{
			Operation:  "create",
			EntityType: "flag",
		}

		err := wn.Send(ctx, notif)
		assert.NoError(t, err)

		assert.Equal(t, "Bearer secret-token", receivedAuth)
		assert.Equal(t, "custom-value", receivedCustomHeader)
	})
}
