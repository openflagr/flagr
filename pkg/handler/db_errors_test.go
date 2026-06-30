package handler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlagKeyUniqueViolation(t *testing.T) {
	assert.False(t, flagKeyUniqueViolation(nil))
	assert.True(t, flagKeyUniqueViolation(errors.New("UNIQUE constraint failed: idx_flag_key")))
	assert.False(t, flagKeyUniqueViolation(errors.New("some other db error")))
}