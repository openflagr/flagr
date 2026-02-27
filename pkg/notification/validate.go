package notification

import (
	"github.com/openflagr/flagr/pkg/config"
	"github.com/sirupsen/logrus"
)

// ValidateConfig checks notification configuration and logs warnings if misconfigured.
// This should be called during application startup.
func ValidateConfig() {
	if config.Config.NotificationSlackEnabled {
		if config.Config.NotificationSlackWebhookURL == "" {
			logrus.Warn("Slack notifications are enabled, but FLAGR_NOTIFICATION_SLACK_WEBHOOK_URL is not set. Slack notifications will be silently dropped.")
		}
	}

	if config.Config.NotificationEmailEnabled {
		if config.Config.NotificationEmailURL == "" || config.Config.NotificationEmailTo == "" || config.Config.NotificationEmailFrom == "" {
			logrus.Warn("Email notifications are enabled, but FLAGR_NOTIFICATION_EMAIL_URL, FLAGR_NOTIFICATION_EMAIL_TO, and FLAGR_NOTIFICATION_EMAIL_FROM should all be set. Email notifications may fail.")
		}
	}

	if config.Config.NotificationWebhookEnabled {
		if config.Config.NotificationWebhookURL == "" {
			logrus.Warn("Webhook notifications are enabled, but FLAGR_NOTIFICATION_WEBHOOK_URL is not set. Webhook notifications will be silently dropped.")
		}
	}
}
