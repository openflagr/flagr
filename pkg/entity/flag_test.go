package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlagPrepareEvaluation(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		f := GenFixtureFlag()
		assert.NoError(t, f.PrepareEvaluation())
		assert.NotNil(t, f.FlagEvaluation.VariantsMap)
	})
}

func TestFlagPreload(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		f := GenFixtureFlag()
		db := PopulateTestDB(f)
		defer db.Close()

		err := f.Preload(db)
		assert.NoError(t, err)
	})
}

func TestFlagBeforeCreate(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		f := GenFixtureFlag()
		f.Key = ""
		db := NewTestDB()
		err := f.Create(db)
		assert.NoError(t, err)
		assert.NotZero(t, f.Key)
	})

	t.Run("invalid key", func(t *testing.T) {
		f := GenFixtureFlag()
		f.Key = "1-2-3"
		db := NewTestDB()
		err := f.Create(db)
		assert.Error(t, err)
	})
}
