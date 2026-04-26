package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sirupsen/logrus"
)

var (
	// Semaphore to limit concurrent notification sends. Default 100.
	notificationSemaphore = make(chan struct{}, 100)
)

func recordNotificationMetrics(provider string, operation Operation, success bool) {
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
		fmt.Sprintf("status:%s", status),
	}
	config.Global.StatsdClient.Incr("notification.sent", tags, 1)
}

// SendNotification dispatches a notification to all configured notifiers asynchronously.
// Notifications are sent in a background goroutine and failures do not affect the caller.
func SendNotification(n Notification) {
	// Capture notifiers BEFORE spawning goroutine to avoid test pollution
	// when Notifiers is modified between test runs
	notifiers := GetNotifiers()
	if len(notifiers) == 0 {
		return
	}

	// Set timestamp if not already set by caller
	if n.Timestamp.IsZero() {
		n.Timestamp = time.Now().UTC()
	}

	go func() {
		// Acquire semaphore slot
		notificationSemaphore <- struct{}{}
		defer func() {
			<-notificationSemaphore
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("panic in SendNotification")
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), config.Config.NotificationTimeout)
		defer cancel()

		// Send to all notifiers concurrently, aggregate errors
		var (
			wg   sync.WaitGroup
			mu   sync.Mutex
			errs []error
		)

		for _, nr := range notifiers {
			wg.Add(1)
			go func(notifier Notifier) {
				defer wg.Done()
				err := notifier.Send(ctx, n)
				recordNotificationMetrics(notifier.Name(), n.Operation, err == nil)
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("%s: %w", notifier.Name(), err))
					mu.Unlock()
				}
			}(nr)
		}

		wg.Wait()

		if len(errs) > 0 {
			logrus.WithFields(logrus.Fields{
				"operation": n.Operation,
				"flagID":    n.FlagID,
				"errors":    errs,
			}).Warn("failed to send notifications to some providers")
		}
	}()
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
