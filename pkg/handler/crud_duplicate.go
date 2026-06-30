package handler

import (
	"errors"
	"fmt"
	"sort"
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

	tx := getDB().Begin()
	if err := tx.Create(created).Error; err != nil {
		tx.Rollback()
		return flag.NewDuplicateFlagDefault(500).WithPayload(ErrorMessage("cannot duplicate flag. %s", err))
	}
	if created.EntityType != "" {
		if err := entity.CreateFlagEntityType(tx, created.EntityType); err != nil {
			tx.Rollback()
			return flag.NewDuplicateFlagDefault(500).WithPayload(ErrorMessage("%s", err))
		}
	}
	if err := cloneFlagGraph(tx, source, created); err != nil {
		tx.Rollback()
		return flag.NewDuplicateFlagDefault(500).WithPayload(ErrorMessage("cannot duplicate flag. %s", err))
	}
	meta, err := entity.WriteFlagSnapshotTx(tx, created.ID, subject)
	if err != nil {
		tx.Rollback()
		return flag.NewDuplicateFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	if err := tx.Commit().Error; err != nil {
		return flag.NewDuplicateFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	entity.NotifyFlagSnapshot(created.ID, subject, notification.OperationCreate, notification.ComponentFlag, created.ID, key, meta)

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

func cloneFlagGraph(tx *gorm.DB, source *entity.Flag, dest *entity.Flag) error {
	variantMap := make(map[uint]uint, len(source.Variants))
	for _, sv := range source.Variants {
		nv := &entity.Variant{
			FlagID:     dest.ID,
			Key:        sv.Key,
			Attachment: sv.Attachment,
		}
		if err := nv.Validate(); err != nil {
			return err
		}
		if err := tx.Create(nv).Error; err != nil {
			return err
		}
		variantMap[sv.ID] = nv.ID
	}

	segments := append([]entity.Segment(nil), source.Segments...)
	sort.SliceStable(segments, func(i, j int) bool {
		if segments[i].Rank != segments[j].Rank {
			return segments[i].Rank < segments[j].Rank
		}
		return segments[i].ID < segments[j].ID
	})

	for _, ss := range segments {
		ns := &entity.Segment{
			FlagID:         dest.ID,
			Description:    ss.Description,
			Rank:           ss.Rank,
			RolloutPercent: ss.RolloutPercent,
		}
		if err := tx.Create(ns).Error; err != nil {
			return err
		}
		for _, sc := range ss.Constraints {
			nc := &entity.Constraint{
				SegmentID: ns.ID,
				Property:  sc.Property,
				Operator:  sc.Operator,
				Value:     sc.Value,
			}
			if err := nc.Validate(); err != nil {
				return err
			}
			if err := tx.Create(nc).Error; err != nil {
				return err
			}
		}
		for _, sd := range ss.Distributions {
			newVID, ok := variantMap[sd.VariantID]
			if !ok {
				return fmt.Errorf("distribution references unknown variant id %d", sd.VariantID)
			}
			nd := &entity.Distribution{
				SegmentID:  ns.ID,
				VariantID:  newVID,
				VariantKey: sd.VariantKey,
				Percent:    sd.Percent,
				Bitmap:     sd.Bitmap,
			}
			if err := tx.Create(nd).Error; err != nil {
				return err
			}
		}
	}

	if len(source.Tags) > 0 {
		flagRef := &entity.Flag{}
		flagRef.ID = dest.ID
		for _, st := range source.Tags {
			t := &entity.Tag{Value: st.Value}
			tx.Where("value = ?", st.Value).FirstOrCreate(t)
			if err := tx.Model(flagRef).Association("Tags").Append(t); err != nil {
				return err
			}
		}
	}
	return nil
}