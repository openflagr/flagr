package notification

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
			FlagID:   1,
			FlagKey:  "test-flag",
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
		}

		err := wn.Send(ctx, notif)
		assert.NoError(t, err)

		assert.Equal(t, "Bearer secret-token", receivedAuth)
		assert.Equal(t, "custom-value", receivedCustomHeader)
	})

	t.Run("sends correctly formatted JSON payload", func(t *testing.T) {
		var receivedBody []byte
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedBody, _ = io.ReadAll(r.Body)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		stubs := gostub.Stub(&config.Config.NotificationWebhookURL, ts.URL)
		defer stubs.Reset()

		wn := NewWebhookNotifier()
		ctx := context.Background()
		now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
		notif := Notification{
			Operation:     OperationUpdate,
			FlagID:        42,
			FlagKey:       "my-feature",
			ComponentType: "segment",
		ComponentID:   7,
		ComponentKey:  "power-users",
			Diff:          "-old\n+new",
			User:          "admin@example.com",
			Timestamp:     now,
		}

		err := wn.Send(ctx, notif)
		assert.NoError(t, err)

		var parsed map[string]any
		assert.NoError(t, json.Unmarshal(receivedBody, &parsed))

		// Verify snake_case JSON keys
		assert.Equal(t, "update", parsed["operation"])
		assert.Equal(t, float64(42), parsed["flag_id"])
		assert.Equal(t, "my-feature", parsed["flag_key"])
		assert.Equal(t, "segment", parsed["component_type"])
		assert.Equal(t, float64(7), parsed["component_id"])
		assert.Equal(t, "power-users", parsed["component_key"])
		assert.Equal(t, "-old\n+new", parsed["diff"])
		assert.Equal(t, "admin@example.com", parsed["user"])
		assert.Equal(t, "2025-01-15T12:00:00Z", parsed["timestamp"])

		// Fields not present
		assert.NotContains(t, parsed, "entity_type") // old name, should be gone
		assert.NotContains(t, parsed, "object")       // removed for simplicity
		assert.NotContains(t, parsed, "pre_value")
		assert.NotContains(t, parsed, "post_value")
	})
}
