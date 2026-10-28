package handler

import (
	"strings"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validFlagWithEnrichers(enrichers []entity.Enricher) entity.Flag {
	return entity.Flag{
		Key:       "enriched",
		Enabled:   true,
		Enrichers: enrichers,
		Variants:  []entity.Variant{{Key: "on"}, {Key: "off"}},
		Segments: []entity.Segment{{
			Description:    "all",
			RolloutPercent: 100,
			Distributions: []entity.Distribution{
				{VariantKey: "on", Percent: 50},
				{VariantKey: "off", Percent: 50},
			},
		}},
	}
}

func jevEnricher(t *testing.T) entity.Enricher {
	t.Helper()
	e := entity.Enricher{}
	require.NoError(t, e.SetConfig(entity.EnricherNamespaceJev, &JevEnricherConfig{
		Questions: map[string]JevQuestion{
			"plan_tier": {Type: JevTypeChoice, Instructions: "plan?", Criteria: map[string]any{"pro": "Pro"}},
		},
	}))
	return e
}

func TestValidateFlagsEnrichersValid(t *testing.T) {
	t.Parallel()

	f := validFlagWithEnrichers([]entity.Enricher{jevEnricher(t)})
	f.Segments[0].Constraints = entity.ConstraintArray{
		{Property: "@jev_plan_tier", Operator: "EQ", Value: `"pro"`},
	}

	r := ValidateFlags([]entity.Flag{f})
	assert.True(t, r.OK(), "errors: %v", r.Errors)
	assert.False(t, r.HasWarnings(), "warnings: %v", r.Warnings)
}

func TestValidateFlagsEnricherErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid config", func(t *testing.T) {
		t.Parallel()
		f := validFlagWithEnrichers([]entity.Enricher{{Namespace: entity.EnricherNamespaceJev, ConfigJSON: `{"questions":{}}`}})
		r := ValidateFlags([]entity.Flag{f})
		assert.False(t, r.OK())
		assert.Contains(t, strings.Join(r.Errors, "\n"), "enricher \"jev\" is invalid")
	})

	t.Run("unknown namespace", func(t *testing.T) {
		t.Parallel()
		f := validFlagWithEnrichers([]entity.Enricher{{Namespace: "nope", ConfigJSON: `{}`}})
		r := ValidateFlags([]entity.Flag{f})
		assert.False(t, r.OK())
		assert.Contains(t, strings.Join(r.Errors, "\n"), "unknown enricher namespace")
	})

	t.Run("duplicate namespace", func(t *testing.T) {
		t.Parallel()
		f := validFlagWithEnrichers([]entity.Enricher{jevEnricher(t), jevEnricher(t)})
		r := ValidateFlags([]entity.Flag{f})
		assert.False(t, r.OK())
		assert.Contains(t, strings.Join(r.Errors, "\n"), "duplicate enricher namespace")
	})
}

func TestValidateFlagsEnricherDanglingReferenceWarns(t *testing.T) {
	t.Parallel()

	f := validFlagWithEnrichers([]entity.Enricher{jevEnricher(t)})
	f.Segments[0].Constraints = entity.ConstraintArray{
		{Property: "@jev_typo", Operator: "EQ", Value: `"pro"`},
	}

	r := ValidateFlags([]entity.Flag{f})
	assert.True(t, r.OK(), "warn-only, not an error: %v", r.Errors)
	require.Len(t, r.Warnings, 1)
	assert.Contains(t, r.Warnings[0], "@jev_typo")
}

func TestValidateEnricher(t *testing.T) {
	t.Parallel()

	e := jevEnricher(t)
	require.NoError(t, validateEnricher(&e))
	assert.Error(t, validateEnricher(&entity.Enricher{Namespace: "nope", ConfigJSON: `{}`}))
	assert.Error(t, validateEnricher(&entity.Enricher{Namespace: entity.EnricherNamespaceTs, ConfigJSON: `{}`}),
		"global namespace is not flag-scoped")
	assert.Error(t, validateEnricher(nil))
}
