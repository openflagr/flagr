package handler

import (
	"testing"

	"github.com/checkr/flagr/pkg/entity"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestGetByFlagKeyOrID(t *testing.T) {
	fixtureFlag := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(fixtureFlag)
	defer db.Close()
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
	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	ec := GetEvalCache()
	ec.reloadMapCache()

	tags := make([]string, len(fixtureFlag.Tags))
	for i, s := range fixtureFlag.Tags {
		tags[i] = s.Value
	}
	f := ec.GetByTags(tags)
	assert.Len(t, f, 1)
	assert.Equal(t, f[0].ID, fixtureFlag.ID)
	assert.Equal(t, f[0].Tags[0].Value, fixtureFlag.Tags[0].Value)
}
