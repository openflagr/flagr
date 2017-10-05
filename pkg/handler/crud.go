package handler

import (
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/mapper/entity_restapi/e2r"
	"github.com/checkr/flagr/pkg/repo"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/go-openapi/runtime/middleware"
)

// CRUD is the CRUD interface
type CRUD interface {
	FindFlags(flag.FindFlagsParams) middleware.Responder
	CreateFlag(flag.CreateFlagParams) middleware.Responder
}

// NewCRUD creates a new CRUD instance
func NewCRUD() CRUD {
	return &crud{}
}

type crud struct{}

func (c *crud) FindFlags(params flag.FindFlagsParams) middleware.Responder {
	return nil
}

func (c *crud) CreateFlag(params flag.CreateFlagParams) middleware.Responder {
	f := &entity.Flag{}
	if params.Body != nil {
		f.Description = util.SafeString(params.Body.Description)
	}
	err := f.Create(repo.GetDB())
	if err != nil {
		return flag.NewCreateFlagDefault(500)
	}

	resp := flag.NewCreateFlagOK()
	resp.Payload = e2r.MapFlag(f)
	return resp
}
