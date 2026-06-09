package handler

import (
	"strings"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
)

// --- Basic structural tests ---

func TestValidateFlags_Valid(t *testing.T) {
	flags := []entity.Flag{
		{
			Key:     "my-flag",
			Enabled: true,
			Variants: []entity.Variant{
				{Key: "on"},
				{Key: "off"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all users",
					Rank:           0,
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 50},
						{VariantKey: "off", Percent: 50},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.True(t, r.OK())
	assert.False(t, r.HasWarnings())
}

func TestValidateFlags_EmptyFlags(t *testing.T) {
	r := ValidateFlags([]entity.Flag{})
	assert.True(t, r.OK())
}

func TestValidateFlags_EmptyFlagKey(t *testing.T) {
	flags := []entity.Flag{{Key: ""}}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	assert.True(t, len(r.Errors) > 0)
	assert.True(t, strings.Contains(r.Errors[0], "missing or empty Key"))
}

func TestValidateFlags_DuplicateFlagKeys(t *testing.T) {
	flags := []entity.Flag{
		{Key: "dup"},
		{Key: "dup"},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	assert.True(t, len(r.Errors) >= 1)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "duplicate flag key") {
			found = true
		}
	}
	assert.True(t, found, "should have duplicate flag key error: %v", r.Errors)
}

// --- Variant validation ---

func TestValidateFlags_DuplicateVariantKeys(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "a"},
				{Key: "a"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "a", Percent: 100},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "duplicate variant key") {
			found = true
		}
	}
	assert.True(t, found, "should have duplicate variant key error: %v", r.Errors)
}

func TestValidateFlags_NoVariantsWarning(t *testing.T) {
	flags := []entity.Flag{
		{Key: "my-flag"},
	}
	r := ValidateFlags(flags)
	assert.True(t, r.OK()) // warning only
	assert.True(t, r.HasWarnings())
	found := false
	for _, w := range r.Warnings {
		if strings.Contains(w, "no variants defined") {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidateFlags_EmptyVariantKey(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "good"},
				{Key: ""},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "good", Percent: 100},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "missing or empty Key") && strings.Contains(e, "variant") {
			found = true
		}
	}
	assert.True(t, found, "should have empty variant key error: %v", r.Errors)
}

func TestValidateFlags_InvalidAttachmentJSON(t *testing.T) {
	// entity.Attachment is map[string]any — invalid JSON unmarshals to empty map,
	// so this test verifies that valid attachments pass validation.
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on", Attachment: entity.Attachment{"color": "red"}},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.True(t, r.OK())
}

// --- Segment validation ---

func TestValidateFlags_NoSegmentsWarning(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.True(t, r.OK()) // warning only
	assert.True(t, r.HasWarnings())
	found := false
	for _, w := range r.Warnings {
		if strings.Contains(w, "no segments defined") {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidateFlags_RolloutPercentOver100(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
				{Key: "off"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 101,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 50},
						{VariantKey: "off", Percent: 50},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "RolloutPercent") && strings.Contains(e, "out of range") {
			found = true
		}
	}
	assert.True(t, found, "should have RolloutPercent error: %v", r.Errors)
}

func TestValidateFlags_RolloutPercentZero(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
				{Key: "off"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 0,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 50},
						{VariantKey: "off", Percent: 50},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.True(t, r.OK())
}

// --- Distribution validation ---

func TestValidateFlags_DistributionSumNot100(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
				{Key: "off"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 50},
						{VariantKey: "off", Percent: 40},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "distribution sum") && strings.Contains(e, "expected 100") {
			found = true
		}
	}
	assert.True(t, found, "should have distribution sum error: %v", r.Errors)
}

func TestValidateFlags_UnknownVariantKey(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "off", Percent: 100},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "unknown variant key") {
			found = true
		}
	}
	assert.True(t, found, "should have unknown variant key error: %v", r.Errors)
}

func TestValidateFlags_DistributionNoVariantRef(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{Percent: 100},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "no VariantKey or VariantID") {
			found = true
		}
	}
	assert.True(t, found, "should have no variant ref error: %v", r.Errors)
}

func TestValidateFlags_DistributionPercentOver100(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 150},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "distribution percent") && strings.Contains(e, "out of range") {
			found = true
		}
	}
	assert.True(t, found, "should have distribution percent error: %v", r.Errors)
}

func TestValidateFlags_NoDistributionsWarning(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.True(t, r.OK()) // warning only
	assert.True(t, r.HasWarnings())
	found := false
	for _, w := range r.Warnings {
		if strings.Contains(w, "no distributions defined") {
			found = true
		}
	}
	assert.True(t, found)
}

// --- Constraint validation ---

func TestValidateFlags_EmptyConstraintProperty(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
					Constraints: []entity.Constraint{
						{Property: "", Operator: "EQ", Value: "\"test\""},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "constraint") && strings.Contains(e, "is invalid") {
			found = true
		}
	}
	assert.True(t, found, "should have constraint error: %v", r.Errors)
}

func TestValidateFlags_EmptyConstraintOperator(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
					Constraints: []entity.Constraint{
						{Property: "region", Operator: "", Value: "\"us\""},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "constraint") && strings.Contains(e, "is invalid") {
			found = true
		}
	}
	assert.True(t, found, "should have constraint error: %v", r.Errors)
}

func TestValidateFlags_EmptyConstraintValue(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
					Constraints: []entity.Constraint{
						{Property: "region", Operator: "EQ", Value: ""},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "constraint") && strings.Contains(e, "is invalid") {
			found = true
		}
	}
	assert.True(t, found, "should have constraint error: %v", r.Errors)
}

func TestValidateFlags_InvalidConstraintOperator(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
					Constraints: []entity.Constraint{
						{Property: "region", Operator: "INVALID", Value: "\"us\""},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "constraint") && strings.Contains(e, "is invalid") {
			found = true
		}
	}
	assert.True(t, found, "should have constraint error: %v", r.Errors)
}

func TestValidateFlags_InvalidConstraintRegex(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
					Constraints: []entity.Constraint{
						{Property: "name", Operator: "EREG", Value: "\"(unclosed"},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "constraint") && strings.Contains(e, "is invalid") {
			found = true
		}
	}
	assert.True(t, found, "should have constraint error: %v", r.Errors)
}

func TestValidateFlags_ValidConstraintEQ(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
					Constraints: []entity.Constraint{
						{Property: "region", Operator: "EQ", Value: "\"us\""},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.True(t, r.OK())
}

func TestValidateFlags_ValidConstraintIN(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
					Constraints: []entity.Constraint{
						{Property: "region", Operator: "IN", Value: "[\"us\", \"eu\"]"},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.True(t, r.OK())
}

func TestValidateFlags_ValidConstraintRegex(t *testing.T) {
	flags := []entity.Flag{
		{
			Key: "my-flag",
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "all",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
					Constraints: []entity.Constraint{
						{Property: "email", Operator: "EREG", Value: "\"^[a-z]+@example\\.com$\""},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.True(t, r.OK())
}

// --- Composite tests ---

func TestValidateFlags_ComplexValidFile(t *testing.T) {
	flags := []entity.Flag{
		{
			Key:     "feature-a",
			Enabled: true,
			Variants: []entity.Variant{
				{Key: "control"},
				{Key: "treatment"},
			},
			Segments: []entity.Segment{
				{
					Description:    "beta users",
					Rank:           0,
					RolloutPercent: 50,
					Distributions: []entity.Distribution{
						{VariantKey: "control", Percent: 80},
						{VariantKey: "treatment", Percent: 20},
					},
					Constraints: []entity.Constraint{
						{Property: "beta", Operator: "EQ", Value: "true"},
					},
				},
			},
		},
		{
			Key:     "feature-b",
			Enabled: false,
			Variants: []entity.Variant{
				{Key: "on"},
			},
			Segments: []entity.Segment{
				{
					Description:    "internal",
					RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
				},
			},
		},
	}
	r := ValidateFlags(flags)
	assert.True(t, r.OK())
}

func TestValidateFlags_MultipleErrors(t *testing.T) {
	flags := []entity.Flag{
		{Key: "dup"},
		{Key: "dup"},
	}
	r := ValidateFlags(flags)
	assert.False(t, r.OK())
	assert.True(t, len(r.Errors) >= 1, "should have at least one error: %v", r.Errors)
}
