package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/sirupsen/logrus"
)

type emailNotifier struct {
	httpClient *http.Client
}

func NewEmailNotifier() Notifier {
	if config.Config.NotificationEmailURL == "" {
		logrus.Warn("NotificationEmailURL is empty, using null notifier")
		return &nullNotifier{}
	}

	return &emailNotifier{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (e *emailNotifier) Send(ctx context.Context, n Notification) error {
	subject := formatEmailSubject(n)
	body := formatEmailBody(n)

	payload := map[string]string{
		"from":    config.Config.NotificationEmailFrom,
		"to":      config.Config.NotificationEmailTo,
		"subject": subject,
		"text":    body,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.Config.NotificationEmailURL, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create email request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if config.Config.NotificationEmailAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+config.Config.NotificationEmailAPIKey)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("email service returned error: %d - %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"status":  resp.StatusCode,
		"to":      config.Config.NotificationEmailTo,
		"from":    config.Config.NotificationEmailFrom,
		"subject": subject,
	}).Info("email notification sent successfully")
	return nil
}

func formatEmailSubject(n Notification) string {
	return fmt.Sprintf("[Flagr] %s %s", n.Operation, n.EntityType)
}

func formatEmailBody(n Notification) string {
	var emoji string
	switch n.Operation {
	case OperationCreate:
		emoji = "🚀"
	case OperationUpdate:
		emoji = "✏️"
	case OperationDelete:
		emoji = "🗑️"
	default:
		emoji = "ℹ️"
	}

	userInfo := "anonymous"
	if n.User != "" {
		userInfo = n.User
	}

	body := fmt.Sprintf(
		"%s %s %s\n\n"+
			"Key: %s\n"+
			"ID: %d\n",
		emoji, n.Operation, n.EntityType, n.EntityKey, n.EntityID,
	)

	if n.Description != "" {
		body += fmt.Sprintf("Description: %s\n", n.Description)
	}

	body += fmt.Sprintf("User: %s\n", userInfo)

	if n.Diff != "" {
		body += fmt.Sprintf("\nDiff:\n%s\n", n.Diff)
	}

	if n.PreValue != "" {
		body += fmt.Sprintf("\nPre-value:\n%s\n", n.PreValue)
	}
	if n.PostValue != "" {
		body += fmt.Sprintf("\nPost-value:\n%s\n", n.PostValue)
	}

	return body
}
