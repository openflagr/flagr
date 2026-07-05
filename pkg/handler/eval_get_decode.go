package handler

import (
	"encoding/json"
	"net/http"

	"github.com/openflagr/flagr/swagger_gen/models"
)

// decodeEvalContextFromGet applies GET /evaluation checks: raw query length (as received),
// non-empty json param, JSON syntax, then POST-parity schema validation.
func decodeEvalContextFromGet(req *http.Request, rawQueryLen int, jsonParam string) (models.EvalContext, *models.Error) {
	var zero models.EvalContext
	if errPayload := evalGetQueryTooLong(rawQueryLen); errPayload != nil {
		return zero, errPayload
	}
	if jsonParam == "" {
		return zero, ErrorMessage("missing required query parameter json")
	}
	var evalContext models.EvalContext
	if err := json.Unmarshal([]byte(jsonParam), &evalContext); err != nil {
		return zero, ErrorMessage("json is not valid evalContext: %v", err)
	}
	if errPayload := validateEvalContextAfterJSON(req, &evalContext); errPayload != nil {
		return zero, errPayload
	}
	return evalContext, nil
}

// decodeEvaluationBatchFromGet applies GET /evaluation/batch checks (same stages as single eval).
func decodeEvaluationBatchFromGet(req *http.Request, rawQueryLen int, jsonParam string) (models.EvaluationBatchRequest, *models.Error) {
	var zero models.EvaluationBatchRequest
	if errPayload := evalGetQueryTooLong(rawQueryLen); errPayload != nil {
		return zero, errPayload
	}
	if jsonParam == "" {
		return zero, ErrorMessage("missing required query parameter json")
	}
	var batchReq models.EvaluationBatchRequest
	if err := json.Unmarshal([]byte(jsonParam), &batchReq); err != nil {
		return zero, ErrorMessage("json is not valid evaluationBatchRequest: %v", err)
	}
	if errPayload := validateEvaluationBatchRequestAfterJSON(req, &batchReq); errPayload != nil {
		return zero, errPayload
	}
	return batchReq, nil
}
