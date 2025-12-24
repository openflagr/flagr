package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

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

		ctx := context.Background()
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

func SendSegmentNotification(operation Operation, segmentID uint, flagID uint, user string) {
	key := fmt.Sprintf("segment-%d-of-flag-%d", segmentID, flagID)
	SendNotification(operation, EntityTypeSegment, segmentID, key, "", "", "", "", user)
}

func SendVariantNotification(operation Operation, variantID uint, flagID uint, variantKey, user string) {
	key := fmt.Sprintf("variant-%s-of-flag-%d", variantKey, flagID)
	SendNotification(operation, EntityTypeVariant, variantID, key, "", "", "", "", user)
}

func SendConstraintNotification(operation Operation, constraintID uint, segmentID uint, flagID uint, user string) {
	key := fmt.Sprintf("constraint-%d-of-segment-%d", constraintID, segmentID)
	SendNotification(operation, EntityTypeConstraint, constraintID, key, "", "", "", "", user)
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
