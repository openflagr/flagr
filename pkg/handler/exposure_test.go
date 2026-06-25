package handler

import (
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	exposureapi "github.com/openflagr/flagr/swagger_gen/restapi/operations/exposure"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestPostExposures_EmptyBody(t *testing.T) {
	h := NewExposure()
	res := h.PostExposures(exposureapi.PostExposuresParams{})
	_, ok := res.(*exposureapi.PostExposuresDefault)
	assert.True(t, ok)
}

func TestPostExposures_BatchLimit(t *testing.T) {
	orig := config.Config.ExposureBatchSize
	config.Config.ExposureBatchSize = 2
	defer func() { config.Config.ExposureBatchSize = orig }()

	h := NewExposure()
	body := &models.ExposuresRequest{
		Exposures: []*models.Exposure{
			{EntityID: strPtr("a")},
			{EntityID: strPtr("b")},
			{EntityID: strPtr("c")},
		},
	}
	res := h.PostExposures(exposureapi.PostExposuresParams{Body: body})
	_, ok := res.(*exposureapi.PostExposuresDefault)
	assert.True(t, ok)
}

func TestPostExposures_PartialAccept(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	defer gostub.Stub(&config.Config.RecorderEnabled, false).Reset()

	flag := GenFixtureEvalCache().GetByFlagKeyOrID(int64(100))
	if !assert.NotNil(t, flag) {
		return
	}

	h := NewExposure()
	eid := "user-1"
	body := &models.ExposuresRequest{
		Exposures: []*models.Exposure{
			{EntityID: &eid, FlagID: int64(flag.ID)},
			{EntityID: strPtr("x")},
		},
	}
	res := h.PostExposures(exposureapi.PostExposuresParams{Body: body})
	okRes, ok := res.(*exposureapi.PostExposuresOK)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, int64(0), okRes.Payload.LoggedCount)
	assert.Len(t, okRes.Payload.Errors, 1)
}

func TestPostExposures_RecordsWhenEnabled(t *testing.T) {
	cache := GenFixtureEvalCache()
	flag := cache.GetByFlagKeyOrID(int64(100))
	if !assert.NotNil(t, flag) {
		return
	}
	flag.DataRecordsEnabled = true

	defer gostub.StubFunc(&GetEvalCache, cache).Reset()
	defer gostub.Stub(&config.Config.RecorderEnabled, true).Reset()

	mock := &mockRecorder{}
	prev := singletonDataRecorder
	singletonDataRecorder = fanOutRecorder{mock}
	defer func() { singletonDataRecorder = prev }()
	h := NewExposure()
	eid := "user-1"
	body := &models.ExposuresRequest{
		Exposures: []*models.Exposure{
			{EntityID: &eid, FlagID: int64(flag.ID)},
		},
	}
	res := h.PostExposures(exposureapi.PostExposuresParams{Body: body})
	okRes, ok := res.(*exposureapi.PostExposuresOK)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, int64(1), okRes.Payload.LoggedCount)
	if assert.Len(t, mock.calls, 1) {
		assert.Equal(t, models.EvalResultRecordSourceExposure, mock.calls[0].RecordSource)
	}
}

func TestValidateAndBuildExposure_FlagResolution(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	fixture := entity.GenFixtureFlag()
	eid := "e1"

	t.Run("by flag id", func(t *testing.T) {
		r, err := validateAndBuildExposure(&models.Exposure{EntityID: &eid, FlagID: int64(fixture.ID)})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(fixture.ID), r.FlagID)
	})

	t.Run("missing flag ref", func(t *testing.T) {
		_, err := validateAndBuildExposure(&models.Exposure{EntityID: &eid})
		assert.Error(t, err)
	})

	t.Run("mismatched id and key", func(t *testing.T) {
		f1 := entity.GenFixtureFlag()
		f2 := entity.GenFixtureFlag()
		f2.ID = 101
		f2.Key = "flag_key_101"
		ec := &EvalCache{cache: &cacheContainer{
			idCache:  map[string]*entity.Flag{"100": &f1, "101": &f2},
			keyCache: map[string]*entity.Flag{f1.Key: &f1, f2.Key: &f2},
		}}
		defer gostub.StubFunc(&GetEvalCache, ec).Reset()
		_, err := validateAndBuildExposure(&models.Exposure{EntityID: &eid, FlagID: int64(f1.ID), FlagKey: f2.Key})
		assert.Error(t, err)
	})
}

func TestExposureVariant(t *testing.T) {
	flag := entity.GenFixtureFlag()
	id, key, err := exposureVariant(&flag, 300, "")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, int64(300), id)
	assert.Equal(t, "control", key)
}

func TestExposureMergeJSONMap(t *testing.T) {
	dst := map[string]interface{}{}
	exposureMergeJSONMap(dst, map[string]any{"a": 1})
	exposureMergeJSONMap(dst, map[string]interface{}{"b": 2})
	assert.Equal(t, float64(1), dst["a"])
	assert.Equal(t, float64(2), dst["b"])
}

func strPtr(s string) *string { return &s }