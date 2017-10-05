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

// MapFlags maps flags
func MapFlags(e []entity.Flag) []*models.Flag {
	ret := make([]*models.Flag, len(e), len(e))
	for i, f := range e {
		ret[i] = MapFlag(&f)
	}
	return ret
}
