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