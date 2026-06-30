package handler

import (
	"net/http"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/variant"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuplicateFlag(t *testing.T) {
	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
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
}

func TestDuplicateFlag_NotFound(t *testing.T) {
	db := entity.NewTestDB()
	defer gostub.StubFunc(&getDB, db).Reset()
	c := &crud{}
	res := c.DuplicateFlag(flag.DuplicateFlagParams{FlagID: 999999})
	_, isDef := res.(*flag.DuplicateFlagDefault)
	assert.True(t, isDef)
}