package e2r

import (
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
)

// MapFlag maps flag
func MapFlag(e *entity.Flag) *models.Flag {
	r := &models.Flag{}
	r.ID = int64(e.ID)
	r.Description = util.StringPtr(e.Description)
	return r
}
