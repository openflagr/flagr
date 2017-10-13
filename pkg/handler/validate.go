package handler

import (
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/repo"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/distribution"
)

func validatePutDistributions(params distribution.PutDistributionsParams) *Error {

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
	err := entity.NewFlagQuerySet(repo.GetDB()).IDEq(uint(params.FlagID)).One(f)
	if err != nil {
		return NewError(400, "error finding flagID %v. reason %s", params.FlagID, err)
	}
	f.Preload(repo.GetDB())

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
			return NewError(400, "error matching variantID %v with variantKey %s. expecting %s.", vID, util.SafeString(v.VariantKey), k)
		}
	}

	return nil
}
