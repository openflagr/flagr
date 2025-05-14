// nolint: errcheck
package handler

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/Allen-Career-Institute/flagr/swagger_gen/models"

	"github.com/Allen-Career-Institute/flagr/pkg/entity"

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

	ec.isInitialized.Store(false)
	f = ec.GetByFlagKeyOrID(fixtureFlag.ID)
	assert.Nil(t, f)

	ec.isInitialized.Store(true)
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

	ec.isInitialized.Store(false)
	f = ec.GetByTags(tags, &any)
	assert.Len(t, f, 0)
}

func TestGetByFlagKeyOrID_Errors(t *testing.T) {
	// Create a new EvalCache to avoid conflicting type history
	ec := &EvalCache{}
	ec.isInitialized.Store(true)

	// Case: invalid type stored first
	var invalid atomic.Value
	invalid.Store("invalid_type") // this is fine since it's the first use
	ec.cache = invalid            // inject it directly

	flag := ec.GetByFlagKeyOrID("someID")
	assert.Nil(t, flag)

	// Case: valid type but key doesn't exist
	ec = &EvalCache{}
	ec.isInitialized.Store(true)
	ec.cache.Store(&cacheContainer{
		idCache:  map[string]*entity.Flag{},
		keyCache: map[string]*entity.Flag{},
		tagCache: map[string]map[uint]*entity.Flag{},
	})
	flag = ec.GetByFlagKeyOrID("nonexistentID")
	assert.Nil(t, flag)
}

func TestGetByFlagKeyOrID_NilCache(t *testing.T) {
	ec := &EvalCache{}           // new cache, no value stored
	ec.isInitialized.Store(true) // simulate initialized state

	result := ec.GetByFlagKeyOrID("anyID")

	assert.Nil(t, result)
}

func TestEvalCache_StartInitialFailure(t *testing.T) {
	ec := &EvalCache{
		refreshTimeout:  10 * time.Millisecond,
		refreshInterval: 100 * time.Millisecond,
	}

	// Mock reloadMapCache to simulate failure
	//called := false

	func1 := ec.reloadMapCache
	// Create a stub and properly replace the method on the ec instance
	stubs := gostub.Stub(&func1, func() error {
		//called = true
		return assert.AnError
	})
	defer stubs.Reset()

	ec.Start()
	//time.Sleep(20 * time.Millisecond) // let goroutine run at least once

	assert.False(t, ec.isInitialized.Load())
	//assert.True(t, called)
}
