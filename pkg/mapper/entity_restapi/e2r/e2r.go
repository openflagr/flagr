package e2r

import (
	"encoding/json"
	"time"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/go-openapi/strfmt"

	"github.com/checkr/flagr/swagger_gen/models"
)

var getDB = entity.GetDB

// MapFlag maps flag
func MapFlag(e *entity.Flag, preload bool) (*models.Flag, error) {
	r := &models.Flag{}
	r.ID = int64(e.ID)
	r.Key = e.Key
	r.Description = util.StringPtr(e.Description)
	r.Enabled = util.BoolPtr(e.Enabled)
	r.DataRecordsEnabled = util.BoolPtr(e.DataRecordsEnabled)
	r.UpdatedAt = strfmt.DateTime(e.UpdatedAt)
	r.CreatedBy = e.CreatedBy
	r.UpdatedBy = e.UpdatedBy

	if preload {
		if err := e.Preload(getDB()); err != nil {
			return nil, err
		}
	}
	r.Segments = MapSegments(e.Segments, preload)
	r.Variants = MapVariants(e.Variants)
	return r, nil
}

// MapFlags maps flags
func MapFlags(e []entity.Flag) ([]*models.Flag, error) {
	ret := make([]*models.Flag, len(e), len(e))
	for i, f := range e {
		rf, err := MapFlag(&f, false)
		if err != nil {
			return nil, err
		}
		ret[i] = rf
	}
	return ret, nil
}

// MapFlagSnapshot maps flag snapshot
func MapFlagSnapshot(e *entity.FlagSnapshot) (*models.FlagSnapshot, error) {
	ef := &entity.Flag{}
	if err := json.Unmarshal(e.Flag, ef); err != nil {
		return nil, err
	}
	f, err := MapFlag(ef, false)
	if err != nil {
		return nil, err
	}
	r := &models.FlagSnapshot{
		Flag:      f,
		ID:        util.Int64Ptr(int64(e.ID)),
		UpdatedBy: e.UpdatedBy,
		UpdatedAt: util.StringPtr(e.UpdatedAt.UTC().Format(time.RFC3339)),
	}
	return r, nil
}

// MapFlagSnapshots maps flag snapshots
func MapFlagSnapshots(e []entity.FlagSnapshot) ([]*models.FlagSnapshot, error) {
	ret := make([]*models.FlagSnapshot, len(e), len(e))
	for i, fs := range e {
		rf, err := MapFlagSnapshot(&fs)
		if err != nil {
			return nil, err
		}
		ret[i] = rf
	}
	return ret, nil
}

// MapSegment maps segment
func MapSegment(e *entity.Segment, preload bool) *models.Segment {
	if preload {
		e.Preload(getDB())
	}

	r := &models.Segment{}
	r.ID = int64(e.ID)
	r.Description = util.StringPtr(e.Description)
	r.Rank = util.Int64Ptr(int64(e.Rank))
	r.RolloutPercent = util.Int64Ptr(int64(e.RolloutPercent))
	r.Constraints = MapConstraints(e.Constraints)
	r.Distributions = MapDistributions(e.Distributions)
	return r
}

// MapSegments maps segments
func MapSegments(e []entity.Segment, preload bool) []*models.Segment {
	ret := make([]*models.Segment, len(e), len(e))
	for i, s := range e {
		ret[i] = MapSegment(&s, preload)
	}
	return ret
}

// MapConstraint maps constraint
func MapConstraint(e *entity.Constraint) *models.Constraint {
	r := &models.Constraint{}
	r.ID = int64(e.ID)
	r.Property = util.StringPtr(e.Property)
	r.Operator = util.StringPtr(e.Operator)
	r.Value = util.StringPtr(e.Value)
	return r
}

// MapConstraints maps constraints
func MapConstraints(e []entity.Constraint) []*models.Constraint {
	ret := make([]*models.Constraint, len(e), len(e))
	for i, c := range e {
		ret[i] = MapConstraint(&c)
	}
	return ret
}

// MapDistribution maps to a distribution
func MapDistribution(e *entity.Distribution) *models.Distribution {
	r := &models.Distribution{
		Bitmap:     e.Bitmap,
		ID:         int64(e.ID),
		Percent:    util.Int64Ptr(int64(e.Percent)),
		VariantID:  util.Int64Ptr(int64(e.VariantID)),
		VariantKey: util.StringPtr(e.VariantKey),
	}
	return r
}

// MapDistributions maps distribution
func MapDistributions(e []entity.Distribution) []*models.Distribution {
	ret := make([]*models.Distribution, len(e), len(e))
	for i, d := range e {
		ret[i] = MapDistribution(&d)
	}
	return ret
}

// MapVariant maps variant
func MapVariant(e *entity.Variant) *models.Variant {
	r := &models.Variant{
		ID:         int64(e.ID),
		Key:        util.StringPtr(e.Key),
		Attachment: e.Attachment,
	}
	return r
}

// MapVariants maps variant
func MapVariants(e []entity.Variant) []*models.Variant {
	ret := make([]*models.Variant, len(e), len(e))
	for i, v := range e {
		ret[i] = MapVariant(&v)
	}
	return ret
}
