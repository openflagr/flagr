package entity

import (
	"testing"
)

func TestSaveFlagSnapshot(t *testing.T) {
	f := GenFixtureFlag()
	db := PopulateTestDB(f)

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()

	t.Run("happy code path", func(t *testing.T) {
		SaveFlagSnapshot(db, f.Model.ID, "flagr-test@example.com")
	})

	t.Run("save on non-existing flag", func(t *testing.T) {
		SaveFlagSnapshot(db, uint(999999), "flagr-test@example.com")
	})
}
