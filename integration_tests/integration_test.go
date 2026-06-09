//go:build integration

// Package flagr_integration provides HTTP-based integration tests for Flagr.
//
// Execution modes:
//   - Local:   go test -tags=integration ./integration_tests/
//              (auto-starts server with SQLite :memory:)
//   - BYO:     FLAGR_SERVER_URL=http://host:18000 go test -tags=integration ./integration_tests/
//   - Docker:  cd integration_tests && make test
//              (builds binary, runs against all 6 compose instances)
package flagr_integration

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestIntegration_Health(t *testing.T) {
	var result map[string]any
	getJSON(t, "/api/v1/health", &result)
	if result["status"] == nil {
		t.Fatal("health response missing status")
	}
}

func TestIntegration_FlagCRUD(t *testing.T) {
	// Create a flag
	key := fmt.Sprintf("crud_flag_%d", time.Now().UnixNano())
	var created map[string]any
	postJSON(t, "/api/v1/flags", map[string]any{
		"key":         key,
		"description": "crud test flag",
	}, &created)
	if created["id"] == nil || created["id"].(float64) == 0 {
		t.Fatal("expected non-zero id")
	}
	flagID := int64(created["id"].(float64))

	// Get flag
	var fetched map[string]any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &fetched)
	if fetched["key"] != key {
		t.Fatalf("expected key %s, got %v", key, fetched["key"])
	}

	// Put flag — description, key, dataRecordsEnabled, entityType
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), map[string]any{
		"description":        "updated description",
		"key":                key,
		"dataRecordsEnabled": true,
		"entityType":         "test_entity",
	}, &fetched)
	if dr, ok := fetched["dataRecordsEnabled"].(bool); !ok || !dr {
		t.Fatalf("expected dataRecordsEnabled=true, got %v", fetched["dataRecordsEnabled"])
	}

	// Set enabled (PUT)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/enabled", flagID), map[string]any{
		"enabled": true,
	}, &fetched)

	// Get flag entity types (should include test_entity)
	var types []any
	getJSON(t, "/api/v1/flags/entity_types", &types)

	// Query flags by description
	var byDesc []any
	getJSON(t, fmt.Sprintf("/api/v1/flags?description=%s", url.QueryEscape("updated description")), &byDesc)
	if len(byDesc) == 0 {
		t.Fatal("expected at least one flag matching description")
	}

	// Query flags by key
	var byKey []any
	getJSON(t, fmt.Sprintf("/api/v1/flags?key=%s", key), &byKey)
	if len(byKey) != 1 {
		t.Fatalf("expected exactly 1 flag matching key, got %d", len(byKey))
	}

	// Query flags with limit/offset
	var limited []any
	getJSON(t, "/api/v1/flags?limit=1&offset=0", &limited)
	if len(limited) > 1 {
		t.Fatalf("expected at most 1 flag with limit=1, got %d", len(limited))
	}

	// Find flags with preload
	var flags []any
	getJSON(t, "/api/v1/flags?preload=true&limit=1", &flags)

	// Delete flag
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d", flagID))

	// Restore flag (PUT, not POST)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/restore", flagID), nil, nil)

	// Get snapshot
	var snapshots []any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/snapshots", flagID), &snapshots)
}

func TestIntegration_SegmentCRUD(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	flagID := seedFlagIDs[0]

	// Create segment
	var seg map[string]any
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments", flagID), map[string]any{
		"description":    "test segment",
		"rolloutPercent": 50,
	}, &seg)
	segID := int64(seg["id"].(float64))

	// Put segment — update description and rollout
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d", flagID, segID), map[string]any{
		"description":    "updated segment",
		"rolloutPercent": 100,
	}, nil)

	// Reorder segments (replace active segments with just the new one)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/reorder", flagID), map[string]any{
		"segmentIDs": []int64{segID},
	}, nil)

	// Verify rank via single flag get
	var flagObj map[string]any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flagObj)
	segs, _ := flagObj["segments"].([]any)
	found := false
	for _, s := range segs {
		sm := s.(map[string]any)
		if int64(sm["id"].(float64)) == segID {
			if rank, ok := sm["rank"].(float64); !ok || rank != 0 {
				t.Fatalf("expected rank 0 after reorder, got %v", sm["rank"])
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("reordered segment not found in flag response")
	}

	// Delete segment
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d", flagID, segID))

	// Verify deletion
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flagObj)
	segs, _ = flagObj["segments"].([]any)
	for _, s := range segs {
		if int64(s.(map[string]any)["id"].(float64)) == segID {
			t.Fatal("segment still present after deletion")
		}
	}
}

func TestIntegration_ConstraintCRUD(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	flagID := seedFlagIDs[0]

	// Create a segment first
	var seg map[string]any
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments", flagID), map[string]any{
		"description":    "constraint test segment",
		"rolloutPercent": 100,
	}, &seg)
	segID := int64(seg["id"].(float64))

	// Create constraint
	var constraint map[string]any
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints", flagID, segID), map[string]any{
		"property": "test_prop",
		"operator": "EQ",
		"value":    `"test_value"`,
	}, &constraint)
	constraintID := int64(constraint["id"].(float64))

	// Update constraint
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints/%d", flagID, segID, constraintID), map[string]any{
		"property": "test_prop",
		"operator": "NEQ",
		"value":    `"other_value"`,
	}, &constraint)

	// List constraints
	var list []any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints", flagID, segID), &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(list))
	}
}

func TestIntegration_VariantCRUD(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	flagID := seedFlagIDs[0]

	// Create variant
	var v map[string]any
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants", flagID), map[string]any{
		"key": "test_variant",
	}, &v)
	variantID := int64(v["id"].(float64))

	// Update variant key
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants/%d", flagID, variantID), map[string]any{
		"key": "test_variant_updated",
	}, &v)

	// Create variant with attachment
	var v2 map[string]any
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants", flagID), map[string]any{
		"key": "variant_with_attachment",
		"attachment": map[string]any{
			"color": "blue",
			"size":  "large",
		},
	}, &v2)
	v2ID := int64(v2["id"].(float64))
	// Verify attachment via GET flag (includes preloaded variants)
	var flagResp map[string]any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flagResp)
	variants, _ := flagResp["variants"].([]any)
	var foundAtt bool
	for _, vr := range variants {
		vm := vr.(map[string]any)
		if int64(vm["id"].(float64)) == v2ID {
			att, ok := vm["attachment"].(map[string]any)
			if !ok || att["color"] != "blue" || att["size"] != "large" {
				t.Fatalf("expected attachment {color:blue,size:large}, got %v", vm["attachment"])
			}
			foundAtt = true
			break
		}
	}
	if !foundAtt {
		t.Fatal("variant with attachment not found in flag response")
	}

	var fetched map[string]any

	// Update variant attachment
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants/%d", flagID, v2ID), map[string]any{
		"key": "variant_with_attachment",
		"attachment": map[string]any{
			"color": "red",
		},
	}, &fetched)
	att2, ok2 := fetched["attachment"].(map[string]any)
	if !ok2 || att2["color"] != "red" {
		t.Fatalf("expected attachment {color:red}, got %v", fetched["attachment"])
	}

	// Delete both variants
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/variants/%d", flagID, variantID))
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/variants/%d", flagID, v2ID))
}

func TestIntegration_DistributionCRUD(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	flagID := seedFlagIDs[0]

	// Get flag to find existing variants+segments
	var flag map[string]any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flag)
	variantsRaw, _ := flag["variants"].([]any)
	if len(variantsRaw) == 0 {
		t.Fatal("no variants on seeded flag")
	}
	firstVariant := variantsRaw[0].(map[string]any)
	variantID := int64(firstVariant["id"].(float64))
	variantKey := firstVariant["key"].(string)

	segsRaw, _ := flag["segments"].([]any)
	if len(segsRaw) == 0 {
		t.Fatal("no segments on seeded flag")
	}
	segID := int64(segsRaw[0].(map[string]any)["id"].(float64))

	// Put distributions (single variant at 100%)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", flagID, segID), map[string]any{
		"distributions": []map[string]any{
			{"percent": 100, "variantID": variantID, "variantKey": variantKey},
		},
	}, nil)

	// Verify via GET distributions endpoint
	var dists []any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", flagID, segID), &dists)
	if len(dists) != 1 {
		t.Fatalf("expected 1 distribution, got %d", len(dists))
	}
	d := dists[0].(map[string]any)
	if int64(d["variantID"].(float64)) != variantID {
		t.Fatalf("expected variantID %d, got %v", variantID, d["variantID"])
	}

	// Update distribution: split 60/40 across two variants
	if len(variantsRaw) >= 2 {
		secondVariant := variantsRaw[1].(map[string]any)
		v2ID := int64(secondVariant["id"].(float64))
		v2Key := secondVariant["key"].(string)

		putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", flagID, segID), map[string]any{
			"distributions": []map[string]any{
				{"percent": 60, "variantID": variantID, "variantKey": variantKey},
				{"percent": 40, "variantID": v2ID, "variantKey": v2Key},
			},
		}, nil)

		getJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", flagID, segID), &dists)
		if len(dists) != 2 {
			t.Fatalf("expected 2 distributions after update, got %d", len(dists))
		}
	}

	// Verify by getting the flag (includes segments with distributions)
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flag)
}

func TestIntegration_Evaluation(t *testing.T) {
	// Use a seeded flag untouched by other tests (all CRUD tests use seedFlagIDs[0]).
	// Index 1 = int_flag_EQ_02 with constraint tier EQ "premium".
	if len(seedFlagIDs) < 2 || len(seedFlagKeys) < 2 {
		t.Fatal("need at least 2 seeded flags")
	}
	flagID := seedFlagIDs[1]
	flagKey := seedFlagKeys[1]

	eval := func(body map[string]any) map[string]any {
		var result map[string]any
		postJSON(t, "/api/v1/evaluation", body, &result)
		return result
	}

	t.Run("eval_by_flag_id", func(t *testing.T) {
		result := eval(map[string]any{
			"flagID":     flagID,
			"entityID":   "test-entity",
			"entityType": "user",
			"entityContext": map[string]any{
				"tier": "premium",
			},
		})
		if id, ok := result["flagID"].(float64); !ok || int64(id) != flagID {
			t.Fatalf("expected flagID %d in eval response, got %v", flagID, result["flagID"])
		}
		if vk, ok := result["variantKey"].(string); !ok || vk == "" {
			t.Fatalf("expected non-empty variantKey in eval response, got %v", result["variantKey"])
		}
	})

	t.Run("eval_by_flag_key", func(t *testing.T) {
		result := eval(map[string]any{
			"flagKey":    flagKey,
			"entityID":   "test-entity",
			"entityType": "user",
			"entityContext": map[string]any{
				"tier": "premium",
			},
		})
		if vk, ok := result["variantKey"].(string); !ok || vk == "" {
			t.Fatalf("expected non-empty variantKey in flagKey eval, got %v", result["variantKey"])
		}
		if result["evalContext"] == nil {
			t.Fatal("eval response missing evalContext")
		}
	})
}

func TestIntegration_Preload(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}

	// Get flags without preload — segments/variants should be empty
	var without []map[string]any
	getJSON(t, "/api/v1/flags", &without)
	for _, f := range without {
		if segs, ok := f["segments"].([]any); ok && len(segs) > 0 {
			t.Log("flag without preload unexpectedly has segments")
		}
	}

	// Get flags WITH preload — should include segments/variants
	var with []map[string]any
	getJSON(t, "/api/v1/flags?preload=true", &with)
	if len(with) == 0 {
		t.Fatal("expected at least one flag")
	}

	// Get single flag — always preloaded, verify variant/segment data present
	var single map[string]any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", seedFlagIDs[0]), &single)
	if segs, ok := single["segments"].([]any); !ok || len(segs) == 0 {
		t.Fatal("expected segments on single flag get (always preloaded)")
	}
}

func TestIntegration_Export(t *testing.T) {
	// Export SQLite
	resp, err := doReq("GET", "/api/v1/export/sqlite", nil)
	if err != nil {
		t.Fatalf("GET /api/v1/export/sqlite: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("export/sqlite: expected 200/204, got %d", resp.StatusCode)
	}

	// Export eval cache json (returns {"Flags": [...]})
	var cache map[string]any
	getJSON(t, "/api/v1/export/eval_cache/json", &cache)
	if cache["Flags"] == nil {
		t.Fatal("eval cache json missing Flags key")
	}
	flags, ok := cache["Flags"].([]any)
	if !ok || len(flags) == 0 {
		t.Fatal("expected non-empty flags in eval cache")
	}
}

func TestIntegration_TagCRUD(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	flagID := seedFlagIDs[0]

	// Create two tags
	tagVal1 := fmt.Sprintf("tag_crud_1_%d", time.Now().UnixNano())
	var tag1 map[string]any
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), map[string]any{
		"value": tagVal1,
	}, &tag1)
	tag1ID := int64(tag1["id"].(float64))

	tagVal2 := fmt.Sprintf("tag_crud_2_%d", time.Now().UnixNano())
	var tag2 map[string]any
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), map[string]any{
		"value": tagVal2,
	}, &tag2)
	tag2ID := int64(tag2["id"].(float64))

	// List tags on flag — both should be present
	var tags []any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), &tags)
	found1, found2 := false, false
	for _, t2 := range tags {
		val, _ := t2.(map[string]any)["value"].(string)
		if val == tagVal1 {
			found1 = true
		}
		if val == tagVal2 {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Fatal("expected both tags in flag tag list")
	}

	// Delete first tag and verify list updated
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/tags/%d", flagID, tag1ID))
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), &tags)
	for _, t2 := range tags {
		if t2.(map[string]any)["value"] == tagVal1 {
			t.Fatal("tag1 still present after deletion")
		}
	}

	// Delete second tag
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/tags/%d", flagID, tag2ID))

	// List all tags — should still contain seeded tags
	var allTags []any
	getJSON(t, "/api/v1/tags", &allTags)
}

func TestIntegration_BatchEval(t *testing.T) {
	if len(seedFlagIDs) < 2 {
		t.Fatal("need at least 2 seeded flags")
	}

	var result map[string]any
	postJSON(t, "/api/v1/evaluation/batch", map[string]any{
		"entities": []map[string]any{
			{
				"entityID":   "batch-entity",
				"entityType": "user",
				"entityContext": map[string]any{
					"region": "us-west",
					"age":    30,
				},
			},
		},
		"flagIDs": []int64{seedFlagIDs[0], seedFlagIDs[1]},
	}, &result)

	if result["evaluationResults"] == nil {
		t.Fatal("batch eval response missing evaluationResults")
	}
}

func TestIntegration_BatchEvalOperator(t *testing.T) {
	if len(seedFlagIDs) < 2 {
		t.Fatal("need at least 2 seeded flags")
	}

	// Batch eval with tag operator ANY
	var result map[string]any
	postJSON(t, "/api/v1/evaluation/batch", map[string]any{
		"entities": []map[string]any{
			{
				"entityID":   "tag-batch-entity",
				"entityType": "user",
				"entityContext": map[string]any{
					"region": "us-west",
				},
			},
		},
		"flagTags":         []string{"int_test"},
		"flagTagsOperator": "ANY",
	}, &result)
	if result["evaluationResults"] == nil {
		t.Fatal("batch eval (ANY) response missing evaluationResults")
	}

	// Batch eval with tag operator ALL
	var resultAll map[string]any
	postJSON(t, "/api/v1/evaluation/batch", map[string]any{
		"entities": []map[string]any{
			{
				"entityID":   "tag-batch-entity",
				"entityType": "user",
				"entityContext": map[string]any{
					"region": "us-west",
				},
			},
		},
		"flagTags":         []string{"int_test"},
		"flagTagsOperator": "ALL",
	}, &resultAll)
	if resultAll["evaluationResults"] == nil {
		t.Fatal("batch eval (ALL) response missing evaluationResults")
	}

	// ALL with multiple tags that must ALL match
	var resultAllMulti map[string]any
	postJSON(t, "/api/v1/evaluation/batch", map[string]any{
		"entities": []map[string]any{
			{
				"entityID":   "tag-batch-entity",
				"entityType": "user",
				"entityContext": map[string]any{
					"region": "us-west",
				},
			},
		},
		"flagTags":         []string{"int_test", "constraint_EQ"},
		"flagTagsOperator": "ALL",
	}, &resultAllMulti)
	if resultAllMulti["evaluationResults"] == nil {
		t.Fatal("batch eval (ALL multi) response missing evaluationResults")
	}

	// ANY with partial match — one of two tags present, should still return results
	var resultAnyPartial map[string]any
	postJSON(t, "/api/v1/evaluation/batch", map[string]any{
		"entities": []map[string]any{
			{
				"entityID":   "tag-batch-entity",
				"entityType": "user",
				"entityContext": map[string]any{
					"region": "us-west",
				},
			},
		},
		"flagTags":         []string{"int_test", "nonexistent_tag_xyz"},
		"flagTagsOperator": "ANY",
	}, &resultAnyPartial)
	if resultAnyPartial["evaluationResults"] == nil {
		t.Fatal("batch eval (ANY partial) response missing evaluationResults")
	}
}
