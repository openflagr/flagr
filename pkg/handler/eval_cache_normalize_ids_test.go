package handler

import (
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNormalizeIDs_AllZero(t *testing.T) {
	t.Parallel()
	// Core use case: hand-edited file with no IDs at all
	flags := []entity.Flag{
		{
			Key: "flag-a", Enabled: true,
			Variants: []entity.Variant{
				{Key: "on", Attachment: entity.Attachment{"value": "1"}},
				{Key: "off", Attachment: entity.Attachment{"value": "0"}},
			},
			Segments: []entity.Segment{
				{
					Description: "all users", Rank: 0, RolloutPercent: 100,
					Constraints: []entity.Constraint{
						{Property: "country", Operator: "EQ", Value: "\"US\""},
					},
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 50},
						{VariantKey: "off", Percent: 50},
					},
				},
			},
			Tags: []entity.Tag{{Value: "frontend"}},
		},
		{
			Key: "flag-b", Enabled: false,
			Variants: []entity.Variant{{Key: "control"}},
			Segments: []entity.Segment{
				{
					Description: "beta users", Rank: 0, RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "control", Percent: 100},
					},
				},
			},
		},
	}

	normalizeIDs(flags)

	// Flags: globally unique
	assert.Equal(t, uint(1), flags[0].ID)
	assert.Equal(t, uint(2), flags[1].ID)

	// Variants: globally unique across flags
	assert.Equal(t, uint(1), flags[0].Variants[0].ID)
	assert.Equal(t, uint(2), flags[0].Variants[1].ID)
	assert.Equal(t, uint(3), flags[1].Variants[0].ID) // continues from flag-a

	// Segments: globally unique
	assert.Equal(t, uint(1), flags[0].Segments[0].ID)
	assert.Equal(t, uint(2), flags[1].Segments[0].ID)

	// FlagID set on segments
	assert.Equal(t, flags[0].ID, flags[0].Segments[0].FlagID)
	assert.Equal(t, flags[1].ID, flags[1].Segments[0].FlagID)

	// Constraints: globally unique
	assert.Equal(t, uint(1), flags[0].Segments[0].Constraints[0].ID)
	assert.Equal(t, flags[0].Segments[0].ID, flags[0].Segments[0].Constraints[0].SegmentID)

	// Distributions: globally unique, VariantID resolved from VariantKey
	assert.Equal(t, flags[0].Variants[0].ID, flags[0].Segments[0].Distributions[0].VariantID)
	assert.Equal(t, flags[0].Variants[1].ID, flags[0].Segments[0].Distributions[1].VariantID)
	assert.Equal(t, flags[1].Variants[0].ID, flags[1].Segments[0].Distributions[0].VariantID)

	// Tags: globally unique
	assert.Equal(t, uint(1), flags[0].Tags[0].ID)
}

func TestNormalizeIDs_ExplicitIDsPreserved(t *testing.T) {
	t.Parallel()
	flags := []entity.Flag{
		{
			Key: "flag-a", Enabled: true,
			Variants: []entity.Variant{
				{Model: gorm.Model{ID: 42}, Key: "on"}, // explicit
				{Key: "off"},                           // auto → 43
			},
			Segments: []entity.Segment{
				{
					Model: gorm.Model{ID: 10}, Description: "all users", Rank: 0, RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "on", Percent: 100},
					},
				},
			},
		},
	}

	normalizeIDs(flags)

	// Flag gets auto ID (no explicit ID)
	assert.Equal(t, uint(1), flags[0].ID)

	// Explicit variant ID preserved; next skips past it
	assert.Equal(t, uint(42), flags[0].Variants[0].ID)
	assert.Equal(t, uint(43), flags[0].Variants[1].ID)

	// Explicit segment ID preserved
	assert.Equal(t, uint(10), flags[0].Segments[0].ID)

	// Distribution resolved via VariantKey
	assert.Equal(t, uint(42), flags[0].Segments[0].Distributions[0].VariantID)
}

func TestNormalizeIDs_PreservesExplicitFlagIDs(t *testing.T) {
	t.Parallel()
	flags := []entity.Flag{
		{Model: gorm.Model{ID: 100}, Key: "flag-a"},
		{Key: "flag-b"}, // auto → 101
		{Model: gorm.Model{ID: 200}, Key: "flag-c"},
	}

	normalizeIDs(flags)
	assert.Equal(t, uint(100), flags[0].ID) // preserved
	assert.Equal(t, uint(201), flags[1].ID) // auto (max existing 200 + 1)
	assert.Equal(t, uint(200), flags[2].ID) // preserved
}

func TestNormalizeIDs_VariantKeyResolution(t *testing.T) {
	t.Parallel()
	// Distribution has VariantKey but VariantID=0 — resolved by key
	flags := []entity.Flag{
		{
			Key: "my-flag", Enabled: true,
			Variants: []entity.Variant{
				{Key: "control"},
				{Key: "treatment"},
			},
			Segments: []entity.Segment{
				{
					Description: "all", Rank: 0, RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "control", Percent: 50},
						{VariantKey: "treatment", Percent: 50},
					},
				},
			},
		},
	}

	normalizeIDs(flags)

	// control=1, treatment=2 (order of appearance)
	assert.Equal(t, uint(1), flags[0].Variants[0].ID)
	assert.Equal(t, uint(2), flags[0].Variants[1].ID)

	d0 := flags[0].Segments[0].Distributions[0]
	d1 := flags[0].Segments[0].Distributions[1]
	assert.Equal(t, uint(1), d0.VariantID) // control
	assert.Equal(t, uint(2), d1.VariantID) // treatment
}

func TestNormalizeIDs_EmptyFlags(t *testing.T) {
	t.Parallel()
	flags := []entity.Flag{}
	normalizeIDs(flags)
	assert.Empty(t, flags)
}

func TestNormalizeIDs_SingleFlagNoSegments(t *testing.T) {
	t.Parallel()
	flags := []entity.Flag{{Key: "simple", Enabled: true}}
	normalizeIDs(flags)
	assert.Equal(t, uint(1), flags[0].ID)
	assert.Empty(t, flags[0].Segments)
}

func TestNormalizeIDs_GlobalVariantIDUniqueness(t *testing.T) {
	t.Parallel()
	// Variant IDs continue across flags (global, not per-flag)
	flags := []entity.Flag{
		{Key: "a", Variants: []entity.Variant{{Key: "v1"}, {Key: "v2"}}},
		{Key: "b", Variants: []entity.Variant{{Key: "v3"}}},
	}

	normalizeIDs(flags)

	assert.Equal(t, uint(1), flags[0].Variants[0].ID)
	assert.Equal(t, uint(2), flags[0].Variants[1].ID)
	assert.Equal(t, uint(3), flags[1].Variants[0].ID) // continues, not reset
}

func TestNormalizeIDs_GlobalSegmentIDUniqueness(t *testing.T) {
	t.Parallel()
	flags := []entity.Flag{
		{Key: "a", Segments: []entity.Segment{{Description: "s1"}, {Description: "s2"}}},
		{Key: "b", Segments: []entity.Segment{{Description: "s3"}}},
	}

	normalizeIDs(flags)

	assert.Equal(t, uint(1), flags[0].Segments[0].ID)
	assert.Equal(t, uint(2), flags[0].Segments[1].ID)
	assert.Equal(t, uint(3), flags[1].Segments[0].ID) // continues
}

func TestNormalizeIDs_GlobalConstraintAndDistributionIDs(t *testing.T) {
	t.Parallel()
	flags := []entity.Flag{
		{
			Key: "a",
			Segments: []entity.Segment{
				{
					Description: "s1", Rank: 0, RolloutPercent: 100,
					Constraints:   []entity.Constraint{{Property: "x", Operator: "EQ", Value: "\"1\""}},
					Distributions: []entity.Distribution{{VariantKey: "v1", Percent: 100}},
				},
			},
			Variants: []entity.Variant{{Key: "v1"}},
		},
		{
			Key: "b",
			Segments: []entity.Segment{
				{
					Description: "s2", Rank: 0, RolloutPercent: 100,
					Constraints: []entity.Constraint{
						{Property: "a", Operator: "EQ", Value: "\"1\""},
						{Property: "b", Operator: "EQ", Value: "\"2\""},
					},
					Distributions: []entity.Distribution{
						{VariantKey: "v2", Percent: 100},
					},
				},
			},
			Variants: []entity.Variant{{Key: "v2"}},
		},
	}

	normalizeIDs(flags)

	// Constraints: globally unique (1 from flag-a, 2-3 from flag-b)
	assert.Equal(t, uint(1), flags[0].Segments[0].Constraints[0].ID)
	assert.Equal(t, uint(2), flags[1].Segments[0].Constraints[0].ID)
	assert.Equal(t, uint(3), flags[1].Segments[0].Constraints[1].ID)

	// Distributions: globally unique
	assert.Equal(t, uint(1), flags[0].Segments[0].Distributions[0].ID)
	assert.Equal(t, uint(2), flags[1].Segments[0].Distributions[0].ID)
}

func TestNormalizeIDs_TagIDGlobalUniqueness(t *testing.T) {
	t.Parallel()
	// Same tag value in two flags — different IDs
	flags := []entity.Flag{
		{Key: "a", Tags: []entity.Tag{{Value: "shared"}, {Value: "x"}}},
		{Key: "b", Tags: []entity.Tag{{Value: "shared"}, {Value: "y"}}},
	}

	normalizeIDs(flags)

	seen := map[uint]string{}
	for i := range flags {
		for j := range flags[i].Tags {
			tg := &flags[i].Tags[j]
			prev, exists := seen[tg.ID]
			assert.False(t, exists, "tag ID %d reused by %q and %q", tg.ID, prev, tg.Value)
			seen[tg.ID] = tg.Value
		}
	}
	assert.Len(t, seen, 4) // 4 unique tag IDs
}

func TestNormalizeIDs_ExplicitTagIDsPreserved(t *testing.T) {
	t.Parallel()
	flags := []entity.Flag{
		{Key: "a", Tags: []entity.Tag{{Model: gorm.Model{ID: 50}, Value: "my-tag"}}},
		{Key: "b", Tags: []entity.Tag{{Value: "auto-tag"}}},
	}

	normalizeIDs(flags)

	assert.Equal(t, uint(50), flags[0].Tags[0].ID) // preserved
	assert.Equal(t, uint(51), flags[1].Tags[0].ID) // auto (50 + 1)
}

func TestNormalizeIDs_MixedExplicitAndAutoIDs(t *testing.T) {
	t.Parallel()
	flags := []entity.Flag{
		{
			Model: gorm.Model{ID: 10}, Key: "flag-a",
			Variants: []entity.Variant{
				{Key: "v1"},                           // auto → 1
				{Model: gorm.Model{ID: 5}, Key: "v2"}, // explicit
				{Key: "v3"},                           // auto → 6 (5 + 1)
			},
			Segments: []entity.Segment{
				{
					Model: gorm.Model{ID: 3}, Description: "seg", Rank: 0, RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "v1", Percent: 34},
						{VariantKey: "v2", Percent: 33},
						{VariantKey: "v3", Percent: 33},
					},
				},
			},
		},
		{
			Key:      "flag-b",
			Variants: []entity.Variant{{Key: "only"}},
			Segments: []entity.Segment{
				{
					Description: "all", Rank: 0, RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "only", Percent: 100},
					},
				},
			},
		},
	}

	normalizeIDs(flags)

	// Flag IDs
	assert.Equal(t, uint(10), flags[0].ID)
	assert.Equal(t, uint(11), flags[1].ID)

	// Variant IDs: max existing is 5, so auto starts at 6
	assert.Equal(t, uint(6), flags[0].Variants[0].ID) // v1: auto (6)
	assert.Equal(t, uint(5), flags[0].Variants[1].ID) // v2: explicit (5)
	assert.Equal(t, uint(7), flags[0].Variants[2].ID) // v3: auto (7)
	// flag-b variant continues globally: next is 8
	assert.Equal(t, uint(8), flags[1].Variants[0].ID)

	// Distribution VariantIDs match resolved variants
	assert.Equal(t, uint(6), flags[0].Segments[0].Distributions[0].VariantID) // v1
	assert.Equal(t, uint(5), flags[0].Segments[0].Distributions[1].VariantID) // v2
	assert.Equal(t, uint(7), flags[0].Segments[0].Distributions[2].VariantID) // v3
}

func TestUnmarshalFlags_NoIDs(t *testing.T) {
	t.Parallel()
	// End-to-end: JSON with zero IDs
	jsonData := `{
		"Flags": [{
			"Key": "my-flag",
			"Description": "test flag",
			"Enabled": true,
			"Variants": [
				{"Key": "on"},
				{"Key": "off"}
			],
			"Segments": [{
				"Description": "all users",
				"Rank": 0,
				"RolloutPercent": 100,
				"Distributions": [
					{"VariantKey": "on", "Percent": 50},
					{"VariantKey": "off", "Percent": 50}
				]
			}]
		}]
	}`

	flags, err := unmarshalFlags([]byte(jsonData))
	assert.NoError(t, err)
	assert.Len(t, flags, 1)

	f := flags[0]
	assert.Equal(t, uint(1), f.ID)
	assert.Equal(t, "my-flag", f.Key)
	assert.Len(t, f.Variants, 2)
	assert.Len(t, f.Segments, 1)

	// Variants got globally unique IDs
	assert.Equal(t, uint(1), f.Variants[0].ID)
	assert.Equal(t, uint(2), f.Variants[1].ID)

	// Distributions resolved VariantIDs
	assert.Equal(t, uint(1), f.Segments[0].Distributions[0].VariantID)
	assert.Equal(t, uint(2), f.Segments[0].Distributions[1].VariantID)
}

func TestUnmarshalFlags_EmptyFlags(t *testing.T) {
	t.Parallel()
	flags, err := unmarshalFlags([]byte(`{"Flags": []}`))
	assert.NoError(t, err)
	assert.Empty(t, flags)
}

func TestNormalizeIDs_DistributionWithExplicitVariantID(t *testing.T) {
	t.Parallel()
	// Distribution has both VariantID and VariantKey — explicit ID wins
	flags := []entity.Flag{
		{
			Key:      "my-flag",
			Variants: []entity.Variant{{Key: "a"}, {Key: "b"}},
			Segments: []entity.Segment{
				{
					Description: "all", Rank: 0, RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{Model: gorm.Model{ID: 99}, VariantID: 42, VariantKey: "a", Percent: 100},
					},
				},
			},
		},
	}

	normalizeIDs(flags)

	// Distribution's own ID is 99 (explicit)
	assert.Equal(t, uint(99), flags[0].Segments[0].Distributions[0].ID)
	// Distribution's VariantID is 42 (explicit, not resolved from key)
	assert.Equal(t, uint(42), flags[0].Segments[0].Distributions[0].VariantID)
}

func TestNormalizeIDs_ComplexHandEditedFile(t *testing.T) {
	t.Parallel()
	// Realistic hand-edited file with no IDs whatsoever
	jsonData := `{
		"Flags": [
			{
				"Key": "feature-x",
				"Description": "New feature rollout",
				"Enabled": true,
				"EntityType": "user",
				"Variants": [
					{"Key": "control", "Attachment": {"version": "old"}},
					{"Key": "variant-a", "Attachment": {"version": "new-a"}},
					{"Key": "variant-b", "Attachment": {"version": "new-b"}}
				],
				"Segments": [
					{
						"Description": "Beta users",
						"Rank": 0,
						"RolloutPercent": 100,
						"Constraints": [
							{"Property": "tier", "Operator": "EQ", "Value": "\"beta\""}
						],
						"Distributions": [
							{"VariantKey": "control", "Percent": 34},
							{"VariantKey": "variant-a", "Percent": 33},
							{"VariantKey": "variant-b", "Percent": 33}
						]
					},
					{
						"Description": "Everyone else",
						"Rank": 1,
						"RolloutPercent": 100,
						"Distributions": [
							{"VariantKey": "control", "Percent": 100}
						]
					}
				],
				"Tags": [
					{"Value": "frontend"},
					{"Value": "experiment"}
				]
			}
		]
	}`

	flags, err := unmarshalFlags([]byte(jsonData))
	assert.NoError(t, err)
	assert.Len(t, flags, 1)
	f := flags[0]

	// Flag got ID
	assert.Equal(t, uint(1), f.ID)

	// All 3 variants got unique IDs
	vids := map[uint]bool{}
	for _, v := range f.Variants {
		assert.NotZero(t, v.ID)
		assert.False(t, vids[v.ID], "duplicate variant ID %d", v.ID)
		vids[v.ID] = true
	}

	// 2 segments: globally unique IDs, correct FlagID
	assert.Len(t, f.Segments, 2)
	assert.Equal(t, uint(1), f.Segments[0].ID)
	assert.Equal(t, uint(2), f.Segments[1].ID)
	assert.Equal(t, f.ID, f.Segments[0].FlagID)
	assert.Equal(t, f.ID, f.Segments[1].FlagID)

	// Constraint got ID and correct SegmentID
	assert.Len(t, f.Segments[0].Constraints, 1)
	assert.Equal(t, uint(1), f.Segments[0].Constraints[0].ID)
	assert.Equal(t, f.Segments[0].ID, f.Segments[0].Constraints[0].SegmentID)

	// All distributions resolved VariantID from VariantKey
	for _, seg := range f.Segments {
		for _, d := range seg.Distributions {
			assert.NotZero(t, d.VariantID, "distribution %s has zero VariantID", d.VariantKey)
			for _, v := range f.Variants {
				if v.Key == d.VariantKey {
					assert.Equal(t, v.ID, d.VariantID)
					break
				}
			}
		}
	}

	// Tags got unique IDs
	assert.Len(t, f.Tags, 2)
	assert.NotEqual(t, f.Tags[0].ID, f.Tags[1].ID)
}

func TestNormalizeIDs_SampleDataUnchanged(t *testing.T) {
	t.Parallel()
	// The existing sample_eval_cache.json has explicit IDs.
	// normalizeIDs must not alter them.
	flags := []entity.Flag{
		{
			Model: gorm.Model{ID: 1}, Key: "kmmcd1nsd6",
			Variants: []entity.Variant{
				{Model: gorm.Model{ID: 1}, Key: "control222"},
				{Model: gorm.Model{ID: 2}, Key: "blue123"},
			},
			Segments: []entity.Segment{
				{
					Model: gorm.Model{ID: 1}, FlagID: 1, Description: "All Users", Rank: 0, RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{Model: gorm.Model{ID: 1}, SegmentID: 1, VariantID: 2, VariantKey: "blue123", Percent: 50},
						{Model: gorm.Model{ID: 2}, SegmentID: 1, VariantID: 3, VariantKey: "red", Percent: 50},
					},
				},
			},
		},
	}

	normalizeIDs(flags)

	// All explicit IDs must be unchanged
	assert.Equal(t, uint(1), flags[0].ID)
	assert.Equal(t, uint(1), flags[0].Variants[0].ID)
	assert.Equal(t, uint(2), flags[0].Variants[1].ID)
	assert.Equal(t, uint(1), flags[0].Segments[0].ID)
	assert.Equal(t, uint(1), flags[0].Segments[0].Distributions[0].ID)
	assert.Equal(t, uint(2), flags[0].Segments[0].Distributions[0].VariantID)
	assert.Equal(t, uint(2), flags[0].Segments[0].Distributions[1].ID)
	assert.Equal(t, uint(3), flags[0].Segments[0].Distributions[1].VariantID)
}

func TestUnmarshalFlags_InvalidJSON(t *testing.T) {
	t.Parallel()
	_, err := unmarshalFlags([]byte(`{bad json`))
	assert.Error(t, err)
}

func TestUnmarshalFlags_ValidationErrors(t *testing.T) {
	t.Parallel()
	// Valid JSON with validation errors must be rejected.
	_, err := unmarshalFlags([]byte(`{"Flags": [{"Key": ""}]}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flag validation failed")
}

func TestUnmarshalFlags_WarningsAreAllowed(t *testing.T) {
	t.Parallel()
	// Warnings (e.g. no segments, no variants) should not prevent loading.
	flags, err := unmarshalFlags([]byte(`{
		"Flags": [{
			"Key": "my-flag",
			"Enabled": true,
			"Variants": [{"Key": "a"}],
			"Segments": []
		}]
	}`))
	assert.NoError(t, err)
	assert.Len(t, flags, 1)
	assert.Equal(t, "my-flag", flags[0].Key)
}

func TestNormalizeIDs_UnknownVariantKey(t *testing.T) {
	t.Parallel()
	// Distribution VariantKey doesn't match any variant — should warn
	// but not crash, and VariantID stays 0.
	flags := []entity.Flag{
		{
			Key: "f", Enabled: true,
			Variants: []entity.Variant{{Key: "a"}},
			Segments: []entity.Segment{
				{
					Description: "all", Rank: 0, RolloutPercent: 100,
					Distributions: []entity.Distribution{
						{VariantKey: "nonexistent", Percent: 100},
					},
				},
			},
		},
	}

	normalizeIDs(flags)

	// VariantID stays 0 (unknown key, nothing to resolve from)
	assert.Equal(t, uint(0), flags[0].Segments[0].Distributions[0].VariantID)
}
