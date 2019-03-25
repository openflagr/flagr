package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/mapper/entity_restapi/e2r"
	"github.com/checkr/flagr/swagger_gen/models"
)

// Webhook is a generic webhook that sends the flag and the event details alongside each other
type Webhook struct {
	client *http.Client
}

// NewWebhook returns a new Webhook
func NewWebhook(c *http.Client) *Webhook {
	return &Webhook{
		client: c,
	}
}

// WebhookMessage defines the JSON object send to webhook endpoints.
type WebhookMessage struct {
	Action  itemAction   `json:"action"`
	Type    itemType     `json:"type"`
	Data    *models.Flag `json:"data"`
	Version string       `json:"version"`
}

// Notify implements the Notifier interface for webhooks
func (w *Webhook) Notify(f *entity.Flag, b itemAction, i itemType) error {
	model, err := e2r.MapFlag(f)

	if err != nil {
		return err
	}

	msg := &WebhookMessage{
		Action:  b,
		Type:    i,
		Version: "1",
		Data:    model,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", config.Config.WebhookURL, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("User-Agent", userAgentHeader)

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(b))
	}

	return nil
}
