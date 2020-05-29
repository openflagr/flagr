package handler

import (
	"fmt"
	"testing"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
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

	t.Run("happy code path and export everything in db", func(t *testing.T) {
		f, done, err := exportSQLiteFile(nil)
		defer done()

		assert.NoError(t, err)
		assert.NotNil(t, f)
	})

	t.Run("happy code path and exclude_snapshots", func(t *testing.T) {
		f, done, err := exportSQLiteFile(util.BoolPtr(true))
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

func TestExportEvalCacheJSONHandler(t *testing.T) {
	fixtureFlag := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(fixtureFlag)
	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	ec := GetEvalCache()
	ec.reloadMapCache()

	t.Run("happy code path", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{})
		assert.IsType(t, res.(*export.GetExportEvalCacheJSONOK), res)
	})
}
