package handler

import (
	"net/http"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/tag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/variant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDuplicateFlag(t *testing.T) {
	_, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}
	req := &http.Request{}

	createRes := c.CreateFlag(flag.CreateFlagParams{
		HTTPRequest: req,
		Body: &models.CreateFlagRequest{
			Description: new("source flag"),
			Key:         "dup_source_key",
		},
	})
	createOK, isCreate := createRes.(*flag.CreateFlagOK)
	require.True(t, isCreate, "create failed: %T", createRes)
	require.NotNil(t, createOK.Payload)
	flagID := createOK.Payload.ID

	c.CreateVariant(variant.CreateVariantParams{
		FlagID: flagID,
		Body:   &models.CreateVariantRequest{Key: new("control")},
	})
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: flagID,
		Body:   &models.CreateVariantRequest{Key: new("treatment")},
	})
	rollout := int64(100)
	c.CreateSegment(segment.CreateSegmentParams{
		FlagID: flagID,
		Body: &models.CreateSegmentRequest{
			Description:    new("seg"),
			RolloutPercent: &rollout,
		},
	})
	c.CreateTag(tag.CreateTagParams{
		FlagID: flagID,
		Body:   &models.CreateTagRequest{Value: new("dup-tag")},
	})

	res := c.DuplicateFlag(flag.DuplicateFlagParams{FlagID: flagID, HTTPRequest: req})
	ok, isOK := res.(*flag.DuplicateFlagOK)
	require.True(t, isOK, "expected DuplicateFlagOK, got %T", res)
	require.NotNil(t, ok.Payload)
	assert.NotEqual(t, flagID, ok.Payload.ID)
	require.NotNil(t, ok.Payload.Description)
	assert.Contains(t, *ok.Payload.Description, "(cloned)")
	assert.NotEmpty(t, ok.Payload.Key)
	assert.Len(t, ok.Payload.Variants, 2)
	assert.Len(t, ok.Payload.Segments, 1)
	require.Len(t, ok.Payload.Tags, 1)
	require.NotNil(t, ok.Payload.Tags[0].Value)
	assert.Equal(t, "dup-tag", *ok.Payload.Tags[0].Value)
}

func TestDuplicateFlag_NotFound(t *testing.T) {
	_, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}
	res := c.DuplicateFlag(flag.DuplicateFlagParams{FlagID: 999999})
	_, isDef := res.(*flag.DuplicateFlagDefault)
	assert.True(t, isDef)
}

func TestCreateTag_ReusesExistingTagValue(t *testing.T) {
	db, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{Description: new("f1")},
	})
	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{Description: new("f2")},
	})
	tagValue := "shared-tag"
	res1 := c.CreateTag(tag.CreateTagParams{
		FlagID: 1,
		Body:   &models.CreateTagRequest{Value: &tagValue},
	})
	ok1 := res1.(*tag.CreateTagOK)
	require.NotNil(t, ok1.Payload)
	firstID := ok1.Payload.ID

	res2 := c.CreateTag(tag.CreateTagParams{
		FlagID: 2,
		Body:   &models.CreateTagRequest{Value: &tagValue},
	})
	ok2 := res2.(*tag.CreateTagOK)
	require.NotNil(t, ok2.Payload)
	assert.Equal(t, firstID, ok2.Payload.ID, "same tag value should reuse tag row")

	var tagRows int64
	require.NoError(t, db.Model(&entity.Tag{}).Where("value = ?", tagValue).Count(&tagRows).Error)
	assert.Equal(t, int64(1), tagRows)
}

func TestPutFlag_InvalidKeyReturns400(t *testing.T) {
	_, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}

	createRes := c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{Description: new("put-key-test")},
	})
	createOK := createRes.(*flag.CreateFlagOK)
	require.NotNil(t, createOK.Payload)

	badKey := " spaces are invalid "
	res := c.PutFlag(flag.PutFlagParams{
		FlagID:      createOK.Payload.ID,
		HTTPRequest: &http.Request{},
		Body: &models.PutFlagRequest{
			Key: &badKey,
		},
	})
	def, ok := res.(*flag.PutFlagDefault)
	require.True(t, ok, "expected PutFlagDefault, got %T", res)
	require.NotNil(t, def.Payload)
	require.NotNil(t, def.Payload.Message)
	assert.Contains(t, *def.Payload.Message, "invalid key")
}

func TestCommitFlagMutation_RollbackOnMutateFailure(t *testing.T) {
	db, cleanup := handlerTestDB(t)
	defer cleanup()
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