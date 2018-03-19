package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlagScan(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		f := GenFixtureFlag()
		b, _ := json.Marshal(f)
		err := f.Scan(b)
		assert.NoError(t, err)
	})

	t.Run("nil bytes", func(t *testing.T) {
		f := GenFixtureFlag()
		err := f.Scan(nil)
		assert.NoError(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		f := GenFixtureFlag()
		err := f.Scan([]byte("{"))
		assert.Error(t, err)
	})

	t.Run("invalid bytes type", func(t *testing.T) {
		f := GenFixtureFlag()
		err := f.Scan(123)
		assert.Error(t, err)
	})
}

func TestFlagValue(t *testing.T) {
	f := GenFixtureFlag()
	v, err := f.Value()
	assert.NoError(t, err)
	assert.NotZero(t, v)
}

func TestSaveFlagSnapshot(t *testing.T) {
	f := GenFixtureFlag()
	db := PopulateTestDB(f)
	defer db.Close()

	t.Run("happy code path", func(t *testing.T) {
		SaveFlagSnapshot(db, f.ID, "flagr-test@example.com")
	})

	t.Run("save on non-existing flag", func(t *testing.T) {
		SaveFlagSnapshot(db, uint(999999), "flagr-test@example.com")
	})
}
