package handler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/mapper/entity_restapi/e2r"
	"github.com/openflagr/flagr/pkg/mapper/entity_restapi/r2e"
	"github.com/openflagr/flagr/pkg/notification"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/constraint"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/tag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/variant"
	"github.com/openflagr/flagr/swagger_gen/models"

	"github.com/go-openapi/runtime/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CRUD is the CRUD interface
type CRUD interface {
	// Flags
	FindFlags(flag.FindFlagsParams) middleware.Responder
	CreateFlag(flag.CreateFlagParams) middleware.Responder
	DuplicateFlag(flag.DuplicateFlagParams) middleware.Responder
	GetFlag(flag.GetFlagParams) middleware.Responder
	PutFlag(flag.PutFlagParams) middleware.Responder
	DeleteFlag(flag.DeleteFlagParams) middleware.Responder
	RestoreFlag(flag.RestoreFlagParams) middleware.Responder
	SetFlagEnabledState(flag.SetFlagEnabledParams) middleware.Responder
	GetFlagSnapshots(params flag.GetFlagSnapshotsParams) middleware.Responder
	GetFlagEntityTypes(params flag.GetFlagEntityTypesParams) middleware.Responder
	GetFlagSnapshotMaxID(params flag.GetFlagSnapshotMaxIDParams) middleware.Responder

	//Tags
	CreateTag(tag.CreateTagParams) middleware.Responder
	DeleteTag(tag.DeleteTagParams) middleware.Responder
	FindTags(tag.FindTagsParams) middleware.Responder
	FindAllTags(params tag.FindAllTagsParams) middleware.Responder

	// Segments
	CreateSegment(segment.CreateSegmentParams) middleware.Responder
	FindSegments(segment.FindSegmentsParams) middleware.Responder
	PutSegment(segment.PutSegmentParams) middleware.Responder
	DeleteSegment(segment.DeleteSegmentParams) middleware.Responder
	PutSegmentsReorder(segment.PutSegmentsReorderParams) middleware.Responder

	// Constraints
	CreateConstraint(constraint.CreateConstraintParams) middleware.Responder
	FindConstraints(constraint.FindConstraintsParams) middleware.Responder
	PutConstraint(params constraint.PutConstraintParams) middleware.Responder
	DeleteConstraint(params constraint.DeleteConstraintParams) middleware.Responder

	// Distributions
	FindDistributions(distribution.FindDistributionsParams) middleware.Responder
	PutDistributions(distribution.PutDistributionsParams) middleware.Responder

	// Variants
	CreateVariant(variant.CreateVariantParams) middleware.Responder
	FindVariants(variant.FindVariantsParams) middleware.Responder
	PutVariant(variant.PutVariantParams) middleware.Responder
	DeleteVariant(variant.DeleteVariantParams) middleware.Responder
}

// NewCRUD creates a new CRUD instance
func NewCRUD() CRUD {
	return &crud{}
}

type crud struct{}

var (
	e2rMapFlag          = e2r.MapFlag
	e2rMapFlags         = e2r.MapFlags
	e2rMapFlagSnapshots = e2r.MapFlagSnapshots

	r2eMapAttachment    = r2e.MapAttachment
	r2eMapDistributions = r2e.MapDistributions
)

func (c *crud) FindFlags(params flag.FindFlagsParams) middleware.Responder {
	// Add Unscoped so GORM doesn't automatically override `deleted_at`
	tx := getDB().Unscoped()
	fs := []entity.Flag{}
	q := entity.Flag{}

	if params.Enabled != nil {
		tx = tx.Where("enabled = ?", *params.Enabled)
	}
	if params.Description != nil {
		q.Description = *params.Description
	}
	if params.Key != nil {
		q.Key = *params.Key
	}
	if params.Offset != nil {
		tx = tx.Offset(int(*params.Offset))
	}
	if params.Limit != nil {
		tx = tx.Limit(int(*params.Limit))
	}
	if params.Preload != nil && *params.Preload {
		tx = entity.PreloadSegmentsVariantsTags(tx)
	} else {
		// Always preload tags for searchability
		tx = entity.PreloadFlagTags(tx)
	}
	if params.DescriptionLike != nil {
		tx = tx.Where(
			"lower(description) like ?",
			fmt.Sprintf("%%%s%%", strings.ToLower(*params.DescriptionLike)),
		)
	}
	if params.Deleted != nil && *params.Deleted {
		tx = tx.Where("deleted_at is not null")
	} else {
		tx = tx.Where("deleted_at is null")
	}

	var err error
	tx = tx.Order("id").Where(q)
	if params.Tags != nil {
		t := []entity.Tag{}
		getDB().Where("value in (?)", strings.Split(*params.Tags, ",")).Find(&t)
		err = tx.Model(&t).Group("flags.id").Association("Flags").Find(&fs)
	} else {
		err = tx.Find(&fs).Error
	}

	if err != nil {
		return flag.NewFindFlagsDefault(500).WithPayload(
			ErrorMessage("cannot query all flags. %s", err))
	}

	resp := flag.NewFindFlagsOK()
	payload, err := e2rMapFlags(fs)
	if err != nil {
		return flag.NewFindFlagsDefault(500).WithPayload(
			ErrorMessage("cannot map flags. %s", err))
	}
	resp.SetPayload(payload)
	return resp
}

func (c *crud) GetFlag(params flag.GetFlagParams) middleware.Responder {
	f := &entity.Flag{}
	result := entity.PreloadSegmentsVariantsTags(getDB()).First(f, params.FlagID).Error

	// Flag with given ID doesn't exist, so we 404
	if errors.Is(result, gorm.ErrRecordNotFound) {
		return flag.NewGetFlagDefault(404).WithPayload(
			ErrorMessage("unable to find flag %v in the database", params.FlagID))
	}

	// Something else happened, return a 500
	if result != nil {
		return flag.NewGetFlagDefault(500).WithPayload(
			ErrorMessage("an unknown error occurred while looking up flag %v: %s", params.FlagID, result))
	}

	resp := flag.NewGetFlagOK()
	payload, err := e2rMapFlag(f)
	if err != nil {
		return flag.NewGetFlagDefault(500).WithPayload(
			ErrorMessage("cannot map flag %v. %s", params.FlagID, err))
	}
	resp.SetPayload(payload)
	return resp
}

func (c *crud) GetFlagSnapshots(params flag.GetFlagSnapshotsParams) middleware.Responder {
	tx := getDB()
	fs := []entity.FlagSnapshot{}

	if params.Limit != nil {
		tx = tx.Limit(int(*params.Limit))
	}
	if params.Offset != nil {
		tx = tx.Offset(int(*params.Offset))
	}

	descending := params.Sort == nil || *params.Sort != "ASC"

	if err := tx.
		Order(clause.OrderByColumn{
			Column: clause.Column{
				Name: "created_at",
			},
			Desc: descending,
		}).
		Where(entity.FlagSnapshot{FlagID: util.SafeUint(params.FlagID)}).
		Find(&fs).Error; err != nil {
		return flag.NewGetFlagSnapshotsDefault(500).WithPayload(
			ErrorMessage("cannot find flag snapshots for %v. %s", params.FlagID, err))
	}
	payload, err := e2rMapFlagSnapshots(fs)
	if err != nil {
		return flag.NewGetFlagSnapshotsDefault(500).WithPayload(
			ErrorMessage("cannot map flag snapshots for flagID %v. %s", params.FlagID, err))
	}
	resp := flag.NewGetFlagSnapshotsOK()
	resp.SetPayload(payload)
	return resp
}

func (c *crud) GetFlagEntityTypes(params flag.GetFlagEntityTypesParams) middleware.Responder {
	entityTypes := []entity.FlagEntityType{}
	if err := getDB().Order("flag_entity_types.key").Find(&entityTypes).Error; err != nil {
		return flag.NewGetFlagEntityTypesDefault(500).WithPayload(
			ErrorMessage("cannot find flag entity types. err:%s", err))

	}

	payload := []string{}
	for _, t := range entityTypes {
		payload = append(payload, t.Key)
	}
	resp := flag.NewGetFlagEntityTypesOK()
	resp.SetPayload(payload)
	return resp
}

func (c *crud) GetFlagSnapshotMaxID(params flag.GetFlagSnapshotMaxIDParams) middleware.Responder {
	var maxID uint
	if err := getDB().Model(&entity.FlagSnapshot{}).
		Select("COALESCE(MAX(id), 0)").
		Scan(&maxID).Error; err != nil {
		maxID = 0
	}
	resp := flag.NewGetFlagSnapshotMaxIDOK()
	resp.SetPayload(&models.FlagSnapshotMaxID{MaxID: int64(maxID)})
	return resp
}

func (c *crud) PutFlag(params flag.PutFlagParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	f := &entity.Flag{}

	err := commitFlagMutation(flagID, subject, notification.OperationUpdate, notification.ComponentFlag, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.First(f, params.FlagID).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		if params.Body.Description != nil {
			f.Description = *params.Body.Description
		}
		if params.Body.DataRecordsEnabled != nil {
			f.DataRecordsEnabled = *params.Body.DataRecordsEnabled
		}
		if params.Body.Key != nil {
			key, err := entity.CreateFlagKey(*params.Body.Key)
			if err != nil {
				return 0, mutationNotify{}, NewError(400, "%s", err)
			}
			f.Key = key
		}
		if params.Body.EntityType != nil {
			et := *params.Body.EntityType
			if err := entity.CreateFlagEntityType(tx, et); err != nil {
				return 0, mutationNotify{}, err
			}
			f.EntityType = et
		}
		if params.Body.Notes != nil {
			f.Notes = *params.Body.Notes
		}
		if err := tx.Save(f).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		if err := entity.PreloadSegmentsVariantsTags(tx).First(f, params.FlagID).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: flagID, ComponentKey: f.Key}, nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return flag.NewPutFlagDefault(404).WithPayload(ErrorMessage("%s", err))
		}
		if herr, ok := err.(*Error); ok {
			return flag.NewPutFlagDefault(herr.StatusCode).WithPayload(ErrorMessage("%s", err))
		}
		if flagKeyUniqueViolation(err) {
			return flag.NewPutFlagDefault(400).WithPayload(
				ErrorMessage("cannot update flag. flag key already exists"))
		}
		return flag.NewPutFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := flag.NewPutFlagOK()
	payload, err := e2rMapFlag(f)
	if err != nil {
		return flag.NewPutFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	resp.SetPayload(payload)
	return resp
}

func (c *crud) SetFlagEnabledState(params flag.SetFlagEnabledParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	f := &entity.Flag{}

	err := commitFlagMutation(flagID, subject, notification.OperationUpdate, notification.ComponentFlag, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.First(f, params.FlagID).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		f.Enabled = *params.Body.Enabled
		if err := tx.Save(f).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: flagID, ComponentKey: f.Key}, nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return flag.NewSetFlagEnabledDefault(404).WithPayload(ErrorMessage("%s", err))
		}
		return flag.NewSetFlagEnabledDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := flag.NewSetFlagEnabledOK()
	payload, err := e2rMapFlag(f)
	if err != nil {
		return flag.NewSetFlagEnabledDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	resp.SetPayload(payload)
	return resp
}

func (c *crud) RestoreFlag(params flag.RestoreFlagParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	f := &entity.Flag{}

	err := commitFlagMutation(flagID, subject, notification.OperationRestore, notification.ComponentFlag, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := entity.PreloadFlagTags(tx.Unscoped()).First(f, params.FlagID).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		f.DeletedAt = gorm.DeletedAt{}
		if err := tx.Unscoped().Save(f).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: flagID, ComponentKey: f.Key}, nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return flag.NewRestoreFlagDefault(404).WithPayload(ErrorMessage("%s", err))
		}
		return flag.NewRestoreFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := flag.NewRestoreFlagOK()
	payload, err := e2rMapFlag(f)
	if err != nil {
		return flag.NewRestoreFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	resp.SetPayload(payload)
	return resp
}

func (c *crud) DeleteFlag(params flag.DeleteFlagParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	f := &entity.Flag{}

	err := commitFlagMutation(flagID, subject, notification.OperationDelete, notification.ComponentFlag, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.First(f, params.FlagID).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		if err := tx.Delete(&entity.Flag{}, params.FlagID).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: flagID, ComponentKey: f.Key}, nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return flag.NewDeleteFlagDefault(404).WithPayload(ErrorMessage("%s", err))
		}
		return flag.NewDeleteFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	return flag.NewDeleteFlagOK()
}

func (c *crud) DeleteTag(params tag.DeleteTagParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	tagID := uint(params.TagID)

	err := commitFlagMutation(flagID, subject, notification.OperationDelete, notification.ComponentTag, func(tx *gorm.DB) (uint, mutationNotify, error) {
		t := &entity.Tag{}
		t.ID = tagID
		s := &entity.Flag{}
		s.ID = flagID
		if err := tx.Model(s).Association("Tags").Delete(t); err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: tagID, ComponentKey: ""}, nil
	})
	if err != nil {
		return tag.NewDeleteTagDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	return tag.NewDeleteTagOK()
}

func (c *crud) FindTags(params tag.FindTagsParams) middleware.Responder {
	ds := []entity.Tag{}

	s := &entity.Flag{}
	s.ID = uint(params.FlagID)

	if err := getDB().Model(s).Association("Tags").Find(&ds); err != nil {
		return tag.NewFindTagsDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := tag.NewFindTagsOK()
	resp.SetPayload(e2r.MapTags(ds))
	return resp
}

func (c *crud) FindAllTags(params tag.FindAllTagsParams) middleware.Responder {
	tx := getDB()
	ds := []entity.Tag{}

	if params.Limit != nil {
		tx = tx.Limit(int(*params.Limit))
	}
	if params.Offset != nil {
		tx = tx.Offset(int(*params.Offset))
	}
	if params.ValueLike != nil {
		tx = tx.Where(
			"lower(value) like ?",
			fmt.Sprintf("%%%s%%", strings.ToLower(*params.ValueLike)),
		)
	}

	if err := tx.Find(&ds).Error; err != nil {
		return tag.NewFindAllTagsDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := tag.NewFindAllTagsOK()
	resp.SetPayload(e2r.MapTags(ds))
	return resp
}

func (c *crud) CreateTag(params tag.CreateTagParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	t := &entity.Tag{}
	t.Value = util.SafeString(params.Body.Value)
	if ok, reason := util.IsSafeValue(t.Value); !ok {
		return tag.NewCreateTagDefault(400).WithPayload(ErrorMessage("%s", reason))
	}

	err := commitFlagMutation(flagID, subject, notification.OperationCreate, notification.ComponentTag, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := entity.AppendTagValueToFlag(tx, flagID, t.Value); err != nil {
			return 0, mutationNotify{}, err
		}
		if err := tx.Where("value = ?", t.Value).First(t).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: t.ID, ComponentKey: t.Value}, nil
	})
	if err != nil {
		return tag.NewCreateTagDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := tag.NewCreateTagOK()
	resp.SetPayload(e2r.MapTag(t))
	return resp
}

func (c *crud) CreateSegment(params segment.CreateSegmentParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	s := &entity.Segment{}
	s.FlagID = flagID
	s.RolloutPercent = uint(*params.Body.RolloutPercent)
	s.Description = util.SafeString(params.Body.Description)
	s.Rank = entity.SegmentDefaultRank

	err := commitFlagMutation(flagID, subject, notification.OperationCreate, notification.ComponentSegment, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.Create(s).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: s.ID, ComponentKey: ""}, nil
	})
	if err != nil {
		return segment.NewCreateSegmentDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := segment.NewCreateSegmentOK()
	resp.SetPayload(e2r.MapSegment(s))
	return resp
}

func (c *crud) FindSegments(params segment.FindSegmentsParams) middleware.Responder {
	ss := []entity.Segment{}
	err := entity.
		PreloadConstraintsDistribution(getDB()).
		Order("segments.rank").
		Order("segments.id").
		Where(entity.Segment{FlagID: uint(params.FlagID)}).
		Find(&ss).
		Error
	if err != nil {
		return segment.NewFindSegmentsDefault(500).
			WithPayload(ErrorMessage("%s", err))
	}

	resp := segment.NewFindSegmentsOK()
	resp.SetPayload(e2r.MapSegments(ss))
	return resp
}

func (c *crud) PutSegment(params segment.PutSegmentParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	segmentID := util.SafeUint(params.SegmentID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	s := &entity.Segment{}

	err := commitFlagMutation(flagID, subject, notification.OperationUpdate, notification.ComponentSegment, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := entity.PreloadConstraintsDistribution(tx).First(s, params.SegmentID).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		s.RolloutPercent = util.SafeUint(params.Body.RolloutPercent)
		s.Description = util.SafeString(params.Body.Description)
		if err := tx.Save(s).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: segmentID, ComponentKey: ""}, nil
	})
	if err != nil {
		return segment.NewPutSegmentDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := segment.NewPutSegmentOK()
	resp.SetPayload(e2r.MapSegment(s))
	return resp
}

func (c *crud) PutSegmentsReorder(params segment.PutSegmentsReorderParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)

	err := commitFlagMutation(flagID, subject, notification.OperationUpdate, notification.ComponentSegment, func(tx *gorm.DB) (uint, mutationNotify, error) {
		for i, segmentID := range params.Body.SegmentIDs {
			s := &entity.Segment{}
			if err := tx.First(s, segmentID).Error; err != nil {
				return 0, mutationNotify{}, err
			}
			s.Rank = uint(i)
			if err := tx.Save(s).Error; err != nil {
				return 0, mutationNotify{}, err
			}
		}
		return flagID, mutationNotify{ComponentID: 0, ComponentKey: ""}, nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return segment.NewPutSegmentsReorderDefault(404).WithPayload(ErrorMessage("%s", err))
		}
		return segment.NewPutSegmentsReorderDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	return segment.NewPutSegmentsReorderOK()
}

func (c *crud) DeleteSegment(params segment.DeleteSegmentParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	segmentID := util.SafeUint(params.SegmentID)
	subject := getSubjectFromRequest(params.HTTPRequest)

	err := commitFlagMutation(flagID, subject, notification.OperationDelete, notification.ComponentSegment, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.Delete(&entity.Segment{}, segmentID).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: segmentID, ComponentKey: ""}, nil
	})
	if err != nil {
		return segment.NewDeleteSegmentDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	return segment.NewDeleteSegmentOK()
}

func (c *crud) CreateConstraint(params constraint.CreateConstraintParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	cons := &entity.Constraint{}
	cons.SegmentID = uint(params.SegmentID)
	if params.Body != nil {
		cons.Property = util.SafeString(params.Body.Property)
		cons.Operator = util.SafeString(params.Body.Operator)
		cons.Value = util.SafeString(params.Body.Value)
	}
	if err := cons.Validate(); err != nil {
		return constraint.NewCreateConstraintDefault(400).WithPayload(ErrorMessage("%s", err))
	}

	err := commitFlagMutation(flagID, subject, notification.OperationCreate, notification.ComponentConstraint, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.Create(cons).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: cons.ID, ComponentKey: ""}, nil
	})
	if err != nil {
		return constraint.NewCreateConstraintDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := constraint.NewCreateConstraintOK()
	resp.SetPayload(e2r.MapConstraint(cons))
	return resp
}

func (c *crud) FindConstraints(params constraint.FindConstraintsParams) middleware.Responder {
	cs := []entity.Constraint{}
	if err := getDB().Order("created_at").Where(entity.Constraint{SegmentID: uint(params.SegmentID)}).Find(&cs).Error; err != nil {
		return constraint.NewFindConstraintsDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := constraint.NewFindConstraintsOK()
	resp.SetPayload(e2r.MapConstraints(cs))
	return resp
}

func (c *crud) PutConstraint(params constraint.PutConstraintParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	constraintID := util.SafeUint(params.ConstraintID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	cons := &entity.Constraint{}

	if err := getDB().First(cons, params.ConstraintID).Error; err != nil {
		return constraint.NewPutConstraintDefault(404).WithPayload(ErrorMessage("%s", err))
	}

	if params.Body != nil {
		cons.Property = util.SafeString(params.Body.Property)
		cons.Operator = util.SafeString(params.Body.Operator)
		cons.Value = util.SafeString(params.Body.Value)
	}
	if err := cons.Validate(); err != nil {
		return constraint.NewPutConstraintDefault(400).WithPayload(ErrorMessage("%s", err))
	}

	err := commitFlagMutation(flagID, subject, notification.OperationUpdate, notification.ComponentConstraint, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.Save(cons).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: constraintID, ComponentKey: ""}, nil
	})
	if err != nil {
		return constraint.NewPutConstraintDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := constraint.NewPutConstraintOK()
	resp.SetPayload(e2r.MapConstraint(cons))
	return resp
}

func (c *crud) DeleteConstraint(params constraint.DeleteConstraintParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	constraintID := util.SafeUint(params.ConstraintID)
	subject := getSubjectFromRequest(params.HTTPRequest)

	err := commitFlagMutation(flagID, subject, notification.OperationDelete, notification.ComponentConstraint, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.Delete(&entity.Constraint{}, params.ConstraintID).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: constraintID, ComponentKey: ""}, nil
	})
	if err != nil {
		return constraint.NewDeleteConstraintDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	return constraint.NewDeleteConstraintOK()
}

// PutDistributions puts the whole distributions and overwrite the old ones
func (c *crud) PutDistributions(params distribution.PutDistributionsParams) middleware.Responder {
	if err := validatePutDistributions(params); err != nil {
		return distribution.NewPutDistributionsDefault(err.StatusCode).WithPayload(ErrorMessage("%s", err))
	}

	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	segmentID := uint(params.SegmentID)
	ds := r2eMapDistributions(params.Body.Distributions, segmentID)

	err := commitFlagMutation(flagID, subject, notification.OperationUpdate, notification.ComponentDistribution, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.Where("segment_id = ?", segmentID).Delete(&entity.Distribution{}).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		for i := range ds {
			if err := tx.Create(&ds[i]).Error; err != nil {
				return 0, mutationNotify{}, err
			}
		}
		return flagID, mutationNotify{ComponentID: 0, ComponentKey: ""}, nil
	})
	if err != nil {
		return distribution.NewPutDistributionsDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := distribution.NewPutDistributionsOK()
	resp.SetPayload(e2r.MapDistributions(ds))
	return resp
}

func (c *crud) FindDistributions(params distribution.FindDistributionsParams) middleware.Responder {
	ds := []entity.Distribution{}
	err := getDB().
		Order("variant_id").
		Where(entity.Distribution{SegmentID: uint(params.SegmentID)}).
		Find(&ds).
		Error
	if err != nil {
		return distribution.NewFindDistributionsDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := distribution.NewFindDistributionsOK()
	resp.SetPayload(e2r.MapDistributions(ds))
	return resp
}

func (c *crud) CreateVariant(params variant.CreateVariantParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	v := &entity.Variant{}
	v.FlagID = flagID
	v.Key = util.SafeString(params.Body.Key)

	a, err := r2eMapAttachment(params.Body.Attachment)
	if err != nil {
		return variant.NewCreateVariantDefault(400).WithPayload(ErrorMessage("%s", err))
	}
	v.Attachment = a

	if err := v.Validate(); err != nil {
		return variant.NewCreateVariantDefault(400).WithPayload(ErrorMessage("%s", err))
	}

	err = commitFlagMutation(flagID, subject, notification.OperationCreate, notification.ComponentVariant, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.Create(v).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: v.ID, ComponentKey: v.Key}, nil
	})
	if err != nil {
		return variant.NewCreateVariantDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := variant.NewCreateVariantOK()
	resp.SetPayload(e2r.MapVariant(v))
	return resp
}

func (c *crud) FindVariants(params variant.FindVariantsParams) middleware.Responder {
	vs := []entity.Variant{}
	err := getDB().
		Order("id").
		Where(entity.Variant{FlagID: uint(params.FlagID)}).
		Find(&vs).
		Error
	if err != nil {
		return variant.NewFindVariantsDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := variant.NewFindVariantsOK()
	resp.SetPayload(e2r.MapVariants(vs))
	return resp
}

func (c *crud) PutVariant(params variant.PutVariantParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	variantID := util.SafeUint(params.VariantID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	v := &entity.Variant{}

	if err := getDB().First(v, params.VariantID).Error; err != nil {
		return variant.NewPutVariantDefault(404).WithPayload(ErrorMessage("%s", err))
	}

	v.Key = util.SafeString(params.Body.Key)
	if params.Body.Attachment != nil {
		a, err := r2eMapAttachment(params.Body.Attachment)
		if err != nil {
			return variant.NewPutVariantDefault(400).WithPayload(ErrorMessage("%s", err))
		}
		v.Attachment = a
	}

	if err := v.Validate(); err != nil {
		return variant.NewPutVariantDefault(400).WithPayload(ErrorMessage("%s", err))
	}

	err := commitFlagMutation(flagID, subject, notification.OperationUpdate, notification.ComponentVariant, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.Save(v).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		if err := validatePutVariantForDistributions(v, tx); err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: variantID, ComponentKey: v.Key}, nil
	})
	if err != nil {
		if herr, ok := err.(*Error); ok {
			return variant.NewPutVariantDefault(herr.StatusCode).WithPayload(ErrorMessage("%s", err))
		}
		return variant.NewPutVariantDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := variant.NewPutVariantOK()
	resp.SetPayload(e2r.MapVariant(v))
	return resp
}

func (c *crud) DeleteVariant(params variant.DeleteVariantParams) middleware.Responder {
	if err := validateDeleteVariant(params); err != nil {
		return variant.NewDeleteVariantDefault(err.StatusCode).WithPayload(ErrorMessage("%s", err))
	}

	flagID := util.SafeUint(params.FlagID)
	variantID := util.SafeUint(params.VariantID)
	subject := getSubjectFromRequest(params.HTTPRequest)

	err := commitFlagMutation(flagID, subject, notification.OperationDelete, notification.ComponentVariant, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if err := tx.Delete(&entity.Variant{}, params.VariantID).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: variantID, ComponentKey: ""}, nil
	})
	if err != nil {
		return variant.NewDeleteVariantDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	return variant.NewDeleteVariantOK()
}

// mutationNotify carries webhook component metadata after a successful mutation.
type mutationNotify struct {
	ComponentID  uint
	ComponentKey string
}

// commitFlagMutation runs mutate in one transaction, writes a flag snapshot on the same tx, commits, then notifies.
// snapshotFlagID is the flag whose history row is updated (use 0 when the new flag ID is assigned inside mutate).
func commitFlagMutation(
	snapshotFlagID uint,
	subject string,
	operation notification.Operation,
	componentType notification.ComponentType,
	mutate func(tx *gorm.DB) (uint, mutationNotify, error),
) error {
	tx := getDB().Begin()
	resolvedID, notify, err := mutate(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	flagIDForSnapshot := snapshotFlagID
	if flagIDForSnapshot == 0 {
		flagIDForSnapshot = resolvedID
	}
	snap, err := writeFlagSnapshotTx(tx, flagIDForSnapshot, subject)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	snap.NotifyAfterCommit(flagIDForSnapshot, subject, operation, componentType, notify.ComponentID, notify.ComponentKey)
	return nil
}

// writeFlagSnapshotTx is the indirection used by commitFlagMutation (stubbable in tests).
var writeFlagSnapshotTx = entity.WriteFlagSnapshotTx
