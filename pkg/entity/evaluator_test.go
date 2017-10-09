package entity

import (
	"testing"

	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/stretchr/testify/assert"
)

func TestSegmentPrepareEvaluation(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		s := Segment{
			FlagID:         0,
			Description:    "",
			Rank:           0,
			RolloutPercent: 0,
			Constraints: []Constraint{
				Constraint{
					SegmentID: 0,
					Property:  "dl_state",
					Operator:  models.ConstraintOperatorEQ,
					Value:     `"CA"`,
				},
			},
			Distributions: []Distribution{
				Distribution{
					SegmentID:  0,
					VariantID:  0,
					VariantKey: "control",
					Percent:    50,
				},
				Distribution{
					SegmentID:  0,
					VariantID:  1,
					VariantKey: "treatment",
					Percent:    50,
				},
			},
		}
		assert.NoError(t, s.PrepareEvaluation())
		assert.NotNil(t, s.SegmentEvaluation.Conditions)
		assert.NotNil(t, s.SegmentEvaluation.DistributionMap)
	})

	t.Run("error code path", func(t *testing.T) {
		s := Segment{
			FlagID:         0,
			Description:    "",
			Rank:           0,
			RolloutPercent: 0,
			Constraints: []Constraint{
				Constraint{
					SegmentID: 0,
					Property:  "dl_state",
					Operator:  models.ConstraintOperatorEQ,
					Value:     `"CA"]`, // invalid value
				},
			},
			Distributions: []Distribution{
				Distribution{
					SegmentID:  0,
					VariantID:  0,
					VariantKey: "control",
					Percent:    50,
				},
				Distribution{
					SegmentID:  0,
					VariantID:  1,
					VariantKey: "treatment",
					Percent:    50,
				},
			},
		}
		assert.Error(t, s.PrepareEvaluation())
		assert.Empty(t, s.SegmentEvaluation.Conditions)
		assert.Empty(t, s.SegmentEvaluation.DistributionMap)
	})
}
