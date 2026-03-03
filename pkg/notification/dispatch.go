package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sirupsen/logrus"
)

var (
	// Semaphore to limit concurrent notification sends. Default 100.
	notificationSemaphore = make(chan struct{}, 100)
)

func recordNotificationMetrics(provider string, operation Operation, entityType EntityType, success bool) {
	if config.Global.StatsdClient == nil {
		return
	}
	status := "failure"
	if success {
		status = "success"
	}
	tags := []string{
		fmt.Sprintf("provider:%s", provider),
		fmt.Sprintf("operation:%s", operation),
		fmt.Sprintf("entity_type:%s", entityType),
		fmt.Sprintf("status:%s", status),
	}
	config.Global.StatsdClient.Incr("notification.sent", tags, 1)
}

func sendNotification(operation Operation, entityType EntityType, entityID uint, entityKey string, description string, preValue string, postValue string, diff string, user string) {
	// Capture notifiers BEFORE spawning goroutine to avoid test pollution
	// when Notifiers is modified between test runs
	notifiers := GetNotifiers()
	if len(notifiers) == 0 {
		return
	}

	go func() {
		// Acquire semaphore slot
		notificationSemaphore <- struct{}{}
		defer func() {
			<-notificationSemaphore
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("panic in sendNotification")
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), config.Config.NotificationTimeout)
		defer cancel()

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

		// Send to all notifiers concurrently, aggregate errors
		var (
			wg   sync.WaitGroup
			mu   sync.Mutex
			errs []error
		)

		for _, notifier := range notifiers {
			wg.Add(1)
			go func(n Notifier) {
				defer wg.Done()
				err := n.Send(ctx, notif)
				recordNotificationMetrics(n.Name(), operation, entityType, err == nil)
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("%s: %w", n.Name(), err))
					mu.Unlock()
				}
			}(notifier)
		}

		wg.Wait()

		if len(errs) > 0 {
			logrus.WithFields(logrus.Fields{
				"operation":  operation,
				"entityType": entityType,
				"entityID":   entityID,
				"errors":     errs,
			}).Warn("failed to send notifications to some providers")
		}
	}()
}

func SendFlagNotification(operation Operation, flagID uint, flagKey string, description string, preValue string, postValue string, diff string, user string) {
	sendNotification(operation, EntityTypeFlag, flagID, flagKey, description, preValue, postValue, diff, user)
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
