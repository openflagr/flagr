package notify

import (
	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/jinzhu/gorm"
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
		integrations = append(integrations, Integration{notifier: NewWebhook(client), name: "webhook"})
	}
	if config.Config.NotifySlackWebhookEnabled {
		integrations = append(integrations, Integration{notifier: NewSlack(client), name: "slack"})
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
