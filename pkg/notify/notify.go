package notify

import (
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/sirupsen/logrus"
)

// A Notifier notifies about flags state
// It returns an error if unsuccessful
type Notifier interface {
	Notify(*entity.Flag, itemAction, itemType, subject) error
}

// Integration holds a notifier and a string name for that notifier
type Integration struct {
	notifier Notifier
	name     string
}

var integrations []Integration

func init() {
	client := NewClient()
	if config.Config.NotifyWebhookEnabled {
		if config.Config.NotifyWebhookHeaders != "" {
			client.Headers = http.Header{}
			pairs := strings.Split(config.Config.NotifyWebhookHeaders, ",")
			for _, v := range pairs {
				kv := strings.Split(v, "=")
				client.Headers[kv[0]] = []string{kv[1]}
			}
		}
		integrations = append(integrations, Integration{notifier: NewWebhook(client), name: "webhook"})
	}

}

type itemAction string

// Notification types
const (
	TOGGLED itemAction = "TOGGLED"
	CREATED itemAction = "CREATED"
	UPDATED itemAction = "UPDATED"
	DELETED itemAction = "DELETED"
)

type itemType string

// Thing being updated
const (
	FLAG         itemType = "FLAG"
	VARIANT      itemType = "VARIANT"
	SEGMENT      itemType = "SEGMENT"
	DISTRIBUTION itemType = "DISTRIBUTION"
	CONSTRAINT   itemType = "CONSTRAINT"
	TAG          itemType = "TAG"
)

type subject string

// All notifies all integrations, and logs an error if any fail
func All(db *gorm.DB, flagID uint, b itemAction, i itemType, s string) {
	f := &entity.Flag{}
	if err := db.First(f, flagID).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": flagID,
		}).Error("failed to find the flag when trying to notify integrations")
		return
	}
	err := f.Preload(db)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": f.ID,
		}).Error("failed to preload flag")
	}
	for _, integration := range integrations {
		fmt.Println(2)
		err := integration.notifier.Notify(f, b, i, subject(s))

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":         err,
				"flagID":      f.ID,
				"flagStatus":  f.Enabled,
				"integration": integration.name,
			}).Error("failed to notify integration")
		}
	}
}
