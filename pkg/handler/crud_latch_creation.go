package handler

import (
	"fmt"

	"github.com/Allen-Career-Institute/flagr/pkg/entity"
	"github.com/Allen-Career-Institute/flagr/pkg/util"
	"github.com/Allen-Career-Institute/flagr/swagger_gen/restapi/operations/flag"
	"github.com/Allen-Career-Institute/flagr/swagger_gen/restapi/operations/latch"
	"github.com/go-openapi/runtime/middleware"
	"gorm.io/gorm"
)

const ErrorCreatingLatch = "cannot create latch. %s"

// Declare a global function variable for LoadSimpleLatchTemplate
var loadLatchTemplateFunc = LoadSimpleLatchTemplate
var associateTagWithFlagFunc = associateTagWithFlag

func (c *crud) CreateLatch(params latch.CreateLatchParams) middleware.Responder {
	f, tx, responder := c.createFlagEntity(flag.CreateFlagParams(params))
	if responder != nil {
		return responder
	}

	if err := loadLatchTemplateFunc(f, tx); err != nil {
		tx.Rollback()
		return flag.NewCreateFlagDefault(500).WithPayload(
			ErrorMessage(ErrorCreatingLatch, err))
	}

	err := tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return flag.NewCreateFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp, m := c.mapResponseAndSaveFlagSnapShot(flag.CreateFlagParams(params), f)
	if m != nil {
		return m
	}

	return resp
}

// LoadSimpleLatchTemplate loads the simple latch template into
// a new flag. It creates a single segment with 100% rollout, variant ('APPLICABLE'),
// and distribution of variant as 100% as well.
func LoadSimpleLatchTemplate(flag *entity.Flag, tx *gorm.DB) error {
	// adding latch tag with each creation in order easily fetch all the latches filtering out AB experiments
	err := associateTagWithFlagFunc(flag, tx, "latch")
	if err != nil {
		return err
	}

	//our default Variant "APPLICABLE" that tells if latch is applicable for set of levers
	v := &entity.Variant{}
	v.FlagID = flag.ID
	v.Key = "APPLICABLE"

	if err := tx.Create(v).Error; err != nil {
		return err
	}

	// Create our default segment with 100% rollout
	seg := &entity.Segment{}
	seg.FlagID = flag.ID
	seg.RolloutPercent = uint(100)
	seg.Rank = entity.SegmentDefaultRank

	if err := tx.Create(seg).Error; err != nil {
		return err
	}

	// default Distribution with 100% for the variant
	dist := &entity.Distribution{}
	dist.SegmentID = seg.ID
	dist.VariantID = v.ID
	dist.VariantKey = v.Key
	dist.Percent = uint(100)

	if err := tx.Create(dist).Error; err != nil {
		return err
	}

	seg.Distributions = append(seg.Distributions, *dist)
	flag.Variants = append(flag.Variants, *v)
	flag.Segments = append(flag.Segments, *seg)

	return nil
}

func associateTagWithFlag(flag *entity.Flag, tx *gorm.DB, tag string) error {
	t := &entity.Tag{Value: tag}
	t.Value = util.SafeString(tag)
	if ok, reason := util.IsSafeValue(t.Value); !ok {
		return fmt.Errorf("error creating tag: %v", reason)
	}

	// Find or create the tag
	if err := tx.Where("value = ?", util.SafeString(tag)).FirstOrCreate(t).Error; err != nil {
		return fmt.Errorf("error finding or creating tag: %v", err)
	}

	// Associate tag to flag
	if err := entity.AddTagToFlag(tx, flag.ID, t.ID); err != nil {
		return fmt.Errorf("error associating tag with flag: %v", err)
	}
	return nil
}
