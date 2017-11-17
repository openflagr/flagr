//go:generate goqueryset -in segment.go

package entity

import (
	"github.com/jinzhu/gorm"
	"github.com/zhouzhuojie/conditions"
)

// Segment is the unit of segmentation
// gen:qs
type Segment struct {
	gorm.Model
	FlagID         uint   `gorm:"index:idx_segment_flagid"`
	Description    string `sql:"type:text"`
	Rank           uint
	RolloutPercent uint
	Constraints    ConstraintArray
	Distributions  []Distribution

	// Purely for evaluation
	SegmentEvaluation SegmentEvaluation `gorm:"-" json:"-"`
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

// SegmentEvaluation is a struct that holds the necessary info for evaluation
type SegmentEvaluation struct {
	ConditionsExpr    conditions.Expr
	DistributionArray DistributionArray
}

// PrepareEvaluation prepares the segment for evaluation by parsing constraints
// and denormalize distributions
func (s *Segment) PrepareEvaluation() error {
	dLen := len(s.Distributions)
	se := SegmentEvaluation{
		DistributionArray: DistributionArray{
			VariantIDs:          make([]uint, dLen, dLen),
			PercentsAccumulated: make([]int, dLen, dLen),
		},
	}

	if len(s.Constraints) != 0 {
		expr, err := s.Constraints.ToExpr()
		if err != nil {
			return err
		}
		se.ConditionsExpr = expr
	}

	for i, d := range s.Distributions {
		se.DistributionArray.VariantIDs[i] = d.VariantID
		if i == 0 {
			se.DistributionArray.PercentsAccumulated[i] = int(d.Percent * PercentMultiplier)
		} else {
			se.DistributionArray.PercentsAccumulated[i] = se.DistributionArray.PercentsAccumulated[i-1] + int(d.Percent*PercentMultiplier)
		}
	}

	s.SegmentEvaluation = se
	return nil
}
