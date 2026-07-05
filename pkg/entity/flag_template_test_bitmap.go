package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyFlagTemplate_CopiesDistributionBitmap(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))

	source := &Flag{Key: "bitmap_src", Description: "s", Enabled: true}
	require.NoError(t, db.Create(source).Error)
	v := &Variant{FlagID: source.ID, Key: "on"}
	require.NoError(t, db.Create(v).Error)
	seg := &Segment{FlagID: source.ID, RolloutPercent: 100, Rank: SegmentDefaultRank}
	require.NoError(t, db.Create(seg).Error)
	wantBitmap := "sticky-bitmap-payload"
	require.NoError(t, db.Create(&Distribution{
		SegmentID:  seg.ID,
		VariantID:  v.ID,
		VariantKey: "on",
		Percent:    100,
		Bitmap:     wantBitmap,
	}).Error)
	require.NoError(t, PreloadSegmentsVariantsTags(db).First(source, source.ID).Error)

	dest := &Flag{Key: "bitmap_dest", Description: "d", Enabled: true}
	require.NoError(t, db.Create(dest).Error)
	require.NoError(t, ApplyFlagTemplate(db, dest.ID, SourceFlagTemplate(source)))

	var dist Distribution
	require.NoError(t, db.Joins("JOIN segments ON segments.id = distributions.segment_id").
		Where("segments.flag_id = ?", dest.ID).First(&dist).Error)
	assert.Equal(t, wantBitmap, dist.Bitmap)
}
