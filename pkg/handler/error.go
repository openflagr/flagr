package handler

import (
	"fmt"

	"github.com/Allen-Career-Institute/flagr/pkg/util"
	"github.com/Allen-Career-Institute/flagr/swagger_gen/models"
)

// Error is the handler error
type Error struct {
	StatusCode int
	Message    string
	Values     []interface{}
}

func (e *Error) Error() string {
	msg := fmt.Sprintf(e.Message, e.Values...)
	return fmt.Sprintf("status_code: %d. %s", e.StatusCode, msg)
}

// NewError creates Error
func NewError(statusCode int, msg string, values ...interface{}) *Error {
	return &Error{
		StatusCode: statusCode,
		Message:    msg,
		Values:     values,
	}
}

// ErrorMessage generates error messages
func ErrorMessage(s string, data ...interface{}) *models.Error {
	return &models.Error{
		Message: util.StringPtr(fmt.Sprintf(s, data...)),
	}
}
