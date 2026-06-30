package handler

import (
	"github.com/openflagr/flagr/pkg/notification"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCommitFlagMutation_RollbackOnMutateFailure(t *testing.T) {
	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
	c := &crud{}

	createRes := c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{Description: new("rollback-test")},
	})
	createOK := createRes.(*flag.CreateFlagOK)
	require.NotNil(t, createOK.Payload)
	flagID := createOK.Payload.ID

	var before int64
	require.NoError(t, db.Model(&entity.FlagSnapshot{}).Where("flag_id = ?", flagID).Count(&before).Error)

	err := commitFlagMutation(uint(flagID), "tester", notification.OperationUpdate, notification.ComponentFlag, func(tx *gorm.DB) (uint, MutationNotify, error) {
		return 0, MutationNotify{}, gorm.ErrInvalidDB
	})
	assert.Error(t, err)

	var after int64
	require.NoError(t, db.Model(&entity.FlagSnapshot{}).Where("flag_id = ?", flagID).Count(&after).Error)
	assert.Equal(t, before, after, "failed mutation must not add a snapshot row")
}