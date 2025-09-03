package handler

import (
	"slices"
	"testing"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/export"

	"github.com/openflagr/flagr/pkg/entity"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestGetByFlagKeyOrID(t *testing.T) {
	fixtureFlag := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(fixtureFlag)

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	ec := GetEvalCache()
	ec.reloadMapCache()
	f := ec.GetByFlagKeyOrID(fixtureFlag.ID)
	assert.Equal(t, f.ID, fixtureFlag.ID)
	assert.Equal(t, f.Tags[0].Value, fixtureFlag.Tags[0].Value)
}

func TestGetByTags(t *testing.T) {
	fixtureFlag := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(fixtureFlag)

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	ec := GetEvalCache()
	ec.reloadMapCache()

	tags := make([]string, len(fixtureFlag.Tags))
	for i, s := range fixtureFlag.Tags {
		tags[i] = s.Value
	}
	any := models.EvalContextFlagTagsOperatorANY
	all := models.EvalContextFlagTagsOperatorALL
	f := ec.GetByTags(tags, &any)
	assert.Len(t, f, 1)
	assert.Equal(t, f[0].ID, fixtureFlag.ID)
	assert.Equal(t, f[0].Tags[0].Value, fixtureFlag.Tags[0].Value)

	tags = make([]string, len(fixtureFlag.Tags)+1)
	for i, s := range fixtureFlag.Tags {
		tags[i] = s.Value
	}
	tags[len(tags)-1] = "tag3"

	f = ec.GetByTags(tags, &any)
	assert.Len(t, f, 1)

	var operator *string
	f = ec.GetByTags(tags, operator)
	assert.Len(t, f, 1)

	f = ec.GetByTags(tags, &all)
	assert.Len(t, f, 0)
}

func TestEvalCacheExport(t *testing.T) {
	ec := GenFixtureEvalCacheWithFlags([]entity.Flag{
		entity.GenFixtureFlagWithTags(1, "first", true, []string{"tag1", "tag2"}),
		entity.GenFixtureFlagWithTags(2, "second", true, []string{"tag2", "tag3"}),
		entity.GenFixtureFlagWithTags(3, "third", false, []string{"tag2", "tag3"}),
		entity.GenFixtureFlagWithTags(4, "fourth", true, []string{}),
	})

	t.Run("should be able to query cache via flag ids", func(t *testing.T) {
		exportedFlags := ec.export(export.GetExportEvalCacheJSONParams{Ids: []int64{1, 3}}).Flags
		assert.Len(t, exportedFlags, 2)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(1)))
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(1)))
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(3)))
	})

	t.Run("should be able to query cache via flag keys", func(t *testing.T) {
		exportedFlags := ec.export(export.GetExportEvalCacheJSONParams{Keys: []string{"second", "fourth"}}).Flags
		assert.Len(t, exportedFlags, 2)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(2)))
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(4)))
	})

	t.Run("should be able to query cache via enabled property", func(t *testing.T) {
		tru := true
		exportedFlags := ec.export(export.GetExportEvalCacheJSONParams{Enabled: &tru}).Flags
		assert.Len(t, exportedFlags, 3)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(1)))
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(2)))
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(4)))

		fals := false
		exportedFlags = ec.export(export.GetExportEvalCacheJSONParams{Enabled: &fals}).Flags
		assert.Len(t, exportedFlags, 1)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(3)))
	})

	t.Run("should be able to query cache via tags with default ANY semantics", func(t *testing.T) {
		exportedFlags := ec.export(export.GetExportEvalCacheJSONParams{Tags: []string{"tag1", "tag2"}}).Flags
		assert.Len(t, exportedFlags, 3)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(1)))
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(2)))
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(3)))

		fals := false
		exportedFlags = ec.export(export.GetExportEvalCacheJSONParams{All: &fals, Tags: []string{"tag1", "tag2"}}).Flags
		assert.Len(t, exportedFlags, 3)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(1)))
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(2)))
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(3)))
	})

	t.Run("should be able to query cache via tags with ALL semantics", func(t *testing.T) {
		tru := true
		exportedFlags := ec.export(export.GetExportEvalCacheJSONParams{All: &tru, Tags: []string{"tag1", "tag2"}}).Flags
		assert.Len(t, exportedFlags, 1)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(1)))
	})

	t.Run("flag ids query should have precedence over other queries", func(t *testing.T) {
		exportedFlags := ec.export(export.GetExportEvalCacheJSONParams{Ids: []int64{4}, Keys: []string{"first", "second"}}).Flags
		assert.Len(t, exportedFlags, 1)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(4)))

		fals := false
		exportedFlags = ec.export(export.GetExportEvalCacheJSONParams{Ids: []int64{4}, Enabled: &fals}).Flags
		assert.Len(t, exportedFlags, 1)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(4)))

		exportedFlags = ec.export(export.GetExportEvalCacheJSONParams{Ids: []int64{4}, Tags: []string{"tag1"}}).Flags
		assert.Len(t, exportedFlags, 1)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(4)))
	})

	t.Run("flag keys query should have precedence over enabled and tags queries", func(t *testing.T) {
		fals := false
		exportedFlags := ec.export(export.GetExportEvalCacheJSONParams{Keys: []string{"fourth"}, Enabled: &fals}).Flags
		assert.Len(t, exportedFlags, 1)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(4)))

		exportedFlags = ec.export(export.GetExportEvalCacheJSONParams{Keys: []string{"fourth"}, Tags: []string{"tag1"}}).Flags
		assert.Len(t, exportedFlags, 1)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(4)))
	})

	t.Run("should be able to combine enabled and tags queries", func(t *testing.T) {
		tru := true
		exportedFlags := ec.export(export.GetExportEvalCacheJSONParams{Enabled: &tru, Tags: []string{"tag2"}}).Flags
		assert.Len(t, exportedFlags, 2)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(1)))
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(2)))

		fals := false
		exportedFlags = ec.export(export.GetExportEvalCacheJSONParams{Enabled: &fals, Tags: []string{"tag2"}}).Flags
		assert.Len(t, exportedFlags, 1)
		assert.True(t, slices.ContainsFunc(exportedFlags, withID(3)))
	})
}

func withID(id uint) func(entity.Flag) bool {
	return func(f entity.Flag) bool {
		return f.ID == id
	}
}
