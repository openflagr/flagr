package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

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

func evalGetQueryTooLong(rawQueryLen int) *models.Error {
	max := config.Config.EvalGetMaxURLBytes
	if max <= 0 || rawQueryLen <= max {
		return nil
	}
	return ErrorMessage("GET evaluation query length %d exceeds maximum of %d; use POST", rawQueryLen, max)
}

func (e *eval) GetEvaluation(params evaluation.GetEvaluationParams) middleware.Responder {
	if errPayload := evalGetQueryTooLong(len(params.HTTPRequest.URL.RawQuery)); errPayload != nil {
		return evaluation.NewGetEvaluationDefault(400).WithPayload(errPayload)
	}
	if params.JSON == "" {
		return evaluation.NewGetEvaluationDefault(400).WithPayload(
			ErrorMessage("missing required query parameter json"))
	}
	var evalContext models.EvalContext
	if err := json.Unmarshal([]byte(params.JSON), &evalContext); err != nil {
		return evaluation.NewGetEvaluationDefault(400).WithPayload(
			ErrorMessage("json is not valid evalContext: %v", err))
	}
	evalResult := EvalFlag(evalContext)
	resp := evaluation.NewGetEvaluationOK()
	resp.SetPayload(evalResult)
	return resp
}

func (e *eval) GetEvaluationBatch(params evaluation.GetEvaluationBatchParams) middleware.Responder {
	if errPayload := evalGetQueryTooLong(len(params.HTTPRequest.URL.RawQuery)); errPayload != nil {
		return evaluation.NewGetEvaluationBatchDefault(400).WithPayload(errPayload)
	}
	if params.JSON == "" {
		return evaluation.NewGetEvaluationBatchDefault(400).WithPayload(
			ErrorMessage("missing required query parameter json"))
	}
	var batchReq models.EvaluationBatchRequest
	if err := json.Unmarshal([]byte(params.JSON), &batchReq); err != nil {
		return evaluation.NewGetEvaluationBatchDefault(400).WithPayload(
			ErrorMessage("json is not valid evaluationBatchRequest: %v", err))
	}
	results, errPayload := EvaluateBatch(&batchReq, nil)
	if errPayload != nil {
		return evaluation.NewGetEvaluationBatchDefault(400).WithPayload(errPayload)
	}
	resp := evaluation.NewGetEvaluationBatchOK()
	resp.SetPayload(results)
	return resp
}

func (e *eval) PostEvaluation(params evaluation.PostEvaluationParams) middleware.Responder {
	evalContext := params.Body
	if evalContext == nil {
		return evaluation.NewPostEvaluationDefault(400).WithPayload(
			ErrorMessage("empty body"))
	}

	// Inject built-in context keys (@ts_*, @http_*) into entityContext
	evalContext.EntityContext = InjectBuiltInContext(evalContext.EntityContext, params.HTTPRequest)

	evalResult := EvalFlag(*evalContext)
	resp := evaluation.NewPostEvaluationOK()
	resp.SetPayload(evalResult)
	return resp
}

func (e *eval) PostEvaluationBatch(params evaluation.PostEvaluationBatchParams) middleware.Responder {
	if params.Body == nil {
		return evaluation.NewPostEvaluationBatchDefault(400).WithPayload(
			ErrorMessage("empty body"))
	}
	results, errPayload := EvaluateBatch(params.Body, params.HTTPRequest)
	if errPayload != nil {
		return evaluation.NewPostEvaluationBatchDefault(400).WithPayload(errPayload)
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
	var flagTags []string
	var dataRecordsEnabled bool
	if f != nil {
		flagID = f.ID
		flagSnapshotID = f.SnapshotID
		flagKey = f.Key
		flagTags = f.FlagEvaluation.TagValues
		dataRecordsEnabled = f.DataRecordsEnabled
	}
	ec := evalContext
	return &models.EvalResult{
		EvalContext: &ec,
		EvalDebugLog: &models.EvalDebugLog{
			Msg:              msg,
			SegmentDebugLogs: nil,
		},
		FlagID:             int64(flagID),
		FlagKey:            flagKey,
		FlagSnapshotID:     int64(flagSnapshotID),
		FlagTags:           flagTags,
		Timestamp:          util.TimeNow(),
		RecordSource:       models.EvalResultRecordSourceEvaluation,
		DataRecordsEnabled: dataRecordsEnabled,
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
	results := make([]*models.EvalResult, 0, len(fs))
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

	var vID int64
	var sID int64
	var logs []*models.SegmentDebugLog
	if config.Config.EvalDebugEnabled && evalContext.EnableDebug {
		logs = make([]*models.SegmentDebugLog, 0, len(flag.Segments))
	}
	for _, segment := range flag.Segments {
		variantID, log, evalNextSegment := evalSegment(evalContext, segment)
		if variantID != nil {
			vID = int64(*variantID)
			sID = int64(segment.ID)
		}
		if config.Config.EvalDebugEnabled && evalContext.EnableDebug {
			logs = append(logs, log)
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

	logEvalResult(evalResult, flag)
	return evalResult
}

var logEvalResult = func(r *models.EvalResult, flag *entity.Flag) {
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

	if dataRecordEnabled(flag) {
		GetDataRecorder().AsyncRecord(*r)
	}
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
	evalContext models.EvalContext,
	segment entity.Segment,
) (
	vID *uint, // returns VariantID
	log *models.SegmentDebugLog,
	evalNextSegment bool,
) {
	debug := config.Config.EvalDebugEnabled && evalContext.EnableDebug

	if len(segment.Constraints) != 0 {
		m, ok := evalContext.EntityContext.(map[string]any)
		if !ok {
			if debug {
				log = &models.SegmentDebugLog{
					Msg:       fmt.Sprintf("constraints are present in the segment_id %v, but got invalid entity_context: %s.", segment.ID, spew.Sdump(evalContext.EntityContext)),
					SegmentID: int64(segment.ID),
				}
			}
			return nil, log, true
		}

		expr := segment.SegmentEvaluation.ConditionsExpr
		match, err := conditions.Evaluate(expr, m)
		if err != nil {
			if debug {
				log = &models.SegmentDebugLog{
					Msg:       err.Error(),
					SegmentID: int64(segment.ID),
				}
			}
			return nil, log, true
		}
		if !match {
			if debug {
				log = &models.SegmentDebugLog{
					Msg:       debugConstraintMsg(true, expr, m),
					SegmentID: int64(segment.ID),
				}
			}
			return nil, log, true
		}
	}

	var debugMsg string
	vID, debugMsg = segment.SegmentEvaluation.DistributionArray.Rollout(
		evalContext.EntityID,
		segment.SegmentEvaluation.FlagIDStr,
		segment.RolloutPercent,
		debug,
	)

	if debug {
		log = &models.SegmentDebugLog{
			Msg:       "matched all constraints. " + debugMsg,
			SegmentID: int64(segment.ID),
		}
	}

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
