package handler

import (
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/models"
	enricherapi "github.com/openflagr/flagr/swagger_gen/restapi/operations/enricher"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jevConfigMap() map[string]any {
	return map[string]any{
		"questions": map[string]any{
			"plan_tier": map[string]any{
				"type":         JevTypeChoice,
				"instructions": "Which plan?",
				"criteria":     map[string]any{"free": "Free", "pro": "Pro"},
			},
		},
	}
}

func newEnricherTestCrud(t *testing.T) *crud {
	t.Helper()
	db := entity.NewTestDB()
	tmpDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { tmpDB.Close() })
	stub := gostub.StubFunc(&getDB, db)
	t.Cleanup(stub.Reset)

	c := &crud{}
	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{Description: new("enricher flag")},
	})
	return c
}

func TestCrudEnrichers(t *testing.T) {
	origURL := config.Config.InjectedContextJevBaseURL
	defer func() { config.Config.InjectedContextJevBaseURL = origURL }()
	config.Config.InjectedContextJevBaseURL = "http://jev.test"

	c := newEnricherTestCrud(t)

	created, ok := c.CreateEnricher(enricherapi.CreateEnricherParams{
		FlagID: 1,
		Body:   &models.CreateEnricherRequest{Namespace: new("jev"), Config: jevConfigMap()},
	}).(*enricherapi.CreateEnricherOK)
	require.True(t, ok, "expected create to succeed")
	assert.Equal(t, "jev", *created.Payload.Namespace)
	assert.Equal(t, "flag", created.Payload.Scope)
	// The write response carries the derived catalog, same shape as the flag read.
	assert.Contains(t, created.Payload.Properties, "@jev_plan_tier")
	require.NotNil(t, created.Payload.Enabled)
	assert.True(t, *created.Payload.Enabled)

	// GET flag returns the effective catalog (flag-scoped + enabled globals).
	got := c.GetFlag(flag.GetFlagParams{FlagID: 1}).(*flag.GetFlagOK).Payload
	require.Len(t, got.Enrichers, 1)
	assert.Equal(t, "jev", *got.Enrichers[0].Namespace)
	assert.Contains(t, got.Enrichers[0].Properties, "@jev_plan_tier")
	assert.NotNil(t, got.Enrichers[0].Config)

	_, ok = c.PutEnricher(enricherapi.PutEnricherParams{
		FlagID:    1,
		Namespace: "jev",
		Body:      &models.PutEnricherRequest{Config: jevConfigMap()},
	}).(*enricherapi.PutEnricherOK)
	require.True(t, ok, "expected update to succeed")

	_, ok = c.DeleteEnricher(enricherapi.DeleteEnricherParams{FlagID: 1, Namespace: "jev"}).(*enricherapi.DeleteEnricherOK)
	require.True(t, ok, "expected delete to succeed")

	// After delete, the effective catalog no longer lists it.
	got = c.GetFlag(flag.GetFlagParams{FlagID: 1}).(*flag.GetFlagOK).Payload
	assert.Empty(t, got.Enrichers)
}

func TestCrudEnrichersWithFailures(t *testing.T) {
	c := newEnricherTestCrud(t)

	require.IsType(t, &enricherapi.CreateEnricherOK{},
		c.CreateEnricher(enricherapi.CreateEnricherParams{
			FlagID: 1,
			Body:   &models.CreateEnricherRequest{Namespace: new("jev"), Config: jevConfigMap()},
		}))

	t.Run("duplicate namespace", func(t *testing.T) {
		res := c.CreateEnricher(enricherapi.CreateEnricherParams{
			FlagID: 1,
			Body:   &models.CreateEnricherRequest{Namespace: new("jev"), Config: jevConfigMap()},
		})
		assert.IsType(t, &enricherapi.CreateEnricherDefault{}, res)
	})

	t.Run("unknown namespace", func(t *testing.T) {
		res := c.CreateEnricher(enricherapi.CreateEnricherParams{
			FlagID: 1,
			Body:   &models.CreateEnricherRequest{Namespace: new("nope"), Config: jevConfigMap()},
		})
		assert.IsType(t, &enricherapi.CreateEnricherDefault{}, res)
	})

	t.Run("global namespace is not flag-scoped", func(t *testing.T) {
		res := c.CreateEnricher(enricherapi.CreateEnricherParams{
			FlagID: 1,
			Body:   &models.CreateEnricherRequest{Namespace: new(entity.EnricherNamespaceTs), Config: jevConfigMap()},
		})
		assert.IsType(t, &enricherapi.CreateEnricherDefault{}, res)
	})

	t.Run("invalid config", func(t *testing.T) {
		res := c.CreateEnricher(enricherapi.CreateEnricherParams{
			FlagID: 1,
			Body:   &models.CreateEnricherRequest{Namespace: new("jev"), Config: map[string]any{"questions": map[string]any{}}},
		})
		assert.IsType(t, &enricherapi.CreateEnricherDefault{}, res)
	})

	t.Run("put missing", func(t *testing.T) {
		res := c.PutEnricher(enricherapi.PutEnricherParams{
			FlagID: 1, Namespace: "ghost",
			Body: &models.PutEnricherRequest{Config: jevConfigMap()},
		})
		assert.IsType(t, &enricherapi.PutEnricherDefault{}, res)
	})

	t.Run("delete missing", func(t *testing.T) {
		res := c.DeleteEnricher(enricherapi.DeleteEnricherParams{FlagID: 1, Namespace: "ghost"})
		assert.IsType(t, &enricherapi.DeleteEnricherDefault{}, res)
	})
}
