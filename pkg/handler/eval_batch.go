package handler

import (
	"net/http"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/swagger_gen/models"
)

// EvaluateBatch runs the same logic as POST/GET /evaluation/batch for the given request body.
// When r is non-nil, built-in context keys (@ts_*, @http_*) are injected per entity (POST path).
func EvaluateBatch(batchReq *models.EvaluationBatchRequest, r *http.Request) (*models.EvaluationBatchResponse, *models.Error) {
	if batchReq == nil {
		return nil, ErrorMessage("empty batch request")
	}
	entities := batchReq.Entities
	flagIDs := batchReq.FlagIDs
	flagKeys := batchReq.FlagKeys
	flagTags := batchReq.FlagTags
	flagTagsOperator := batchReq.FlagTagsOperator

	if len(flagKeys) > 1 {
		seen := make(map[string]struct{}, len(flagKeys))
		uniqueFlagKeys := make([]string, 0, len(flagKeys))
		for _, k := range flagKeys {
			if _, exists := seen[k]; !exists {
				seen[k] = struct{}{}
				uniqueFlagKeys = append(uniqueFlagKeys, k)
			}
		}
		flagKeys = uniqueFlagKeys
	}

	if len(flagIDs) > 1 {
		seen := make(map[int64]struct{}, len(flagIDs))
		uniqueFlagIDs := make([]int64, 0, len(flagIDs))
		for _, id := range flagIDs {
			if _, exists := seen[id]; !exists {
				seen[id] = struct{}{}
				uniqueFlagIDs = append(uniqueFlagIDs, id)
			}
		}
		flagIDs = uniqueFlagIDs
	}

	if maxBatchSize := config.Config.EvalBatchSize; maxBatchSize > 0 {
		flagsPerEntity := len(flagIDs) + len(flagKeys)
		if len(flagTags) > 0 {
			flagsPerEntity++
		}
		if total := len(entities) * flagsPerEntity; total > maxBatchSize {
			return nil, ErrorMessage("batch evaluation size %d exceeds maximum allowed size of %d", total, maxBatchSize)
		}
	}

	est := len(entities) * (len(flagIDs) + len(flagKeys))
	if len(flagTags) > 0 {
		est += len(entities)
	}
	results := &models.EvaluationBatchResponse{
		EvaluationResults: make([]*models.EvalResult, 0, est),
	}

	enableDebug := batchReq.EnableDebug
	for _, entity := range entities {
		if entity != nil {
			entity.EntityContext = InjectBuiltInContext(entity.EntityContext, r)
		}
		if len(flagTags) > 0 {
			evalContext := models.EvalContext{
				EnableDebug:      enableDebug,
				EntityContext:    entity.EntityContext,
				EntityID:         entity.EntityID,
				EntityType:       entity.EntityType,
				FlagTags:         flagTags,
				FlagTagsOperator: flagTagsOperator,
			}
			evalResults := EvalFlagsByTags(evalContext)
			results.EvaluationResults = append(results.EvaluationResults, evalResults...)
		}
		for _, flagID := range flagIDs {
			evalContext := models.EvalContext{
				EnableDebug:   enableDebug,
				EntityContext: entity.EntityContext,
				EntityID:      entity.EntityID,
				EntityType:    entity.EntityType,
				FlagID:        flagID,
			}
			evalResult := EvalFlag(evalContext)
			results.EvaluationResults = append(results.EvaluationResults, evalResult)
		}
		for _, flagKey := range flagKeys {
			evalContext := models.EvalContext{
				EnableDebug:   enableDebug,
				EntityContext: entity.EntityContext,
				EntityID:      entity.EntityID,
				EntityType:    entity.EntityType,
				FlagKey:       flagKey,
			}
			evalResult := EvalFlag(evalContext)
			results.EvaluationResults = append(results.EvaluationResults, evalResult)
		}
	}

	return results, nil
}
