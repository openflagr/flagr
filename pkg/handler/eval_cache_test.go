package handler

import (
	"github.com/openflagr/flagr/swagger_gen/models"
	"testing"

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
