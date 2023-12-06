package handler

import (
	"fmt"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/variant"
	"gorm.io/gorm"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestValidatePutDistributions(t *testing.T) {
	db := entity.NewTestDB()
	c := &crud{}

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("control"),
		},
	})

	t.Run("happy code path", func(t *testing.T) {
		param := distribution.PutDistributionsParams{
			FlagID:    int64(1),
			SegmentID: int64(1),
			Body: &models.PutDistributionsRequest{
				Distributions: []*models.Distribution{
					{
						Percent:    util.Int64Ptr(int64(100)),
						VariantID:  util.Int64Ptr(int64(1)),
						VariantKey: util.StringPtr("control"),
					},
				},
			},
		}
		err := validatePutDistributions(param)
		assert.Nil(t, err)
	})

	t.Run("percent is nil", func(t *testing.T) {
		param := distribution.PutDistributionsParams{
			FlagID:    int64(1),
			SegmentID: int64(1),
			Body: &models.PutDistributionsRequest{
				Distributions: []*models.Distribution{
					{
						Percent:    nil,
						VariantID:  util.Int64Ptr(int64(1)),
						VariantKey: util.StringPtr("control"),
					},
				},
			},
		}
		err := validatePutDistributions(param)
		assert.NotZero(t, err)
	})

	t.Run("try to operate on a non-existing flag", func(t *testing.T) {
		param := distribution.PutDistributionsParams{
			FlagID:    int64(999999),
			SegmentID: int64(1),
			Body: &models.PutDistributionsRequest{
				Distributions: []*models.Distribution{
					{
						Percent:    util.Int64Ptr(int64(100)),
						VariantID:  util.Int64Ptr(int64(1)),
						VariantKey: util.StringPtr("control"),
					},
				},
			},
		}
		err := validatePutDistributions(param)
		assert.NotZero(t, err)
	})

	t.Run("try to operate on a non-existing variantID of the flag", func(t *testing.T) {
		param := distribution.PutDistributionsParams{
			FlagID:    int64(1),
			SegmentID: int64(1),
			Body: &models.PutDistributionsRequest{
				Distributions: []*models.Distribution{
					{
						Percent:    util.Int64Ptr(int64(100)),
						VariantID:  util.Int64Ptr(int64(999999)),
						VariantKey: util.StringPtr("control"),
					},
				},
			},
		}
		err := validatePutDistributions(param)
		assert.NotZero(t, err)
	})

	t.Run("try to operate on variant with inconsistent ID/Key pair", func(t *testing.T) {
		param := distribution.PutDistributionsParams{
			FlagID:    int64(1),
			SegmentID: int64(1),
			Body: &models.PutDistributionsRequest{
				Distributions: []*models.Distribution{
					{
						Percent:    util.Int64Ptr(int64(100)),
						VariantID:  util.Int64Ptr(int64(1)),
						VariantKey: util.StringPtr("treatment"),
					},
				},
			},
		}
		err := validatePutDistributions(param)
		assert.NotZero(t, err)
	})
}

func TestValidateDeleteVariant(t *testing.T) {
	db := entity.NewTestDB()
	c := &crud{}

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("control"),
		},
	})
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("treatment"),
		},
	})
	c.PutDistributions(distribution.PutDistributionsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.PutDistributionsRequest{
			Distributions: []*models.Distribution{
				{
					Percent:    util.Int64Ptr(int64(100)),
					VariantID:  util.Int64Ptr(int64(1)),
					VariantKey: util.StringPtr("control"),
				},
				{
					Percent:    util.Int64Ptr(int64(0)),
					VariantID:  util.Int64Ptr(int64(2)),
					VariantKey: util.StringPtr("treatment"),
				},
			},
		},
	})

	t.Run("happy code path - try to delete a variant with 0 percent distribution", func(t *testing.T) {
		param := variant.DeleteVariantParams{
			FlagID:    int64(1),
			VariantID: int64(2),
		}
		err := validateDeleteVariant(param)
		assert.Nil(t, err)
	})

	t.Run("try to delete a variant that's used in a distribution", func(t *testing.T) {
		param := variant.DeleteVariantParams{
			FlagID:    int64(1),
			VariantID: int64(1),
		}
		err := validateDeleteVariant(param)
		assert.NotZero(t, err)
	})
}

func TestValidatePutVariantForDistributions(t *testing.T) {
	db := entity.NewTestDB()
	c := &crud{}

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("control"),
		},
	})
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("treatment"),
		},
	})
	c.PutDistributions(distribution.PutDistributionsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.PutDistributionsRequest{
			Distributions: []*models.Distribution{
				{
					Percent:    util.Int64Ptr(int64(100)),
					VariantID:  util.Int64Ptr(int64(1)),
					VariantKey: util.StringPtr("control"),
				},
				{
					Percent:    util.Int64Ptr(int64(0)),
					VariantID:  util.Int64Ptr(int64(2)),
					VariantKey: util.StringPtr("treatment"),
				},
			},
		},
	})

	t.Run("happy code path", func(t *testing.T) {
		v := &entity.Variant{
			Model: gorm.Model{
				ID: 1,
			},
			FlagID: 1,
			Key:    "control",
		}
		err := validatePutVariantForDistributions(v)
		assert.Nil(t, err)
	})

	t.Run("validatePutVariantForDistributions - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		v := &entity.Variant{
			Model: gorm.Model{
				ID: 1,
			},
			FlagID: 1,
			Key:    "control",
		}
		err := validatePutVariantForDistributions(v)
		assert.NotZero(t, err)
		db.Error = nil
	})
}
