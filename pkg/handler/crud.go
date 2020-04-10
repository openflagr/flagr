package handler

import (
	"fmt"
	"strings"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/mapper/entity_restapi/e2r"
	"github.com/checkr/flagr/pkg/mapper/entity_restapi/r2e"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/constraint"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/variant"

	"github.com/go-openapi/runtime/middleware"
)

// CRUD is the CRUD interface
type CRUD interface {
	// Flags
	FindFlags(flag.FindFlagsParams) middleware.Responder
	CreateFlag(flag.CreateFlagParams) middleware.Responder
	GetFlag(flag.GetFlagParams) middleware.Responder
	PutFlag(flag.PutFlagParams) middleware.Responder
	DeleteFlag(flag.DeleteFlagParams) middleware.Responder
	SetFlagEnabledState(flag.SetFlagEnabledParams) middleware.Responder
	GetFlagSnapshots(params flag.GetFlagSnapshotsParams) middleware.Responder
	GetFlagEntityTypes(params flag.GetFlagEntityTypesParams) middleware.Responder

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
	tx := getDB()
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
		tx = entity.PreloadSegmentsVariants(tx)
	}
	if params.DescriptionLike != nil {
		tx = tx.Where(
			"lower(description) like ?",
			fmt.Sprintf("%%%s%%", strings.ToLower(*params.DescriptionLike)),
		)
	}

	err := tx.Order("id").Where(q).Find(&fs).Error
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
	result := entity.PreloadSegmentsVariants(getDB()).First(f, params.FlagID)

	// Flag with given ID doesn't exist, so we 404
	if result.RecordNotFound() {
		return flag.NewGetFlagDefault(404).WithPayload(
			ErrorMessage("unable to find flag %v in the database", params.FlagID))
	}

	// Something else happened, return a 500
	if err := result.Error; err != nil {
		return flag.NewGetFlagDefault(500).WithPayload(
			ErrorMessage("an unknown error occurred while looking up flag %v: %s", params.FlagID, err))
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
	fs := []entity.FlagSnapshot{}
	err := getDB().
		Order("created_at desc").
		Where(entity.FlagSnapshot{FlagID: util.SafeUint(params.FlagID)}).
		Find(&fs).Error
	if err != nil {
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
	if err := getDB().Order("key").Find(&entityTypes).Error; err != nil {
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

func (c *crud) PutFlag(params flag.PutFlagParams) middleware.Responder {
	f := &entity.Flag{}
	tx := getDB()

	if err := tx.First(f, params.FlagID).Error; err != nil {
		return flag.NewPutFlagDefault(404).WithPayload(ErrorMessage("%s", err))
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
			return flag.NewPutFlagDefault(400).WithPayload(ErrorMessage("%s", err))
		}
		f.Key = key
	}
	if params.Body.EntityType != nil {
		et := *params.Body.EntityType
		if err := entity.CreateFlagEntityType(tx, et); err != nil {
			return flag.NewPutFlagDefault(400).WithPayload(ErrorMessage("%s", err))
		}
		f.EntityType = et
	}

	if params.Body.Notes != nil {
		f.Notes = *params.Body.Notes
	}

	if err := tx.Save(f).Error; err != nil {
		return flag.NewPutFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	if err := entity.PreloadSegmentsVariants(tx).First(f, params.FlagID).Error; err != nil {
		return flag.NewPutFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := flag.NewPutFlagOK()
	payload, err := e2rMapFlag(f)
	if err != nil {
		return flag.NewPutFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	resp.SetPayload(payload)

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
	return resp
}

func (c *crud) SetFlagEnabledState(params flag.SetFlagEnabledParams) middleware.Responder {
	f := &entity.Flag{}
	if err := getDB().First(f, params.FlagID).Error; err != nil {
		return flag.NewSetFlagEnabledDefault(404).WithPayload(ErrorMessage("%s", err))
	}

	f.Enabled = *params.Body.Enabled

	if err := getDB().Save(f).Error; err != nil {
		return flag.NewSetFlagEnabledDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := flag.NewSetFlagEnabledOK()
	payload, err := e2rMapFlag(f)
	if err != nil {
		return flag.NewSetFlagEnabledDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	resp.SetPayload(payload)

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
	return resp
}

func (c *crud) DeleteFlag(params flag.DeleteFlagParams) middleware.Responder {
	if err := getDB().Delete(&entity.Flag{}, params.FlagID).Error; err != nil {
		return flag.NewDeleteFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	return flag.NewDeleteFlagOK()
}

func (c *crud) CreateSegment(params segment.CreateSegmentParams) middleware.Responder {
	s := &entity.Segment{}
	s.FlagID = uint(params.FlagID)
	s.RolloutPercent = uint(*params.Body.RolloutPercent)
	s.Description = util.SafeString(params.Body.Description)
	s.Rank = entity.SegmentDefaultRank

	err := getDB().Create(s).Error
	if err != nil {
		return segment.NewCreateSegmentDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := segment.NewCreateSegmentOK()
	resp.SetPayload(e2r.MapSegment(s))

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
	return resp
}

func (c *crud) FindSegments(params segment.FindSegmentsParams) middleware.Responder {
	ss := []entity.Segment{}
	err := entity.
		PreloadConstraintsDistribution(getDB()).
		Order("rank").
		Order("id").
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
	s := &entity.Segment{}
	err := entity.
		PreloadConstraintsDistribution(getDB()).
		First(s, params.SegmentID).
		Error
	if err != nil {
		return segment.NewPutSegmentDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	s.RolloutPercent = util.SafeUint(params.Body.RolloutPercent)
	s.Description = util.SafeString(params.Body.Description)

	if err := getDB().Save(s).Error; err != nil {
		return segment.NewPutSegmentDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := segment.NewPutSegmentOK()
	resp.SetPayload(e2r.MapSegment(s))

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
	return resp
}

func (c *crud) PutSegmentsReorder(params segment.PutSegmentsReorderParams) middleware.Responder {
	tx := getDB().Begin()
	for i, segmentID := range params.Body.SegmentIDs {
		s := &entity.Segment{}
		if err := tx.First(s, segmentID).Error; err != nil {
			tx.Rollback()
			return segment.NewPutSegmentsReorderDefault(404).WithPayload(ErrorMessage("%s", err))
		}
		s.Rank = uint(i)
		if err := tx.Save(s).Error; err != nil {
			tx.Rollback()
			return segment.NewPutSegmentsReorderDefault(500).WithPayload(ErrorMessage("%s", err))
		}
	}
	err := tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return segment.NewPutSegmentsReorderDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))

	return segment.NewPutSegmentsReorderOK()
}

func (c *crud) DeleteSegment(params segment.DeleteSegmentParams) middleware.Responder {
	if err := getDB().Delete(&entity.Segment{}, util.SafeUint(params.SegmentID)).Error; err != nil {
		return segment.NewDeleteSegmentDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
	return segment.NewDeleteSegmentOK()
}

func (c *crud) CreateConstraint(params constraint.CreateConstraintParams) middleware.Responder {
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
	if err := getDB().Create(cons).Error; err != nil {
		return constraint.NewCreateConstraintDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := constraint.NewCreateConstraintOK()
	resp.SetPayload(e2r.MapConstraint(cons))

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
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

	if err := getDB().Save(&cons).Error; err != nil {
		return constraint.NewPutConstraintDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := constraint.NewPutConstraintOK()
	resp.SetPayload(e2r.MapConstraint(cons))

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
	return resp
}

func (c *crud) DeleteConstraint(params constraint.DeleteConstraintParams) middleware.Responder {
	if err := getDB().Delete(entity.Constraint{}, params.ConstraintID).Error; err != nil {
		return constraint.NewDeleteConstraintDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := constraint.NewDeleteConstraintOK()

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
	return resp
}

// PutDistributions puts the whole distributions and overwrite the old ones
func (c *crud) PutDistributions(params distribution.PutDistributionsParams) middleware.Responder {
	if err := validatePutDistributions(params); err != nil {
		return distribution.NewPutDistributionsDefault(err.StatusCode).WithPayload(ErrorMessage("%s", err))
	}

	segmentID := uint(params.SegmentID)

	tx := getDB().Begin()
	err := tx.Delete(entity.Distribution{}, "segment_id = ?", segmentID).Error
	if err != nil {
		tx.Rollback()
		return distribution.NewPutDistributionsDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	ds := r2eMapDistributions(params.Body.Distributions, segmentID)
	for _, d := range ds {
		err1 := tx.Create(&d).Error
		if err1 != nil {
			tx.Rollback()
			return distribution.NewPutDistributionsDefault(500).WithPayload(ErrorMessage("%s", err))
		}
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return distribution.NewPutDistributionsDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := distribution.NewPutDistributionsOK()
	resp.SetPayload(e2r.MapDistributions(ds))

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
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
	v := &entity.Variant{}
	v.FlagID = uint(params.FlagID)
	v.Key = util.SafeString(params.Body.Key)

	a, err := r2eMapAttachment(params.Body.Attachment)
	if err != nil {
		return variant.NewCreateVariantDefault(400).WithPayload(ErrorMessage("%s", err))
	}
	v.Attachment = a

	if err := v.Validate(); err != nil {
		return variant.NewCreateVariantDefault(400).WithPayload(ErrorMessage("%s", err))
	}

	if err := getDB().Create(v).Error; err != nil {
		return variant.NewCreateVariantDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := variant.NewCreateVariantOK()
	resp.SetPayload(e2r.MapVariant(v))

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
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

	if err := getDB().Save(&v).Error; err != nil {
		return variant.NewPutVariantDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	if err := validatePutVariantForDistributions(v); err != nil {
		return variant.NewPutVariantDefault(err.StatusCode).WithPayload(ErrorMessage("%s", err))
	}

	resp := variant.NewPutVariantOK()
	resp.SetPayload(e2r.MapVariant(v))

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
	return resp
}

func (c *crud) DeleteVariant(params variant.DeleteVariantParams) middleware.Responder {
	if err := validateDeleteVariant(params); err != nil {
		return variant.NewDeleteVariantDefault(err.StatusCode).WithPayload(ErrorMessage("%s", err))
	}

	if err := getDB().Delete(entity.Variant{}, params.VariantID).Error; err != nil {
		return variant.NewDeleteVariantDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	entity.SaveFlagSnapshot(getDB(), util.SafeUint(params.FlagID), getSubjectFromRequest(params.HTTPRequest))
	return variant.NewDeleteVariantOK()
}
