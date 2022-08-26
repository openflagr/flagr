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
		assert.NotNil(t, f.Tags)
	})
}

func TestFlagPreload(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		f := GenFixtureFlag()
		db := PopulateTestDB(f)

		tmpDB, dbErr := db.DB()
		if dbErr != nil {
			t.Errorf("Failed to get database")
		}

		defer tmpDB.Close()

		err := f.Preload(db)
		assert.NoError(t, err)
	})
}

func TestFlagPreloadTags(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		f := GenFixtureFlag()
		db := PopulateTestDB(f)

		tmpDB, dbErr := db.DB()
		if dbErr != nil {
			t.Errorf("Failed to get database")
		}

		defer tmpDB.Close()

		err := f.PreloadTags(db)
		assert.NoError(t, err)
	})
}

func TestCreateFlagKey(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		key, err := CreateFlagKey("")
		assert.NoError(t, err)
		assert.NotZero(t, key)
	})

	t.Run("invalid key", func(t *testing.T) {
		key, err := CreateFlagKey(" spaces in key are not allowed 1-2-3")
		assert.Error(t, err)
		assert.Zero(t, key)
	})
}

func TestCreateFlagEntityType(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		f := GenFixtureFlag()
		db := PopulateTestDB(f)

		err := CreateFlagEntityType(db, "")
		assert.NoError(t, err)
	})

	t.Run("invalid key", func(t *testing.T) {
		f := GenFixtureFlag()
		db := PopulateTestDB(f)

		err := CreateFlagEntityType(db, " spaces in key are not allowed 123-invalid-key")
		assert.Error(t, err)
	})
}
