package handler

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"encoding/json"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/evaluation"
	"gorm.io/gorm"

	"github.com/bsm/ratelimit"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-openapi/runtime/middleware"
	"github.com/zhouzhuojie/conditions"
)

// Eval is the Eval interface
type Eval interface {
	GetEvaluation(evaluation.GetEvaluationParams) middleware.Responder
	GetEvaluationBatch(evaluation.GetEvaluationBatchParams) middleware.Responder
	PostEvaluation(evaluation.PostEvaluationParams) middleware.Responder
	PostEvaluationBatch(evaluation.PostEvaluationBatchParams) middleware.Responder
}

// NewEval creates a new Eval instance
func NewEval() Eval {
	return &eval{}
}

type eval struct{}

func (e *eval) GetEvaluation(params evaluation.GetEvaluationParams) middleware.Responder {
	var evaluationEntity models.GetEvaluationEntity
	err := json.Unmarshal([]byte(params.Entity), &evaluationEntity)
	if err != nil {
		return evaluation.NewGetEvaluationDefault(400).WithPayload(
			ErrorMessage("entity in not a valid evaluation entity"))
	}
	flagTagsOperator := "ANY"
	if params.All != nil && *params.All {
		flagTagsOperator = "ALL"
	}
	evalContext := models.EvalContext{
		EntityID:         evaluationEntity.ID,
		EntityType:       evaluationEntity.Type,
		EntityContext:    evaluationEntity.Ctx,
		FlagTags:         params.Tags,
		FlagTagsOperator: &flagTagsOperator,
	}
	if params.Dbg != nil && *params.Dbg {
		evalContext.EnableDebug = true
	}
	if params.ID != nil {
		evalContext.FlagID = *params.ID
	}
	if params.Key != nil {
		evalContext.FlagKey = *params.Key
	}

	evalResult := EvalFlag(evalContext)
	resp := evaluation.NewPostEvaluationOK()
	resp.SetPayload(evalResult)
	return resp
}

func (e *eval) GetEvaluationBatch(params evaluation.GetEvaluationBatchParams) middleware.Responder {
	var evaluationEntities []*models.EvaluationEntity
	for _, rawEntity := range params.Entity {
		var getEvaluationEntity models.GetEvaluationEntity
		err := json.Unmarshal([]byte(rawEntity), &getEvaluationEntity)
		evaluationEntity := models.EvaluationEntity{
			EntityID:      getEvaluationEntity.ID,
			EntityContext: getEvaluationEntity.Ctx,
			EntityType:    getEvaluationEntity.Type,
		}
		if err != nil {
			return evaluation.NewGetEvaluationBatchDefault(400).WithPayload(
				ErrorMessage("entity is not a valid evaluation entity: %s", rawEntity))
		}
		evaluationEntities = append(evaluationEntities, &evaluationEntity)
	}
	flagTagsOperator := "ANY"
	if params.All != nil && *params.All {
		flagTagsOperator = "ALL"
	}
	var enableDebug = false
	if params.Dbg != nil && *params.Dbg {
		enableDebug = true
	}

	results := EvaluateBatch(
		evaluationEntities,
		params.Ids,
		params.Keys,
		params.Tags,
		&flagTagsOperator,
		enableDebug,
	)

	resp := evaluation.NewPostEvaluationBatchOK()
	resp.SetPayload(results)
	return resp
}

func (e *eval) PostEvaluation(params evaluation.PostEvaluationParams) middleware.Responder {
	evalContext := params.Body
	if evalContext == nil {
		return evaluation.NewPostEvaluationDefault(400).WithPayload(
			ErrorMessage("empty body"))
	}

	evalResult := EvalFlag(*evalContext)
	resp := evaluation.NewPostEvaluationOK()
	resp.SetPayload(evalResult)
	return resp
}

func (e *eval) PostEvaluationBatch(params evaluation.PostEvaluationBatchParams) middleware.Responder {
	results := EvaluateBatch(
		params.Body.Entities,
		params.Body.FlagIDs,
		params.Body.FlagKeys,
		params.Body.FlagTags,
		params.Body.FlagTagsOperator,
		params.Body.EnableDebug,
	)

	resp := evaluation.NewPostEvaluationBatchOK()
	resp.SetPayload(results)
	return resp
}

func EvaluateBatch(
	evaluationEntities []*models.EvaluationEntity,
	flagIDs []int64,
	flagKeys []string,
	flagTags []string,
	flagTagsOperator *string,
	enableDebug bool,
) *models.EvaluationBatchResponse {
	results := &models.EvaluationBatchResponse{}

	// Deduplicate flagKeys to prevent DoS via repeated keys
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

	// Deduplicate flagIDs to prevent DoS via repeated IDs
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

	// Validate batch size to prevent DoS attacks via resource exhaustion (if enabled)
	if maxBatchSize := config.Config.EvalBatchSize; maxBatchSize > 0 {
		// Calculate total evaluations: entities * (flagIDs + flagKeys + flagTags)
		// For flagTags, we count each tag as potentially matching one flag (conservative estimate)
		flagsPerEntity := len(flagIDs) + len(flagKeys)
		if len(flagTags) > 0 {
			flagsPerEntity++ // flagTags is evaluated once per entity regardless of count
		}
		if total := len(entities) * flagsPerEntity; total > maxBatchSize {
			return evaluation.NewPostEvaluationBatchDefault(400).WithPayload(
				ErrorMessage("batch evaluation size %d exceeds maximum allowed size of %d", total, maxBatchSize))
		}
	}

	// TODO make it concurrent
	for _, evaluationEntity := range evaluationEntities {
		if len(flagTags) > 0 {
			evalContext := models.EvalContext{
				EnableDebug:      enableDebug,
				EntityContext:    evaluationEntity.EntityContext,
				EntityID:         evaluationEntity.EntityID,
				EntityType:       evaluationEntity.EntityType,
				FlagTags:         flagTags,
				FlagTagsOperator: flagTagsOperator,
			}
			evalResults := EvalFlagsByTags(evalContext)
			results.EvaluationResults = append(results.EvaluationResults, evalResults...)
		}
		for _, flagID := range flagIDs {
			evalContext := models.EvalContext{
				EnableDebug:   enableDebug,
				EntityContext: evaluationEntity.EntityContext,
				EntityID:      evaluationEntity.EntityID,
				EntityType:    evaluationEntity.EntityType,
				FlagID:        flagID,
			}

			evalResult := EvalFlag(evalContext)
			results.EvaluationResults = append(results.EvaluationResults, evalResult)
		}
		for _, flagKey := range flagKeys {
			evalContext := models.EvalContext{
				EnableDebug:   enableDebug,
				EntityContext: evaluationEntity.EntityContext,
				EntityID:      evaluationEntity.EntityID,
				EntityType:    evaluationEntity.EntityType,
				FlagKey:       flagKey,
			}

			evalResult := EvalFlag(evalContext)
			results.EvaluationResults = append(results.EvaluationResults, evalResult)
		}
	}

	return results
}

// BlankResult creates a blank result
func BlankResult(f *entity.Flag, evalContext models.EvalContext, msg string) *models.EvalResult {
	flagID := uint(0)
	flagKey := ""
	flagSnapshotID := uint(0)
	flagTags := []string{}
	if f != nil {
		flagID = f.ID
		flagSnapshotID = f.SnapshotID
		flagKey = f.Key
		if len(f.Tags) > 0 {
			for _, tag := range f.Tags {
				flagTags = append(flagTags, tag.Value)
			}
		}
	}
	return &models.EvalResult{
		EvalContext: &evalContext,
		EvalDebugLog: &models.EvalDebugLog{
			Msg:              msg,
			SegmentDebugLogs: nil,
		},
		FlagID:         int64(flagID),
		FlagKey:        flagKey,
		FlagSnapshotID: int64(flagSnapshotID),
		FlagTags:       flagTags,
		Timestamp:      util.TimeNow(),
	}
}

var LookupFlag = func(evalContext models.EvalContext) *entity.Flag {
	cache := GetEvalCache()
	flagID := util.SafeUint(evalContext.FlagID)
	flagKey := util.SafeString(evalContext.FlagKey)
	f := cache.GetByFlagKeyOrID(flagID)
	if f == nil {
		f = cache.GetByFlagKeyOrID(flagKey)
	}
	return f
}

var EvalFlagsByTags = func(evalContext models.EvalContext) []*models.EvalResult {
	cache := GetEvalCache()
	fs := cache.GetByTags(evalContext.FlagTags, evalContext.FlagTagsOperator)
	results := []*models.EvalResult{}
	for _, f := range fs {
		results = append(results, EvalFlagWithContext(f, evalContext))
	}
	return results
}

var EvalFlag = func(evalContext models.EvalContext) *models.EvalResult {
	flag := LookupFlag(evalContext)
	return EvalFlagWithContext(flag, evalContext)
}

var EvalFlagWithContext = func(flag *entity.Flag, evalContext models.EvalContext) *models.EvalResult {
	flagID := util.SafeUint(evalContext.FlagID)
	flagKey := util.SafeString(evalContext.FlagKey)

	if flag == nil {
		emptyFlag := &entity.Flag{Model: gorm.Model{ID: flagID}, Key: flagKey}
		return BlankResult(emptyFlag, evalContext, fmt.Sprintf("flagID %v not found or deleted", flagID))
	}

	if !flag.Enabled {
		return BlankResult(flag, evalContext, fmt.Sprintf("flagID %v is not enabled", flag.ID))
	}

	if len(flag.Segments) == 0 {
		return BlankResult(flag, evalContext, fmt.Sprintf("flagID %v has no segments", flag.ID))
	}

	if evalContext.EntityID == "" {
		evalContext.EntityID = fmt.Sprintf("randomly_generated_%d", rand.Int31())
	}

	if flag.EntityType != "" {
		evalContext.EntityType = flag.EntityType
	}

	logs := []*models.SegmentDebugLog{}
	var vID int64
	var sID int64

	for _, segment := range flag.Segments {
		sID = int64(segment.ID)
		variantID, log, evalNextSegment := evalSegment(flag.ID, evalContext, segment)
		if config.Config.EvalDebugEnabled && evalContext.EnableDebug {
			logs = append(logs, log)
		}
		if variantID != nil {
			vID = int64(*variantID)
		}
		if !evalNextSegment {
			break
		}
	}
	evalResult := BlankResult(flag, evalContext, "")
	evalResult.EvalDebugLog.SegmentDebugLogs = logs
	evalResult.SegmentID = sID
	evalResult.VariantID = vID
	v := flag.FlagEvaluation.VariantsMap[util.SafeUint(vID)]
	if v != nil {
		evalResult.VariantAttachment = v.Attachment
		evalResult.VariantKey = v.Key
	}

	logEvalResult(evalResult, flag.DataRecordsEnabled)
	return evalResult
}

var logEvalResult = func(r *models.EvalResult, dataRecordsEnabled bool) {
	if r == nil {
		// this is just a safety check, r is from BlankResult,
		// and usually it cannot be nil
		return
	}

	if config.Config.EvalLoggingEnabled {
		rateLimitPerFlagConsoleLogging(r)
	}

	logEvalResultToDatadog(r)
	logEvalResultToPrometheus(r)

	if !config.Config.RecorderEnabled || !dataRecordsEnabled {
		return
	}
	rec := GetDataRecorder()
	rec.AsyncRecord(*r)
}

var logEvalResultToDatadog = func(r *models.EvalResult) {
	if config.Global.StatsdClient == nil {
		return
	}

	config.Global.StatsdClient.Incr(
		"evaluation",
		[]string{
			fmt.Sprintf("FlagID:%d", util.SafeUint(r.FlagID)),
			fmt.Sprintf("VariantID:%d", util.SafeUint(r.VariantID)),
			fmt.Sprintf("VariantKey:%s", util.SafeStringWithDefault(r.VariantKey, "null")),
		},
		float64(1),
	)
}

var logEvalResultToPrometheus = func(r *models.EvalResult) {
	if config.Global.Prometheus.EvalCounter == nil {
		return
	}
	config.Global.Prometheus.EvalCounter.WithLabelValues(
		util.SafeStringWithDefault(r.EvalContext.EntityType, "null"),
		util.SafeStringWithDefault(r.FlagID, "null"),
		util.SafeStringWithDefault(r.FlagKey, "null"),
		util.SafeStringWithDefault(r.VariantID, "null"),
		util.SafeStringWithDefault(r.VariantKey, "null"),
	).Inc()

}

var evalSegment = func(
	flagID uint,
	evalContext models.EvalContext,
	segment entity.Segment,
) (
	vID *uint, // returns VariantID
	log *models.SegmentDebugLog,
	evalNextSegment bool,
) {
	if len(segment.Constraints) != 0 {
		m, ok := evalContext.EntityContext.(map[string]any)
		if !ok {
			log = &models.SegmentDebugLog{
				Msg:       fmt.Sprintf("constraints are present in the segment_id %v, but got invalid entity_context: %s.", segment.ID, spew.Sdump(evalContext.EntityContext)),
				SegmentID: int64(segment.ID),
			}
			return nil, log, true
		}

		expr := segment.SegmentEvaluation.ConditionsExpr
		match, err := conditions.Evaluate(expr, m)
		if err != nil {
			log = &models.SegmentDebugLog{
				Msg:       err.Error(),
				SegmentID: int64(segment.ID),
			}
			return nil, log, true
		}
		if !match {
			log = &models.SegmentDebugLog{
				Msg:       debugConstraintMsg(evalContext.EnableDebug, expr, m),
				SegmentID: int64(segment.ID),
			}
			return nil, log, true
		}
	}

	vID, debugMsg := segment.SegmentEvaluation.DistributionArray.Rollout(
		evalContext.EntityID,
		fmt.Sprint(flagID), // default use the flagID as salt
		segment.RolloutPercent,
	)

	log = &models.SegmentDebugLog{
		Msg:       "matched all constraints. " + debugMsg,
		SegmentID: int64(segment.ID),
	}

	// at this point, all constraints are matched, so we shouldn't go to next segment
	// thus setting evalNextSegment = false
	return vID, log, false
}

func debugConstraintMsg(enableDebug bool, expr conditions.Expr, m map[string]any) string {
	if !enableDebug {
		return ""
	}
	return fmt.Sprintf("constraint not match. constraint: %s, entity_context: %+v.", expr, m)
}

var rateLimitMap = sync.Map{}

var rateLimitPerFlagConsoleLogging = func(r *models.EvalResult) {
	flagID := util.SafeUint(r.FlagID)
	rl, _ := rateLimitMap.LoadOrStore(flagID, ratelimit.New(
		config.Config.RateLimiterPerFlagPerSecondConsoleLogging,
		time.Second,
	))
	if !rl.(*ratelimit.RateLimiter).Limit() {
		jsonStr, _ := json.Marshal(struct{ FlagEvalResult *models.EvalResult }{FlagEvalResult: r})
		fmt.Println(string(jsonStr))
	}
}
