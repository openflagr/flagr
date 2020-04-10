package handler

import (
	"fmt"
	"testing"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/go-openapi/runtime/middleware"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestCrudCreateFlag(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("it should be able to create one flag", func(t *testing.T) {
		res = c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("funny flag"),
				Key:         "some_random_flag_key",
			},
		})
		assert.NotZero(t, res.(*flag.CreateFlagOK).Payload.ID)
		assert.Equal(t, "some_random_flag_key", res.(*flag.CreateFlagOK).Payload.Key)

		flagID := uint(res.(*flag.CreateFlagOK).Payload.ID)
		segment := entity.Segment{FlagID: flagID}
		db.First(&segment)
		assert.Zero(t, segment.ID)
	})

	t.Run("it should be able to create simple_boolean_flag template", func(t *testing.T) {
		res = c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("simple flag"),
				Key:         "simple_boolean_flag_key",
				Template:    "simple_boolean_flag",
			},
		})
		res := res.(*flag.CreateFlagOK)
		assert.NotZero(t, res.Payload.ID)
		assert.Equal(t, "simple_boolean_flag_key", res.Payload.Key)
		assert.Equal(t, len(res.Payload.Variants), 1)
		assert.Equal(t, len(res.Payload.Segments), 1)
		assert.Equal(t, len(res.Payload.Segments[0].Distributions), 1)
		flagID := uint(res.Payload.ID)
		segment := entity.Segment{FlagID: flagID}
		db.First(&segment)
		assert.NotZero(t, segment.ID)
		assert.Equal(t, segment.Rank, entity.SegmentDefaultRank)

		variant := entity.Variant{FlagID: flagID}
		db.First(&variant)
		assert.NotZero(t, variant.ID)
		assert.Equal(t, variant.Key, "on")

		distribution := entity.Distribution{VariantID: variant.ID}
		db.First(&distribution)
		assert.NotZero(t, distribution.ID)
		assert.Equal(t, distribution.Percent, uint(100))
		assert.Equal(t, distribution.SegmentID, segment.ID)
		assert.Equal(t, distribution.VariantKey, variant.Key)
	})
}

func TestCrudCreateFlagWithFailures(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("CreateFlag - got e2r MapFlag error", func(t *testing.T) {
		defer gostub.StubFunc(&e2rMapFlag, nil, fmt.Errorf("e2r MapFlag error")).Reset()
		res = c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("funny flag"),
			},
		})
		assert.NotZero(t, res.(*flag.CreateFlagDefault).Payload)
	})

	t.Run("CreateFlag - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("funny flag"),
			},
		})
		assert.NotZero(t, res.(*flag.CreateFlagDefault).Payload)
		db.Error = nil
	})

	t.Run("CreateFlag - invalid key error", func(t *testing.T) {
		res = c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr(" flag with a space"),
				Key:         " 1-2-3", // invalid key
			},
		})
		assert.NotZero(t, res.(*flag.CreateFlagDefault).Payload)
	})

	t.Run("CreateFlag - invalid template error", func(t *testing.T) {
		res = c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr(" flag with a space"),
				Key:         "invalid_template",
				Template:    "invalid_template",
			},
		})
		assert.NotZero(t, res.(*flag.CreateFlagDefault).Payload)
	})
}
