package handler

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	exposureapi "github.com/openflagr/flagr/swagger_gen/restapi/operations/exposure"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestPostExposures_EmptyBody(t *testing.T) {
	t.Parallel()
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

func TestExposureEvalResult_WireContainsRecordSource(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	eid := "wire-test-entity"
	r, _, err := buildExposureDataRecord(&models.Exposure{
		EntityID: &eid,
		FlagID:   100,
	})
	if !assert.NoError(t, err) {
		return
	}
	raw, err := r.MarshalBinary()
	if !assert.NoError(t, err) {
		return
	}
	var payload map[string]any
	if !assert.NoError(t, json.Unmarshal(raw, &payload)) {
		return
	}
	assert.Equal(t, models.EvalResultRecordSourceExposure, payload["recordSource"])
	assert.Equal(t, int64(0), r.SegmentID)
}

func TestExposureDataRecordFrame_OutputContainsRecordSource(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	eid := "frame-test-entity"
	r, _, err := buildExposureDataRecord(&models.Exposure{
		EntityID: &eid,
		FlagID:   100,
	})
	if !assert.NoError(t, err) {
		return
	}
	frame := DataRecordFrame{
		evalResult: r,
		options:    DataRecordFrameOptions{FrameOutputMode: frameOutputModePayloadRawJSON},
	}
	out, err := frame.Output()
	if !assert.NoError(t, err) {
		return
	}
	var outer struct {
		Payload json.RawMessage `json:"payload"`
	}
	if !assert.NoError(t, json.Unmarshal(out, &outer)) {
		return
	}
	var inner map[string]any
	if !assert.NoError(t, json.Unmarshal(outer.Payload, &inner)) {
		return
	}
	assert.Equal(t, models.EvalResultRecordSourceExposure, inner["recordSource"])
}

func TestBuildExposureDataRecord(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	fixture := entity.GenFixtureFlag()
	eid := "e1"

	t.Run("requires entityID", func(t *testing.T) {
		empty := ""
		_, _, err := buildExposureDataRecord(&models.Exposure{EntityID: &empty, FlagID: int64(fixture.ID)})
		assert.Error(t, err)
	})

	t.Run("nil entityID", func(t *testing.T) {
		_, _, err := buildExposureDataRecord(&models.Exposure{FlagID: int64(fixture.ID)})
		assert.Error(t, err)
	})

	t.Run("flag not found", func(t *testing.T) {
		_, _, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagID: 999999})
		assert.Error(t, err)
	})

	t.Run("flag key only", func(t *testing.T) {
		r, _, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagKey: fixture.Key})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, fixture.Key, r.FlagKey)
	})

	t.Run("client timestamp", func(t *testing.T) {
		ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
		r, _, err := buildExposureDataRecord(&models.Exposure{
			EntityID: &eid, FlagID: int64(fixture.ID), Timestamp: strfmt.DateTime(ts),
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "2024-06-01T12:00:00Z", r.Timestamp)
	})

	t.Run("no context leaves evalContext empty", func(t *testing.T) {
		r, _, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagID: int64(fixture.ID)})
		if !assert.NoError(t, err) {
			return
		}
		assert.Nil(t, r.EvalContext.EntityContext)
	})

	t.Run("by flag id", func(t *testing.T) {
		r, _, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagID: int64(fixture.ID)})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(fixture.ID), r.FlagID)
		assert.Equal(t, models.EvalResultRecordSourceExposure, r.RecordSource)
		assert.Equal(t, int64(0), r.SegmentID)
	})

	t.Run("by flag key", func(t *testing.T) {
		r, _, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagKey: fixture.Key})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, fixture.Key, r.FlagKey)
	})

	t.Run("missing flag ref", func(t *testing.T) {
		_, _, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid})
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
		_, _, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagID: int64(f1.ID), FlagKey: f2.Key})
		assert.Error(t, err)
	})

	t.Run("optional variant", func(t *testing.T) {
		r, _, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagID: int64(fixture.ID)})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(0), r.VariantID)
	})

	t.Run("variant by id", func(t *testing.T) {
		r, _, err := buildExposureDataRecord(&models.Exposure{
			EntityID: &eid, FlagID: int64(fixture.ID), VariantID: 300,
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(300), r.VariantID)
		assert.Equal(t, "control", r.VariantKey)
	})

	t.Run("variant by key", func(t *testing.T) {
		r, _, err := buildExposureDataRecord(&models.Exposure{
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
		r, _, err := buildExposureDataRecord(&models.Exposure{
			EntityID: &eid, FlagID: int64(f.ID), EntityType: clientType,
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "from_flag", r.EvalContext.EntityType)
	})

	t.Run("entityContext on record", func(t *testing.T) {
		r, _, err := buildExposureDataRecord(&models.Exposure{
			EntityID:      &eid,
			FlagID:        int64(fixture.ID),
			EntityContext: map[string]any{"country": "US", "page": "/home"},
		})
		if !assert.NoError(t, err) {
			return
		}
		ctx, ok := r.EvalContext.EntityContext.(map[string]any)
		if !assert.True(t, ok) {
			return
		}
		assert.Equal(t, "US", ctx["country"])
		assert.Equal(t, "/home", ctx["page"])
	})

	t.Run("flagSnapshotID from client", func(t *testing.T) {
		r, _, err := buildExposureDataRecord(&models.Exposure{
			EntityID: &eid, FlagID: int64(fixture.ID), FlagSnapshotID: 4242,
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(4242), r.FlagSnapshotID)
	})

	t.Run("flagSnapshotID defaults to cache", func(t *testing.T) {
		r, _, err := buildExposureDataRecord(&models.Exposure{EntityID: &eid, FlagID: int64(fixture.ID)})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(fixture.SnapshotID), r.FlagSnapshotID)
	})
}

func TestResolveExposureVariant(t *testing.T) {
	t.Parallel()
	flag := entity.GenFixtureFlag()

	t.Run("optional empty", func(t *testing.T) {
		t.Parallel()
		id, key, err := resolveExposureVariant(&flag, 0, "")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(0), id)
		assert.Empty(t, key)
	})

	t.Run("by id", func(t *testing.T) {
		t.Parallel()
		id, key, err := resolveExposureVariant(&flag, 300, "")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(300), id)
		assert.Equal(t, "control", key)
	})

	t.Run("by key", func(t *testing.T) {
		t.Parallel()
		id, key, err := resolveExposureVariant(&flag, 0, "treatment")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(301), id)
		assert.Equal(t, "treatment", key)
	})

	t.Run("id and key mismatch", func(t *testing.T) {
		t.Parallel()
		_, _, err := resolveExposureVariant(&flag, 300, "treatment")
		assert.Error(t, err)
	})

	t.Run("invalid variant id", func(t *testing.T) {
		t.Parallel()
		_, _, err := resolveExposureVariant(&flag, 999, "")
		assert.Error(t, err)
	})

	t.Run("invalid variant key", func(t *testing.T) {
		t.Parallel()
		_, _, err := resolveExposureVariant(&flag, 0, "missing")
		assert.Error(t, err)
	})

	t.Run("id not found when both set", func(t *testing.T) {
		t.Parallel()
		_, _, err := resolveExposureVariant(&flag, 999, "control")
		assert.Error(t, err)
	})

	t.Run("key not found when both set", func(t *testing.T) {
		t.Parallel()
		_, _, err := resolveExposureVariant(&flag, 300, "missing")
		assert.Error(t, err)
	})

	t.Run("matching id and key", func(t *testing.T) {
		t.Parallel()
		id, key, err := resolveExposureVariant(&flag, 300, "control")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(300), id)
		assert.Equal(t, "control", key)
	})
}

func TestResolveExposureFlag(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	ec := GenFixtureEvalCache()
	fixture := ec.GetByFlagKeyOrID(int64(100))
	if !assert.NotNil(t, fixture) {
		return
	}

	t.Run("by id", func(t *testing.T) {
		f, err := resolveExposureFlag(ec, &models.Exposure{FlagID: int64(fixture.ID)})
		assert.NoError(t, err)
		assert.Equal(t, fixture.ID, f.ID)
	})

	t.Run("by key", func(t *testing.T) {
		f, err := resolveExposureFlag(ec, &models.Exposure{FlagKey: fixture.Key})
		assert.NoError(t, err)
		assert.Equal(t, fixture.ID, f.ID)
	})

	t.Run("missing ref", func(t *testing.T) {
		_, err := resolveExposureFlag(ec, &models.Exposure{})
		assert.Error(t, err)
	})
}

func TestLogExposureStatsd_Stubbed(t *testing.T) {
	t.Parallel()
	var lastStatus string
	orig := logExposureStatsd
	logExposureStatsd = func(status string, flagID int64, flagKey string) {
		lastStatus = status
		_ = flagID
		_ = flagKey
	}
	defer func() { logExposureStatsd = orig }()

	logExposureStatsd("accepted", 100, "k")
	assert.Equal(t, "accepted", lastStatus)
	logExposureStatsd("recorded", 100, "k")
	assert.Equal(t, "recorded", lastStatus)
}

func TestLogExposureStatsd_NoClient(t *testing.T) {
	orig := config.Global.StatsdClient
	config.Global.StatsdClient = nil
	defer func() { config.Global.StatsdClient = orig }()
	assert.NotPanics(t, func() { logExposureStatsd("rejected", 0, "") })
}

func strPtr(s string) *string { return &s }
