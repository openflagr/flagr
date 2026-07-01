package handler

import (
	"fmt"
	"strings"

	"github.com/openflagr/flagr/swagger_gen/models"
)

// Error is the handler error
type Error struct {
	StatusCode int
	Message    string
	Values     []any
}

func (e *Error) Error() string {
	msg := fmt.Sprintf(e.Message, e.Values...)
	return fmt.Sprintf("status_code: %d. %s", e.StatusCode, msg)
}

// NewError creates Error
func NewError(statusCode int, msg string, values ...any) *Error {
	return &Error{
		StatusCode: statusCode,
		Message:    msg,
		Values:     values,
	}
}

// ErrorMessage generates error messages
func ErrorMessage(s string, data ...any) *models.Error {
	return &models.Error{
		Message: new(fmt.Sprintf(s, data...)),
	}
}

// flagKeyUniqueViolation reports whether err is a DB unique violation on flags.key (idx_flag_key).
func flagKeyUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "unique") && !strings.Contains(msg, "duplicate") {
		return false
	}
	return strings.Contains(msg, "idx_flag_key") ||
		strings.Contains(msg, "flags.key") ||
		(strings.Contains(msg, "constraint failed") && strings.Contains(msg, "key"))
}
