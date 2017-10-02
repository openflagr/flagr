package entity

import "github.com/jinzhu/gorm"

// Flag is the unit of flags
type Flag struct {
	gorm.Model
	Description string
	CreatedBy   string
	UpdatedBy   string

	Segments    []*Segment
	Experiences []*Experience
}
