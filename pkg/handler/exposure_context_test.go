package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeExposureEntityContext(t *testing.T) {
	m := mergeExposureEntityContext(
		map[string]any{"a": 1},
		map[string]interface{}{"b": 2},
	)
	assert.Equal(t, float64(1), m["a"]) // JSON numbers decode as float64
	assert.Equal(t, float64(2), m["b"])
}