package handler

import (
	"context"
	"net/http"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/openflagr/flagr/swagger_gen/models"
)

// validateEvalContextAfterJSON applies the same checks as POST /evaluation BindRequest
// (body.Validate + body.ContextValidate) after GET json has been unmarshaled.
func validateEvalContextAfterJSON(r *http.Request, evalContext *models.EvalContext) *models.Error {
	if evalContext == nil {
		return ErrorMessage("json is not valid evalContext: empty object")
	}
	formats := strfmt.Default
	if err := evalContext.Validate(formats); err != nil {
		return ErrorMessage("json is not valid evalContext: %v", err)
	}
	ctx := context.Background()
	if r != nil {
		ctx = validate.WithOperationRequest(r.Context())
	}
	if err := evalContext.ContextValidate(ctx, formats); err != nil {
		return ErrorMessage("json is not valid evalContext: %v", err)
	}
	return nil
}

// validateEvaluationBatchRequestAfterJSON applies the same checks as POST /evaluation/batch BindRequest.
func validateEvaluationBatchRequestAfterJSON(r *http.Request, batchReq *models.EvaluationBatchRequest) *models.Error {
	if batchReq == nil {
		return ErrorMessage("json is not valid evaluationBatchRequest: empty object")
	}
	formats := strfmt.Default
	if err := batchReq.Validate(formats); err != nil {
		return ErrorMessage("json is not valid evaluationBatchRequest: %v", err)
	}
	ctx := context.Background()
	if r != nil {
		ctx = validate.WithOperationRequest(r.Context())
	}
	if err := batchReq.ContextValidate(ctx, formats); err != nil {
		return ErrorMessage("json is not valid evaluationBatchRequest: %v", err)
	}
	return nil
}
