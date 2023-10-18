package notify

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
)

func TestWebhookSendsRequest(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		config.Config.NotifyWebhookURL = "https://foo.com/1ASDA"
		client := NewTestClient(func(req *http.Request) *http.Response {
			assert.Equal(t, config.Config.NotifyWebhookURL, req.URL.String())

			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: io.NopCloser(bytes.NewBufferString(`OK`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		})

		webhook := NewWebhook(client)

		flag := entity.GenFixtureFlag()
		err := webhook.Notify(&flag, TOGGLED, FLAG, "")
		assert.NoError(t, err)
	})

	t.Run("failing webhook", func(t *testing.T) {
		config.Config.NotifyWebhookURL = "https://foo.com/1ASDA"
		client := NewTestClient(func(req *http.Request) *http.Response {
			assert.Equal(t, config.Config.NotifyWebhookURL, req.URL.String())
			return &http.Response{
				StatusCode: 500,
				// Send response to be tested
				Body: io.NopCloser(bytes.NewBufferString(`NOT OK`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		})

		webhook := NewWebhook(client)

		flag := entity.GenFixtureFlag()
		err := webhook.Notify(&flag, TOGGLED, FLAG, "")
		assert.Error(t, err)
	})
}
