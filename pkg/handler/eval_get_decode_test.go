package handler

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeEvalContextFromGet(t *testing.T) {
	t.Parallel()
	req := &http.Request{URL: &url.URL{RawQuery: "json=%7B%7D"}}

	t.Run("missing json", func(t *testing.T) {
		t.Parallel()
		_, err := decodeEvalContextFromGet(req, 10, "")
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "missing required query parameter json")
	})

	t.Run("invalid syntax", func(t *testing.T) {
		t.Parallel()
		_, err := decodeEvalContextFromGet(req, 10, "{")
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "json is not valid evalContext")
	})

	t.Run("schema violation", func(t *testing.T) {
		t.Parallel()
		_, err := decodeEvalContextFromGet(req, 10, `{"flagID":-1,"entityID":"e1"}`)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "flagID")
	})

	t.Run("query too long", func(t *testing.T) {
		t.Parallel()
		stubs := gostub.Stub(&config.Config.EvalGetMaxURLBytes, 8)
		defer stubs.Reset()
		_, err := decodeEvalContextFromGet(req, 9, `{"flagID":1}`)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "exceeds maximum")
	})

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		ec, err := decodeEvalContextFromGet(req, 50, `{"flagID":1,"entityID":"e1"}`)
		require.Nil(t, err)
		assert.Equal(t, int64(1), ec.FlagID)
	})
}

func TestDecodeEvaluationBatchFromGet(t *testing.T) {
	t.Parallel()
	req := &http.Request{URL: &url.URL{RawQuery: "json=%7B%7D"}}

	t.Run("missing entities", func(t *testing.T) {
		t.Parallel()
		_, err := decodeEvaluationBatchFromGet(req, 10, `{"flagIDs":[1]}`)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "entities")
	})

	t.Run("missing json", func(t *testing.T) {
		t.Parallel()
		_, err := decodeEvaluationBatchFromGet(req, 10, "")
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "missing required query parameter json")
	})

	t.Run("query too long", func(t *testing.T) {
		t.Parallel()
		stubs := gostub.Stub(&config.Config.EvalGetMaxURLBytes, 8)
		defer stubs.Reset()
		_, err := decodeEvaluationBatchFromGet(req, 9, `{"entities":[{"entityID":"e1"}]}`)
		require.NotNil(t, err)
		assert.Contains(t, *err.Message, "exceeds maximum")
	})

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		b, err := decodeEvaluationBatchFromGet(req, 80, `{"entities":[{"entityID":"e1"}],"flagIDs":[1]}`)
		require.Nil(t, err)
		require.Len(t, b.Entities, 1)
	})
}

func TestValidateSwaggerModelAfterJSON_nil(t *testing.T) {
	t.Parallel()
	err := validateSwaggerModelAfterJSON(nil, "evalContext", nil)
	require.NotNil(t, err)
	assert.Contains(t, *err.Message, "empty object")
}
