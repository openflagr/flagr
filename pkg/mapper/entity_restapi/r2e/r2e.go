package r2e

import (
	"fmt"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/davecgh/go-spew/spew"
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
func MapAttachment(a interface{}) (entity.Attachment, error) {
	e := entity.Attachment{}

	if a != nil {
		m, ok := a.(map[string]interface{})
		if !ok {
			return e, fmt.Errorf("all key/value pairs should be string/string. invalid attachment format %s", spew.Sdump(a))
		}
		for k, v := range m {
			s, ok := v.(string)
			if !ok {
				return e, fmt.Errorf("all key/value pairs should be string/string. invalid attachment format %s", spew.Sdump(a))
			}
			e[k] = s
		}
	}
	return e, nil
}
