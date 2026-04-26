package notification

import (
	"context"
	"sync"
	"time"

	"github.com/openflagr/flagr/pkg/config"
)

type Notifier interface {
	Send(ctx context.Context, n Notification) error
	Name() string
}

type Operation string

const (
	OperationCreate  Operation = "create"
	OperationUpdate  Operation = "update"
	OperationDelete  Operation = "delete"
	OperationRestore Operation = "restore"
)

// ComponentType identifies which part of a flag was modified.
type ComponentType string

const (
	ComponentFlag         ComponentType = "flag"
	ComponentSegment      ComponentType = "segment"
	ComponentVariant      ComponentType = "variant"
	ComponentConstraint   ComponentType = "constraint"
	ComponentDistribution ComponentType = "distribution"
	ComponentTag          ComponentType = "tag"
)

type Notification struct {
	Operation     Operation `json:"operation"`
	FlagID        uint      `json:"flag_id"`
	FlagKey       string    `json:"flag_key"`
	ComponentType ComponentType `json:"component_type,omitempty"`
	ComponentID   uint      `json:"component_id,omitempty"`
	ComponentKey  string    `json:"component_key,omitempty"`
	PreValue      string    `json:"pre_value,omitempty"`
	PostValue     string    `json:"post_value,omitempty"`
	Diff          string    `json:"diff,omitempty"`
	User          string    `json:"user,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

var (
	// Notifiers is the list of configured notifiers. Set directly for testing.
	Notifiers []Notifier
	once      sync.Once
)

// GetNotifiers returns the list of configured notifiers.
// It initializes the notifiers on first call using sync.Once.
// For testing, set Notifiers directly before calling GetNotifiers.
func GetNotifiers() []Notifier {
	// If already set (e.g., by tests), return immediately
	if len(Notifiers) > 0 {
		return Notifiers
	}

	once.Do(func() {
		if config.Config.NotificationWebhookEnabled {
			if wn := NewWebhookNotifier(); wn != nil {
				Notifiers = append(Notifiers, wn)
			}
		}
	})

	return Notifiers
}

type nullNotifier struct{}

func (n *nullNotifier) Send(ctx context.Context, notification Notification) error {
	return nil
}

func (n *nullNotifier) Name() string {
	return "null"
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

func (m *MockNotifier) Name() string {
	return "mock"
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
