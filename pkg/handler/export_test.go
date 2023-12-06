package handler

import (
	"fmt"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/export"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestExportFlags(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)

	tmpDB1, dbErr1 := db.DB()
	if dbErr1 != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB1.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("happy code path", func(t *testing.T) {
		tmpDB := entity.NewTestDB()
		tmpDB2, dbErr2 := tmpDB.DB()
		if dbErr2 != nil {
			t.Errorf("Failed to get database")
		}

		defer tmpDB2.Close()

		exportFlags(tmpDB)
		tmpFlag := entity.Flag{}
		tmpDB.First(&tmpFlag)
		assert.NotZero(t, tmpFlag.ID)
	})

	t.Run("fetchAllFlags error code path", func(t *testing.T) {
		defer gostub.StubFunc(&fetchAllFlags, nil, fmt.Errorf("error")).Reset()
		tmpDB := entity.NewTestDB()
		tmpDB2, dbErr2 := tmpDB.DB()
		if dbErr2 != nil {
			t.Errorf("Failed to get database")
		}

		defer tmpDB2.Close()

		err := exportFlags(tmpDB)
		assert.Error(t, err)
	})
}

func TestExportFlagSnapshots(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)
	entity.SaveFlagSnapshot(db, f.ID, "flagr-test@example.com")

	tmpDB1, dbErr1 := db.DB()
	if dbErr1 != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB1.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("happy code path", func(t *testing.T) {
		tmpDB := entity.NewTestDB()
		tmpDB2, dbErr2 := tmpDB.DB()
		if dbErr2 != nil {
			t.Errorf("Failed to get database")
		}

		defer tmpDB2.Close()

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

	tmpDB1, dbErr1 := db.DB()
	if dbErr1 != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB1.Close()
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

	tmpDB1, dbErr1 := db.DB()
	if dbErr1 != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB1.Close()
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
	tmpDB1, dbErr1 := db.DB()
	if dbErr1 != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB1.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	ec := GetEvalCache()
	ec.reloadMapCache()

	t.Run("happy code path", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{})
		assert.IsType(t, res.(*export.GetExportEvalCacheJSONOK), res)
	})
}
