package handler

import (
	"fmt"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/evaluation"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-openapi/runtime/middleware"
	"github.com/zhouzhuojie/conditions"
)

// Eval is the Eval interface
type Eval interface {
	PostEvaluation(evaluation.PostEvaluationParams) middleware.Responder
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

	evalResult, err := evalFlag(evalContext)
	if err != nil {
		return evaluation.NewPostEvaluationDefault(err.StatusCode).WithPayload(
			ErrorMessage("error evaluating. reason: %s. context: %s", err, spew.Sdump(evalContext)))
	}
	resp := evaluation.NewPostEvaluationOK()
	resp.SetPayload(evalResult)
	return resp
}

func evalFlag(evalContext *models.EvalContext) (*models.EvalResult, *Error) {
	if evalContext == nil {
		return nil, NewError(400, "empty evalContext")
	}

	cache := GetEvalCache()
	flagID := util.SafeUint(evalContext.FlagID)
	f := cache.GetByFlagID(flagID)
	if f == nil {
		return nil, NewError(404, "flagID not found: %v", flagID)
	}

	logs := []*models.SegmentDebugLog{}

	var vID *int64
	var sID *int64

	for _, segment := range f.Segments {
		variantID, log, err := evalSegment(evalContext, segment)
		if err != nil {
			return nil, err
		}
		if evalContext.EnableDebug {
			logs = append(logs, log)
		}
		if variantID != nil {
			sID = util.Int64Ptr(int64(segment.ID))
			vID = util.Int64Ptr(int64(*variantID))
			break
		}
	}

	evalResult := &models.EvalResult{
		EvalContext: evalContext,
		EvalDebugLog: &models.EvalDebugLog{
			Msg:              "",
			SegmentDebugLogs: logs,
		},
		FlagID:    util.Int64Ptr(int64(f.ID)),
		SegmentID: sID,
		VariantID: vID,
		Timestamp: util.StringPtr(util.TimeNow()),
	}

	return evalResult, nil
}

func evalSegment(
	evalContext *models.EvalContext,
	segment entity.Segment,
) (
	vID *uint, // returns VariantID
	log *models.SegmentDebugLog,
	evalErr *Error,
) {
	if len(segment.SegmentEvaluation.Conditions) != 0 {
		m, ok := evalContext.EntityContext.(map[string]interface{})
		if !ok {
			evalErr = NewError(400, "constraints are present in the segment_id %v, but got invalid entity_context: %s.", segment.ID, spew.Sdump(evalContext.EntityContext))
			return nil, nil, evalErr
		}

		for _, expr := range segment.SegmentEvaluation.Conditions {

			match, err := conditions.Evaluate(expr, m)
			if err != nil {
				evalErr = NewError(400, "invalid entity_context: %s. reason: %s.", spew.Sdump(evalContext.EntityContext), err)
				return nil, nil, evalErr
			}
			if !match {
				log = &models.SegmentDebugLog{
					Msg:       debugConstraintMsg(expr, m),
					SegmentID: int64(segment.ID),
				}
				return nil, log, nil
			}
		}
	}

	vID, debugMsg := segment.SegmentEvaluation.DistributionArray.Rollout(
		evalContext.EntityID,
		fmt.Sprint(evalContext.FlagID), // default use the flagID as salt
		segment.RolloutPercent,
	)

	log = &models.SegmentDebugLog{
		Msg:       "matched all constraints. " + debugMsg,
		SegmentID: int64(segment.ID),
	}

	return vID, log, evalErr
}

func debugConstraintMsg(expr conditions.Expr, m map[string]interface{}) string {
	return fmt.Sprintf("constraint not match. constraint: %s, entity_context: %+v.", expr, m)
}
