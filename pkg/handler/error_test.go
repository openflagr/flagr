package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	err := NewError(500, "%s some error", "err message")
	s := err.Error()
	assert.NotNil(t, err)
	assert.NotEmpty(t, s)
}

func TestErrorMessage(t *testing.T) {
	msg := ErrorMessage("%s some error", "err message")
	assert.NotNil(t, msg)
}
