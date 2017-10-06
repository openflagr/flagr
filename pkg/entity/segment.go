//go:generate goqueryset -in segment.go

package entity

import "github.com/jinzhu/gorm"

// Segment is the unit of segmentation
// gen:qs
type Segment struct {
	gorm.Model

	FlagID         uint   `gorm:"index:idx_flagid"`
	Description    string `sql:"type:text"`
	Rank           uint
	RolloutPercent uint

	Constraints  []*Constraint
	Distribution []*Distribution
}
