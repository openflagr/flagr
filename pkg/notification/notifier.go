package notification

import (
	"context"
	"sync"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/sirupsen/logrus"
)

type Notifier interface {
	Send(ctx context.Context, n Notification) error
}

type Operation string

const (
	OperationCreate  Operation = "create"
	OperationUpdate  Operation = "update"
	OperationDelete  Operation = "delete"
	OperationRestore Operation = "restore"
)

type EntityType string

const (
	EntityTypeFlag       EntityType = "flag"
	EntityTypeSegment    EntityType = "segment"
	EntityTypeVariant    EntityType = "variant"
	EntityTypeConstraint EntityType = "constraint"
	EntityTypeTag        EntityType = "tag"
)

type Notification struct {
	Operation   Operation
	EntityType  EntityType
	EntityID    uint
	EntityKey   string
	Description string
	PreValue    string
	PostValue   string
	Diff        string
	User        string
	Details     map[string]any
}

var (
	singletonNotifier Notifier
	once              sync.Once
)

// SetNotifier sets the global notifier, useful for testing
func SetNotifier(n Notifier) {
	singletonNotifier = n
}

func GetNotifier() Notifier {
	if singletonNotifier != nil {
		return singletonNotifier
	}

	once.Do(func() {
		if !config.Config.NotificationEnabled {
			singletonNotifier = &nullNotifier{}
			return
		}

		switch config.Config.NotificationProvider {
		case "slack":
			singletonNotifier = NewSlackNotifier()
		case "email":
			singletonNotifier = NewEmailNotifier()
		default:
			logrus.Warnf("unknown notification provider: %s, using null notifier", config.Config.NotificationProvider)
			singletonNotifier = &nullNotifier{}
		}
	})

	if singletonNotifier == nil {
		return &nullNotifier{}
	}

	return singletonNotifier
}

type nullNotifier struct{}

func (n *nullNotifier) Send(ctx context.Context, notification Notification) error {
	return nil
}

type MockNotifier struct {
	sent      []Notification
	mu        sync.Mutex
	sendError error
}

func NewMockNotifier() *MockNotifier {
	return &MockNotifier{
		sent: make([]Notification, 0),
	}
}

func (m *MockNotifier) Send(ctx context.Context, n Notification) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, n)
	return m.sendError
}

func (m *MockNotifier) SetSendError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendError = err
}

func (m *MockNotifier) GetSentNotifications() []Notification {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]Notification, len(m.sent))
	copy(result, m.sent)
	return result
}

func (m *MockNotifier) ClearSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = make([]Notification, 0)
}
