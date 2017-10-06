//go:generate goqueryset -in segment.go

package entity

import (
	"github.com/jinzhu/gorm"
)

// Segment is the unit of segmentation
// gen:qs
type Segment struct {
	gorm.Model

	FlagID         uint   `gorm:"index:idx_flagid"`
	Description    string `sql:"type:text"`
	Rank           uint
	RolloutPercent uint

	Constraints   []Constraint
	Distributions []Distribution
}

// Preload preloads the segment
func (s *Segment) Preload(db *gorm.DB) error {
	cs := []Constraint{}
	constraintQuery := NewConstraintQuerySet(db)
	err := constraintQuery.SegmentIDEq(s.ID).OrderAscByCreatedAt().All(&cs)
	if err != nil {
		return err
	}
	s.Constraints = cs

	ds := []Distribution{}
	distributionQuery := NewDistributionQuerySet(db)
	err = distributionQuery.SegmentIDEq(s.ID).OrderAscByVariantID().All(&ds)
	if err != nil {
		return err
	}
	s.Distributions = ds

	return nil
}
