package handler

import (
	"fmt"

	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
)

// ErrorMessage generates error messages
func ErrorMessage(s string, data ...interface{}) *models.Error {
	return &models.Error{
		Message: util.StringPtr(fmt.Sprintf(s, data...)),
	}
}
