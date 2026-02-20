package notification

import (
	"github.com/openflagr/flagr/pkg/config"
	"github.com/sirupsen/logrus"
)

// ValidateConfig checks notification configuration and logs warnings if misconfigured.
// This should be called during application startup when notifications are enabled.
func ValidateConfig() {
	if !config.Config.NotificationEnabled {
		return
	}

	provider := config.Config.NotificationProvider
	var configured bool

	switch provider {
	case "slack":
		configured = config.Config.NotificationSlackWebhookURL != ""
		if !configured {
			logrus.Warn("Notifications are enabled with provider 'slack', but FLAGR_NOTIFICATION_SLACK_WEBHOOK_URL is not set. Notifications will be silently dropped.")
		}
	case "email":
		configured = config.Config.NotificationEmailURL != "" && config.Config.NotificationEmailTo != "" && config.Config.NotificationEmailFrom != ""
		if !configured {
			logrus.Warn("Notifications are enabled with provider 'email', but FLAGR_NOTIFICATION_EMAIL_URL, FLAGR_NOTIFICATION_EMAIL_TO, and FLAGR_NOTIFICATION_EMAIL_FROM must all be set. Notifications will be silently dropped.")
		}
	case "webhook":
		configured = config.Config.NotificationWebhookURL != ""
		if !configured {
			logrus.Warn("Notifications are enabled with provider 'webhook', but FLAGR_NOTIFICATION_WEBHOOK_URL is not set. Notifications will be silently dropped.")
		}
	default:
		logrus.Warnf("Unknown notification provider: %s. Notifications will be silently dropped.", provider)
	}
}
