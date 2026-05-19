package handler

import (
	"github.com/openflagr/flagr/swagger_gen/models"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"

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
	ec.lastSnapshotMaxID = 0
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
	ec.lastSnapshotMaxID = 0
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

func TestReloadMapCacheShortCircuit(t *testing.T) {
	fixtureFlag := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(fixtureFlag)

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Fatalf("Failed to get database")
	}
	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	// Create an initial snapshot so MAX(id) > 0 and the short-circuit
	// guard (lastSnapshotMaxID > 0) can engage.
	entity.SaveFlagSnapshot(db, fixtureFlag.ID, "test",
		notification.OperationCreate, notification.ComponentFlag, fixtureFlag.ID, fixtureFlag.Key)

	ec := GetEvalCache()
	ec.lastSnapshotMaxID = 0

	// Wrap fetchAllFlags to count how many times it's called.
	fetchCount := 0
	origFetch := fetchAllFlags
	fetchAllFlags = func() ([]entity.Flag, error) {
		fetchCount++
		return origFetch()
	}
	defer func() { fetchAllFlags = origFetch }()

	// 1st call: must fetch (no prior MAX id tracked).
	err := ec.reloadMapCache()
	assert.NoError(t, err)
	assert.Equal(t, 1, fetchCount, "first call should fetch full data")
	assert.Greater(t, ec.lastSnapshotMaxID, uint(0), "should track snapshot max ID")

	// 2nd call: must short-circuit (no new snapshot created).
	err = ec.reloadMapCache()
	assert.NoError(t, err)
	assert.Equal(t, 1, fetchCount, "second call should short-circuit")
	assert.Equal(t, ec.lastSnapshotMaxID, ec.lastSnapshotMaxID,
		"snapshot max ID must not change when no new snapshot exists")

	// Create another snapshot to simulate a mutation via the API.
	entity.SaveFlagSnapshot(db, fixtureFlag.ID, "test",
		notification.OperationUpdate, notification.ComponentFlag, fixtureFlag.ID, fixtureFlag.Key)

	// 3rd call: must fetch again (new snapshot invalidated the cache).
	err = ec.reloadMapCache()
	assert.NoError(t, err)
	assert.Equal(t, 2, fetchCount, "third call should fetch (new snapshot)")
}
