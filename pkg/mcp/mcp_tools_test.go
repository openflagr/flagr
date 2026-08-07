package mcp

import (
	"testing"

	"github.com/openflagr/flagr/pkg/handler"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	result := callTool(t, cc, "health", map[string]any{})
	assert.Equal(t, "OK", result["status"])
}

func TestCreateAndGetFlag(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flag := callTool(t, cc, "create_flag", map[string]any{
		"description": "test flag",
		"key":         "test_flag",
		"enabled":     false,
	})
	assert.NotZero(t, flag["id"])
	assert.Equal(t, "test_flag", flag["key"])
	assert.Equal(t, "test flag", flag["description"])

	flagID := flag["id"].(float64)
	got := callTool(t, cc, "get_flag", map[string]any{"flag_id": int(flagID)})
	assert.Equal(t, "test_flag", got["key"])
}

func TestListFlags(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	createFlag(t, cc, "flag_a")
	createFlag(t, cc, "flag_b")

	result := callTool(t, cc, "list_flags", map[string]any{})
	flags := result["flags"].([]any)
	assert.Equal(t, 2, len(flags))
}

func TestUpdateFlag(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flagID := createFlag(t, cc, "orig")
	updated := callTool(t, cc, "update_flag", map[string]any{
		"flag_id": int(flagID), "description": "new",
	})
	assert.Equal(t, "new", updated["description"])
	assert.Equal(t, "orig", updated["key"])
}

func TestSetFlagEnabled(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flagID := createFlag(t, cc, "toggle")
	enabled := callTool(t, cc, "set_flag_enabled", map[string]any{
		"flag_id": int(flagID), "enabled": true,
	})
	assert.Equal(t, true, enabled["enabled"])
}

func TestDeleteFlag(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flagID := createFlag(t, cc, "to_delete")
	msg, isError := callToolRaw(t, cc, "delete_flag", map[string]any{"flag_id": int(flagID)})
	assert.False(t, isError)
	assert.Contains(t, msg, "deleted")

	_, isError = callToolRaw(t, cc, "get_flag", map[string]any{"flag_id": int(flagID)})
	assert.True(t, isError)
}

func TestVariantCRUD(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flagID := createFlag(t, cc, "vcr")
	v := callTool(t, cc, "create_variant", map[string]any{
		"flag_id": int(flagID), "key": "on",
	})
	assert.Equal(t, "on", v["key"])

	vs := callToolArray(t, cc, "list_variants", map[string]any{"flag_id": int(flagID)})
	assert.Len(t, vs, 1)

	variantID := v["id"].(float64)
	updated := callTool(t, cc, "update_variant", map[string]any{
		"flag_id": int(flagID), "variant_id": int(variantID),
		"key": "on_v2", "attachment": map[string]any{"color": "blue"},
	})
	assert.Equal(t, "on_v2", updated["key"])

	msg, _ := callToolRaw(t, cc, "delete_variant", map[string]any{
		"flag_id": int(flagID), "variant_id": int(variantID),
	})
	assert.Contains(t, msg, "deleted")
}

func TestSegmentCRUD(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flagID := createFlag(t, cc, "seg")
	seg := callTool(t, cc, "create_segment", map[string]any{
		"flag_id": int(flagID), "rollout_percent": 100, "description": "everyone",
	})
	assert.Equal(t, "everyone", seg["description"])

	segs := callToolArray(t, cc, "list_segments", map[string]any{"flag_id": int(flagID)})
	assert.Len(t, segs, 1)

	segID := seg["id"].(float64)
	updated := callTool(t, cc, "update_segment", map[string]any{
		"flag_id": int(flagID), "segment_id": int(segID), "rollout_percent": 50,
	})
	assert.Equal(t, float64(50), updated["rollout_percent"])

	msg, _ := callToolRaw(t, cc, "delete_segment", map[string]any{
		"flag_id": int(flagID), "segment_id": int(segID),
	})
	assert.Contains(t, msg, "deleted")
}

func TestConstraintCRUD(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flagID := createFlagWithSegment(t, cc)
	segs := callToolArray(t, cc, "list_segments", map[string]any{"flag_id": int(flagID)})
	segID := segs[0].(map[string]any)["id"].(float64)

	c := callTool(t, cc, "create_constraint", map[string]any{
		"flag_id": int(flagID), "segment_id": int(segID),
		"property": "country", "operator": "EQ", "value": "\"US\"",
	})
	assert.Equal(t, "country", c["property"])

	cs := callToolArray(t, cc, "list_constraints", map[string]any{
		"flag_id": int(flagID), "segment_id": int(segID),
	})
	assert.Len(t, cs, 1)

	constraintID := c["id"].(float64)
	msg, _ := callToolRaw(t, cc, "delete_constraint", map[string]any{
		"flag_id": int(flagID), "constraint_id": int(constraintID),
	})
	assert.Contains(t, msg, "deleted")
}

func TestDistributionCRUD(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flagID := createFlagWithSegment(t, cc)
	v1 := callTool(t, cc, "create_variant", map[string]any{
		"flag_id": int(flagID), "key": "control",
	})
	v2 := callTool(t, cc, "create_variant", map[string]any{
		"flag_id": int(flagID), "key": "treatment",
	})
	v1ID := v1["id"].(float64)
	v2ID := v2["id"].(float64)

	segs := callToolArray(t, cc, "list_segments", map[string]any{"flag_id": int(flagID)})
	segID := segs[0].(map[string]any)["id"].(float64)

	msg, isError := callToolRaw(t, cc, "set_distributions", map[string]any{
		"flag_id": int(flagID), "segment_id": int(segID),
		"distributions": []map[string]any{
			{"variant_id": int(v1ID), "variant_key": "control", "percent": 50},
			{"variant_id": int(v2ID), "variant_key": "treatment", "percent": 50},
		},
	})
	assert.False(t, isError)
	assert.Contains(t, msg, "updated")

	ds := callToolArray(t, cc, "list_distributions", map[string]any{
		"flag_id": int(flagID), "segment_id": int(segID),
	})
	assert.Len(t, ds, 2)
}

func TestTagCRUD(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flagID := createFlag(t, cc, "tagged")
	tag := callTool(t, cc, "add_tag", map[string]any{
		"flag_id": int(flagID), "value": "release-2024",
	})
	assert.Equal(t, "release-2024", tag["value"])

	tags := callToolArray(t, cc, "list_tags", map[string]any{"flag_id": int(flagID)})
	assert.Len(t, tags, 1)

	allTags := callToolArray(t, cc, "list_all_tags", map[string]any{})
	assert.GreaterOrEqual(t, len(allTags), 1)

	tagID := tag["id"].(float64)
	msg, _ := callToolRaw(t, cc, "delete_tag", map[string]any{
		"flag_id": int(flagID), "tag_id": int(tagID),
	})
	assert.Contains(t, msg, "removed")
}

func TestDuplicateFlag(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flagID := createFlag(t, cc, "original")
	v := callTool(t, cc, "create_variant", map[string]any{
		"flag_id": int(flagID), "key": "on",
	})
	vID := v["id"].(float64)

	segs := callToolArray(t, cc, "list_segments", map[string]any{"flag_id": int(flagID)})
	if len(segs) == 0 {
		callTool(t, cc, "create_segment", map[string]any{
			"flag_id": int(flagID), "rollout_percent": 100,
		})
		segs = callToolArray(t, cc, "list_segments", map[string]any{"flag_id": int(flagID)})
	}
	segID := segs[0].(map[string]any)["id"].(float64)

	callTool(t, cc, "set_distributions", map[string]any{
		"flag_id": int(flagID), "segment_id": int(segID),
		"distributions": []map[string]any{
			{"variant_id": int(vID), "variant_key": "on", "percent": 100},
		},
	})

	callTool(t, cc, "add_tag", map[string]any{
		"flag_id": int(flagID), "value": "release",
	})

	dup := callTool(t, cc, "duplicate_flag", map[string]any{"flag_id": int(flagID)})
	assert.NotZero(t, dup["id"])
	assert.NotEqual(t, float64(flagID), dup["id"])
	dupID := dup["id"].(float64)
	assert.Contains(t, dup["key"].(string), "copy")

	dupVariants := callToolArray(t, cc, "list_variants", map[string]any{"flag_id": int(dupID)})
	assert.Len(t, dupVariants, 1)

	dupTags := callToolArray(t, cc, "list_tags", map[string]any{"flag_id": int(dupID)})
	assert.Len(t, dupTags, 1)
}

func TestEvaluateFlag(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	flagID := createFlag(t, cc, "eval_flag")
	callTool(t, cc, "set_flag_enabled", map[string]any{"flag_id": int(flagID), "enabled": true})
	callTool(t, cc, "create_variant", map[string]any{"flag_id": int(flagID), "key": "on"})

	// Stub handler.EvalFlag since the eval cache singleton can't be seeded from outside.
	stub := gostub.StubFunc(&handler.EvalFlag, &models.EvalResult{
		FlagID:     int64(flagID),
		FlagKey:    "eval_flag",
		VariantID:  1,
		VariantKey: "on",
	})
	defer stub.Reset()

	result := callTool(t, cc, "evaluate_flag", map[string]any{
		"flag_id":   int(flagID),
		"entity_id": "user-1",
	})
	assert.Equal(t, "on", result["variant_key"])
	assert.Equal(t, "eval_flag", result["flag_key"])
}

func TestGetFlagNotFound(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	_, isError := callToolRaw(t, cc, "get_flag", map[string]any{"flag_id": 99999})
	assert.True(t, isError)
}

func TestListFlagsWithFilter(t *testing.T) {
	cc, cleanup := setupTest(t)
	defer cleanup()

	createFlag(t, cc, "enabled_flag")
	createFlag(t, cc, "disabled_flag")

	callTool(t, cc, "set_flag_enabled", map[string]any{"flag_id": 1, "enabled": true})

	result := callTool(t, cc, "list_flags", map[string]any{"enabled": true})
	flags := result["flags"].([]any)
	assert.Equal(t, 1, len(flags))
}
