package handler

import (
	"errors"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"gorm.io/gorm"
)

func (c *crud) DuplicateFlag(params flag.DuplicateFlagParams) middleware.Responder {
	source := &entity.Flag{}
	if err := entity.PreloadSegmentsVariantsTags(getDB()).First(source, params.FlagID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return flag.NewDuplicateFlagDefault(404).WithPayload(
				ErrorMessage("unable to find flag %v in the database", params.FlagID))
		}
		return flag.NewDuplicateFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	keyInput := ""
	descInput := ""
	if params.Body != nil {
		keyInput = params.Body.Key
		descInput = params.Body.Description
	}

	key, err := entity.CreateFlagKey(keyInput)
	if err != nil {
		return flag.NewDuplicateFlagDefault(400).WithPayload(
			ErrorMessage("cannot duplicate flag. %s", err))
	}

	description := descInput
	if strings.TrimSpace(description) == "" {
		base := util.SafeString(source.Description)
		if strings.TrimSpace(base) == "" {
			description = "(cloned)"
		} else {
			description = base + " (cloned)"
		}
	}

	subject := getSubjectFromRequest(params.HTTPRequest)
	created := &entity.Flag{
		Description:        description,
		Key:                key,
		Enabled:            source.Enabled,
		Notes:              source.Notes,
		DataRecordsEnabled: source.DataRecordsEnabled,
		EntityType:         source.EntityType,
		CreatedBy:          subject,
	}

	err = commitFlagMutation(0, subject, notification.OperationCreate, notification.ComponentFlag, func(tx *gorm.DB) (uint, MutationNotify, error) {
		if err := tx.Create(created).Error; err != nil {
			return 0, MutationNotify{}, err
		}
		if created.EntityType != "" {
			if err := entity.CreateFlagEntityType(tx, created.EntityType); err != nil {
				return 0, MutationNotify{}, err
			}
		}
		if err := entity.ApplyFlagTemplate(tx, created.ID, entity.SourceFlagTemplate(source)); err != nil {
			return 0, MutationNotify{}, err
		}
		return created.ID, MutationNotify{ComponentID: created.ID, ComponentKey: key}, nil
	})
	if err != nil {
		return flag.NewDuplicateFlagDefault(500).WithPayload(ErrorMessage("cannot duplicate flag. %s", err))
	}

	if err := entity.PreloadSegmentsVariantsTags(getDB()).First(created, created.ID).Error; err != nil {
		return flag.NewDuplicateFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := flag.NewDuplicateFlagOK()
	payload, err := e2rMapFlag(created)
	if err != nil {
		return flag.NewDuplicateFlagDefault(500).WithPayload(ErrorMessage("cannot map flag. %s", err))
	}
	resp.SetPayload(payload)
	return resp
}