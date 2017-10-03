//go:generate goqueryset -in variant.go

package entity

import "github.com/jinzhu/gorm"

// Variant is the struct that represent the experience/variant of the evaluation entity
// gen:qs
type Variant struct {
	gorm.Model
	FlagID uint

	Key        string
	Attachment map[string]string
}
