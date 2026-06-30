package handler

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/constraint"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/tag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/variant"
	"github.com/prashantv/gostub"
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

func TestDuplicateFlag_DefaultDescriptionWhenSourceEmpty(t *testing.T) {
	_, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}

	createRes := c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{Key: "empty_desc_src"},
	})
	createOK := createRes.(*flag.CreateFlagOK)
	require.NotNil(t, createOK.Payload)

	dupRes := c.DuplicateFlag(flag.DuplicateFlagParams{FlagID: createOK.Payload.ID})
	ok := dupRes.(*flag.DuplicateFlagOK)
	require.NotNil(t, ok.Payload.Description)
	assert.Equal(t, "(cloned)", *ok.Payload.Description)
}

func TestDuplicateFlag_CopiesEnabledState(t *testing.T) {
	_, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}

	createRes := c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: new("off flag"),
			Key:         "dup_enabled_src",
		},
	})
	createOK := createRes.(*flag.CreateFlagOK)
	require.NotNil(t, createOK.Payload)
	flagID := createOK.Payload.ID

	c.SetFlagEnabledState(flag.SetFlagEnabledParams{
		FlagID: flagID,
		Body:   &models.SetFlagEnabledRequest{Enabled: new(false)},
	})

	dupRes := c.DuplicateFlag(flag.DuplicateFlagParams{FlagID: flagID})
	ok := dupRes.(*flag.DuplicateFlagOK)
	require.NotNil(t, ok.Payload.Enabled)
	assert.False(t, *ok.Payload.Enabled)
}

func TestDuplicateFlag_OptionalKeyAndDescription(t *testing.T) {
	_, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}

	createRes := c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: new("custom dup source"),
			Key:         "dup_body_src",
		},
	})
	createOK := createRes.(*flag.CreateFlagOK)
	require.NotNil(t, createOK.Payload)

	customKey := "my_clone_key"
	customDesc := "my clone description"
	dupRes := c.DuplicateFlag(flag.DuplicateFlagParams{
		FlagID: createOK.Payload.ID,
		Body: &models.DuplicateFlagRequest{
			Key:         customKey,
			Description: customDesc,
		},
	})
	ok := dupRes.(*flag.DuplicateFlagOK)
	require.NotNil(t, ok.Payload)
	assert.Equal(t, customKey, ok.Payload.Key)
	require.NotNil(t, ok.Payload.Description)
	assert.Equal(t, customDesc, *ok.Payload.Description)
}

func TestDuplicateFlag_InvalidOptionalKeyReturns400(t *testing.T) {
	_, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}

	createRes := c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{Description: new("bad key dup"), Key: "dup_bad_key_src"},
	})
	createOK := createRes.(*flag.CreateFlagOK)
	require.NotNil(t, createOK.Payload)

	res := c.DuplicateFlag(flag.DuplicateFlagParams{
		FlagID: createOK.Payload.ID,
		Body:   &models.DuplicateFlagRequest{Key: " spaces invalid "},
	})
	def, ok := res.(*flag.DuplicateFlagDefault)
	require.True(t, ok, "expected DuplicateFlagDefault, got %T", res)
	require.NotNil(t, def.Payload)
	require.NotNil(t, def.Payload.Message)
	assert.Contains(t, *def.Payload.Message, "cannot duplicate flag")
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

func TestDuplicateFlag_DeletedSourceReturns404(t *testing.T) {
	_, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}
	req := &http.Request{}

	createRes := c.CreateFlag(flag.CreateFlagParams{
		HTTPRequest: req,
		Body:        &models.CreateFlagRequest{Key: "dup_del_src", Description: new("src")},
	})
	createOK := createRes.(*flag.CreateFlagOK)
	require.NotNil(t, createOK.Payload)
	flagID := createOK.Payload.ID

	delRes := c.DeleteFlag(flag.DeleteFlagParams{FlagID: flagID, HTTPRequest: req})
	_, isDelOK := delRes.(*flag.DeleteFlagOK)
	require.True(t, isDelOK, "delete failed: %T", delRes)

	dupRes := c.DuplicateFlag(flag.DuplicateFlagParams{FlagID: flagID, HTTPRequest: req})
	_, ok := dupRes.(*flag.DuplicateFlagDefault)
	require.True(t, ok, "expected 404 default, got %T", dupRes)
}

func TestDuplicateFlag_DuplicateKeyReturns400(t *testing.T) {
	_, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}
	req := &http.Request{}

	takenKey := "dup_taken_key"
	c.CreateFlag(flag.CreateFlagParams{
		HTTPRequest: req,
		Body:        &models.CreateFlagRequest{Key: takenKey, Description: new("existing")},
	})
	createRes := c.CreateFlag(flag.CreateFlagParams{
		HTTPRequest: req,
		Body:        &models.CreateFlagRequest{Key: "dup_src_for_key", Description: new("src")},
	})
	sourceID := createRes.(*flag.CreateFlagOK).Payload.ID

	dupRes := c.DuplicateFlag(flag.DuplicateFlagParams{
		FlagID:      sourceID,
		HTTPRequest: req,
		Body:        &models.DuplicateFlagRequest{Key: takenKey},
	})
	def, ok := dupRes.(*flag.DuplicateFlagDefault)
	require.True(t, ok, "expected error response, got %T", dupRes)
	require.NotNil(t, def.Payload)
	require.NotNil(t, def.Payload.Message)
	assert.Contains(t, *def.Payload.Message, "key")
}

func TestCommitFlagMutation_RollbackOnWriteFlagSnapshotFailure(t *testing.T) {
	db, cleanup := handlerTestDB(t)
	defer cleanup()

	createRes := (&crud{}).CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{Description: new("snap-rollback"), Key: "snap_rb_src"},
	})
	createOK := createRes.(*flag.CreateFlagOK)
	require.NotNil(t, createOK.Payload)
	flagID := createOK.Payload.ID

	var before int64
	require.NoError(t, db.Model(&entity.FlagSnapshot{}).Where("flag_id = ?", flagID).Count(&before).Error)

	stub := gostub.Stub(&writeFlagSnapshotTx, func(tx *gorm.DB, flagID uint, updatedBy string) (entity.FlagSnapshotCommitMeta, error) {
		return entity.FlagSnapshotCommitMeta{}, fmt.Errorf("snapshot write failed")
	})
	defer stub.Reset()

	err := commitFlagMutation(uint(flagID), "tester", notification.OperationUpdate, notification.ComponentFlag, func(tx *gorm.DB) (uint, MutationNotify, error) {
		return uint(flagID), MutationNotify{ComponentID: uint(flagID), ComponentKey: "snap_rb_src"}, nil
	})
	assert.Error(t, err)

	var after int64
	require.NoError(t, db.Model(&entity.FlagSnapshot{}).Where("flag_id = ?", flagID).Count(&after).Error)
	assert.Equal(t, before, after)
}

func TestDuplicateFlag_CopiesConstraintsAndDistributions(t *testing.T) {
	_, cleanup := handlerTestDB(t)
	defer cleanup()
	c := &crud{}
	req := &http.Request{}

	createRes := c.CreateFlag(flag.CreateFlagParams{
		HTTPRequest: req,
		Body:        &models.CreateFlagRequest{Key: "dup_graph_src", Description: new("graph")},
	})
	createOK := createRes.(*flag.CreateFlagOK)
	require.NotNil(t, createOK.Payload)
	flagID := createOK.Payload.ID
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: flagID,
		Body:   &models.CreateVariantRequest{Key: new("on")},
	})
	variantRes := c.FindVariants(variant.FindVariantsParams{FlagID: flagID})
	variantOK := variantRes.(*variant.FindVariantsOK)
	require.NotEmpty(t, variantOK.Payload)
	variantID := variantOK.Payload[0].ID

	rollout := int64(100)
	segRes := c.CreateSegment(segment.CreateSegmentParams{
		FlagID: flagID,
		Body: &models.CreateSegmentRequest{
			Description:    new("seg-with-constraint"),
			RolloutPercent: &rollout,
		},
	})
	segOK := segRes.(*segment.CreateSegmentOK)
	require.NotNil(t, segOK.Payload)
	segID := segOK.Payload.ID

	c.CreateConstraint(constraint.CreateConstraintParams{
		FlagID:    flagID,
		SegmentID: segID,
		Body: &models.CreateConstraintRequest{
			Property: new("country"),
			Operator: new("EQ"),
			Value:    new(`"US"`),
		},
	})

	pct := int64(100)
	c.PutDistributions(distribution.PutDistributionsParams{
		FlagID:    flagID,
		SegmentID: segID,
		Body: &models.PutDistributionsRequest{
			Distributions: []*models.Distribution{{
				Percent:    &pct,
				VariantID:  &variantID,
				VariantKey: new("on"),
			}},
		},
	})

	dupRes := c.DuplicateFlag(flag.DuplicateFlagParams{FlagID: flagID, HTTPRequest: req})
	ok := dupRes.(*flag.DuplicateFlagOK)
	require.NotNil(t, ok.Payload)
	require.Len(t, ok.Payload.Segments, 1)
	require.NotEmpty(t, ok.Payload.Segments[0].Constraints)
	require.NotNil(t, ok.Payload.Segments[0].Constraints[0].Property)
	assert.Equal(t, "country", *ok.Payload.Segments[0].Constraints[0].Property)
	assert.NotEmpty(t, ok.Payload.Segments[0].Distributions)
}
