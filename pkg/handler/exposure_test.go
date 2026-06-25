package handler

import (
	"fmt"
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

func TestPostExposures_NullRow(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	h := NewExposure()
	body := &models.ExposuresRequest{
		Exposures: []*models.Exposure{nil},
	}
	res := h.PostExposures(exposureapi.PostExposuresParams{Body: body})
	okRes, ok := res.(*exposureapi.PostExposuresOK)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, int64(0), okRes.Payload.LoggedCount)
	if assert.Len(t, okRes.Payload.Errors, 1) {
		assert.Equal(t, int64(0), okRes.Payload.Errors[0].Index)
	}
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

func TestPostExposures_RecorderOn_DataRecordsOff(t *testing.T) {
	cache := GenFixtureEvalCache()
	flag := cache.GetByFlagKeyOrID(int64(100))
	if !assert.NotNil(t, flag) {
		return
	}
	flag.DataRecordsEnabled = false

	defer gostub.StubFunc(&GetEvalCache, cache).Reset()
	defer gostub.Stub(&config.Config.RecorderEnabled, true).Reset()
	defer gostub.Stub(&config.Config.RecorderType, []string{}).Reset()
	GetDataRecorder()

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
	assert.Equal(t, int64(0), okRes.Payload.LoggedCount)
	assert.Empty(t, mock.calls)
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

	defer gostub.Stub(&config.Config.RecorderType, []string{}).Reset()
	GetDataRecorder()
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
		assert.Equal(t, int64(0), mock.calls[0].SegmentID)
	}
}

func TestBuildExposureDataRecord(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	fixture := entity.GenFixtureFlag()
	eid := "e1"

	t.Run("requires entityID", func(t *testing.T) {
		empty := ""
		_, err := buildExposureDataRecord(&models.Exposure{EntityID: &empty, FlagID: int64(fixture.ID)})
		assert.Error(t, err)
	})

	t.Run("by flag id", func(t *testing.T) {
		r, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagID: int64(fixture.ID)})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(fixture.ID), r.FlagID)
		assert.Equal(t, models.EvalResultRecordSourceExposure, r.RecordSource)
		assert.Equal(t, int64(0), r.SegmentID)
	})

	t.Run("by flag key", func(t *testing.T) {
		r, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagKey: fixture.Key})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, fixture.Key, r.FlagKey)
	})

	t.Run("missing flag ref", func(t *testing.T) {
		_, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid})
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
		_, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagID: int64(f1.ID), FlagKey: f2.Key})
		assert.Error(t, err)
	})

	t.Run("optional variant", func(t *testing.T) {
		r, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagID: int64(fixture.ID)})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(0), r.VariantID)
	})

	t.Run("variant by id", func(t *testing.T) {
		r, err := buildExposureDataRecord(&models.Exposure{
			EntityID: &eid, FlagID: int64(fixture.ID), VariantID: 300,
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(300), r.VariantID)
		assert.Equal(t, "control", r.VariantKey)
	})

	t.Run("variant by key", func(t *testing.T) {
		r, err := buildExposureDataRecord(&models.Exposure{
			EntityID: &eid, FlagKey: fixture.Key, VariantKey: "treatment",
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(301), r.VariantID)
		assert.Equal(t, "treatment", r.VariantKey)
	})

	t.Run("entity type from flag overrides client", func(t *testing.T) {
		f := entity.GenFixtureFlag()
		f.EntityType = "from_flag"
		ec := &EvalCache{cache: &cacheContainer{
			idCache:  map[string]*entity.Flag{fmt.Sprintf("%d", f.ID): &f},
			keyCache: map[string]*entity.Flag{f.Key: &f},
		}}
		defer gostub.StubFunc(&GetEvalCache, ec).Reset()
		clientType := "client_type"
		r, err := buildExposureDataRecord(&models.Exposure{
			EntityID: &eid, FlagID: int64(f.ID), EntityType: clientType,
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "from_flag", r.EvalContext.EntityType)
	})

	t.Run("merges entityContext and metadata", func(t *testing.T) {
		r, err := buildExposureDataRecord(&models.Exposure{
			EntityID:       &eid,
			FlagID:         int64(fixture.ID),
			EntityContext:  map[string]any{"country": "US"},
			Metadata:       map[string]interface{}{"page": "/home"},
		})
		if !assert.NoError(t, err) {
			return
		}
		ctx, ok := r.EvalContext.EntityContext.(map[string]interface{})
		if !assert.True(t, ok) {
			return
		}
		assert.Equal(t, "US", ctx["country"])
		assert.Equal(t, "/home", ctx["page"])
	})

	t.Run("flagSnapshotID from client", func(t *testing.T) {
		r, err := buildExposureDataRecord(&models.Exposure{
			EntityID: &eid, FlagID: int64(fixture.ID), FlagSnapshotID: 4242,
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(4242), r.FlagSnapshotID)
	})

	t.Run("flagSnapshotID defaults to cache", func(t *testing.T) {
		r, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagID: int64(fixture.ID)})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(fixture.SnapshotID), r.FlagSnapshotID)
	})
}

func TestResolveExposureVariant(t *testing.T) {
	flag := entity.GenFixtureFlag()

	t.Run("optional empty", func(t *testing.T) {
		id, key, err := resolveExposureVariant(&flag, 0, "")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(0), id)
		assert.Empty(t, key)
	})

	t.Run("by id", func(t *testing.T) {
		id, key, err := resolveExposureVariant(&flag, 300, "")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(300), id)
		assert.Equal(t, "control", key)
	})

	t.Run("by key", func(t *testing.T) {
		id, key, err := resolveExposureVariant(&flag, 0, "treatment")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(301), id)
		assert.Equal(t, "treatment", key)
	})

	t.Run("id and key mismatch", func(t *testing.T) {
		_, _, err := resolveExposureVariant(&flag, 300, "treatment")
		assert.Error(t, err)
	})
}

func TestMergeJSONIntoMap(t *testing.T) {
	dst := map[string]interface{}{}
	mergeJSONIntoMap(dst, map[string]any{"a": 1})
	mergeJSONIntoMap(dst, map[string]interface{}{"b": 2})
	assert.Equal(t, float64(1), dst["a"])
	assert.Equal(t, float64(2), dst["b"])
}

func strPtr(s string) *string { return &s }
