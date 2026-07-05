package handler

import (
	"net/http"
	"testing"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateEvalContextAfterJSON(t *testing.T) {
	t.Parallel()
	req := &http.Request{}

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		ec := models.EvalContext{FlagID: 1, EntityID: "e1"}
		assert.Nil(t, validateEvalContextAfterJSON(req, &ec))
	})

	t.Run("flagID negative", func(t *testing.T) {
		t.Parallel()
		ec := models.EvalContext{FlagID: -1, EntityID: "e1"}
		err := validateEvalContextAfterJSON(req, &ec)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "flagID")
	})
}

func TestValidateEvaluationBatchRequestAfterJSON(t *testing.T) {
	t.Parallel()
	req := &http.Request{}

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		b := models.EvaluationBatchRequest{
			Entities: []*models.EvaluationEntity{{EntityID: "e1"}},
			FlagIDs:  []int64{1},
		}
		assert.Nil(t, validateEvaluationBatchRequestAfterJSON(req, &b))
	})

	t.Run("missing entities", func(t *testing.T) {
		t.Parallel()
		b := models.EvaluationBatchRequest{FlagIDs: []int64{1}}
		err := validateEvaluationBatchRequestAfterJSON(req, &b)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "entities")
	})
}
