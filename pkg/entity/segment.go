package entity

import (
	"github.com/zhouzhuojie/conditions"
	"gorm.io/gorm"
)

// SegmentDefaultRank is the default rank when we create the segment
const SegmentDefaultRank = uint(999)

// Segment is the unit of segmentation
type Segment struct {
	gorm.Model
	FlagID         uint   `gorm:"index:idx_segment_flagid"`
	Description    string `gorm:"type:text"`
	Rank           uint
	RolloutPercent uint
	Constraints    ConstraintArray
	Distributions  []Distribution

	// Purely for evaluation
	SegmentEvaluation SegmentEvaluation `gorm:"-" json:"-"`
}

// PreloadConstraintsDistribution preloads constraints and distributions
// for segment
func PreloadConstraintsDistribution(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Distributions", func(db *gorm.DB) *gorm.DB {
			return db.Order("variant_id")
		}).
		Preload("Constraints", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at")
		})
}

// Preload preloads the segment
func (s *Segment) Preload(db *gorm.DB) error {
	return PreloadConstraintsDistribution(db).First(s, s.Model.ID).Error
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
			VariantIDs:          make([]uint, dLen),
			PercentsAccumulated: make([]int, dLen),
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
