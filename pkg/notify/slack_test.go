package notify

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
)

func TestSlackSendsRequest(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		config.Config.NotifySlackURL = "https://foo.com/1ASDA"
		client := NewTestClient(func(req *http.Request) *http.Response {
			assert.Equal(t, config.Config.NotifySlackURL, req.URL.String())
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		})

		slackHook := NewSlack(client)
		f := entity.GenFixtureFlag()
		err := slackHook.Notify(&f, "TOGGLED", "FLAG", "author@example.com")
		assert.NoError(t, err)
	})

	t.Run("failing webhook", func(t *testing.T) {
		config.Config.NotifySlackURL = "https://foo.com/1ASDA"
		client := NewTestClient(func(req *http.Request) *http.Response {
			assert.Equal(t, config.Config.NotifySlackURL, req.URL.String())
			return &http.Response{
				StatusCode: 500,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(`NOT OK`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		})

		slackHook := NewSlack(client)
		f := entity.GenFixtureFlag()
		err := slackHook.Notify(&f, "TOGGLED", "FLAG", "author@example.com")
		assert.Error(t, err)
	})

	t.Run("slack req is built properly for toggle notifications", func(t *testing.T) {
		f := entity.GenFixtureFlag()

		slackReq := buildSlackRequest(&f, TOGGLED, FLAG, "")
		fmt.Printf("%+v\n", slackReq)
		assert.Equal(t, "Flag #100 () was updated", slackReq.Text)
		assert.Contains(t, slackReq.Blocks[0].Text.Text, "Flag #100 () has been enabled at")
		assert.Equal(t, "Flagr", slackReq.Username)

		slackReq = buildSlackRequest(&f, TOGGLED, FLAG, "author@example.com")
		fmt.Printf("%+v\n", slackReq)
		assert.Equal(t, "Flag #100 () by author@example.com was updated", slackReq.Text)
		assert.Contains(t, slackReq.Blocks[0].Text.Text, "Flag #100 () by author@example.com has been enabled at")
		assert.Equal(t, "Flagr", slackReq.Username)
	})

	t.Run("slack req is built properly for crud notifications", func(t *testing.T) {
		f := entity.GenFixtureFlag()

		slackReq := buildSlackRequest(&f, UPDATED, SEGMENT, "")
		assert.Equal(t, "Flag #100 () was updated", slackReq.Text)
		assert.Equal(t, "*Flag #100 ()*\n Segment was Updated", slackReq.Blocks[0].Text.Text)
		assert.Equal(t, "Flagr", slackReq.Username)

		slackReq = buildSlackRequest(&f, UPDATED, SEGMENT, "author@example.com")
		assert.Equal(t, "Flag #100 () by author@example.com was updated", slackReq.Text)
		assert.Equal(t, "*Flag #100 () by author@example.com*\n Segment was Updated", slackReq.Blocks[0].Text.Text)
		assert.Equal(t, "Flagr", slackReq.Username)
	})
}
