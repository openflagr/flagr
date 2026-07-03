package handler

import (
	"fmt"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
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
	entity.SaveFlagSnapshot(db, f.ID, "flagr-test@example.com", notification.OperationUpdate, notification.ComponentFlag, f.ID, f.Key)

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
	entity.SaveFlagSnapshot(db, f.ID, "flagr-test@example.com", notification.OperationUpdate, notification.ComponentFlag, f.ID, f.Key)

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
		f, done, err := exportSQLiteFile(new(true))
		defer done()

		assert.NoError(t, err)
		assert.NotNil(t, f)
	})
}

func TestExportSQLiteHandler(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)
	entity.SaveFlagSnapshot(db, f.ID, "flagr-test@example.com", notification.OperationUpdate, notification.ComponentFlag, f.ID, f.Key)

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
	ec.lastSnapshotMaxID = 0
	ec.reloadMapCache()

	t.Run("happy code path", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{})
		assert.IsType(t, res.(*export.GetExportEvalCacheJSONOK), res)
	})
}

func boolPtr(b bool) *bool { return &b }

func TestExportEvalCacheQuery(t *testing.T) {
	// Use multiple flags to test filtering properly
	ec := GenFixtureEvalCacheWithFlags([]entity.Flag{
		GenFixtureFlagWithTags(1, "first", true, []string{"tag1", "tag2"}),
		GenFixtureFlagWithTags(2, "second", true, []string{"tag2", "tag3"}),
		GenFixtureFlagWithTags(3, "third", false, []string{"tag2", "tag3"}),
		GenFixtureFlagWithTags(4, "fourth", true, []string{}),
	})

	t.Run("no params returns all flags", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{})
		assert.Len(t, result.Flags, 4)
	})

	t.Run("filter by ids", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{Ids: []int64{1, 3}})
		assert.Len(t, result.Flags, 2)
		assert.True(t, containsID(result.Flags, 1))
		assert.True(t, containsID(result.Flags, 3))
	})

	t.Run("filter by ids with non-existent id", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{Ids: []int64{999}})
		assert.Len(t, result.Flags, 0)
	})

	t.Run("filter by keys", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{Keys: []string{"second", "fourth"}})
		assert.Len(t, result.Flags, 2)
		assert.True(t, containsID(result.Flags, 2))
		assert.True(t, containsID(result.Flags, 4))
	})

	t.Run("filter by keys with non-existent key", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{Keys: []string{"nonexistent"}})
		assert.Len(t, result.Flags, 0)
	})

	t.Run("filter by enabled true", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{Enabled: boolPtr(true)})
		assert.Len(t, result.Flags, 3)
		assert.True(t, containsID(result.Flags, 1))
		assert.True(t, containsID(result.Flags, 2))
		assert.True(t, containsID(result.Flags, 4))
	})

	t.Run("filter by enabled false", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{Enabled: boolPtr(false)})
		assert.Len(t, result.Flags, 1)
		assert.True(t, containsID(result.Flags, 3))
	})

	t.Run("filter by tags ANY", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{Tags: []string{"tag1", "tag999"}})
		assert.Len(t, result.Flags, 1) // only flag 1 has tag1
		assert.True(t, containsID(result.Flags, 1))
	})

	t.Run("filter by tags ALL", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{Tags: []string{"tag2", "tag3"}, All: boolPtr(true)})
		assert.Len(t, result.Flags, 2) // flags 2 and 3 have both tag2 and tag3
		assert.True(t, containsID(result.Flags, 2))
		assert.True(t, containsID(result.Flags, 3))
	})

	t.Run("ids override enabled and tags", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{
			Ids:     []int64{1},
			Enabled: boolPtr(false), // flag 1 is enabled=true, but ids take precedence
			Tags:    []string{"nonexistent"},
		})
		assert.Len(t, result.Flags, 1)
		assert.True(t, containsID(result.Flags, 1))
	})

	t.Run("keys override enabled and tags", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{
			Keys:    []string{"first"},
			Enabled: boolPtr(false), // flag first is enabled=true, but keys take precedence
			Tags:    []string{"nonexistent"},
		})
		assert.Len(t, result.Flags, 1)
		assert.True(t, containsID(result.Flags, 1))
	})

	t.Run("combined enabled AND tags", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{
			Enabled: boolPtr(true),
			Tags:    []string{"tag2"},
		})
		assert.Len(t, result.Flags, 2) // flags 1 and 2 are enabled and have tag2
		assert.True(t, containsID(result.Flags, 1))
		assert.True(t, containsID(result.Flags, 2))
	})

	t.Run("no match returns empty", func(t *testing.T) {
		result := ec.export(export.GetExportEvalCacheJSONParams{Ids: []int64{999}})
		assert.Len(t, result.Flags, 0)
	})
}

func containsID(flags []entity.Flag, id uint) bool {
	for _, f := range flags {
		if f.ID == id {
			return true
		}
	}
	return false
}
