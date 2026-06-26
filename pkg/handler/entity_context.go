package handler

import (
	"encoding/json"
	"fmt"

	"github.com/openflagr/flagr/swagger_gen/models"
)

// normalizeEntityContext converts decoded entityContext (any) into map[string]any
// for constraint evaluation. json.Number leaves from UseNumber become float64.
func normalizeEntityContext(raw any) (map[string]any, error) {
	if raw == nil {
		return nil, nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("entityContext must be a JSON object, got %T", raw)
	}
	return normalizeMap(m), nil
}

func normalizeMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = normalizeValue(v)
	}
	return out
}

func normalizeValue(v any) any {
	switch x := v.(type) {
	case json.Number:
		f, err := x.Float64()
		if err != nil {
			return x.String()
		}
		return f
	case map[string]any:
		return normalizeMap(x)
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = normalizeValue(e)
		}
		return out
	default:
		return v
	}
}

func applyNormalizedEntityContext(ec *models.EvalContext) error {
	if ec == nil || ec.EntityContext == nil {
		return nil
	}
	m, err := normalizeEntityContext(ec.EntityContext)
	if err != nil {
		return err
	}
	ec.EntityContext = m
	return nil
}