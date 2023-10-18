package notify

import (
	"bytes"
	"encoding/json"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/mapper/entity_restapi/e2r"
	"github.com/openflagr/flagr/swagger_gen/models"
)

// Webhook is a generic webhook that sends the flag and the event details alongside each other
type Webhook struct {
	client *Client
}

// NewWebhook returns a new Webhook
func NewWebhook(c *Client) *Webhook {
	return &Webhook{
		client: c,
	}
}

// WebhookMessage defines the JSON object send to webhook endpoints.
type WebhookMessage struct {
	Action  itemAction   `json:"action"`
	Type    itemType     `json:"type"`
	Subject subject      `json:"subject,omitempty"`
	Data    *models.Flag `json:"data"`
	Version string       `json:"version"`
}

// Notify implements the Notifier interface for webhooks
func (w *Webhook) Notify(f *entity.Flag, b itemAction, i itemType, s subject) error {
	model, err := e2r.MapFlag(f)
	if err != nil {
		return err
	}

	msg := &WebhookMessage{
		Action:  b,
		Type:    i,
		Subject: s,
		Data:    model,
		Version: "1",
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return err
	}

	_, err = w.client.Post(config.Config.NotifyWebhookURL, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}

	return nil
}
