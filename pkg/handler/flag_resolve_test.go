package handler

import (
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestResolveFlagFromExposure(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	fixture := entity.GenFixtureFlag()

	t.Run("by flag id", func(t *testing.T) {
		f, err := resolveFlagFromExposure(&models.Exposure{FlagID: int64(fixture.ID)})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, fixture.ID, f.ID)
	})

	t.Run("by flag key", func(t *testing.T) {
		f, err := resolveFlagFromExposure(&models.Exposure{FlagKey: fixture.Key})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, fixture.ID, f.ID)
	})

	t.Run("mismatched id and key", func(t *testing.T) {
		f1 := entity.GenFixtureFlag()
		f2 := entity.GenFixtureFlag()
		f2.ID = 101
		f2.Key = "flag_key_101"
		ec := &EvalCache{cache: &cacheContainer{
			idCache: map[string]*entity.Flag{
				"100": &f1,
				"101": &f2,
			},
			keyCache: map[string]*entity.Flag{
				f1.Key: &f1,
				f2.Key: &f2,
			},
		}}
		defer gostub.StubFunc(&GetEvalCache, ec).Reset()
		_, err := resolveFlagFromExposure(&models.Exposure{FlagID: int64(f1.ID), FlagKey: f2.Key})
		assert.Error(t, err)
	})

	t.Run("missing flag ref", func(t *testing.T) {
		_, err := resolveFlagFromExposure(&models.Exposure{})
		assert.Error(t, err)
	})
}

func TestResolveVariantOnFlag(t *testing.T) {
	flag := entity.GenFixtureFlag()

	t.Run("optional empty", func(t *testing.T) {
		id, key, err := resolveVariantOnFlag(&flag, 0, "")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(0), id)
		assert.Empty(t, key)
	})

	t.Run("by id", func(t *testing.T) {
		id, key, err := resolveVariantOnFlag(&flag, 300, "")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, int64(300), id)
		assert.Equal(t, "control", key)
	})

	t.Run("id and key mismatch", func(t *testing.T) {
		_, _, err := resolveVariantOnFlag(&flag, 300, "treatment")
		assert.Error(t, err)
	})
}