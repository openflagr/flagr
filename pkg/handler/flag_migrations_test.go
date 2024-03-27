package handler

import (

	"github.com/go-openapi/runtime/middleware"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlagMigrations(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB.Close()

	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("FlagMigrationWithValidPath", func(t *testing.T) {
		_, err := FlagMigrations("./testdata/migrations")
		assert.Nil(t, err)

		res = c.FindFlags(flag.FindFlagsParams{
			Preload: util.BoolPtr(true),
			Key: util.StringPtr("FLAGS-123"),
		})

		flags := res.(*flag.FindFlagsOK).Payload

		assert.NotZero(t, len(flags))
		assert.Equal(t, len(flags), 1)
		assert.NotZero(t, flags[0].ID)
		assert.Equal(t, flags[0].Key, "FLAGS-123")
		assert.Equal(t, len(flags[0].Variants), 1)
		assert.Equal(t, len(flags[0].Segments), 1)
		assert.Equal(t, len(flags[0].Segments[0].Distributions), 1)

		res = c.FindFlags(flag.FindFlagsParams{
			Preload: util.BoolPtr(true),
			Key: util.StringPtr("FLAGS-124"),
		})

		flags = res.(*flag.FindFlagsOK).Payload

		assert.NotZero(t, len(flags))
		assert.Equal(t, len(flags), 1)
		assert.NotZero(t, flags[0].ID)
		assert.Equal(t, flags[0].Key, "FLAGS-124")
		assert.Equal(t, len(flags[0].Variants), 2)
		assert.Equal(t, len(flags[0].Segments), 2)
		assert.Equal(t, len(flags[0].Segments[0].Distributions), 1)

	})

	t.Run("FlagMigrationWithInvalidPath", func(t *testing.T) {
		_, err := FlagMigrations("./testdata/migration")
		assert.NotNil(t, err)
	})
}
