package entity

import (
	"gorm.io/gorm"
)

// Webhook represents a global webhook configuration for system events
type Webhook struct {
	gorm.Model
	URL         string `gorm:"type:text"`
	Description string `gorm:"type:text"`
	Events      string `gorm:"type:text"` // Comma-separated list of events to trigger on
	Secret      string `gorm:"type:text"` // Optional secret for webhook signature
	Enabled     bool   `gorm:"default:true"`
}

// WebhookEvent represents a webhook event that was triggered
type WebhookEvent struct {
	gorm.Model
	WebhookID uint   `gorm:"index:idx_webhookevent_webhookid"`
	Event     string `gorm:"type:text"`
	Payload   string `gorm:"type:text"`
	Status    string `gorm:"type:text"` // success, failed, pending
	Error     string `gorm:"type:text"` // Error message if status is failed
}
