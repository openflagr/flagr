package handler

import (
	"fmt"
	"math"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/evaluation"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestEvalSegment(t *testing.T) {
	t.Run("test empty evalContext", func(t *testing.T) {
		s := entity.GenFixtureSegment()
		vID, log, evalNextSegment := evalSegment(100, models.EvalContext{}, s)

		assert.Nil(t, vID)
		assert.NotEmpty(t, log)
		assert.True(t, evalNextSegment)
	})

	t.Run("test happy code path", func(t *testing.T) {
		s := entity.GenFixtureSegment()
		s.RolloutPercent = uint(100)
		vID, log, evalNextSegment := evalSegment(100, models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"dl_state": "CA"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		}, s)

		assert.NotNil(t, vID)
		assert.NotEmpty(t, log)
		assert.False(t, evalNextSegment)
	})

	t.Run("test constraint evaluation error", func(t *testing.T) {
		s := entity.GenFixtureSegment()
		s.RolloutPercent = uint(100)
		vID, log, evalNextSegment := evalSegment(100, models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		}, s)

		assert.Nil(t, vID)
		assert.NotEmpty(t, log)
		assert.True(t, evalNextSegment)
	})

	t.Run("test constraint not match", func(t *testing.T) {
		s := entity.GenFixtureSegment()
		s.RolloutPercent = uint(100)
		vID, log, evalNextSegment := evalSegment(100, models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"dl_state": "NY"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		}, s)

		assert.Nil(t, vID)
		assert.NotEmpty(t, log)
		assert.True(t, evalNextSegment)
	})

	t.Run("test evalContext wrong format", func(t *testing.T) {
		s := entity.GenFixtureSegment()
		s.RolloutPercent = uint(100)
		vID, log, evalNextSegment := evalSegment(100, models.EvalContext{
			EnableDebug:   true,
			EntityContext: nil,
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		}, s)

		assert.Nil(t, vID)
		assert.NotEmpty(t, log)
		assert.True(t, evalNextSegment)
	})

	t.Run("test float comparison - 9990403>=9990404 evals to be false", func(t *testing.T) {
		s := entity.GenFixtureSegment()
		s.RolloutPercent = uint(100)
		s.Constraints = []entity.Constraint{
			{
				Model:     gorm.Model{ID: 500},
				SegmentID: 200,
				Property:  "foo",
				Operator:  models.ConstraintOperatorGTE,
				Value:     `9990404`,
			},
		}
		s.PrepareEvaluation()

		vID, log, evalNextSegment := evalSegment(100, models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"foo": float64(9990403)},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		}, s)

		assert.Nil(t, vID)
		assert.NotZero(t, log)
		assert.True(t, evalNextSegment)
	})

	t.Run("test float comparison - 9990404>=9990403 evals to be true", func(t *testing.T) {
		s := entity.GenFixtureSegment()
		s.RolloutPercent = uint(100)
		s.Constraints = []entity.Constraint{
			{
				Model:     gorm.Model{ID: 500},
				SegmentID: 200,
				Property:  "foo",
				Operator:  models.ConstraintOperatorGTE,
				Value:     `9990403`,
			},
		}
		s.PrepareEvaluation()

		vID, log, evalNextSegment := evalSegment(100, models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"foo": float64(9990404)},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		}, s)

		assert.NotZero(t, vID)
		assert.NotZero(t, log)
		assert.False(t, evalNextSegment)
	})
}

func TestEvalFlag(t *testing.T) {
	defer gostub.StubFunc(&logEvalResult).Reset()

	t.Run("test empty evalContext", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		result := EvalFlag(models.EvalContext{FlagID: int64(100)})
		assert.Zero(t, result.VariantID)
		assert.NotZero(t, result.FlagID)
		assert.NotEmpty(t, result.EvalContext.EntityID)
	})

	t.Run("test happy code path", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		result := EvalFlag(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"dl_state": "CA"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		})
		assert.NotNil(t, result)
		assert.NotZero(t, result.VariantID)
	})

	t.Run("test happy code path with flagKey", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		result := EvalFlag(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"dl_state": "CA"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagKey:       "flag_key_100",
		})
		assert.NotNil(t, result)
		assert.NotZero(t, result.VariantID)
	})

	t.Run("test happy code path with flagKey", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		result := EvalFlag(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"dl_state": "CA"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagKey:       "flag_key_100",
		})
		assert.NotNil(t, result)
		assert.NotZero(t, result.VariantID)
	})

	t.Run("test happy code path with multiple constraints", func(t *testing.T) {
		f := entity.GenFixtureFlag()
		f.Segments[0].Constraints = []entity.Constraint{
			{
				Model:     gorm.Model{ID: 500},
				SegmentID: 200,
				Property:  "dl_state",
				Operator:  models.ConstraintOperatorEQ,
				Value:     `"CA"`,
			},
			{
				Model:     gorm.Model{ID: 501},
				SegmentID: 200,
				Property:  "state",
				Operator:  models.ConstraintOperatorEQ,
				Value:     `{dl_state}`,
			},
			{
				Model:     gorm.Model{ID: 502},
				SegmentID: 200,
				Property:  "rate",
				Operator:  models.ConstraintOperatorGT,
				Value:     `1000`,
			},
			{
				Model:     gorm.Model{ID: 503},
				SegmentID: 200,
				Property:  "city-name",
				Operator:  models.ConstraintOperatorEQ,
				Value:     `"SF"`,
			},
		}
		f.PrepareEvaluation()
		ec := &EvalCache{
			cache: &cacheContainer{idCache: map[string]*entity.Flag{"100": &f}},
		}
		defer gostub.StubFunc(&GetEvalCache, ec).Reset()
		result := EvalFlag(models.EvalContext{
			EnableDebug: true,
			EntityContext: map[string]interface{}{
				"dl_state":  "CA",
				"state":     "CA",
				"rate":      2000,
				"city-name": "SF",
			},
			EntityID:   "entityID1",
			EntityType: "entityType1",
			FlagID:     int64(100),
		})
		assert.NotNil(t, result)
		assert.NotZero(t, result.VariantID)
	})

	t.Run("test multiple segments with the first segment 0% rollout", func(t *testing.T) {
		f := entity.GenFixtureFlag()
		f.Segments = append(f.Segments, entity.GenFixtureSegment())
		f.Segments[0].Constraints = []entity.Constraint{}
		f.Segments[0].RolloutPercent = uint(0)

		f.PrepareEvaluation()
		ec := &EvalCache{
			cache: &cacheContainer{idCache: map[string]*entity.Flag{"100": &f}},
		}
		defer gostub.StubFunc(&GetEvalCache, ec).Reset()
		result := EvalFlag(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"dl_state": "CA", "state": "CA", "rate": 2000},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		})
		assert.NotNil(t, result)
		assert.Zero(t, result.VariantID)
	})

	t.Run("test no match path with multiple constraints", func(t *testing.T) {
		f := entity.GenFixtureFlag()
		f.Segments[0].Constraints = []entity.Constraint{
			{
				Model:     gorm.Model{ID: 500},
				SegmentID: 200,
				Property:  "dl_state",
				Operator:  models.ConstraintOperatorEQ,
				Value:     `"CA"`,
			},
			{
				Model:     gorm.Model{ID: 500},
				SegmentID: 200,
				Property:  "state",
				Operator:  models.ConstraintOperatorEQ,
				Value:     `{dl_state}`,
			},
		}
		f.PrepareEvaluation()

		ec := &EvalCache{
			cache: &cacheContainer{idCache: map[string]*entity.Flag{"100": &f}},
		}
		defer gostub.StubFunc(&GetEvalCache, ec).Reset()
		result := EvalFlag(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"dl_state": "CA", "state": "NY"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		})
		assert.NotNil(t, result)
		assert.Zero(t, result.VariantID)
	})

	t.Run("test enabled=false", func(t *testing.T) {
		f := entity.GenFixtureFlag()
		f.Enabled = false
		ec := &EvalCache{
			cache: &cacheContainer{idCache: map[string]*entity.Flag{"100": &f}},
		}
		defer gostub.StubFunc(&GetEvalCache, ec).Reset()
		result := EvalFlag(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"dl_state": "CA"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		})
		assert.NotNil(t, result)
		assert.Zero(t, result.VariantID)
	})

	t.Run("test entityType override", func(t *testing.T) {
		t.Run("empty entityType case", func(t *testing.T) {
			f := entity.GenFixtureFlag()
			f.EntityType = ""
			ec := &EvalCache{
				cache: &cacheContainer{idCache: map[string]*entity.Flag{"100": &f}},
			}
			defer gostub.StubFunc(&GetEvalCache, ec).Reset()
			result := EvalFlag(models.EvalContext{
				EnableDebug:   true,
				EntityContext: map[string]interface{}{"dl_state": "CA"},
				EntityID:      "entityID1",
				EntityType:    "entityType1",
				FlagID:        int64(100),
			})
			assert.NotNil(t, result)
			assert.NotZero(t, result.VariantID)
			assert.Equal(t, "entityType1", result.EvalContext.EntityType)
		})
		t.Run("override case", func(t *testing.T) {
			f := entity.GenFixtureFlag()
			f.EntityType = "some_entity_type"
			ec := &EvalCache{
				cache: &cacheContainer{idCache: map[string]*entity.Flag{"100": &f}},
			}
			defer gostub.StubFunc(&GetEvalCache, ec).Reset()
			result := EvalFlag(models.EvalContext{
				EnableDebug:   true,
				EntityContext: map[string]interface{}{"dl_state": "CA"},
				EntityID:      "entityID1",
				EntityType:    "entityType1",
				FlagID:        int64(100),
			})
			assert.NotNil(t, result)
			assert.NotZero(t, result.VariantID)
			assert.NotEqual(t, "entityType1", result.EvalContext.EntityType)
			assert.Equal(t, "some_entity_type", result.EvalContext.EntityType)
		})
	})
}

func TestEvalFlagDistribution(t *testing.T) {
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()

	// vID1 and vID2 are the variants generated from GenFixtureEvalCache
	vID1, vID2 := int64(300), int64(301)

	// we are testing the `num` cases with the relative distribution differences between the two variants
	// in this case, we set it to be 0.5% of 1e6 samples
	num, threshold := int(1e6), 0.005

	t.Run("test distribution on integers", func(t *testing.T) {
		cnt := make(map[int64]int)
		for i := 0; i < num; i++ {
			result := EvalFlag(models.EvalContext{
				EnableDebug:   false,
				EntityContext: map[string]interface{}{"dl_state": "CA"},
				EntityID:      fmt.Sprintf("%d", i),
				EntityType:    "entityType1",
				FlagID:        int64(100),
			})
			cnt[result.VariantID]++
		}
		assert.Len(t, cnt, 2)
		assert.Less(t,
			math.Abs(float64(cnt[vID1]-cnt[vID2])/float64(cnt[vID1]+cnt[vID2])),
			threshold,
			"Expected distribution to be uniform",
		)
	})

	t.Run("test distribution on secure random key generator", func(t *testing.T) {
		cnt := make(map[int64]int)
		for i := 0; i < num; i++ {
			result := EvalFlag(models.EvalContext{
				EnableDebug:   false,
				EntityContext: map[string]interface{}{"dl_state": "CA"},
				EntityID:      util.NewSecureRandomKey(),
				EntityType:    "entityType1",
				FlagID:        int64(100),
			})
			cnt[result.VariantID]++
		}
		assert.Len(t, cnt, 2)
		assert.Less(t,
			math.Abs(float64(cnt[vID1]-cnt[vID2])/float64(cnt[vID1]+cnt[vID2])),
			threshold,
			"Expected distribution to be uniform",
		)
	})

	t.Run("test distribution on uuid", func(t *testing.T) {
		cnt := make(map[int64]int)
		for i := 0; i < num; i++ {
			result := EvalFlag(models.EvalContext{
				EnableDebug:   false,
				EntityContext: map[string]interface{}{"dl_state": "CA"},
				EntityID:      uniuri.NewLen(uniuri.UUIDLen),
				EntityType:    "entityType1",
				FlagID:        int64(100),
			})
			cnt[result.VariantID]++
		}
		assert.Len(t, cnt, 2)
		assert.Less(t,
			math.Abs(float64(cnt[vID1]-cnt[vID2])/float64(cnt[vID1]+cnt[vID2])),
			threshold,
			"Expected distribution to be uniform",
		)
	})

	t.Run("test distribution on random string + int", func(t *testing.T) {
		cnt := make(map[int64]int)
		for i := 0; i < num; i++ {
			result := EvalFlag(models.EvalContext{
				EnableDebug:   false,
				EntityContext: map[string]interface{}{"dl_state": "CA"},
				EntityID:      fmt.Sprintf("random_int%d%s", i, util.NewSecureRandomKey()),
				EntityType:    "entityType1",
				FlagID:        int64(100),
			})
			cnt[result.VariantID]++
		}
		assert.Len(t, cnt, 2)
		assert.Less(t,
			math.Abs(float64(cnt[vID1]-cnt[vID2])/float64(cnt[vID1]+cnt[vID2])),
			threshold,
			"Expected distribution to be uniform",
		)
	})
}

func TestEvalFlagsByTags(t *testing.T) {
	defer gostub.StubFunc(&logEvalResult).Reset()

	t.Run("test happy code path", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		results := EvalFlagsByTags(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]interface{}{"dl_state": "CA"},
			FlagTags:      []string{"tag1", "tag2"},
		})
		assert.NotZero(t, len(results))
		assert.NotZero(t, results[0].VariantID)
	})
}

func TestPostEvaluation(t *testing.T) {
	t.Run("test empty body", func(t *testing.T) {
		defer gostub.StubFunc(&EvalFlag, &models.EvalResult{}).Reset()
		e := NewEval()
		resp := e.PostEvaluation(evaluation.PostEvaluationParams{})
		assert.NotNil(t, resp)
	})

	t.Run("test happy code path", func(t *testing.T) {
		defer gostub.StubFunc(&EvalFlag, &models.EvalResult{}).Reset()
		e := NewEval()
		resp := e.PostEvaluation(evaluation.PostEvaluationParams{
			Body: &models.EvalContext{
				EnableDebug:   true,
				EntityContext: map[string]interface{}{"dl_state": "CA", "state": "NY"},
				EntityID:      "entityID1",
				EntityType:    "entityType1",
				FlagID:        int64(100),
			},
		})
		assert.NotNil(t, resp)
	})
}

func TestPostEvaluationBatch(t *testing.T) {
	t.Run("test happy code path", func(t *testing.T) {
		defer gostub.StubFunc(&EvalFlag, &models.EvalResult{}).Reset()
		e := NewEval()
		resp := e.PostEvaluationBatch(evaluation.PostEvaluationBatchParams{
			Body: &models.EvaluationBatchRequest{
				EnableDebug: true,
				Entities: []*models.EvaluationEntity{
					{
						EntityContext: map[string]interface{}{"dl_state": "CA", "state": "NY"},
						EntityID:      "entityID1",
						EntityType:    "entityType1",
					},
				},
				FlagIDs:  []int64{100, 200},
				FlagKeys: []string{"flag_key_1", "flag_key_2"},
			},
		})
		assert.NotNil(t, resp)
	})
}
func TestTagsPostEvaluationBatch(t *testing.T) {
	t.Run("test happy code path", func(t *testing.T) {
		defer gostub.StubFunc(&EvalFlag, &models.EvalResult{}).Reset()
		e := NewEval()
		resp := e.PostEvaluationBatch(evaluation.PostEvaluationBatchParams{
			Body: &models.EvaluationBatchRequest{
				EnableDebug: true,
				FlagTags:    []string{"tag1", "tag2"},
				Entities: []*models.EvaluationEntity{
					{
						EntityContext: map[string]interface{}{"dl_state": "CA", "state": "NY"},
						EntityID:      "entityID1",
						EntityType:    "entityType1",
					},
				},
			},
		})
		assert.NotNil(t, resp)
	})
}

func TestRateLimitPerFlagConsoleLogging(t *testing.T) {
	r := &models.EvalResult{FlagID: 1}
	t.Run("running fast triggers rate limiting", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			rateLimitPerFlagConsoleLogging(r)
		}
	})
}

func BenchmarkEvalFlag(b *testing.B) {
	b.StopTimer()
	defer gostub.StubFunc(&logEvalResult).Reset()
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		EvalFlag(models.EvalContext{
			EntityContext: map[string]interface{}{"dl_state": "CA"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		})
	}
}

func BenchmarkEvalFlagsByTags(b *testing.B) {
	b.StopTimer()
	defer gostub.StubFunc(&logEvalResult).Reset()
	defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		EvalFlagsByTags(models.EvalContext{
			EntityContext: map[string]interface{}{"dl_state": "CA"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagTags:      []string{"tag1", "tag2"},
		})
	}
}
