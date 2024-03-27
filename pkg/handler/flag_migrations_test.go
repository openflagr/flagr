package handler

import (
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlagMigrations(t *testing.T) {
	db := entity.NewTestDB()

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()

	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("FlagMigrationWithValidPath", func(t *testing.T) {
		err := FlagMigrations("./testdata/migrations")
		assert.Nil(t, err)
	})

	t.Run("FlagMigrationWithInvalidPath", func(t *testing.T) {
		err := FlagMigrations("./testdata/migration")
		assert.NotNil(t, err)
	})
}
