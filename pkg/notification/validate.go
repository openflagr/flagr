package notification

import (
	"github.com/openflagr/flagr/pkg/config"
	"github.com/sirupsen/logrus"
)

// ValidateConfig checks notification configuration and logs warnings if misconfigured.
// This should be called during application startup.
func ValidateConfig() {
	if config.Config.NotificationWebhookEnabled {
		if config.Config.NotificationWebhookURL == "" {
			logrus.Warn("Webhook notifications are enabled, but FLAGR_NOTIFICATION_WEBHOOK_URL is not set. Webhook notifications will be silently dropped.")
		}
	}
}
