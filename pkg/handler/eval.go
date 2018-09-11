package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/evaluation"

	"github.com/bsm/ratelimit"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-openapi/runtime/middleware"
	"github.com/zhouzhuojie/conditions"
)

// Eval is the Eval interface
type Eval interface {
	PostEvaluation(evaluation.PostEvaluationParams) middleware.Responder
	PostEvaluationBatch(evaluation.PostEvaluationBatchParams) middleware.Responder
}

// NewEval creates a new Eval instance
func NewEval() Eval {
	return &eval{}
}

type eval struct{}

func (e *eval) PostEvaluation(params evaluation.PostEvaluationParams) middleware.Responder {
	evalContext := params.Body
	if evalContext == nil {
		return evaluation.NewPostEvaluationDefault(400).WithPayload(
			ErrorMessage("empty body"))
	}

	evalResult := evalFlag(*evalContext)
	resp := evaluation.NewPostEvaluationOK()
	resp.SetPayload(evalResult)
	return resp
}

func (e *eval) PostEvaluationBatch(params evaluation.PostEvaluationBatchParams) middleware.Responder {
	entities := params.Body.Entities
	flagIDs := params.Body.FlagIds
	flagKeys := params.Body.FlagKeys
	results := &models.EvaluationBatchResponse{}

	// TODO make it concurrent
	for _, entity := range entities {
		for _, flagID := range flagIDs {
			evalContext := models.EvalContext{
				EnableDebug:   params.Body.EnableDebug,
				EntityContext: entity.EntityContext,
				EntityID:      entity.EntityID,
				EntityType:    entity.EntityType,
				FlagID:        flagID,
			}
			evalResult := evalFlag(evalContext)
			results.EvaluationResults = append(results.EvaluationResults, evalResult)
		}
		for _, flagKey := range flagKeys {
			evalContext := models.EvalContext{
				EnableDebug:   params.Body.EnableDebug,
				EntityContext: entity.EntityContext,
				EntityID:      entity.EntityID,
				EntityType:    entity.EntityType,
				FlagKey:       flagKey,
			}
			evalResult := evalFlag(evalContext)
			results.EvaluationResults = append(results.EvaluationResults, evalResult)
		}
	}

	resp := evaluation.NewPostEvaluationBatchOK()
	resp.SetPayload(results)
	return resp
}

// BlankResult creates a blank result
func BlankResult(f *entity.Flag, evalContext models.EvalContext, msg string) *models.EvalResult {
	flagID := uint(0)
	flagKey := ""
	flagSnapshotID := uint(0)
	if f != nil {
		flagID = f.ID
		flagSnapshotID = f.SnapshotID
		flagKey = f.Key
	}
	return &models.EvalResult{
		EvalContext: &evalContext,
		EvalDebugLog: &models.EvalDebugLog{
			Msg:              msg,
			SegmentDebugLogs: nil,
		},
		FlagID:         util.Int64Ptr(int64(flagID)),
		FlagKey:        util.StringPtr(flagKey),
		FlagSnapshotID: int64(flagSnapshotID),
		SegmentID:      nil,
		VariantID:      nil,
		Timestamp:      util.StringPtr(util.TimeNow()),
	}
}

var evalFlag = func(evalContext models.EvalContext) *models.EvalResult {
	cache := GetEvalCache()
	flagID := util.SafeUint(evalContext.FlagID)
	flagKey := util.SafeString(evalContext.FlagKey)
	f := cache.GetByFlagKeyOrID(flagID)
	if f == nil {
		f = cache.GetByFlagKeyOrID(flagKey)
	}

	if f == nil {
		return BlankResult(nil, evalContext, fmt.Sprintf("flagID %v not found", flagID))
	}

	if !f.Enabled {
		return BlankResult(f, evalContext, fmt.Sprintf("flagID %v is not enabled", f.ID))
	}

	if len(f.Segments) == 0 {
		return BlankResult(f, evalContext, fmt.Sprintf("flagID %v has no segments", f.ID))
	}

	if evalContext.EntityID == "" {
		evalContext.EntityID = fmt.Sprintf("randomly_generated_%d", rand.Int31())
	}

	logs := []*models.SegmentDebugLog{}
	var vID *int64
	var sID *int64

	for _, segment := range f.Segments {
		variantID, log := evalSegment(f.ID, evalContext, segment)
		if evalContext.EnableDebug {
			logs = append(logs, log)
		}
		if variantID != nil {
			sID = util.Int64Ptr(int64(segment.ID))
			vID = util.Int64Ptr(int64(*variantID))
			break
		}
	}
	evalResult := BlankResult(f, evalContext, "")
	evalResult.EvalDebugLog.SegmentDebugLogs = logs
	evalResult.SegmentID = sID
	evalResult.VariantID = vID
	v := f.FlagEvaluation.VariantsMap[util.SafeUint(vID)]
	if v != nil {
		evalResult.VariantAttachment = v.Attachment
		evalResult.VariantKey = util.StringPtr(v.Key)
	}

	logEvalResult(evalResult, f.DataRecordsEnabled)
	return evalResult
}

var logEvalResult = func(r *models.EvalResult, dataRecordsEnabled bool) {
	if config.Config.EvalLoggingEnabled {
		rateLimitPerFlagConsoleLogging(r)
	}

	if !config.Config.RecorderEnabled || !dataRecordsEnabled {
		return
	}
	rec := GetDataRecorder()
	rec.AsyncRecord(r)
}

var evalSegment = func(
	flagID uint,
	evalContext models.EvalContext,
	segment entity.Segment,
) (
	vID *uint, // returns VariantID
	log *models.SegmentDebugLog,
) {
	if len(segment.Constraints) != 0 {
		m, ok := evalContext.EntityContext.(map[string]interface{})
		if !ok {
			log = &models.SegmentDebugLog{
				Msg:       fmt.Sprintf("constraints are present in the segment_id %v, but got invalid entity_context: %s.", segment.ID, spew.Sdump(evalContext.EntityContext)),
				SegmentID: int64(segment.ID),
			}
			return nil, log
		}

		expr := segment.SegmentEvaluation.ConditionsExpr
		match, err := conditions.Evaluate(expr, m)
		if err != nil {
			log = &models.SegmentDebugLog{
				Msg:       err.Error(),
				SegmentID: int64(segment.ID),
			}
			return nil, log
		}
		if !match {
			log = &models.SegmentDebugLog{
				Msg:       debugConstraintMsg(expr, m),
				SegmentID: int64(segment.ID),
			}
			return nil, log
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

	return vID, log
}

func debugConstraintMsg(expr conditions.Expr, m map[string]interface{}) string {
	return fmt.Sprintf("constraint not match. constraint: %s, entity_context: %+v.", expr, m)
}

var rateLimitMap = make(map[uint]*ratelimit.RateLimiter)

var rateLimitPerFlagConsoleLogging = func(r *models.EvalResult) {
	flagID := util.SafeUint(r.FlagID)
	rl, ok := rateLimitMap[flagID]
	if !ok {
		rl = ratelimit.New(
			config.Config.RateLimiterPerFlagPerSecondConsoleLogging,
			time.Second,
		)
		rateLimitMap[flagID] = rl
	}
	if !rl.Limit() {
		jsonStr, _ := json.Marshal(struct{ FlagEvalResult *models.EvalResult }{FlagEvalResult: r})
		fmt.Println(string(jsonStr))
	}
}
