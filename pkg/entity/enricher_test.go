package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleConfig struct {
	Value string `json:"value"`
}

func TestEnricherSetConfigRoundTrip(t *testing.T) {
	t.Parallel()

	e := &Enricher{}
	require.NoError(t, e.SetConfig(EnricherNamespaceJev, &sampleConfig{Value: "x"}))
	assert.Equal(t, EnricherNamespaceJev, e.Namespace)
	assert.JSONEq(t, `{"value":"x"}`, e.ConfigJSON)
	assert.NoError(t, e.Validate())

	got := &sampleConfig{}
	require.NoError(t, e.DecodeConfig(got))
	assert.Equal(t, "x", got.Value)
}

func TestEnricherValidateStorageInvariants(t *testing.T) {
	t.Parallel()

	assert.Error(t, (&Enricher{}).Validate(), "namespace is required")
	assert.Error(t, (&Enricher{Namespace: "  "}).Validate(), "blank namespace is required")
	assert.NoError(t, (&Enricher{Namespace: "anything", ConfigJSON: `{"a":1}`}).Validate())

	// Namespace scope and config schema are the handler registry's concern; the
	// entity only rejects malformed JSON.
	assert.Error(t, (&Enricher{Namespace: EnricherNamespaceJev, ConfigJSON: "not json"}).Validate())
}

func TestEnricherDecodeConfig(t *testing.T) {
	t.Parallel()

	e := &Enricher{Namespace: EnricherNamespaceJev, ConfigJSON: `{"value":"y"}`}
	got := &sampleConfig{}
	require.NoError(t, e.DecodeConfig(got))
	assert.Equal(t, "y", got.Value)

	assert.Error(t, (&Enricher{}).DecodeConfig(got), "empty config must error")
	assert.Error(t, (&Enricher{ConfigJSON: "{"}).DecodeConfig(got), "malformed JSON must error")
}

func TestEnricherSetConfigRejectsEmptyNamespace(t *testing.T) {
	t.Parallel()

	assert.Error(t, (&Enricher{}).SetConfig("", &sampleConfig{Value: "x"}))
}
