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
	GetFlag(flag.GetFlagParams) middleware.Responder
	PutFlag(flag.PutFlagParams) middleware.Responder
	DeleteFlag(flag.DeleteFlagParams) middleware.Responder
}

// NewCRUD creates a new CRUD instance
func NewCRUD() CRUD {
	return &crud{}
}

type crud struct{}

func (c *crud) FindFlags(params flag.FindFlagsParams) middleware.Responder {
	fs := []entity.Flag{}
	q := entity.NewFlagQuerySet(repo.GetDB())
	err := q.All(&fs)
	if err != nil {
		return flag.NewFindFlagsDefault(500)
	}
	resp := flag.NewFindFlagsOK()
	resp.SetPayload(e2r.MapFlags(fs))
	return resp
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
	resp.SetPayload(e2r.MapFlag(f))
	return resp
}

func (c *crud) GetFlag(params flag.GetFlagParams) middleware.Responder {
	f := &entity.Flag{}
	q := entity.NewFlagQuerySet(repo.GetDB())
	err := q.IDEq(uint(params.FlagID)).One(f)
	if err != nil {
		return flag.NewGetFlagDefault(500)
	}
	resp := flag.NewGetFlagOK()
	resp.SetPayload(e2r.MapFlag(f))
	return resp
}

func (c *crud) PutFlag(params flag.PutFlagParams) middleware.Responder {
	q := entity.NewFlagQuerySet(repo.GetDB())

	err := q.IDEq(uint(params.FlagID)).
		GetUpdater().
		SetDescription(util.SafeString(params.Body.Description)).
		Update()
	if err != nil {
		return flag.NewGetFlagDefault(500)
	}

	f := &entity.Flag{}
	err = q.IDEq(uint(params.FlagID)).One(f)
	if err != nil {
		return flag.NewGetFlagDefault(500)
	}

	resp := flag.NewGetFlagOK()
	resp.SetPayload(e2r.MapFlag(f))
	return resp
}

func (c *crud) DeleteFlag(params flag.DeleteFlagParams) middleware.Responder {
	q := entity.NewFlagQuerySet(repo.GetDB())

	err := q.IDEq(uint(params.FlagID)).Delete()
	if err != nil {
		return flag.NewGetFlagDefault(500)
	}
	return flag.NewDeleteFlagOK()
}
