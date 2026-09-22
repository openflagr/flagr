package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyFlagTemplate_PreservesJevQuestion(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))

	threshold := 0.8
	source := &Flag{Key: "jev_src", Description: "s", Enabled: true}
	require.NoError(t, db.Create(source).Error)
	require.NoError(t, db.Create(&Variant{FlagID: source.ID, Key: "on"}).Error)
	seg := &Segment{FlagID: source.ID, RolloutPercent: 100, Rank: SegmentDefaultRank}
	require.NoError(t, db.Create(seg).Error)
	require.NoError(t, db.Create(&Constraint{
		SegmentID:              seg.ID,
		Property:               "@jev.risk",
		Operator:               "GTE",
		Value:                  "0.5",
		JevType:                JevTypeScore,
		JevInstructions:        `"How risky?"`,
		JevCriteria:            `["low","high"]`,
		JevConfidenceThreshold: &threshold,
	}).Error)
	require.NoError(t, PreloadSegmentsVariantsTags(db).First(source, source.ID).Error)

	dest := &Flag{Key: "jev_dest", Description: "d", Enabled: true}
	require.NoError(t, db.Create(dest).Error)
	require.NoError(t, ApplyFlagTemplate(db, dest.ID, SourceFlagTemplate(source)))

	var got Constraint
	require.NoError(t, db.Joins("JOIN segments ON segments.id = constraints.segment_id").
		Where("segments.flag_id = ?", dest.ID).First(&got).Error)
	assert.Equal(t, JevTypeScore, got.JevType)
	assert.Equal(t, `"How risky?"`, got.JevInstructions)
	assert.Equal(t, `["low","high"]`, got.JevCriteria)
	require.NotNil(t, got.JevConfidenceThreshold)
	assert.Equal(t, 0.8, *got.JevConfidenceThreshold)
}
