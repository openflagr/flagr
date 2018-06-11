package handler

import (
	"testing"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/query"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestQueryFlags(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	qAPI := &queryAPI{}
	crudAPI := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	// Initialize 2 flags
	res = crudAPI.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Name:        util.StringPtr("flag1"),
			Description: util.StringPtr("1st flag"),
		},
	})
	flag1 := res.(*flag.CreateFlagOK).Payload

	res = crudAPI.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Name:        util.StringPtr("flag2"),
			Description: util.StringPtr("2nd flag"),
		},
	})
	flag2 := res.(*flag.CreateFlagOK).Payload

	// Should return an error if flag with name doesn't exist
	res = qAPI.GetFlagByName(query.GetFlagByNameParams{FlagName: "flag-6"})
	assert.NotZero(t, res.(*query.GetFlagByNameDefault).Payload)

	// Should correctly find the flag if it does exist
	res = qAPI.GetFlagByName(query.GetFlagByNameParams{FlagName: "flag1"})
	assert.Equal(t, flag1.ID, res.(*query.GetFlagByNameOK).Payload.ID)

	// Batch query should be able to find multiple flags
	res = qAPI.GetFlagByNameBatch(query.GetFlagByNameBatchParams{
		FlagNames: []string{"flag1", "flag2"},
	})
	assert.Equal(t, 2, len(res.(*query.GetFlagByNameBatchOK).Payload))

	// Batch query should ignore names that don't exist
	res = qAPI.GetFlagByNameBatch(query.GetFlagByNameBatchParams{
		FlagNames: []string{"flag2", "flag6"},
	})
	flags := res.(*query.GetFlagByNameBatchOK).Payload
	assert.Equal(t, 1, len(flags))
	assert.Equal(t, flag2.ID, flags[0].ID)
}
