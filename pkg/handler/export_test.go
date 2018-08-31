package handler

import (
	"fmt"
	"testing"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/export"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestExportFlags(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)
	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("happy code path", func(t *testing.T) {
		tmpDB := entity.NewTestDB()
		defer tmpDB.Close()

		exportFlags(tmpDB)
		tmpFlag := entity.Flag{}
		tmpDB.First(&tmpFlag)
		assert.NotZero(t, tmpFlag.ID)
	})

	t.Run("fetchAllFlags error code path", func(t *testing.T) {
		defer gostub.StubFunc(&fetchAllFlags, nil, fmt.Errorf("error")).Reset()
		tmpDB := entity.NewTestDB()
		defer tmpDB.Close()

		err := exportFlags(tmpDB)
		assert.Error(t, err)
	})
}

func TestExportFlagSnapshots(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)
	entity.SaveFlagSnapshot(db, f.ID, "flagr-test@example.com")

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("happy code path", func(t *testing.T) {
		tmpDB := entity.NewTestDB()
		defer tmpDB.Close()

		exportFlagSnapshots(tmpDB)
		fs := entity.FlagSnapshot{}
		tmpDB.First(&fs)
		assert.NotZero(t, fs.ID)
	})
}

func TestExportSQLiteFile(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)
	entity.SaveFlagSnapshot(db, f.ID, "flagr-test@example.com")

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("happy code path", func(t *testing.T) {
		f, done, err := exportSQLiteFile()
		defer done()

		assert.NoError(t, err)
		assert.NotNil(t, f)
	})
}

func TestExportSQLiteHandler(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)
	entity.SaveFlagSnapshot(db, f.ID, "flagr-test@example.com")

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("happy code path", func(t *testing.T) {
		res := exportSQLiteHandler(export.GetExportSqliteParams{})
		assert.IsType(t, res.(*export.GetExportSqliteOK), res)
	})

	t.Run("fetchAllFlags error code path", func(t *testing.T) {
		defer gostub.StubFunc(&fetchAllFlags, nil, fmt.Errorf("error")).Reset()

		res := exportSQLiteHandler(export.GetExportSqliteParams{})
		assert.IsType(t, res.(*export.GetExportSqliteDefault), res)
	})
}
