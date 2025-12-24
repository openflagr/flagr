package notification

import (
	"context"
	"fmt"

	notify "github.com/nikoksr/notify"
	notifySlack "github.com/nikoksr/notify/service/slack"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/sirupsen/logrus"
)

type slackNotifier struct {
	client *notify.Notify
}

func NewSlackNotifier() Notifier {
	if config.Config.NotificationSlackWebhookURL == "" {
		logrus.Warn("NotificationSlackWebhookURL is empty, using null notifier")
		return &nullNotifier{}
	}

	slackService := notifySlack.New(config.Config.NotificationSlackWebhookURL)

	if config.Config.NotificationSlackChannel != "" {
		slackService.AddReceivers(config.Config.NotificationSlackChannel)
	}

	n := notify.New()
	n.UseServices(slackService)

	return &slackNotifier{client: n}
}

func (s *slackNotifier) Send(ctx context.Context, n Notification) error {
	subject := fmt.Sprintf("%s %s", n.Operation, n.EntityType)
	message := formatNotification(n)
	return s.client.Send(ctx, subject, message)
}

func formatNotification(n Notification) string {
	var emoji string
	switch n.Operation {
	case OperationCreate:
		emoji = ":rocket:"
	case OperationUpdate:
		emoji = ":pencil2:"
	case OperationDelete:
		emoji = ":wastebasket:"
	default:
		emoji = ":information_source:"
	}

	userInfo := "anonymous"
	if n.User != "" {
		userInfo = n.User
	}

	msg := fmt.Sprintf("%s *%s %s*\n", emoji, n.Operation, n.EntityType)
	msg += fmt.Sprintf("*Key:* %s\n", n.EntityKey)
	msg += fmt.Sprintf("*ID:* %d\n", n.EntityID)
	if n.Description != "" {
		msg += fmt.Sprintf("*Description:* %s\n", n.Description)
	}
	msg += fmt.Sprintf("*User:* %s\n", userInfo)

	if n.Diff != "" {
		msg += fmt.Sprintf("*Diff:*\n```diff\n%s\n```\n", n.Diff)
	}

	if n.PreValue != "" {
		msg += fmt.Sprintf("*Pre-value:*\n```json\n%s\n```\n", n.PreValue)
	}
	if n.PostValue != "" {
		msg += fmt.Sprintf("*Post-value:*\n```json\n%s\n```\n", n.PostValue)
	}

	return msg
}
