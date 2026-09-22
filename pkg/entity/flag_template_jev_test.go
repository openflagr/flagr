package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyFlagTemplate_PreservesJevQuestion(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))

	source := &Flag{Key: "jev_src", Description: "s", Enabled: true}
	require.NoError(t, db.Create(source).Error)
	require.NoError(t, db.Create(&Variant{FlagID: source.ID, Key: "on"}).Error)
	seg := &Segment{FlagID: source.ID, RolloutPercent: 100, Rank: SegmentDefaultRank}
	require.NoError(t, db.Create(seg).Error)

	sc := &Constraint{SegmentID: seg.ID, Property: "@jev.risk", Operator: "GTE", Value: "0.5"}
	require.NoError(t, sc.SetJevQuestion(&JevQuestion{
		Type:                JevTypeScore,
		Instructions:        "How risky?",
		Criteria:            []any{"low", "high"},
		ConfidenceThreshold: f64(0.8),
	}))
	require.NoError(t, db.Create(sc).Error)
	require.NoError(t, PreloadSegmentsVariantsTags(db).First(source, source.ID).Error)

	dest := &Flag{Key: "jev_dest", Description: "d", Enabled: true}
	require.NoError(t, db.Create(dest).Error)
	require.NoError(t, ApplyFlagTemplate(db, dest.ID, SourceFlagTemplate(source)))

	var got Constraint
	require.NoError(t, db.Joins("JOIN segments ON segments.id = constraints.segment_id").
		Where("segments.flag_id = ?", dest.ID).First(&got).Error)
	assert.True(t, got.IsJev())
	q, err := got.JevQuestion()
	require.NoError(t, err)
	require.NotNil(t, q)
	assert.Equal(t, JevTypeScore, q.Type)
	assert.Equal(t, "How risky?", q.Instructions)
	assert.Equal(t, []any{"low", "high"}, q.Criteria)
	require.NotNil(t, q.ConfidenceThreshold)
	assert.Equal(t, 0.8, *q.ConfidenceThreshold)
}
