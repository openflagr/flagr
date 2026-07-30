package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/constraint"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/health"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/tag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/variant"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readOnlyFixtureJSON is a hand-authored JSON flag source (no explicit IDs)
// exercising auto-ID assignment: flag_a → ID 1, flag_b → ID 2.
const readOnlyFixtureJSON = `{
  "Flags": [
    {
      "Key": "flag_a",
      "Description": "alpha flag",
      "Enabled": true,
      "Tags": [{"Value": "team-a"}, {"Value": "shared"}],
      "Variants": [{"Key": "on"}, {"Key": "off"}],
      "Segments": [
        {
          "Description": "everyone",
          "RolloutPercent": 100,
          "Constraints": [{"Property": "env", "Operator": "EQ", "Value": "\"prod\""}],
          "Distributions": [{"VariantKey": "on", "Percent": 100}]
        }
      ]
    },
    {
      "Key": "flag_b",
      "Description": "beta flag",
      "Enabled": false,
      "Tags": [{"Value": "team-b"}, {"Value": "shared"}]
    }
  ]
}`

// seedReadOnlyCache points the EvalCache at a json_file source with the given
// content and reloads it. The returned reset restores config and cache state.
// Tests using it mutate global state (config.Config, the EvalCache singleton),
// so they must not run with t.Parallel().
func seedReadOnlyCache(t *testing.T, jsonContent string) (ec *EvalCache, reset func()) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "flags.json")
	require.NoError(t, os.WriteFile(path, []byte(jsonContent), 0o600))

	old := config.Config
	config.Config.DBDriver = "json_file"
	config.Config.EvalOnlyMode = true
	config.Config.DBConnectionStr = path

	ec = GetEvalCache()
	ec.fetcher = nil
	ec.lastSnapshotMaxID = 0
	require.NoError(t, ec.reloadMapCache())

	return ec, func() {
		ec.fetcher = nil
		ec.lastSnapshotMaxID = 0
		config.Config = old
	}
}

// responderStatusCode renders the responder and returns the HTTP status code,
// since go-swagger Default responders don't expose their status code directly.
func responderStatusCode(t *testing.T, res middleware.Responder) int {
	t.Helper()
	rw := httptest.NewRecorder()
	res.WriteResponse(rw, runtime.JSONProducer())
	return rw.Code
}

func TestReadOnlyFindFlags(t *testing.T) {
	_, reset := seedReadOnlyCache(t, readOnlyFixtureJSON)
	defer reset()
	c := NewReadOnlyCRUD()

	t.Run("returns all flags sorted by ID", func(t *testing.T) {
		res := c.FindFlags(flag.FindFlagsParams{})
		payload := res.(*flag.FindFlagsOK).Payload
		require.Len(t, payload, 2)
		assert.Equal(t, "flag_a", payload[0].Key)
		assert.Equal(t, "flag_b", payload[1].Key)
		// Full flag payloads, like preload=true in DB mode.
		assert.NotEmpty(t, payload[0].Segments)
		assert.NotEmpty(t, payload[0].Variants)
		assert.NotEmpty(t, payload[0].Tags)
	})

	t.Run("filters by enabled", func(t *testing.T) {
		res := c.FindFlags(flag.FindFlagsParams{Enabled: new(true)})
		payload := res.(*flag.FindFlagsOK).Payload
		require.Len(t, payload, 1)
		assert.Equal(t, "flag_a", payload[0].Key)
	})

	t.Run("filters by key", func(t *testing.T) {
		res := c.FindFlags(flag.FindFlagsParams{Key: new("flag_b")})
		payload := res.(*flag.FindFlagsOK).Payload
		require.Len(t, payload, 1)
		assert.Equal(t, "flag_b", payload[0].Key)
	})

	t.Run("filters by description_like case-insensitively", func(t *testing.T) {
		res := c.FindFlags(flag.FindFlagsParams{DescriptionLike: new("BETA")})
		payload := res.(*flag.FindFlagsOK).Payload
		require.Len(t, payload, 1)
		assert.Equal(t, "flag_b", payload[0].Key)
	})

	t.Run("filters by tags with ANY semantics", func(t *testing.T) {
		res := c.FindFlags(flag.FindFlagsParams{Tags: new("team-a,no-such-tag")})
		payload := res.(*flag.FindFlagsOK).Payload
		require.Len(t, payload, 1)
		assert.Equal(t, "flag_a", payload[0].Key)

		res = c.FindFlags(flag.FindFlagsParams{Tags: new("shared")})
		assert.Len(t, res.(*flag.FindFlagsOK).Payload, 2)
	})

	t.Run("applies offset and limit", func(t *testing.T) {
		res := c.FindFlags(flag.FindFlagsParams{Offset: new(int64(1)), Limit: new(int64(10))})
		payload := res.(*flag.FindFlagsOK).Payload
		require.Len(t, payload, 1)
		assert.Equal(t, "flag_b", payload[0].Key)

		res = c.FindFlags(flag.FindFlagsParams{Offset: new(int64(5))})
		assert.Empty(t, res.(*flag.FindFlagsOK).Payload)
	})

	t.Run("deleted=true returns empty", func(t *testing.T) {
		res := c.FindFlags(flag.FindFlagsParams{Deleted: new(true)})
		assert.Empty(t, res.(*flag.FindFlagsOK).Payload)
	})
}

func TestReadOnlyGetFlag(t *testing.T) {
	_, reset := seedReadOnlyCache(t, readOnlyFixtureJSON)
	defer reset()
	c := NewReadOnlyCRUD()

	t.Run("returns the cached flag", func(t *testing.T) {
		res := c.GetFlag(flag.GetFlagParams{FlagID: 1})
		payload := res.(*flag.GetFlagOK).Payload
		assert.Equal(t, "flag_a", payload.Key)
		require.Len(t, payload.Segments, 1)
		require.Len(t, payload.Segments[0].Distributions, 1)
		assert.Equal(t, "on", *payload.Segments[0].Distributions[0].VariantKey)
	})

	t.Run("404 for unknown flag", func(t *testing.T) {
		res := c.GetFlag(flag.GetFlagParams{FlagID: 999})
		assert.Equal(t, http.StatusNotFound, responderStatusCode(t, res))
	})
}

func TestReadOnlyFlagMetaEndpoints(t *testing.T) {
	ec, reset := seedReadOnlyCache(t, readOnlyFixtureJSON)
	defer reset()
	c := NewReadOnlyCRUD()

	t.Run("snapshots are empty", func(t *testing.T) {
		res := c.GetFlagSnapshots(flag.GetFlagSnapshotsParams{FlagID: 1})
		assert.Empty(t, res.(*flag.GetFlagSnapshotsOK).Payload)
	})

	t.Run("entity types are empty", func(t *testing.T) {
		res := c.GetFlagEntityTypes(flag.GetFlagEntityTypesParams{})
		assert.Empty(t, res.(*flag.GetFlagEntityTypesOK).Payload)
	})

	t.Run("snapshot max_id serves the content fingerprint", func(t *testing.T) {
		res := c.GetFlagSnapshotMaxID(flag.GetFlagSnapshotMaxIDParams{})
		payload := res.(*flag.GetFlagSnapshotMaxIDOK).Payload
		assert.NotZero(t, payload.MaxID)
		assert.Equal(t, ec.ContentFingerprint(), payload.MaxID)
	})
}

func TestReadOnlyTags(t *testing.T) {
	_, reset := seedReadOnlyCache(t, readOnlyFixtureJSON)
	defer reset()
	c := NewReadOnlyCRUD()

	t.Run("FindTags returns the flag's tags", func(t *testing.T) {
		res := c.FindTags(tag.FindTagsParams{FlagID: 1})
		payload := res.(*tag.FindTagsOK).Payload
		values := make([]string, len(payload))
		for i, tg := range payload {
			values[i] = *tg.Value
		}
		assert.ElementsMatch(t, []string{"team-a", "shared"}, values)
	})

	t.Run("FindTags for unknown flag is empty", func(t *testing.T) {
		res := c.FindTags(tag.FindTagsParams{FlagID: 999})
		assert.Empty(t, res.(*tag.FindTagsOK).Payload)
	})

	t.Run("FindAllTags dedupes by value", func(t *testing.T) {
		res := c.FindAllTags(tag.FindAllTagsParams{})
		payload := res.(*tag.FindAllTagsOK).Payload
		values := make([]string, len(payload))
		for i, tg := range payload {
			values[i] = *tg.Value
		}
		assert.ElementsMatch(t, []string{"team-a", "team-b", "shared"}, values)
	})

	t.Run("FindAllTags filters by value_like", func(t *testing.T) {
		res := c.FindAllTags(tag.FindAllTagsParams{ValueLike: new("team")})
		assert.Len(t, res.(*tag.FindAllTagsOK).Payload, 2)
	})
}

func TestReadOnlySubResources(t *testing.T) {
	_, reset := seedReadOnlyCache(t, readOnlyFixtureJSON)
	defer reset()
	c := NewReadOnlyCRUD()

	t.Run("FindSegments", func(t *testing.T) {
		res := c.FindSegments(segment.FindSegmentsParams{FlagID: 1})
		payload := res.(*segment.FindSegmentsOK).Payload
		require.Len(t, payload, 1)
		assert.Equal(t, "everyone", *payload[0].Description)
	})

	t.Run("FindVariants", func(t *testing.T) {
		res := c.FindVariants(variant.FindVariantsParams{FlagID: 1})
		assert.Len(t, res.(*variant.FindVariantsOK).Payload, 2)
	})

	t.Run("FindConstraints", func(t *testing.T) {
		res := c.FindConstraints(constraint.FindConstraintsParams{FlagID: 1, SegmentID: 1})
		payload := res.(*constraint.FindConstraintsOK).Payload
		require.Len(t, payload, 1)
		assert.Equal(t, "env", *payload[0].Property)
	})

	t.Run("FindDistributions", func(t *testing.T) {
		res := c.FindDistributions(distribution.FindDistributionsParams{FlagID: 1, SegmentID: 1})
		payload := res.(*distribution.FindDistributionsOK).Payload
		require.Len(t, payload, 1)
		assert.Equal(t, "on", *payload[0].VariantKey)
	})

	t.Run("unknown flag or segment yields empty results", func(t *testing.T) {
		assert.Empty(t, c.FindSegments(segment.FindSegmentsParams{FlagID: 999}).(*segment.FindSegmentsOK).Payload)
		assert.Empty(t, c.FindVariants(variant.FindVariantsParams{FlagID: 999}).(*variant.FindVariantsOK).Payload)
		assert.Empty(t, c.FindConstraints(constraint.FindConstraintsParams{FlagID: 1, SegmentID: 999}).(*constraint.FindConstraintsOK).Payload)
		assert.Empty(t, c.FindDistributions(distribution.FindDistributionsParams{FlagID: 999, SegmentID: 1}).(*distribution.FindDistributionsOK).Payload)
	})
}

func TestReadOnlyWritesForbidden(t *testing.T) {
	_, reset := seedReadOnlyCache(t, readOnlyFixtureJSON)
	defer reset()
	c := NewReadOnlyCRUD()

	responders := map[string]middleware.Responder{
		"CreateFlag":         c.CreateFlag(flag.CreateFlagParams{}),
		"DuplicateFlag":      c.DuplicateFlag(flag.DuplicateFlagParams{}),
		"PutFlag":            c.PutFlag(flag.PutFlagParams{}),
		"DeleteFlag":         c.DeleteFlag(flag.DeleteFlagParams{}),
		"RestoreFlag":        c.RestoreFlag(flag.RestoreFlagParams{}),
		"SetFlagEnabled":     c.SetFlagEnabledState(flag.SetFlagEnabledParams{}),
		"CreateTag":          c.CreateTag(tag.CreateTagParams{}),
		"DeleteTag":          c.DeleteTag(tag.DeleteTagParams{}),
		"CreateSegment":      c.CreateSegment(segment.CreateSegmentParams{}),
		"PutSegment":         c.PutSegment(segment.PutSegmentParams{}),
		"DeleteSegment":      c.DeleteSegment(segment.DeleteSegmentParams{}),
		"PutSegmentsReorder": c.PutSegmentsReorder(segment.PutSegmentsReorderParams{}),
		"CreateConstraint":   c.CreateConstraint(constraint.CreateConstraintParams{}),
		"PutConstraint":      c.PutConstraint(constraint.PutConstraintParams{}),
		"DeleteConstraint":   c.DeleteConstraint(constraint.DeleteConstraintParams{}),
		"PutDistributions":   c.PutDistributions(distribution.PutDistributionsParams{}),
		"CreateVariant":      c.CreateVariant(variant.CreateVariantParams{}),
		"PutVariant":         c.PutVariant(variant.PutVariantParams{}),
		"DeleteVariant":      c.DeleteVariant(variant.DeleteVariantParams{}),
	}

	for name, res := range responders {
		t.Run(name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			res.WriteResponse(rw, runtime.JSONProducer())
			assert.Equal(t, http.StatusForbidden, rw.Code)
			assert.Contains(t, rw.Body.String(), "read-only")
		})
	}
}

func TestReadOnlyFingerprint(t *testing.T) {
	ec, reset := seedReadOnlyCache(t, readOnlyFixtureJSON)
	defer reset()

	first := ec.ContentFingerprint()
	assert.NotZero(t, first)

	// Same content → same fingerprint.
	require.NoError(t, ec.reloadMapCache())
	assert.Equal(t, first, ec.ContentFingerprint())

	// Changed content → changed fingerprint.
	changed, resetChanged := seedReadOnlyCache(t,
		`{"Flags": [{"Key": "flag_a", "Description": "alpha flag v2", "Enabled": true}]}`)
	defer resetChanged()
	assert.NotEqual(t, first, changed.ContentFingerprint())
}

func TestHealthReportsEvalOnlyMode(t *testing.T) {
	old := config.Config.EvalOnlyMode
	defer func() { config.Config.EvalOnlyMode = old }()

	api := &operations.FlagrAPI{}
	setupHealth(api)

	config.Config.EvalOnlyMode = true
	res := api.HealthGetHealthHandler.Handle(health.GetHealthParams{})
	payload := res.(*health.GetHealthOK).Payload
	assert.Equal(t, "OK", payload.Status)
	assert.True(t, payload.EvalOnlyMode)

	config.Config.EvalOnlyMode = false
	res = api.HealthGetHealthHandler.Handle(health.GetHealthParams{})
	assert.False(t, res.(*health.GetHealthOK).Payload.EvalOnlyMode)
}

func TestSetupEvalOnlyRegistersReadOnlyCRUD(t *testing.T) {
	_, reset := seedReadOnlyCache(t, readOnlyFixtureJSON)
	defer reset()

	api := &operations.FlagrAPI{}
	assert.NotPanics(t, func() { Setup(api) })

	assert.NotNil(t, api.FlagFindFlagsHandler)
	assert.NotNil(t, api.FlagGetFlagHandler)
	assert.NotNil(t, api.TagFindAllTagsHandler)
	assert.NotNil(t, api.FlagCreateFlagHandler)

	// Registered write handlers must deny, not mutate.
	res := api.FlagCreateFlagHandler.Handle(flag.CreateFlagParams{})
	assert.Equal(t, http.StatusForbidden, responderStatusCode(t, res))
}
