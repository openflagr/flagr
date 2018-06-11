package handler

import (
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/mapper/entity_restapi/e2r"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/query"

	"github.com/go-openapi/runtime/middleware"
)

// QueryAPI is the QueryAPI interface
type QueryAPI interface {
	GetFlagByName(query.GetFlagByNameParams) middleware.Responder
	GetFlagByNameBatch(query.GetFlagByNameBatchParams) middleware.Responder
}

// NewQueryAPI creates a new NewQueryAPI instance
func NewQueryAPI() QueryAPI {
	return &queryAPI{}
}

type queryAPI struct{}

func (qAPI *queryAPI) GetFlagByName(params query.GetFlagByNameParams) middleware.Responder {
	f := &entity.Flag{}
	q := entity.NewFlagQuerySet(getDB())

	err := q.NameEq(params.FlagName).One(f)
	if err != nil {
		return query.NewGetFlagByNameDefault(404).WithPayload(
			ErrorMessage("cannot find flag %v. %s", params.FlagName, err))
	}

	resp := query.NewGetFlagByNameOK()
	payload, err := e2r.MapFlag(f, true)
	if err != nil {
		return query.NewGetFlagByNameDefault(404).WithPayload(
			ErrorMessage("cannot map flag %v. %s", params.FlagName, err))
	}
	resp.SetPayload(payload)
	return resp
}

func (qAPI *queryAPI) GetFlagByNameBatch(params query.GetFlagByNameBatchParams) middleware.Responder {
	flagNames := params.FlagNames

	fs := []entity.Flag{}
	q := entity.NewFlagQuerySet(getDB())
	err := q.NameIn(flagNames...).All(&fs)
	if err != nil {
		return query.NewGetFlagByNameBatchDefault(404).WithPayload(
			ErrorMessage("Error finding flags. %s", err))
	}

	resp := query.NewGetFlagByNameBatchOK()
	payload, err := e2r.MapFlags(fs)
	if err != nil {
		return query.NewGetFlagByNameDefault(404).WithPayload(
			ErrorMessage("Error mapping flags. %s", err))
	}
	resp.SetPayload(payload)
	return resp
}
