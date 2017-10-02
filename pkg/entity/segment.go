package entity

import "github.com/jinzhu/gorm"

// Segment is the unit of segmentation
type Segment struct {
	gorm.Model
	FlagID uint

	Description string
	Constraints []*Constraint
}
