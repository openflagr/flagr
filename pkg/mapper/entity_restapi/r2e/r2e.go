package r2e

import (
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
)

// MapDistributions maps distribution
func MapDistributions(r models.PutDistributionsRequestDistributions, segmentID uint) []entity.Distribution {
	e := make([]entity.Distribution, len(r), len(r))
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
		Bitmap:     r.Bitmap,
	}
	return e
}
