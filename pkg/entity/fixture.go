package entity

import (
	"github.com/openflagr/flagr/swagger_gen/models"

	"gorm.io/gorm"
)

// GenFixtureFlag is a fixture
func GenFixtureFlag() Flag {
	f := Flag{
		Model:       gorm.Model{ID: 100},
		Key:         "flag_key_100",
		Description: "",
		Enabled:     true,
		Segments:    []Segment{GenFixtureSegment()},
		Variants: []Variant{
			{
				Model:  gorm.Model{ID: 300},
				FlagID: 100,
				Key:    "control",
			},
			{
				Model:  gorm.Model{ID: 301},
				FlagID: 100,
				Key:    "treatment",
				Attachment: map[string]any{
					"value": "321",
				},
			},
		},
		Tags: []Tag{
			{
				Value: "tag1",
			},
			{
				Value: "tag2",
			},
		},
	}
	f.PrepareEvaluation()
	return f
}

// GenFixtureSegment is a fixture
func GenFixtureSegment() Segment {
	s := Segment{
		Model:          gorm.Model{ID: 200},
		FlagID:         100,
		Description:    "",
		Rank:           0,
		RolloutPercent: 100,
		Constraints: []Constraint{
			{
				Model:     gorm.Model{ID: 500},
				SegmentID: 200,
				Property:  "dl_state",
				Operator:  models.ConstraintOperatorEQ,
				Value:     `"CA"`,
			},
		},
		Distributions: []Distribution{
			{
				Model:      gorm.Model{ID: 400},
				SegmentID:  200,
				VariantID:  300,
				VariantKey: "control",
				Percent:    50,
			},
			{
				Model:      gorm.Model{ID: 401},
				SegmentID:  200,
				VariantID:  301,
				VariantKey: "treatment",
				Percent:    50,
			},
		},
	}
	s.PrepareEvaluation()
	return s
}
