package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendTagValueToFlag(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))
	f := &Flag{Key: "tag_flag", Description: "d", Enabled: true}
	require.NoError(t, db.Create(f).Error)

	tx := db.Begin()
	require.NoError(t, AppendTagValueToFlag(tx, f.ID, "alpha"))
	require.NoError(t, AppendTagValueToFlag(tx, f.ID, "alpha"))
	require.NoError(t, tx.Commit().Error)

	var loaded Flag
	require.NoError(t, PreloadFlagTags(db).First(&loaded, f.ID).Error)
	require.Len(t, loaded.Tags, 1)
	assert.Equal(t, "alpha", loaded.Tags[0].Value)
}

func TestApplyFlagTemplate_SimpleBoolean(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))

	f := &Flag{Key: "bool_mat", Description: "d", Enabled: true}
	require.NoError(t, db.Create(f).Error)

	require.NoError(t, ApplyFlagTemplate(db, f.ID, SimpleBooleanFlagTemplate()))

	var variant Variant
	require.NoError(t, db.Where("flag_id = ?", f.ID).First(&variant).Error)
	assert.Equal(t, "on", variant.Key)

	var segment Segment
	require.NoError(t, db.Where("flag_id = ?", f.ID).First(&segment).Error)
	assert.Equal(t, uint(100), segment.RolloutPercent)
	assert.Equal(t, SegmentDefaultRank, segment.Rank)

	var dist Distribution
	require.NoError(t, db.Where("segment_id = ?", segment.ID).First(&dist).Error)
	assert.Equal(t, variant.ID, dist.VariantID)
	assert.Equal(t, uint(100), dist.Percent)
}

func TestApplyFlagTemplate_UnknownVariantKey(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))
	f := &Flag{Key: "bad_tpl", Description: "d", Enabled: true}
	require.NoError(t, db.Create(f).Error)

	tpl := Flag{
		Segments: []Segment{{
			RolloutPercent: 100,
			Rank:           SegmentDefaultRank,
			Distributions:  []Distribution{{VariantKey: "missing", Percent: 100}},
		}},
	}
	err := ApplyFlagTemplate(db, f.ID, tpl)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestApplyFlagTemplate_InvalidVariantKey(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))
	f := &Flag{Key: "invalid_var", Description: "d", Enabled: true}
	require.NoError(t, db.Create(f).Error)

	tpl := Flag{
		Variants: []Variant{{Key: " bad key "}},
	}
	err := ApplyFlagTemplate(db, f.ID, tpl)
	require.Error(t, err)
}

func TestSourceFlagTemplate_PreservesGraphShape(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))
	source := GenFixtureFlag()
	require.NoError(t, db.Create(&source).Error)
	require.NoError(t, PreloadSegmentsVariantsTags(db).First(&source, source.ID).Error)

	tpl := SourceFlagTemplate(&source)
	assert.Len(t, tpl.Variants, len(source.Variants))
	assert.Len(t, tpl.Segments, len(source.Segments))
	require.NotEmpty(t, tpl.Segments)
	assert.NotEmpty(t, tpl.Segments[0].Constraints)
	assert.NotEmpty(t, tpl.Segments[0].Distributions)
	for i, v := range tpl.Variants {
		assert.Equal(t, source.Variants[i].Key, v.Key)
		assert.Zero(t, v.ID)
	}
}

func TestSourceFlagTemplate_RoundTripViaApply(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))
	source := GenFixtureFlag()
	require.NoError(t, db.Create(&source).Error)
	require.NoError(t, PreloadSegmentsVariantsTags(db).First(&source, source.ID).Error)

	dest := &Flag{Key: "roundtrip", Description: "d", Enabled: true}
	require.NoError(t, db.Create(dest).Error)

	tx := db.Begin()
	require.NoError(t, ApplyFlagTemplate(tx, dest.ID, SourceFlagTemplate(&source)))
	require.NoError(t, tx.Commit().Error)

	var loaded Flag
	require.NoError(t, PreloadSegmentsVariantsTags(db).First(&loaded, dest.ID).Error)
	assert.Len(t, loaded.Variants, len(source.Variants))
	assert.Len(t, loaded.Segments, len(source.Segments))
	require.NotEmpty(t, loaded.Segments[0].Constraints)
	require.NotEmpty(t, loaded.Segments[0].Distributions)
}

func TestApplyFlagTemplate_FromSourceVariantsAndSegments(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))
	source := GenFixtureFlag()
	require.NoError(t, db.Create(&source).Error)
	require.NoError(t, PreloadSegmentsVariantsTags(db).First(&source, source.ID).Error)

	dest := &Flag{Key: "clone_dest", Description: "dest", Enabled: true}
	require.NoError(t, db.Create(dest).Error)

	tx := db.Begin()
	require.NoError(t, ApplyFlagTemplate(tx, dest.ID, SourceFlagTemplate(&source)))
	require.NoError(t, tx.Commit().Error)

	var loaded Flag
	require.NoError(t, PreloadSegmentsVariantsTags(db).First(&loaded, dest.ID).Error)
	assert.Len(t, loaded.Variants, len(source.Variants))
	assert.Len(t, loaded.Segments, len(source.Segments))
}

func TestApplyFlagTemplate_FromSourceTags(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))
	source := &Flag{Key: "src_tags", Description: "s", Enabled: true}
	require.NoError(t, db.Create(source).Error)
	tag := &Tag{Value: "team-a"}
	require.NoError(t, db.Where("value = ?", tag.Value).FirstOrCreate(tag).Error)
	require.NoError(t, db.Model(source).Association("Tags").Append(tag))

	dest := &Flag{Key: "dst_tags", Description: "d", Enabled: true}
	require.NoError(t, db.Create(dest).Error)
	require.NoError(t, PreloadFlagTags(db).First(source, source.ID).Error)

	tx := db.Begin()
	require.NoError(t, ApplyFlagTemplate(tx, dest.ID, SourceFlagTemplate(source)))
	require.NoError(t, tx.Commit().Error)

	var loaded Flag
	require.NoError(t, PreloadFlagTags(db).First(&loaded, dest.ID).Error)
	require.Len(t, loaded.Tags, 1)
	assert.Equal(t, "team-a", loaded.Tags[0].Value)
}