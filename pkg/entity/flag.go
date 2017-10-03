//go:generate goqueryset -in flag.go

package entity

import "github.com/jinzhu/gorm"

// Flag is the unit of flags
// gen:qs
type Flag struct {
	gorm.Model
	Description string
	CreatedBy   string
	UpdatedBy   string

	Segments []*Segment
	Variants []*Variant
}
