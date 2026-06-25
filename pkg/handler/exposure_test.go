package handler

import (
	"testing"

	"github.com/openflagr/flagr/pkg/config"
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

	var recorded []models.EvalResult
	defer gostub.Stub(&recordPipelineEvent, func(r models.EvalResult) {
		recorded = append(recorded, r)
	}).Reset()

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
	if assert.Len(t, recorded, 1) {
		assert.Equal(t, models.EvalResultRecordSourceExposure, recorded[0].RecordSource)
	}
}

func TestProcessExposureRow_NullRow(t *testing.T) {
	n, err := processExposureRow(0, nil)
	assert.Equal(t, int64(0), n)
	assert.NotNil(t, err)
}

func strPtr(s string) *string { return &s }