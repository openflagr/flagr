package handler

import (
	"github.com/Allen-Career-Institute/flagr/pkg/entity"
	"github.com/Allen-Career-Institute/flagr/pkg/util"
	"github.com/Allen-Career-Institute/flagr/swagger_gen/restapi/operations/flag"
	"github.com/go-openapi/runtime/middleware"
	"gorm.io/gorm"
)

const ErrorCreatingFlag = "cannot create flag. %s"

func (c *crud) CreateFlag(params flag.CreateFlagParams) middleware.Responder {
	f, tx, errCreateFlag := c.createFlagEntity(params)
	if errCreateFlag != nil {
		return errCreateFlag
	}

	if params.Body.Template == "simple_boolean_flag" {
		if err := LoadSimpleBooleanFlagTemplate(f, tx); err != nil {
			tx.Rollback()
			return flag.NewCreateFlagDefault(500).WithPayload(
				ErrorMessage(ErrorCreatingFlag, err))
		}
	} else if params.Body.Template != "" {
		return flag.NewCreateFlagDefault(400).WithPayload(
			ErrorMessage("unknown value for template: %s", params.Body.Template))
	}

	// adding AB tag in order to easily fetch only AB Experiments and ignore latch
	err := associateTagWithFlagFunc(f, tx, "AB")
	if err != nil {
		return flag.NewCreateFlagDefault(500).WithPayload(
			ErrorMessage("cannot associate AB tag to flag. %s", err))
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return flag.NewCreateFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp, m := c.mapResponseAndSaveFlagSnapShot(params, f)
	if m != nil {
		return m
	}

	return resp
}

func (c *crud) mapResponseAndSaveFlagSnapShot(params flag.CreateFlagParams, f *entity.Flag) (*flag.CreateFlagOK, middleware.Responder) {
	// Reload the flag with all relationships (segments, variants, tags) to include in response
	if err := entity.PreloadSegmentsVariantsTags(getDB()).First(f, f.ID).Error; err != nil {
		return nil, flag.NewCreateFlagDefault(500).WithPayload(
			ErrorMessage("cannot reload flag with relationships. %s", err))
	}

	resp := flag.NewCreateFlagOK()
	payload, err := e2rMapFlag(f)
	if err != nil {
		return nil, flag.NewCreateFlagDefault(500).WithPayload(
			ErrorMessage("cannot map flag. %s", err))
	}
	resp.SetPayload(payload)

	entity.SaveFlagSnapshot(getDB(), f.ID, getSubjectFromRequest(params.HTTPRequest))
	return resp, nil
}

func (c *crud) createFlagEntity(params flag.CreateFlagParams) (*entity.Flag, *gorm.DB, middleware.Responder) {
	f := &entity.Flag{}
	if params.Body != nil {
		f.Description = util.SafeString(params.Body.Description)
		f.CreatedBy = getSubjectFromRequest(params.HTTPRequest)

		key, err := entity.CreateFlagKey(params.Body.Key)
		if err != nil {
			return nil, nil, flag.NewCreateFlagDefault(400).WithPayload(
				ErrorMessage(ErrorCreatingFlag, err))
		}
		f.Key = key
	}

	tx := getDB().Begin()

	if err := tx.Create(f).Error; err != nil {
		tx.Rollback()
		return nil, nil, flag.NewCreateFlagDefault(500).WithPayload(
			ErrorMessage(ErrorCreatingFlag, err))
	}
	return f, tx, nil
}

// LoadSimpleBooleanFlagTemplate loads the simple boolean flag template into
// a new flag. It creates a single segment, variant ('on'), and distribution.
func LoadSimpleBooleanFlagTemplate(flag *entity.Flag, tx *gorm.DB) error {
	// Create our default segment
	s := &entity.Segment{}
	s.FlagID = flag.ID
	s.RolloutPercent = uint(100)
	s.Rank = entity.SegmentDefaultRank

	if err := tx.Create(s).Error; err != nil {
		return err
	}

	// .. and our default Variant
	v := &entity.Variant{}
	v.FlagID = flag.ID
	v.Key = "on"

	if err := tx.Create(v).Error; err != nil {
		return err
	}

	// .. and our default Distribution
	d := &entity.Distribution{}
	d.SegmentID = s.ID
	d.VariantID = v.ID
	d.VariantKey = v.Key
	d.Percent = uint(100)

	if err := tx.Create(d).Error; err != nil {
		return err
	}

	s.Distributions = append(s.Distributions, *d)
	flag.Variants = append(flag.Variants, *v)
	flag.Segments = append(flag.Segments, *s)

	return nil
}
