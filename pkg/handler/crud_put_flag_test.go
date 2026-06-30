package handler

import (
	"net/http"
	"testing"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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