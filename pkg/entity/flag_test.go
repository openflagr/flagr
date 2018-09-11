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

func TestCreateFlagKey(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		key, err := CreateFlagKey("")
		assert.NoError(t, err)
		assert.NotZero(t, key)
	})

	t.Run("invalid key", func(t *testing.T) {
		key, err := CreateFlagKey("1-2-3")
		assert.Error(t, err)
		assert.Zero(t, key)
	})
}
