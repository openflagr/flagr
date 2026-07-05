package handler

import (
	"context"
	"net/http"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/openflagr/flagr/swagger_gen/models"
)

type swaggerValidated interface {
	Validate(strfmt.Registry) error
	ContextValidate(context.Context, strfmt.Registry) error
}

func validateSwaggerModelAfterJSON(r *http.Request, label string, v swaggerValidated) *models.Error {
	if v == nil {
		return ErrorMessage("json is not valid %s: empty object", label)
	}
	formats := strfmt.Default
	if err := v.Validate(formats); err != nil {
		return ErrorMessage("json is not valid %s: %v", label, err)
	}
	ctx := context.Background()
	if r != nil {
		ctx = validate.WithOperationRequest(r.Context())
	}
	if err := v.ContextValidate(ctx, formats); err != nil {
		return ErrorMessage("json is not valid %s: %v", label, err)
	}
	return nil
}

// validateEvalContextAfterJSON applies the same checks as POST /evaluation BindRequest
// (body.Validate + body.ContextValidate) after GET json has been unmarshaled.
func validateEvalContextAfterJSON(r *http.Request, evalContext *models.EvalContext) *models.Error {
	if evalContext == nil {
		return ErrorMessage("json is not valid evalContext: empty object")
	}
	return validateSwaggerModelAfterJSON(r, "evalContext", evalContext)
}

func validateEvaluationBatchRequestAfterJSON(r *http.Request, batchReq *models.EvaluationBatchRequest) *models.Error {
	if batchReq == nil {
		return ErrorMessage("json is not valid evaluationBatchRequest: empty object")
	}
	return validateSwaggerModelAfterJSON(r, "evaluationBatchRequest", batchReq)
}
