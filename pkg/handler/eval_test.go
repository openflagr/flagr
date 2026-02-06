package handler

import (
	"fmt"
	"math"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/openflagr/flagr/pkg/config"
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
			EntityContext: map[string]any{"dl_state": "CA"},
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
			EntityContext: map[string]any{},
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
			EntityContext: map[string]any{"dl_state": "NY"},
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
			EntityContext: map[string]any{"foo": float64(9990403)},
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
			EntityContext: map[string]any{"foo": float64(9990404)},
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
			EntityContext: map[string]any{"dl_state": "CA"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagID:        int64(100),
		})
		assert.NotNil(t, result)
		assert.NotZero(t, result.VariantID)
		assert.NotEmpty(t, result.FlagTags)
		assert.Len(t, result.FlagTags, 2)
		assert.Contains(t, result.FlagTags, "tag1")
		assert.Contains(t, result.FlagTags, "tag2")
	})

	t.Run("test happy code path with flagKey", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		result := EvalFlag(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]any{"dl_state": "CA"},
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
			EntityContext: map[string]any{"dl_state": "CA"},
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
			EntityContext: map[string]any{
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
			EntityContext: map[string]any{"dl_state": "CA", "state": "CA", "rate": 2000},
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
			EntityContext: map[string]any{"dl_state": "CA", "state": "NY"},
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
			EntityContext: map[string]any{"dl_state": "CA"},
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
				EntityContext: map[string]any{"dl_state": "CA"},
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
				EntityContext: map[string]any{"dl_state": "CA"},
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
				EntityContext: map[string]any{"dl_state": "CA"},
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
				EntityContext: map[string]any{"dl_state": "CA"},
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
				EntityContext: map[string]any{"dl_state": "CA"},
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
				EntityContext: map[string]any{"dl_state": "CA"},
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
			EntityContext: map[string]any{"dl_state": "CA"},
			FlagTags:      []string{"tag1", "tag2"},
		})
		assert.NotZero(t, len(results))
		assert.NotZero(t, results[0].VariantID)
		assert.NotEmpty(t, results[0].FlagTags)
		assert.Len(t, results[0].FlagTags, 2)
		assert.Contains(t, results[0].FlagTags, "tag1")
		assert.Contains(t, results[0].FlagTags, "tag2")
	})

	t.Run("test happy code path with ALL operator", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		op := models.EvaluationBatchRequestFlagTagsOperatorALL
		results := EvalFlagsByTags(models.EvalContext{
			EnableDebug:      true,
			EntityContext:    map[string]any{"dl_state": "CA"},
			FlagTags:         []string{"tag1", "tag2"},
			FlagTagsOperator: &op,
		})
		assert.NotZero(t, len(results))
		assert.NotZero(t, results[0].VariantID)
		assert.NotEmpty(t, results[0].FlagTags)
		assert.Len(t, results[0].FlagTags, 2)
		assert.Contains(t, results[0].FlagTags, "tag1")
		assert.Contains(t, results[0].FlagTags, "tag2")
	})

	t.Run("test happy code path with ANY operator", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		op := models.EvaluationBatchRequestFlagTagsOperatorANY
		results := EvalFlagsByTags(models.EvalContext{
			EnableDebug:      true,
			EntityContext:    map[string]any{"dl_state": "CA"},
			FlagTags:         []string{"tag1"},
			FlagTagsOperator: &op,
		})
		assert.NotZero(t, len(results))
		assert.NotZero(t, results[0].VariantID)
		assert.NotEmpty(t, results[0].FlagTags)
		assert.Len(t, results[0].FlagTags, 2)
		assert.Contains(t, results[0].FlagTags, "tag1")
		assert.Contains(t, results[0].FlagTags, "tag2")
	})

	t.Run("test mixed match with ALL operator", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		op := models.EvaluationBatchRequestFlagTagsOperatorALL
		results := EvalFlagsByTags(models.EvalContext{
			EnableDebug:      true,
			EntityContext:    map[string]any{"dl_state": "CA"},
			FlagTags:         []string{"tag1", "tag_not_exist"},
			FlagTagsOperator: &op,
		})
		assert.Zero(t, len(results))
	})

	t.Run("test mixed match with ANY operator", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		op := models.EvaluationBatchRequestFlagTagsOperatorANY
		results := EvalFlagsByTags(models.EvalContext{
			EnableDebug:      true,
			EntityContext:    map[string]any{"dl_state": "CA"},
			FlagTags:         []string{"tag1", "tag_not_exist"},
			FlagTagsOperator: &op,
		})
		assert.NotZero(t, len(results))
		assert.NotZero(t, results[0].VariantID)
		assert.Contains(t, results[0].FlagTags, "tag1")
	})

	t.Run("test no match", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		results := EvalFlagsByTags(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]any{"dl_state": "CA"},
			FlagTags:      []string{"tag_not_exist"},
		})
		assert.Zero(t, len(results))
	})

	t.Run("test empty tags", func(t *testing.T) {
		defer gostub.StubFunc(&GetEvalCache, GenFixtureEvalCache()).Reset()
		results := EvalFlagsByTags(models.EvalContext{
			EnableDebug:   true,
			EntityContext: map[string]any{"dl_state": "CA"},
			FlagTags:      []string{},
		})
		assert.Zero(t, len(results))
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
				EntityContext: map[string]any{"dl_state": "CA", "state": "NY"},
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
						EntityContext: map[string]any{"dl_state": "CA", "state": "NY"},
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

	t.Run("test duplicate flagKeys are deduplicated", func(t *testing.T) {
		evalCount := 0
		originalEvalFlag := EvalFlag
		EvalFlag = func(evalContext models.EvalContext) *models.EvalResult {
			evalCount++
			return &models.EvalResult{}
		}
		defer func() { EvalFlag = originalEvalFlag }()

		e := NewEval()
		// Send 100 duplicate flagKeys - should only evaluate once
		flagKeys := make([]string, 100)
		for i := range flagKeys {
			flagKeys[i] = "same_flag_key"
		}
		resp := e.PostEvaluationBatch(evaluation.PostEvaluationBatchParams{
			Body: &models.EvaluationBatchRequest{
				EnableDebug: true,
				Entities: []*models.EvaluationEntity{
					{
						EntityContext: map[string]any{"dl_state": "CA"},
						EntityID:      "entityID1",
						EntityType:    "entityType1",
					},
				},
				FlagKeys: flagKeys,
			},
		})
		_, ok := resp.(*evaluation.PostEvaluationBatchOK)
		assert.True(t, ok, "expected PostEvaluationBatchOK response")
		assert.Equal(t, 1, evalCount, "expected only 1 evaluation after deduplication")
	})

	t.Run("test duplicate flagIDs are deduplicated", func(t *testing.T) {
		evalCount := 0
		originalEvalFlag := EvalFlag
		EvalFlag = func(evalContext models.EvalContext) *models.EvalResult {
			evalCount++
			return &models.EvalResult{}
		}
		defer func() { EvalFlag = originalEvalFlag }()

		e := NewEval()
		// Send 100 duplicate flagIDs - should only evaluate once
		flagIDs := make([]int64, 100)
		for i := range flagIDs {
			flagIDs[i] = 123
		}
		resp := e.PostEvaluationBatch(evaluation.PostEvaluationBatchParams{
			Body: &models.EvaluationBatchRequest{
				EnableDebug: true,
				Entities: []*models.EvaluationEntity{
					{
						EntityContext: map[string]any{"dl_state": "CA"},
						EntityID:      "entityID1",
						EntityType:    "entityType1",
					},
				},
				FlagIDs: flagIDs,
			},
		})
		_, ok := resp.(*evaluation.PostEvaluationBatchOK)
		assert.True(t, ok, "expected PostEvaluationBatchOK response")
		assert.Equal(t, 1, evalCount, "expected only 1 evaluation after deduplication")
	})

	t.Run("test mixed duplicates are deduplicated", func(t *testing.T) {
		evalCount := 0
		originalEvalFlag := EvalFlag
		EvalFlag = func(evalContext models.EvalContext) *models.EvalResult {
			evalCount++
			return &models.EvalResult{}
		}
		defer func() { EvalFlag = originalEvalFlag }()

		e := NewEval()
		// 50 duplicates of key_1, 50 duplicates of key_2 = 2 unique keys
		flagKeys := make([]string, 100)
		for i := range flagKeys {
			if i%2 == 0 {
				flagKeys[i] = "flag_key_1"
			} else {
				flagKeys[i] = "flag_key_2"
			}
		}
		resp := e.PostEvaluationBatch(evaluation.PostEvaluationBatchParams{
			Body: &models.EvaluationBatchRequest{
				EnableDebug: true,
				Entities: []*models.EvaluationEntity{
					{
						EntityContext: map[string]any{"dl_state": "CA"},
						EntityID:      "entityID1",
						EntityType:    "entityType1",
					},
				},
				FlagKeys: flagKeys,
			},
		})
		_, ok := resp.(*evaluation.PostEvaluationBatchOK)
		assert.True(t, ok, "expected PostEvaluationBatchOK response")
		assert.Equal(t, 2, evalCount, "expected 2 evaluations after deduplication")
	})

	t.Run("test batch size limit exceeded", func(t *testing.T) {
		originalBatchSize := config.Config.EvalBatchSize
		config.Config.EvalBatchSize = 10 // Set max batch size to 10
		defer func() { config.Config.EvalBatchSize = originalBatchSize }()

		e := NewEval()
		// 10 entities * 2 flagIDs = 20 evaluations (exceeds limit of 10)
		entities := make([]*models.EvaluationEntity, 10)
		for i := range entities {
			entities[i] = &models.EvaluationEntity{
				EntityContext: map[string]any{"dl_state": "CA"},
				EntityID:      fmt.Sprintf("entity%d", i),
				EntityType:    "entityType1",
			}
		}
		resp := e.PostEvaluationBatch(evaluation.PostEvaluationBatchParams{
			Body: &models.EvaluationBatchRequest{
				EnableDebug: true,
				Entities:    entities,
				FlagIDs:     []int64{100, 200},
			},
		})
		_, ok := resp.(*evaluation.PostEvaluationBatchDefault)
		assert.True(t, ok, "expected PostEvaluationBatchDefault response for size exceeded")
	})

	t.Run("test batch size limit not exceeded", func(t *testing.T) {
		originalBatchSize := config.Config.EvalBatchSize
		config.Config.EvalBatchSize = 100 // Set max batch size to 100
		defer func() { config.Config.EvalBatchSize = originalBatchSize }()

		defer gostub.StubFunc(&EvalFlag, &models.EvalResult{}).Reset()

		e := NewEval()
		// 10 entities * 2 flagIDs = 20 evaluations (within limit of 100)
		entities := make([]*models.EvaluationEntity, 10)
		for i := range entities {
			entities[i] = &models.EvaluationEntity{
				EntityContext: map[string]any{"dl_state": "CA"},
				EntityID:      fmt.Sprintf("entity%d", i),
				EntityType:    "entityType1",
			}
		}
		resp := e.PostEvaluationBatch(evaluation.PostEvaluationBatchParams{
			Body: &models.EvaluationBatchRequest{
				EnableDebug: true,
				Entities:    entities,
				FlagIDs:     []int64{100, 200},
			},
		})
		_, ok := resp.(*evaluation.PostEvaluationBatchOK)
		assert.True(t, ok, "expected PostEvaluationBatchOK response when within limit")
	})

	t.Run("test batch size limit disabled when zero", func(t *testing.T) {
		originalBatchSize := config.Config.EvalBatchSize
		config.Config.EvalBatchSize = 0 // Disable batch size limit
		defer func() { config.Config.EvalBatchSize = originalBatchSize }()

		defer gostub.StubFunc(&EvalFlag, &models.EvalResult{}).Reset()

		e := NewEval()
		// Large batch that would normally exceed limit
		entities := make([]*models.EvaluationEntity, 1000)
		for i := range entities {
			entities[i] = &models.EvaluationEntity{
				EntityContext: map[string]any{"dl_state": "CA"},
				EntityID:      fmt.Sprintf("entity%d", i),
				EntityType:    "entityType1",
			}
		}
		resp := e.PostEvaluationBatch(evaluation.PostEvaluationBatchParams{
			Body: &models.EvaluationBatchRequest{
				EnableDebug: true,
				Entities:    entities,
				FlagIDs:     []int64{100, 200},
			},
		})
		_, ok := resp.(*evaluation.PostEvaluationBatchOK)
		assert.True(t, ok, "expected PostEvaluationBatchOK response when limit is disabled")
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
						EntityContext: map[string]any{"dl_state": "CA", "state": "NY"},
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
			EntityContext: map[string]any{"dl_state": "CA"},
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
			EntityContext: map[string]any{"dl_state": "CA"},
			EntityID:      "entityID1",
			EntityType:    "entityType1",
			FlagTags:      []string{"tag1", "tag2"},
		})
	}
}

func genBenchmarkEvalCache(numFlags int) (*EvalCache, []int64, []string) {
	idCache := make(map[string]*entity.Flag, numFlags)
	keyCache := make(map[string]*entity.Flag, numFlags)
	tagCache := make(map[string]map[uint]*entity.Flag)
	flagIDs := make([]int64, 0, numFlags)
	flagKeys := make([]string, 0, numFlags)

	for i := 0; i < numFlags; i++ {
		f := entity.GenFixtureFlag()
		f.ID = uint(100 + i)
		f.Key = fmt.Sprintf("flag_key_%d", 100+i)
		f.Tags = []entity.Tag{
			{Value: "tag1"},
			{Value: "tag2"},
			{Value: fmt.Sprintf("cohort_%d", i%3)},
		}

		for vi := range f.Variants {
			f.Variants[vi].ID = uint(300 + i*10 + vi)
			f.Variants[vi].FlagID = f.ID
		}
		for si := range f.Segments {
			f.Segments[si].FlagID = f.ID
			for di := range f.Segments[si].Distributions {
				if di < len(f.Variants) {
					f.Segments[si].Distributions[di].VariantID = f.Variants[di].ID
					f.Segments[si].Distributions[di].VariantKey = f.Variants[di].Key
				}
			}
		}
		if err := f.PrepareEvaluation(); err != nil {
			panic(err)
		}

		idCache[util.SafeString(f.ID)] = &f
		keyCache[f.Key] = &f
		for _, tag := range f.Tags {
			if _, ok := tagCache[tag.Value]; !ok {
				tagCache[tag.Value] = map[uint]*entity.Flag{}
			}
			tagCache[tag.Value][f.ID] = &f
		}

		flagIDs = append(flagIDs, int64(f.ID))
		flagKeys = append(flagKeys, f.Key)
	}

	return &EvalCache{
		cache: &cacheContainer{
			idCache:  idCache,
			keyCache: keyCache,
			tagCache: tagCache,
		},
	}, flagIDs, flagKeys
}

func BenchmarkPostEvaluationBatch(b *testing.B) {
	b.StopTimer()
	defer gostub.StubFunc(&logEvalResult).Reset()
	evalCache, flagIDs, flagKeys := genBenchmarkEvalCache(10)
	defer gostub.StubFunc(&GetEvalCache, evalCache).Reset()

	numEntities := 10
	entities := make([]*models.EvaluationEntity, numEntities)
	for i := range entities {
		entities[i] = &models.EvaluationEntity{
			EntityContext: map[string]any{"dl_state": "CA", "state": "NY"},
			EntityID:      fmt.Sprintf("entityID%d", i),
			EntityType:    "entityType1",
		}
	}

	e := NewEval()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		e.PostEvaluationBatch(evaluation.PostEvaluationBatchParams{
			Body: &models.EvaluationBatchRequest{
				EnableDebug: false,
				Entities:    entities,
				FlagIDs:     flagIDs,
				FlagKeys:    flagKeys,
			},
		})
	}
}

func BenchmarkPostEvaluationBatchWithTags(b *testing.B) {
	b.StopTimer()
	defer gostub.StubFunc(&logEvalResult).Reset()
	evalCache, _, _ := genBenchmarkEvalCache(10)
	defer gostub.StubFunc(&GetEvalCache, evalCache).Reset()

	numEntities := 10

	entities := make([]*models.EvaluationEntity, numEntities)
	for i := range entities {
		entities[i] = &models.EvaluationEntity{
			EntityContext: map[string]any{"dl_state": "CA", "state": "NY"},
			EntityID:      fmt.Sprintf("entityID%d", i),
			EntityType:    "entityType1",
		}
	}

	e := NewEval()
	op := models.EvaluationBatchRequestFlagTagsOperatorALL

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		e.PostEvaluationBatch(evaluation.PostEvaluationBatchParams{
			Body: &models.EvaluationBatchRequest{
				EnableDebug:      false,
				Entities:         entities,
				FlagTags:         []string{"tag1", "tag2"},
				FlagTagsOperator: &op,
			},
		})
	}
}
