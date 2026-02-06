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

// MapAttachment maps attachment
func MapAttachment(a any) (entity.Attachment, error) {
	e := entity.Attachment{}

	if a != nil {
		m, ok := a.(map[string]any)
		if !ok {
			return e, fmt.Errorf("make sure JSON is properly formatted into key/value pairs. Invalid attachment format %s", spew.Sdump(a))
		}
		e = m
	}
	return e, nil
}
