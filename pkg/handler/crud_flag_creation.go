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
