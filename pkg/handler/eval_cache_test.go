package handler

import (
	"testing"

	"github.com/checkr/flagr/pkg/entity"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestGetEvalCacheStart(t *testing.T) {
	db := entity.PopulateTestDB(entity.GenFixtureFlag())
	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	ec := GetEvalCache()
	assert.NotPanics(t, func() {
		ec.Start()
	})
}

func TestGetByFlagIDs(t *testing.T) {
	fixtureFlag := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(fixtureFlag)
	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	ec := GetEvalCache()
	ec.Start()
	f := ec.GetByFlagID(fixtureFlag.ID)
	assert.Equal(t, f.ID, fixtureFlag.ID)

	fs := ec.GetByFlagIDs([]uint{fixtureFlag.ID})
	assert.NotZero(t, len(fs))
	assert.Equal(t, fs[fixtureFlag.ID].ID, fixtureFlag.ID)
}
