package handler

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"gorm.io/gorm"
)

func (c *crud) CreateFlag(params flag.CreateFlagParams) middleware.Responder {
	f := &entity.Flag{}
	if params.Body != nil {
		f.Description = util.SafeString(params.Body.Description)
		f.CreatedBy = getSubjectFromRequest(params.HTTPRequest)

		key, err := entity.CreateFlagKey(params.Body.Key)
		if err != nil {
			return flag.NewCreateFlagDefault(400).WithPayload(
				ErrorMessage("cannot create flag. %s", err))
		}
		f.Key = key
	}

	subject := getSubjectFromRequest(params.HTTPRequest)

	if params.Body != nil && params.Body.Template != "" && params.Body.Template != "simple_boolean_flag" {
		return flag.NewCreateFlagDefault(400).WithPayload(
			ErrorMessage("unknown value for template: %s", params.Body.Template))
	}

	err := commitFlagMutation(0, subject, notification.OperationCreate, notification.ComponentFlag, func(tx *gorm.DB) (uint, MutationNotify, error) {
		if err := tx.Create(f).Error; err != nil {
			return 0, MutationNotify{}, err
		}
		if params.Body != nil && params.Body.Template == "simple_boolean_flag" {
			if err := LoadSimpleBooleanFlagTemplate(f, tx); err != nil {
				return 0, MutationNotify{}, err
			}
		}
		return f.ID, MutationNotify{ComponentID: f.ID, ComponentKey: f.Key}, nil
	})
	if err != nil {
		if flagKeyUniqueViolation(err) {
			return flag.NewCreateFlagDefault(400).WithPayload(
				ErrorMessage("cannot create flag. flag key already exists"))
		}
		return flag.NewCreateFlagDefault(500).WithPayload(
			ErrorMessage("cannot create flag. %s", err))
	}

	if err := entity.PreloadSegmentsVariantsTags(getDB()).First(f, f.ID).Error; err != nil {
		return flag.NewCreateFlagDefault(500).WithPayload(
			ErrorMessage("cannot create flag. %s", err))
	}

	resp := flag.NewCreateFlagOK()
	payload, err := e2rMapFlag(f)
	if err != nil {
		return flag.NewCreateFlagDefault(500).WithPayload(
			ErrorMessage("cannot create flag. %s", err))
	}
	resp.SetPayload(payload)
	return resp
}

// LoadSimpleBooleanFlagTemplate loads the simple boolean flag template into a new flag.
func LoadSimpleBooleanFlagTemplate(flag *entity.Flag, tx *gorm.DB) error {
	return entity.ApplyFlagTemplate(tx, flag.ID, entity.SimpleBooleanFlagTemplate())
}
