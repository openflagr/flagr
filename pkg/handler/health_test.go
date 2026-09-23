package handler

import (
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/health"
	"github.com/stretchr/testify/assert"
)

func TestHealthReportsSnapshotsDefaultLimit(t *testing.T) {
	api := &operations.FlagrAPI{}
	setupHealth(api)

	original := config.Config.SnapshotsDefaultLimit
	defer func() { config.Config.SnapshotsDefaultLimit = original }()

	config.Config.SnapshotsDefaultLimit = 50
	res := api.HealthGetHealthHandler.Handle(health.GetHealthParams{})
	payload := res.(*health.GetHealthOK).Payload
	assert.Equal(t, "OK", payload.Status)
	assert.Equal(t, int64(50), payload.SnapshotsDefaultLimit)

	// Default (0) means the endpoint serves the full history; UI does not paginate.
	config.Config.SnapshotsDefaultLimit = 0
	res = api.HealthGetHealthHandler.Handle(health.GetHealthParams{})
	assert.Equal(t, int64(0), res.(*health.GetHealthOK).Payload.SnapshotsDefaultLimit)
}
