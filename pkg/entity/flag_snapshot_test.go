package entity

import (
	"testing"
)

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
