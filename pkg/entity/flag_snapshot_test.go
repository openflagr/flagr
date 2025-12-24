package entity

import (
	"testing"

	"github.com/openflagr/flagr/pkg/notification"
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
		SaveFlagSnapshot(db, f.ID, "flagr-test@example.com", notification.OperationUpdate)
	})

	t.Run("save on non-existing flag", func(t *testing.T) {
		SaveFlagSnapshot(db, uint(999999), "flagr-test@example.com", "test")
	})
}
