package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneFlagGraph_CopiesVariantsAndSegments(t *testing.T) {
	db := NewTestDB()
	require.NoError(t, db.AutoMigrate(AutoMigrateTables...))
	source := GenFixtureFlag()
	require.NoError(t, db.Create(&source).Error)
	require.NoError(t, PreloadSegmentsVariantsTags(db).First(&source, source.ID).Error)

	dest := &Flag{Key: "clone_dest", Description: "dest", Enabled: true}
	require.NoError(t, db.Create(dest).Error)

	tx := db.Begin()
	require.NoError(t, CloneFlagGraph(tx, &source, dest))
	require.NoError(t, tx.Commit().Error)

	var loaded Flag
	require.NoError(t, PreloadSegmentsVariantsTags(db).First(&loaded, dest.ID).Error)
	assert.Len(t, loaded.Variants, len(source.Variants))
	assert.Len(t, loaded.Segments, len(source.Segments))
}

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

func TestCloneFlagGraph_CopiesTags(t *testing.T) {
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
	require.NoError(t, CloneFlagGraph(tx, source, dest))
	require.NoError(t, tx.Commit().Error)

	var loaded Flag
	require.NoError(t, PreloadFlagTags(db).First(&loaded, dest.ID).Error)
	require.Len(t, loaded.Tags, 1)
	assert.Equal(t, "team-a", loaded.Tags[0].Value)
}