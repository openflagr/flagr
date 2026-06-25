package handler

import (
	"fmt"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
)

// resolveFlagFromExposure resolves a flag from exposure row identifiers (catalog validation).
// At least one of flagID or flagKey is required; when both are set they must refer to the same flag.
func resolveFlagFromExposure(row *models.Exposure) (*entity.Flag, error) {
	if row == nil {
		return nil, fmt.Errorf("exposure row is null")
	}
	hasID := row.FlagID > 0
	hasKey := row.FlagKey != ""
	if !hasID && !hasKey {
		return nil, fmt.Errorf("flagID or flagKey is required")
	}

	ec := GetEvalCache()
	var flag *entity.Flag
	if hasID {
		flag = ec.GetByFlagKeyOrID(row.FlagID)
	}
	if hasKey {
		byKey := ec.GetByFlagKeyOrID(row.FlagKey)
		switch {
		case byKey == nil && flag == nil:
			return nil, fmt.Errorf("flag not found")
		case byKey == nil:
			// flag from ID only
		case flag == nil:
			flag = byKey
		case flag.ID != byKey.ID:
			return nil, fmt.Errorf("flagID and flagKey refer to different flags")
		}
	}
	if flag == nil {
		return nil, fmt.Errorf("flag not found")
	}
	return flag, nil
}

// resolveVariantOnFlag returns variant id/key for an optional variant reference on a flag.
// Zero id and empty key means no variant (allowed for exposure).
func resolveVariantOnFlag(flag *entity.Flag, variantID int64, variantKey string) (id int64, key string, err error) {
	if variantID <= 0 && variantKey == "" {
		return 0, "", nil
	}
	if variantID > 0 && variantKey != "" {
		var matchID, matchKey bool
		for _, v := range flag.Variants {
			if v.ID == uint(variantID) {
				matchID = true
			}
			if v.Key == variantKey {
				matchKey = true
			}
		}
		if !matchID {
			return 0, "", fmt.Errorf("variantID %d not found on flag", variantID)
		}
		if !matchKey {
			return 0, "", fmt.Errorf("variantKey %q not found on flag", variantKey)
		}
		for _, v := range flag.Variants {
			if v.ID == uint(variantID) && v.Key == variantKey {
				return int64(v.ID), v.Key, nil
			}
		}
		return 0, "", fmt.Errorf("variantID and variantKey do not match")
	}
	if variantID > 0 {
		for _, v := range flag.Variants {
			if v.ID == uint(variantID) {
				return int64(v.ID), v.Key, nil
			}
		}
		return 0, "", fmt.Errorf("variantID %d not found on flag", variantID)
	}
	for _, v := range flag.Variants {
		if v.Key == variantKey {
			return int64(v.ID), v.Key, nil
		}
	}
	return 0, "", fmt.Errorf("variantKey %q not found on flag", variantKey)
}