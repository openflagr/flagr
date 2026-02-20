package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/sirupsen/logrus"
)

type webhookNotifier struct {
	httpClient *http.Client
}

func NewWebhookNotifier() Notifier {
	if config.Config.NotificationWebhookURL == "" {
		logrus.Warn("NotificationWebhookURL is empty, using null notifier")
		return &nullNotifier{}
	}

	return &webhookNotifier{
		httpClient: &http.Client{Timeout: config.Config.NotificationTimeout},
	}
}

func (w *webhookNotifier) Send(ctx context.Context, n Notification) error {
	jsonPayload, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.Config.NotificationWebhookURL, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range util.ParseHeaders(config.Config.NotificationWebhookHeaders) {
		req.Header.Set(k, v)
	}

	// Execute request with retry
	resp, err := doRequestWithRetry(ctx, w.httpClient, req, config.Config.NotificationMaxRetries, config.Config.NotificationRetryBase, config.Config.NotificationRetryMax)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook service returned error: %d - %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"status":    resp.StatusCode,
		"operation": n.Operation,
		"entityID":  n.EntityID,
	}).Info("webhook notification sent successfully")

	return nil
}

func (w *webhookNotifier) Name() string {
	return "webhook"
}
