package r2e

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/models"
)

// MapDistributions maps distribution
func MapDistributions(r []*models.Distribution, segmentID uint) []entity.Distribution {
	e := make([]entity.Distribution, len(r))
	for i, d := range r {
		e[i] = MapDistribution(d, segmentID)
	}
	return e
}

// MapDistribution maps distribution
func MapDistribution(r *models.Distribution, segmentID uint) entity.Distribution {
	e := entity.Distribution{
		SegmentID:  segmentID,
		VariantID:  uint(*r.VariantID),
		VariantKey: util.SafeString(r.VariantKey),
		Percent:    uint(*r.Percent),
	}
	return e
}

func MapConstraints(r []*models.Constraint, segmentID uint) []entity.Constraint {
	e := make([]entity.Constraint, len(r))
	for i, d := range r {
		e[i] = MapConstraint(d, segmentID)
	}
	return e
}

// MapDistribution maps distribution
func MapConstraint(r *models.Constraint, segmentID uint) entity.Constraint {
	e := entity.Constraint{
		SegmentID: segmentID,
		Property:  util.SafeString(r.Property),
		Operator:  util.SafeString(r.Operator),
		Value:     util.SafeString(r.Value),
	}
	return e
}

// MapAttachment maps attachment
func MapAttachment(a interface{}) (entity.Attachment, error) {
	e := entity.Attachment{}

	if a != nil {
		m, ok := a.(map[string]interface{})
		if !ok {
			return e, fmt.Errorf("Make sure JSON is properly formatted into key/value pairs. Invalid attachment format %s", spew.Sdump(a))
		}
		e = m
	}
	return e, nil
}
