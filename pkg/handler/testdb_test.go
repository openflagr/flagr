package handler

import (
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// handlerTestDB returns a migrated in-memory DB with getDB stubbed; call cleanup when done.
func handlerTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	db := entity.NewTestDB()
	require.NoError(t, db.AutoMigrate(entity.AutoMigrateTables...))
	stub := gostub.StubFunc(&getDB, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	return db, func() {
		stub.Reset()
		_ = sqlDB.Close()
	}
}
