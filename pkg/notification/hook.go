package notification

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sirupsen/logrus"
)

func SendNotification(operation Operation, entityType EntityType, entityID uint, entityKey string, description string, preValue string, postValue string, diff string, user string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("panic in SendNotification")
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), config.Config.NotificationTimeout)
		defer cancel()
		notifier := GetNotifier()

		notif := Notification{
			Operation:   operation,
			EntityType:  entityType,
			EntityID:    entityID,
			EntityKey:   entityKey,
			Description: description,
			PreValue:    preValue,
			PostValue:   postValue,
			Diff:        diff,
			User:        user,
			Details:     make(map[string]any),
		}

		if err := notifier.Send(ctx, notif); err != nil {
			logrus.WithFields(logrus.Fields{
				"operation":  operation,
				"entityType": entityType,
				"entityID":   entityID,
				"error":      err,
			}).Warn("failed to send notification")
		}
	}()
}

func SendFlagNotification(operation Operation, flagID uint, flagKey string, description string, preValue string, postValue string, diff string, user string) {
	SendNotification(operation, EntityTypeFlag, flagID, flagKey, description, preValue, postValue, diff, user)
}

func CalculateDiff(pre, post string) string {
	if pre == "" || post == "" {
		return ""
	}

	prePretty := prettyPrintJSON(pre)
	postPretty := prettyPrintJSON(post)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(prePretty),
		B:        difflib.SplitLines(postPretty),
		FromFile: "Previous",
		ToFile:   "Current",
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	return text
}

func prettyPrintJSON(s string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(s), "", "  ")
	if err != nil {
		return s
	}
	return out.String()
}
