package handler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	t.Parallel()
	err := NewError(500, "%s some error", "err message")
	s := err.Error()
	assert.NotNil(t, err)
	assert.NotEmpty(t, s)
}

func TestErrorMessage(t *testing.T) {
	t.Parallel()
	msg := ErrorMessage("%s some error", "err message")
	assert.NotNil(t, msg)
}

func TestFlagKeyUniqueViolation(t *testing.T) {
	t.Parallel()
	assert.False(t, flagKeyUniqueViolation(nil))
	assert.True(t, flagKeyUniqueViolation(errors.New("UNIQUE constraint failed: idx_flag_key")))
	assert.False(t, flagKeyUniqueViolation(errors.New("some other db error")))
	assert.True(t, flagKeyUniqueViolation(errors.New("duplicate key value violates unique constraint on flags.key")))
	assert.False(t, flagKeyUniqueViolation(errors.New("unique constraint failed: other_table")))
}

func TestIsDuplicateClientError(t *testing.T) {
	t.Parallel()
	assert.False(t, isDuplicateClientError(nil))
	assert.True(t, isDuplicateClientError(NewError(400, "bad")))
	assert.False(t, isDuplicateClientError(NewError(500, "bad")))
	assert.True(t, isDuplicateClientError(errors.New("distribution references unknown variant key \"x\"")))
	assert.True(t, isDuplicateClientError(errors.New("cannot create flag due to invalid key. reason: x")))
	assert.False(t, isDuplicateClientError(errors.New("connection reset")))
}
