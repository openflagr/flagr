package handler

import (
	"encoding/json"
	"testing"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeEntityContext_jsonNumber(t *testing.T) {
	raw := map[string]any{"foo": json.Number("42")}
	m, err := normalizeEntityContext(raw)
	require.NoError(t, err)
	assert.Equal(t, float64(42), m["foo"])
}

func TestNormalizeEntityContext_nested(t *testing.T) {
	raw := map[string]any{
		"user": map[string]any{
			"age": json.Number("25"),
		},
		"roles": []any{json.Number("1"), "admin"},
	}
	m, err := normalizeEntityContext(raw)
	require.NoError(t, err)
	user := m["user"].(map[string]any)
	assert.Equal(t, float64(25), user["age"])
	roles := m["roles"].([]any)
	assert.Equal(t, float64(1), roles[0])
}

func TestNormalizeEntityContext_invalidType(t *testing.T) {
	_, err := normalizeEntityContext("not-a-map")
	assert.Error(t, err)
}

func TestApplyNormalizedEntityContext(t *testing.T) {
	ec := &models.EvalContext{
		EntityContext: map[string]any{"n": json.Number("3.14")},
	}
	require.NoError(t, applyNormalizedEntityContext(ec))
	assert.Equal(t, float64(3.14), ec.EntityContext.(map[string]any)["n"])
}