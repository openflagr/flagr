package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneFlagGraph_CopiesVariantsAndSegments(t *testing.T) {
	db := NewTestDB()
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