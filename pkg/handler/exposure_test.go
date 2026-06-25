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

func TestDatarRecorder_SkipsExposure(t *testing.T) {
	defer ResetDatar()
	defer gostub.Stub(&config.Config.RecorderType, []string{"datar"}).Reset()
	defer gostub.Stub(&config.Config.RecorderEnabled, true).Reset()

	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
	db.AutoMigrate(entity.AutoMigrateTables...)

	_ = GetDatar()
	r := NewDatarRecorder()
	r.AsyncRecord(models.EvalResult{
		FlagID:       1,
		VariantID:    10,
		SegmentID:    0,
		RecordSource: models.EvalResultRecordSourceExposure,
	})
	assert.Equal(t, 0, GetDatar().Len())

	r.AsyncRecord(models.EvalResult{FlagID: 1, VariantID: 10, SegmentID: 20})
	assert.Equal(t, 1, GetDatar().Len())
}

func strPtr(s string) *string { return &s }