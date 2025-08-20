package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/sirupsen/logrus"
)

// Service handles webhook operations
type Service struct {
	client *http.Client
}

// NewService creates a new webhook service
func NewService() *Service {
	return &Service{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// TriggerWebhooks triggers webhooks for system events
func (s *Service) TriggerWebhooks(event string, data interface{}) error {
	logrus.WithFields(logrus.Fields{
		"event": event,
	}).Info("Triggering webhooks for system event")

	var webhooks []entity.Webhook
	query := entity.GetDB().Where("enabled = ?", true)

	if err := query.Find(&webhooks).Error; err != nil {
		logrus.WithError(err).Error("Failed to find webhooks")
		return fmt.Errorf("failed to find webhooks: %v", err)
	}

	logrus.WithField("webhook_count", len(webhooks)).Info("Found webhooks to trigger")

	for _, webhook := range webhooks {
		if !strings.Contains(webhook.Events, event) {
			logrus.WithFields(logrus.Fields{
				"webhook_id": webhook.ID,
				"events":     webhook.Events,
				"event":      event,
			}).Debug("Skipping webhook - event not subscribed")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"webhook_id": webhook.ID,
			"url":        webhook.URL,
			"event":      event,
		}).Info("Preparing webhook payload")

		payload := map[string]interface{}{
			"event":     event,
			"timestamp": time.Now().UTC(),
			"data":      data,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			logrus.WithError(err).Error("Failed to marshal webhook payload")
			continue
		}

		req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			logrus.WithError(err).Error("Failed to create webhook request")
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Flagr-Event", event)

		if webhook.Secret != "" {
			signature := s.generateSignature(payloadBytes, webhook.Secret)
			req.Header.Set("X-Flagr-Signature", signature)
		}

		webhookEvent := &entity.WebhookEvent{
			WebhookID: webhook.ID,
			Event:     event,
			Payload:   string(payloadBytes),
			Status:    "pending",
		}

		if err := entity.GetDB().Create(webhookEvent).Error; err != nil {
			logrus.WithError(err).Error("Failed to create webhook event")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"webhook_id": webhook.ID,
			"url":        webhook.URL,
		}).Info("Sending webhook request")

		resp, err := s.client.Do(req)
		if err != nil {
			webhookEvent.Status = "failed"
			webhookEvent.Error = err.Error()
			entity.GetDB().Save(webhookEvent)
			logrus.WithError(err).Error("Failed to send webhook")
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			webhookEvent.Status = "success"
		} else {
			webhookEvent.Status = "failed"
			webhookEvent.Error = fmt.Sprintf("Webhook returned status code %d", resp.StatusCode)
		}

		entity.GetDB().Save(webhookEvent)
	}

	return nil
}

// generateSignature generates an HMAC SHA-256 signature for the webhook payload
func (s *Service) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
