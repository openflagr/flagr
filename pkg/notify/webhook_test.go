package notify

import (
	"bytes"
	"fmt"
	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"testing"
)

func TestWebhookSendsRequest(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		config.Config.WebhookUrl = "https://foo.com/1ASDA"
		client := NewTestClient(func(req *http.Request) * http.Response {
			assert.Equal(t, config.Config.WebhookUrl, req.URL.String())
			res, _ := httputil.DumpRequestOut(req, true)
			fmt.Println(string(res))

			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
				// Must be set to non-nil value or it panics
				Header:     make(http.Header),
			}
		})

		webhook := NewWebhook(client)

		flag := entity.GenFixtureFlag()
		err := webhook.Notify(&flag, TOGGLED, FLAG)
		assert.NoError(t, err)
	})

	t.Run("failing webhook", func(t *testing.T) {
		config.Config.WebhookUrl = "https://foo.com/1ASDA"
		client := NewTestClient(func(req *http.Request) * http.Response {
			assert.Equal(t, config.Config.WebhookUrl, req.URL.String())
			return &http.Response{
				StatusCode: 500,
				// Send response to be tested
				Body:       ioutil.NopCloser(bytes.NewBufferString(`NOT OK`)),
				// Must be set to non-nil value or it panics
				Header:     make(http.Header),
			}
		})

		webhook := NewWebhook(client)

		flag := entity.GenFixtureFlag()
		err := webhook.Notify(&flag, TOGGLED, FLAG)
		assert.Error(t, err)
	})
}