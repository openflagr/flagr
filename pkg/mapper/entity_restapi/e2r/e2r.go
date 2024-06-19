package e2r

import (
	"time"

	"encoding/json"

	"github.com/go-openapi/strfmt"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"

	"github.com/openflagr/flagr/swagger_gen/models"
)

// MapFlag maps flag
func MapFlag(e *entity.Flag) (*models.Flag, error) {
	r := &models.Flag{}
	r.ID = int64(e.ID)
	r.Key = e.Key
	r.CreatedBy = e.CreatedBy
	r.DataRecordsEnabled = util.BoolPtr(e.DataRecordsEnabled)
	r.EntityType = e.EntityType
	r.Description = util.StringPtr(e.Description)
	r.Notes = e.Notes
	r.Enabled = util.BoolPtr(e.Enabled)
	r.UpdatedAt = strfmt.DateTime(e.UpdatedAt)
	r.UpdatedBy = e.UpdatedBy
	r.Segments = MapSegments(e.Segments)
	r.Variants = MapVariants(e.Variants)
	r.Tags = MapTags(e.Tags)

	return r, nil
}

// MapFlags maps flags
func MapFlags(e []entity.Flag) ([]*models.Flag, error) {
	ret := make([]*models.Flag, len(e))
	for i, f := range e {
		rf, err := MapFlag(&f)
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
	f, err := MapFlag(ef)
	if err != nil {
		return nil, err
	}
	r := &models.FlagSnapshot{
		Flag:      f,
		ID:        int64(e.ID),
		UpdatedBy: e.UpdatedBy,
		UpdatedAt: util.StringPtr(e.UpdatedAt.UTC().Format(time.RFC3339)),
	}
	return r, nil
}

// MapFlagSnapshots maps flag snapshots
func MapFlagSnapshots(e []entity.FlagSnapshot) ([]*models.FlagSnapshot, error) {
	ret := make([]*models.FlagSnapshot, len(e))
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
func MapSegment(e *entity.Segment) *models.Segment {
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
func MapSegments(e []entity.Segment) []*models.Segment {
	ret := make([]*models.Segment, len(e))
	for i, s := range e {
		ret[i] = MapSegment(&s)
	}
	return ret
}

// MapTagEntity maps tag entity
func MapTag(e *entity.Tag) *models.Tag {
	r := &models.Tag{}
	r.ID = int64(e.ID)
	r.Value = util.StringPtr(e.Value)
	return r
}

// MapTags maps tags
func MapTags(e []entity.Tag) []*models.Tag {
	ret := make([]*models.Tag, len(e))
	for i, s := range e {
		ret[i] = MapTag(&s)
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
	ret := make([]*models.Constraint, len(e))
	for i, c := range e {
		ret[i] = MapConstraint(&c)
	}
	return ret
}

// MapDistribution maps to a distribution
func MapDistribution(e *entity.Distribution) *models.Distribution {
	r := &models.Distribution{
		ID:         int64(e.ID),
		Percent:    util.Int64Ptr(int64(e.Percent)),
		VariantID:  util.Int64Ptr(int64(e.VariantID)),
		VariantKey: util.StringPtr(e.VariantKey),
	}
	return r
}

// MapDistributions maps distribution
func MapDistributions(e []entity.Distribution) []*models.Distribution {
	ret := make([]*models.Distribution, len(e))
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
	ret := make([]*models.Variant, len(e))
	for i, v := range e {
		ret[i] = MapVariant(&v)
	}
	return ret
}
