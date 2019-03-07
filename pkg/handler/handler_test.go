package handler

import (
	"testing"

	"github.com/rexmont/flagr/pkg/entity"
	"github.com/rexmont/flagr/swagger_gen/restapi/operations"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	defer gostub.StubFunc(&getDB, entity.NewTestDB()).Reset()
	assert.NotPanics(t, func() {
		Setup(&operations.FlagrAPI{})
	})
}
