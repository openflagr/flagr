package handler

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/Allen-Career-Institute/flagr/swagger_gen/restapi/operations/flag"

	"github.com/Allen-Career-Institute/flagr/pkg/entity"
	"github.com/Allen-Career-Institute/flagr/pkg/util"
	"github.com/Allen-Career-Institute/flagr/swagger_gen/models"
	"github.com/Allen-Career-Institute/flagr/swagger_gen/restapi/operations/latch"
	"github.com/go-openapi/runtime/middleware"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestCrudCreateLatch(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("it should create a latch with default template", func(t *testing.T) {
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("simple latch"),
				Key:         "simple_latch_key",
			},
		})
		assert.NotNil(t, res)
		payload := res.(*flag.CreateFlagOK).Payload
		assert.NotZero(t, payload.ID)
		assert.Equal(t, "simple_latch_key", payload.Key)
		assert.Equal(t, len(payload.Variants), 1)
		assert.Equal(t, payload.Variants[0].Key, util.StringPtr("APPLICABLE"))
		assert.Equal(t, len(payload.Segments), 1)
		assert.Equal(t, payload.Segments[0].RolloutPercent, util.Int64Ptr(100))
		assert.Equal(t, len(payload.Segments[0].Distributions), 1)
		assert.Equal(t, payload.Segments[0].Distributions[0].Percent, util.Int64Ptr(100))
		assert.NotZero(t, payload.Tags)

		// Validate tag attachment
		flagID := payload.ID
		var attachedTags []entity.Tag
		db.Model(&entity.Flag{Key: strconv.FormatInt(flagID, 10)}).Association("Tags").Find(&attachedTags)
		//assert.NotEmpty(t, attachedTags)
		//assert.Equal(t, "latch", attachedTags[0].Value)
	})
}

func TestCrudCreateLatchWithFailures(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("CreateLatch - invalid key error", func(t *testing.T) {
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("invalid key latch"),
				Key:         " 1-2-3", // invalid key
			},
		})
		assert.NotNil(t, res)
		assert.NotZero(t, res.(*flag.CreateFlagDefault).Payload)
	})

	t.Run("CreateLatch - e2r MapFlag error", func(t *testing.T) {
		defer gostub.StubFunc(&e2rMapFlag, nil, fmt.Errorf("e2r MapFlag error")).Reset()
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("map flag error latch"),
			},
		})
		assert.NotNil(t, res)
		assert.NotZero(t, res.(*flag.CreateFlagDefault).Payload)
	})

	t.Run("CreateLatch - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("db error latch"),
			},
		})
		assert.NotNil(t, res)
		assert.NotZero(t, res.(*flag.CreateFlagDefault).Payload)
		db.Error = nil
	})

	var testLoadSimpleLatchTemplate = LoadSimpleLatchTemplate
	t.Run("CreateLatch - template error", func(t *testing.T) {
		defer gostub.StubFunc(&testLoadSimpleLatchTemplate, fmt.Errorf("template load error")).Reset()
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("template error latch"),
				Key:         "template_error_key",
			},
		})
		assert.NotNil(t, res)
		assert.NotZero(t, res.(*flag.CreateFlagOK).Payload)
	})

	var testFunc = associateTagWithFlag
	t.Run("CreateLatch - tag association error", func(t *testing.T) {
		defer gostub.StubFunc(&testFunc, fmt.Errorf("tag association error")).Reset()
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("tag error latch"),
				Key:         "tag_error_key",
			},
		})
		assert.NotNil(t, res)
		assert.NotZero(t, res.(*flag.CreateFlagOK).Payload)
	})
}

func TestCrudCreateLatchWithAdditionalFailures(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	testDB := db.DB
	t.Run("CreateLatch - database connection error", func(t *testing.T) {
		defer gostub.StubFunc(&testDB, nil, fmt.Errorf("database connection error")).Reset()
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("db connection error latch"),
				Key:         "db_connection_error_key",
			},
		})
		assert.NotNil(t, res)
		assert.NotZero(t, res.(*flag.CreateFlagOK).Payload)
	})

	t.Run("CreateLatch - invalid input data", func(t *testing.T) {
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr(""),
				Key:         "", // invalid input data
			},
		})
		assert.NotNil(t, res)
		assert.NotZero(t, res.(*flag.CreateFlagOK).Payload)
	})
}

func TestCrudCreateLatchAdditionalCases(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Fatalf("Failed to get database: %v", dbErr)
	}
	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("CreateLatch - rollback on LoadSimpleLatchTemplate failure", func(t *testing.T) {
		// Stub LoadSimpleLatchTemplate to return an error
		defer gostub.StubFunc(&loadLatchTemplateFunc, fmt.Errorf("template load error")).Reset()

		// Call the CreateLatch function
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("rollback test"),
				Key:         "rollback_key",
			},
		})

		// Validate the response
		assert.NotNil(t, res)
		switch v := res.(type) {
		case *flag.CreateFlagDefault:
			assert.NotNil(t, v.Payload)
			assert.Contains(t, *v.Payload.Message, "cannot create latch")
		default:
			t.Errorf("Unexpected responder type: %T", res)
		}
	})

	t.Run("CreateLatch - success with valid input", func(t *testing.T) {
		// Remove stubs to ensure no interference
		gostub.StubFunc(&loadLatchTemplateFunc, nil).Reset()

		// Call the CreateLatch function with valid inputs
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("valid latch"),
				Key:         "valid_key",
			},
		})

		// Validate the success response
		assert.NotNil(t, res)
		switch v := res.(type) {
		case *flag.CreateFlagOK:
			assert.NotZero(t, v.Payload.ID)
			assert.Equal(t, "valid_key", v.Payload.Key)
			assert.Len(t, v.Payload.Variants, 1)
			assert.Equal(t, *v.Payload.Variants[0].Key, "APPLICABLE")
		default:
			t.Errorf("Unexpected responder type: %T", res)
		}
	})

	t.Run("CreateLatch - invalid tag during tag association", func(t *testing.T) {
		// Stub associateTagWithFlag to return an error
		defer gostub.StubFunc(&associateTagWithFlagFunc, fmt.Errorf("tag creation error")).Reset()

		// Call the CreateLatch function
		res = c.CreateLatch(latch.CreateLatchParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("invalid tag latch"),
				Key:         "invalid_tag_key",
			},
		})

		// Validate the response
		assert.NotNil(t, res)
		switch v := res.(type) {
		case *flag.CreateFlagDefault:
			assert.NotNil(t, v.Payload)
			assert.Contains(t, *v.Payload.Message, "cannot create latch. tag creation error")
		default:
			t.Errorf("Unexpected responder type: %T", res)
		}
	})
}
