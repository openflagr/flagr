package handler

import (
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/variant"
)

var validatePutDistributions = func(params distribution.PutDistributionsParams) *Error {
	sum := int64(0)
	for _, d := range params.Body.Distributions {
		if d.Percent == nil {
			return NewError(400, "the percent of distribution %v is empty", d.ID)
		}
		sum += *d.Percent
	}
	if sum != 100 {
		return NewError(400, "the sum of distributions' percent %v is not 100", sum)
	}

	f := &entity.Flag{}
	err := entity.NewFlagQuerySet(getDB()).IDEq(uint(params.FlagID)).One(f)
	if err != nil {
		return NewError(400, "error finding flagID %v. reason %s", params.FlagID, err)
	}
	f.Preload(getDB())

	vMap := make(map[uint]string)
	vIDs := []uint{}
	for _, v := range f.Variants {
		vMap[v.ID] = v.Key
		vIDs = append(vIDs, v.ID)
	}

	for _, v := range params.Body.Distributions {
		vID := util.SafeUint(v.VariantID)
		k, ok := vMap[vID]
		if !ok {
			return NewError(400, "error finding variantID %v under this flag. expecting %v", vID, vIDs)
		}
		if k != util.SafeString(v.VariantKey) {
			return NewError(400, "error matching variantID %v with variantKey %s. expecting %s", vID, util.SafeString(v.VariantKey), k)
		}
	}

	return nil
}

var validateDeleteVariant = func(params variant.DeleteVariantParams) *Error {
	f := &entity.Flag{}
	err := entity.NewFlagQuerySet(getDB()).IDEq(uint(params.FlagID)).One(f)
	if err != nil {
		return NewError(400, "error finding flagID %v. reason %s", params.FlagID, err)
	}
	f.Preload(getDB())

	q := entity.NewDistributionQuerySet(getDB())
	for _, s := range f.Segments {
		for _, d := range s.Distributions {
			if d.VariantID == util.SafeUint(params.VariantID) {
				if d.Percent != uint(0) {
					return NewError(400, "error deleting variant %v. distribution %v still has non-zero distribution %v", params.VariantID, d.ID, d.Percent)
				}
				if err := q.IDEq(d.ID).Delete(); err != nil {
					return NewError(500, "error deleting distribution %v. reason: %s", d.ID, err)
				}
			}
		}
	}

	return nil
}

var validatePutVariantForDistributions = func(v *entity.Variant) *Error {
	q := entity.NewDistributionQuerySet(getDB())
	if err := q.VariantIDEq(v.ID).GetUpdater().SetVariantKey(v.Key).Update(); err != nil {
		return NewError(500, "error updating distribution to sync with variantID %v with variantKey %v. reason: %s", v.ID, v.Key, err)
	}
	return nil
}
