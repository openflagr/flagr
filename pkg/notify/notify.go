package notify

import (
	"net/http"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

const contentTypeJSON = "application/json"
const userAgentHeader = "checkr/flagr"

// A Notifier notifies about flags state
// It returns an error if unsuccessful
type Notifier interface {
	Notify(*entity.Flag, notify, itemType) error
}

// Integration holds a notifier and a string name for that notifier
type Integration struct {
	notifier Notifier
	name     string
}

var integrations []Integration

// NewClient returns a new http client
func NewClient() *http.Client {
	return &http.Client{}
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}

func init() {
	client := NewClient()
	if config.Config.WebhookEnabled {
		integrations = append(integrations, Integration{notifier: NewWebhook(client), name: "webhook"})
	}
	if config.Config.SlackWebhookEnabled {
		integrations = append(integrations, Integration{notifier: NewSlack(client), name: "slack"})
	}
}

type notify string

// Notification types
const (
	TOGGLED notify = "TOGGLED"
	CREATED        = "CREATED"
	UPDATED        = "UPDATED"
	DELETED        = "DELETED"
)

type itemType string

// Thing being updated
const (
	FLAG         itemType = "FLAG"
	VARIANT               = "VARIANT"
	SEGMENT               = "SEGMENT"
	DISTRIBUTION          = "DISTRIBUTION"
	CONSTRAINT            = "CONSTRAINT"
)

// All notifies all integrations, and logs an error if any fail
func All(db *gorm.DB, flagID uint, b notify, i itemType) {
	f := &entity.Flag{}
	if err := db.First(f, flagID).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"flagID": flagID,
		}).Error("failed to find the flag when trying to notify integrations")
		return
	}
	f.Preload(db)

	for _, integration := range integrations {
		err := integration.notifier.Notify(f, b, i)

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
