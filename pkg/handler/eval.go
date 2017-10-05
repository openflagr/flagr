package handler

import (
	"github.com/checkr/flagr/swagger_gen/restapi/operations/evaluation"
	"github.com/go-openapi/runtime/middleware"
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

func (e *eval) PostEvaluation(evaluation.PostEvaluationParams) middleware.Responder {
	return nil
}
