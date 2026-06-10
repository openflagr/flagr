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
	var created flagResponse
	postJSON(t, "/api/v1/flags", map[string]any{
		"key":         key,
		"description": "crud test flag",
	}, &created)
	if created.ID == 0 {
		t.Fatal("expected non-zero id")
	}

	// Get flag
	var fetched flagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", created.ID), &fetched)
	if fetched.Key != key {
		t.Fatalf("expected key %s, got %s", key, fetched.Key)
	}

	// Put flag — description, key, dataRecordsEnabled, entityType
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d", created.ID), map[string]any{
		"description":        "updated description",
		"key":                key,
		"dataRecordsEnabled": true,
		"entityType":         "test_entity",
	}, &fetched)
	if !fetched.DataRecordsEnabled {
		t.Fatalf("expected dataRecordsEnabled=true, got %v", fetched.DataRecordsEnabled)
	}

	// Set enabled (PUT)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/enabled", created.ID), map[string]any{
		"enabled": true,
	}, &fetched)

	// Get flag entity types (should include test_entity)
	var types []string
	getJSON(t, "/api/v1/flags/entity_types", &types)

	// Query flags by description
	var byDesc []flagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags?description=%s", url.QueryEscape("updated description")), &byDesc)
	if len(byDesc) == 0 {
		t.Fatal("expected at least one flag matching description")
	}

	// Query flags by key
	var byKey []flagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags?key=%s", key), &byKey)
	if len(byKey) != 1 {
		t.Fatalf("expected exactly 1 flag matching key, got %d", len(byKey))
	}

	// Query flags with limit/offset
	var limited []flagResponse
	getJSON(t, "/api/v1/flags?limit=1&offset=0", &limited)
	if len(limited) > 1 {
		t.Fatalf("expected at most 1 flag with limit=1, got %d", len(limited))
	}

	// Find flags with preload
	var flags []flagResponse
	getJSON(t, "/api/v1/flags?preload=true&limit=1", &flags)

	// Delete flag
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d", created.ID))

	// Restore flag (PUT, not POST)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/restore", created.ID), nil, nil)

	// Get snapshot
	var snapshots []any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/snapshots", created.ID), &snapshots)
}

func TestIntegration_SegmentCRUD(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	flagID := seedFlagIDs[0]

	// Create segment
	var seg segmentResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments", flagID), map[string]any{
		"description":    "test segment",
		"rolloutPercent": 50,
	}, &seg)

	// Put segment — update description and rollout
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d", flagID, seg.ID), map[string]any{
		"description":    "updated segment",
		"rolloutPercent": 100,
	}, nil)

	// Reorder segments (replace active segments with just the new one)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/reorder", flagID), map[string]any{
		"segmentIDs": []int64{seg.ID},
	}, nil)

	// Verify rank via single flag get
	var flagObj flagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flagObj)
	found := false
	for _, s := range flagObj.Segments {
		if s.ID == seg.ID {
			if s.Rank != 0 {
				t.Fatalf("expected rank 0 after reorder, got %d", s.Rank)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("reordered segment not found in flag response")
	}

	// Delete segment
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d", flagID, seg.ID))

	// Verify deletion
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flagObj)
	for _, s := range flagObj.Segments {
		if s.ID == seg.ID {
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
	var seg segmentResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments", flagID), map[string]any{
		"description":    "constraint test segment",
		"rolloutPercent": 100,
	}, &seg)

	// Create constraint
	var constraint constraintResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints", flagID, seg.ID), map[string]any{
		"property": "test_prop",
		"operator": "EQ",
		"value":    `"test_value"`,
	}, &constraint)

	// Update constraint
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints/%d", flagID, seg.ID, constraint.ID), map[string]any{
		"property": "test_prop",
		"operator": "NEQ",
		"value":    `"other_value"`,
	}, &constraint)

	// List constraints
	var list []constraintResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints", flagID, seg.ID), &list)
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
	var v variantResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants", flagID), map[string]any{
		"key": "test_variant",
	}, &v)
	variantID := v.ID

	// Update variant key
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants/%d", flagID, variantID), map[string]any{
		"key": "test_variant_updated",
	}, &v)

	// Create variant with attachment
	var v2 variantResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants", flagID), map[string]any{
		"key": "variant_with_attachment",
		"attachment": map[string]any{
			"color": "blue",
			"size":  "large",
		},
	}, &v2)

	// Verify attachment via GET flag (includes preloaded variants)
	var flagResp flagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flagResp)
	var foundAtt bool
	for _, vr := range flagResp.Variants {
		if vr.ID == v2.ID {
			if vr.Attachment == nil || vr.Attachment["color"] != "blue" || vr.Attachment["size"] != "large" {
				t.Fatalf("expected attachment {color:blue,size:large}, got %v", vr.Attachment)
			}
			foundAtt = true
			break
		}
	}
	if !foundAtt {
		t.Fatal("variant with attachment not found in flag response")
	}

	// Update variant attachment
	var fetched variantResponse
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants/%d", flagID, v2.ID), map[string]any{
		"key": "variant_with_attachment",
		"attachment": map[string]any{
			"color": "red",
		},
	}, &fetched)
	if fetched.Attachment == nil || fetched.Attachment["color"] != "red" {
		t.Fatalf("expected attachment {color:red}, got %v", fetched.Attachment)
	}

	// Delete both variants
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/variants/%d", flagID, variantID))
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/variants/%d", flagID, v2.ID))
}

func TestIntegration_DistributionCRUD(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	flagID := seedFlagIDs[0]

	// Get flag to find existing variants+segments
	var flag flagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flag)
	if len(flag.Variants) == 0 {
		t.Fatal("no variants on seeded flag")
	}
	if len(flag.Segments) == 0 {
		t.Fatal("no segments on seeded flag")
	}
	firstVariant := flag.Variants[0]
	segID := flag.Segments[0].ID

	// Put distributions (single variant at 100%)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", flagID, segID), map[string]any{
		"distributions": []map[string]any{
			{"percent": 100, "variantID": firstVariant.ID, "variantKey": firstVariant.Key},
		},
	}, nil)

	// Verify via GET distributions endpoint
	var dists []distributionResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", flagID, segID), &dists)
	if len(dists) != 1 {
		t.Fatalf("expected 1 distribution, got %d", len(dists))
	}
	if dists[0].VariantID != firstVariant.ID {
		t.Fatalf("expected variantID %d, got %d", firstVariant.ID, dists[0].VariantID)
	}

	// Update distribution: split 60/40 across two variants
	if len(flag.Variants) >= 2 {
		secondVariant := flag.Variants[1]
		putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", flagID, segID), map[string]any{
			"distributions": []map[string]any{
				{"percent": 60, "variantID": firstVariant.ID, "variantKey": firstVariant.Key},
				{"percent": 40, "variantID": secondVariant.ID, "variantKey": secondVariant.Key},
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

	eval := func(body map[string]any) evalResponse {
		var result evalResponse
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
		if result.FlagID != flagID {
			t.Fatalf("expected flagID %d in eval response, got %d", flagID, result.FlagID)
		}
		if result.VariantKey == "" {
			t.Fatal("expected non-empty variantKey in eval response")
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
		if result.VariantKey == "" {
			t.Fatal("expected non-empty variantKey in flagKey eval")
		}
		if result.EvalContext == nil {
			t.Fatal("eval response missing evalContext")
		}
	})
}

func TestIntegration_Preload(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}

	// Get flags without preload — segments/variants should be empty
	var without []flagResponse
	getJSON(t, "/api/v1/flags", &without)
	for _, f := range without {
		if len(f.Segments) > 0 {
			t.Log("flag without preload unexpectedly has segments")
		}
	}

	// Get flags WITH preload — should include segments/variants
	var with []flagResponse
	getJSON(t, "/api/v1/flags?preload=true", &with)
	if len(with) == 0 {
		t.Fatal("expected at least one flag")
	}

	// Get single flag — always preloaded, verify variant/segment data present
	var single flagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", seedFlagIDs[0]), &single)
	if len(single.Segments) == 0 {
		t.Fatal("expected segments on single flag get (always preloaded)")
	}
}

func TestIntegration_Export(t *testing.T) {
	// Export SQLite — doReqOK drains body to avoid broken pipe.
	doReqOK(t, "GET", "/api/v1/export/sqlite", nil)

	// Export eval cache json (returns {"Flags": [...]})
	var cache struct {
		Flags []flagResponse `json:"Flags"`
	}
	getJSON(t, "/api/v1/export/eval_cache/json", &cache)
	if len(cache.Flags) == 0 {
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
	var tag1 tagResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), map[string]any{
		"value": tagVal1,
	}, &tag1)

	tagVal2 := fmt.Sprintf("tag_crud_2_%d", time.Now().UnixNano())
	var tag2 tagResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), map[string]any{
		"value": tagVal2,
	}, &tag2)

	// List tags on flag — both should be present
	var tags []tagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), &tags)
	found1, found2 := false, false
	for _, tg := range tags {
		if tg.Value == tagVal1 {
			found1 = true
		}
		if tg.Value == tagVal2 {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Fatal("expected both tags in flag tag list")
	}

	// Delete first tag and verify list updated
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/tags/%d", flagID, tag1.ID))
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), &tags)
	for _, tg := range tags {
		if tg.Value == tagVal1 {
			t.Fatal("tag1 still present after deletion")
		}
	}

	// Delete second tag
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/tags/%d", flagID, tag2.ID))

	// List all tags — should still contain seeded tags
	var allTags []tagResponse
	getJSON(t, "/api/v1/tags", &allTags)
}

func TestIntegration_BatchEval(t *testing.T) {
	if len(seedFlagIDs) < 2 {
		t.Fatal("need at least 2 seeded flags")
	}

	var result batchEvalResponse
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

	if len(result.EvaluationResults) == 0 {
		t.Fatal("batch eval response missing evaluationResults")
	}
}

func TestIntegration_BatchEvalOperator(t *testing.T) {
	if len(seedFlagIDs) < 2 {
		t.Fatal("need at least 2 seeded flags")
	}

	cases := []struct {
		name     string
		tags     []string
		operator string
	}{
		{"ANY matching", []string{"int_test"}, "ANY"},
		{"ALL matching", []string{"int_test"}, "ALL"},
		{"ALL multi-tag", []string{"int_test", "constraint_EQ"}, "ALL"},
		{"ANY partial", []string{"int_test", "nonexistent_tag_xyz"}, "ANY"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var result batchEvalResponse
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
				"flagTags":         tc.tags,
				"flagTagsOperator": tc.operator,
			}, &result)
			if len(result.EvaluationResults) == 0 {
				t.Fatalf("batch eval (%s) response missing evaluationResults", tc.operator)
			}
		})
	}
}
