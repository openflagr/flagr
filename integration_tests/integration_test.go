// Tests run sequentially within this package (no t.Parallel) because they share
// a single server instance and mutate seeded state. Each CRUD test uses
// seedFlagIDs[0]; Evaluation uses seedFlagIDs[1]. Do not reorder or parallelize
// without isolating per-test state.

//go:build integration

// Package flagr_integration provides HTTP-based integration tests for Flagr.
//
// Execution modes:
//   - Local:   go test -tags=integration ./integration_tests/
//              (auto-starts server: SQLite :memory:, recorder on, Datar flush 500ms, eval cache 1s)
//   - BYO:     FLAGR_SERVER_URL=http://host:18000 go test -tags=integration ./integration_tests/
//   - Docker:  cd integration_tests && make test
//              (builds binary, runs against all 6 compose instances)
//
// TestIntegration_Exposures asserts POST /exposures recording via loggedCount (no test-process FLAGR_RECORDER_ENABLED).
package flagr_integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	maxBeforeCreate := getSnapshotMaxID(t)
	var created flagResponse
	postJSON(t, "/api/v1/flags", map[string]any{
		"key":         key,
		"description": "crud test flag",
	}, &created)
	if created.ID == 0 {
		t.Fatal("expected non-zero id")
	}
	maxAfterCreate := getSnapshotMaxID(t)
	if maxAfterCreate <= maxBeforeCreate {
		t.Fatalf("expected global max snapshot id to increase after create: before=%d after=%d", maxBeforeCreate, maxAfterCreate)
	}
	if n := countFlagSnapshots(t, created.ID); n < 1 {
		t.Fatalf("expected at least 1 snapshot after create, got %d", n)
	}

	// Get flag
	var fetched flagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", created.ID), &fetched)
	if fetched.Key != key {
		t.Fatalf("expected key %s, got %s", key, fetched.Key)
	}

	snapsBeforePut := countFlagSnapshots(t, created.ID)

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
	if n := countFlagSnapshots(t, created.ID); n <= snapsBeforePut {
		t.Fatalf("expected new snapshot after put: before=%d after=%d", snapsBeforePut, n)
	}

	// Set enabled (PUT)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/enabled", created.ID), map[string]any{
		"enabled": true,
	}, &fetched)

	// Get flag entity types (should include test_entity)
	var types []string
	getJSON(t, "/api/v1/flags/entity_types", &types)
	found := false
	for _, et := range types {
		if et == "test_entity" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected test_entity in entity types")
	}

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
	if len(flags) == 0 {
		t.Fatal("expected at least one flag with preload")
	}
	if len(flags[0].Segments) == 0 {
		t.Fatal("expected segments on preloaded flag")
	}

	// Delete flag
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d", created.ID))

	// Restore flag (PUT, not POST)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/restore", created.ID), nil, nil)

	// Each mutation above should append one history row for this flag
	var snapshots []any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/snapshots", created.ID), &snapshots)
	if len(snapshots) < 5 {
		t.Fatalf("expected at least 5 flag snapshots (create, put, enabled, delete, restore), got %d", len(snapshots))
	}
}

func TestIntegration_SegmentCRUD(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	flagID := seedFlagIDs[0]
	snapsBefore := countFlagSnapshots(t, flagID)

	// Create segment
	var seg segmentResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments", flagID), map[string]any{
		"description":    "test segment",
		"rolloutPercent": 50,
	}, &seg)
	if n := countFlagSnapshots(t, flagID); n <= snapsBefore {
		t.Fatalf("expected snapshot after segment create: before=%d after=%d", snapsBefore, n)
	}

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

	snapsBeforeDelete := countFlagSnapshots(t, flagID)
	// Delete segment
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d", flagID, seg.ID))
	if n := countFlagSnapshots(t, flagID); n <= snapsBeforeDelete {
		t.Fatalf("expected snapshot after segment delete: before=%d after=%d", snapsBeforeDelete, n)
	}

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
	snapsBefore := countFlagSnapshots(t, flagID)

	// Create variant
	var v variantResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants", flagID), map[string]any{
		"key": "test_variant",
	}, &v)
	if n := countFlagSnapshots(t, flagID); n <= snapsBefore {
		t.Fatalf("expected snapshot after variant create: before=%d after=%d", snapsBefore, n)
	}
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
func TestIntegration_DuplicateFlag(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	sourceID := seedFlagIDs[0]
	var source flagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", sourceID), &source)

	sourceSnapsBefore := countFlagSnapshots(t, sourceID)
	maxBefore := getSnapshotMaxID(t)

	var clone flagResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/duplicate", sourceID), map[string]any{}, &clone)
	if clone.ID == sourceID {
		t.Fatalf("clone id must differ from source, both %d", sourceID)
	}
	if clone.Key == source.Key {
		t.Fatalf("clone key must differ from source")
	}
	if !strings.Contains(clone.Description, "(cloned)") {
		t.Fatalf("expected cloned description suffix, got %q", clone.Description)
	}
	if len(clone.Variants) != len(source.Variants) {
		t.Fatalf("variant count: source %d clone %d", len(source.Variants), len(clone.Variants))
	}
	if len(clone.Segments) != len(source.Segments) {
		t.Fatalf("segment count: source %d clone %d", len(source.Segments), len(clone.Segments))
	}

	if sourceSnapsAfter := countFlagSnapshots(t, sourceID); sourceSnapsAfter != sourceSnapsBefore {
		t.Fatalf("duplicate must not append snapshots to source flag: before=%d after=%d", sourceSnapsBefore, sourceSnapsAfter)
	}

	cloneKey := fmt.Sprintf("dup_custom_key_%d", time.Now().UnixNano())
	cloneDesc := "integration custom clone description"
	var customClone flagResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/duplicate", sourceID), map[string]any{
		"key":         cloneKey,
		"description": cloneDesc,
	}, &customClone)
	if customClone.Key != cloneKey {
		t.Fatalf("expected clone key %q, got %q", cloneKey, customClone.Key)
	}
	if customClone.Description != cloneDesc {
		t.Fatalf("expected clone description %q, got %q", cloneDesc, customClone.Description)
	}
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d", customClone.ID))
	if cloneSnaps := countFlagSnapshots(t, clone.ID); cloneSnaps != 1 {
		t.Fatalf("duplicate must create exactly one snapshot on the new flag, got %d", cloneSnaps)
	}
	maxAfter := getSnapshotMaxID(t)
	if maxAfter <= maxBefore {
		t.Fatalf("expected global max snapshot id to increase after duplicate: before=%d after=%d", maxBefore, maxAfter)
	}

	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", clone.ID), &clone)
	if clone.ID == 0 {
		t.Fatal("GET clone failed")
	}
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d", clone.ID))
}

func TestIntegration_DuplicateFlag_Errors(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	sourceID := seedFlagIDs[0]

	postJSONExpectStatus(t, "/api/v1/flags/999999999/duplicate", map[string]any{}, 404, nil)

	key := fmt.Sprintf("dup_err_src_%d", time.Now().UnixNano())
	var created flagResponse
	postJSON(t, "/api/v1/flags", map[string]any{
		"key":         key,
		"description": "dup error source",
	}, &created)
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d", created.ID))
	postJSONExpectStatus(t, fmt.Sprintf("/api/v1/flags/%d/duplicate", created.ID), map[string]any{}, 404, nil)

	takenKey := fmt.Sprintf("dup_taken_%d", time.Now().UnixNano())
	postJSON(t, "/api/v1/flags", map[string]any{"key": takenKey, "description": "holder"}, nil)
	postJSONExpectStatus(t, fmt.Sprintf("/api/v1/flags/%d/duplicate", sourceID), map[string]any{
		"key": takenKey,
	}, 400, nil)

	postJSONExpectStatus(t, fmt.Sprintf("/api/v1/flags/%d/duplicate", sourceID), map[string]any{
		"key": " spaces invalid ",
	}, 400, nil)
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
			t.Errorf("flag %d without preload unexpectedly has %d segments", f.ID, len(f.Segments))
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
	snapsBefore := countFlagSnapshots(t, flagID)

	// Create two tags
	tagVal1 := fmt.Sprintf("tag_crud_1_%d", time.Now().UnixNano())
	var tag1 tagResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), map[string]any{
		"value": tagVal1,
	}, &tag1)
	if n := countFlagSnapshots(t, flagID); n <= snapsBefore {
		t.Fatalf("expected snapshot after tag create: before=%d after=%d", snapsBefore, n)
	}

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

// assertExposureRecorderGate polls until a valid exposure row is written (loggedCount==1).
// The Flagr server must have FLAGR_RECORDER_ENABLED and per-flag dataRecordsEnabled.
func assertExposureRecorderGate(t *testing.T, flagID int64) {
	t.Helper()
	err := pollUntil("exposure recorder gate", "/api/v1/exposures", exposureRecorderGateTimeout, func() bool {
		r, err := doReq("POST", "/api/v1/exposures", map[string]any{
			"exposures": []map[string]any{{
				"flagID":   flagID,
				"entityID": "exposure-cache-warmup",
			}},
		})
		if err != nil {
			return false
		}
		defer r.Body.Close()
		if r.StatusCode < 200 || r.StatusCode >= 300 {
			return false
		}
		var warm struct {
			LoggedCount int64 `json:"loggedCount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&warm); err != nil {
			return false
		}
		return warm.LoggedCount == 1
	})
	if err != nil {
		t.Fatalf("exposure logging not recording after dataRecordsEnabled (server needs FLAGR_RECORDER_ENABLED): %v", err)
	}
	t.Log("exposure recorder gate: loggedCount=1")
}

func TestIntegration_Exposures(t *testing.T) {
	if len(seedFlagIDs) < 2 {
		t.Fatal("need at least 2 seeded flags")
	}
	flagID := seedFlagIDs[1]

	avail, err := doReq("POST", "/api/v1/exposures", map[string]any{
		"exposures": []map[string]any{{"flagID": flagID, "entityID": "exposure-probe"}},
	})
	if err != nil {
		t.Fatalf("POST /exposures probe: %v", err)
	}
	io.Copy(io.Discard, avail.Body)
	avail.Body.Close()
	if avail.StatusCode == http.StatusNotFound {
		t.Skip("POST /exposures not available on this server (e.g. checkr/flagr:1.1.12)")
	}

	doReqOK(t, "PUT", fmt.Sprintf("/api/v1/flags/%d", flagID), map[string]any{
		"dataRecordsEnabled": true,
	})
	var flag flagResponse
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flag)
	if !flag.DataRecordsEnabled {
		t.Fatalf("expected dataRecordsEnabled=true on flag %d", flagID)
	}

	assertExposureRecorderGate(t, flagID)

	body := map[string]any{
		"exposures": []map[string]any{
			{"flagID": flagID, "entityID": "exposure-test-entity"},
			{"entityID": "bad-row-no-flag"},
		},
	}
	probe, err := doReq("POST", "/api/v1/exposures", body)
	if err != nil {
		t.Fatalf("POST /exposures: %v", err)
	}
	defer probe.Body.Close()
	if probe.StatusCode < 200 || probe.StatusCode >= 300 {
		b, _ := io.ReadAll(probe.Body)
		t.Fatalf("POST /exposures: expected 2xx, got %d: %s", probe.StatusCode, b)
	}

	var resp struct {
		LoggedCount int64 `json:"loggedCount"`
		Errors      []struct {
			Index   int64  `json:"index"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(probe.Body).Decode(&resp); err != nil {
		t.Fatalf("decode exposures response: %v", err)
	}
	if len(resp.Errors) != 1 {
		t.Fatalf("expected 1 row error, got %d", len(resp.Errors))
	}
	if resp.Errors[0].Index != 1 {
		t.Fatalf("expected error index 1, got %d", resp.Errors[0].Index)
	}
	if resp.LoggedCount != 1 {
		t.Fatalf("loggedCount want 1 after recorder gate, got %d", resp.LoggedCount)
	}
}

// ---------------------------------------------------------------------------
// Datar integration tests
// ---------------------------------------------------------------------------

// datarSummaryFlag mirrors the swagger DatarSummaryFlag response.
type datarSummaryFlag struct {
	FlagID          int64  `json:"flagID"`
	FlagKey         string `json:"flagKey"`
	TotalEvalCount  int64  `json:"totalEvalCount"`
	Enabled         bool   `json:"enabled"`
	LastEvaluatedAt string `json:"lastEvaluatedAt"`
}

type datarSummaryResponse struct {
	Flags []datarSummaryFlag `json:"flags"`
}

type datarVariantEntry struct {
	VariantID int64 `json:"variantID"`
	Count     int64 `json:"count"`
}

type datarSegmentEntry struct {
	SegmentID int64 `json:"segmentID"`
	Count     int64 `json:"count"`
}

type datarDayEntry struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type datarFlagSummaryResponse struct {
	FlagID           int64               `json:"flagID"`
	TrafficByVariant []datarVariantEntry `json:"trafficByVariant"`
	TrafficBySegment []datarSegmentEntry `json:"trafficBySegment"`
	TrafficByDay     []datarDayEntry     `json:"trafficByDay"`
}

func waitForDatarFlushAfterEvals() {
	time.Sleep(integrationDatarFlushWait())
}

func postDatarEvalBurst(t *testing.T, flagID int64, idPrefix string) {
	t.Helper()
	for i := 0; i < datarPollEvalsPerAttempt; i++ {
		doReqOK(t, "POST", "/api/v1/evaluation", datarEvalBody(flagID, fmt.Sprintf("%s-%d", idPrefix, i)))
	}
}

// datarEvalBody matches TestIntegration_Evaluation (seedFlagIDs[1] = int_flag_EQ_02, tier EQ premium).
func datarEvalBody(flagID int64, entityID string) map[string]any {
	return map[string]any{
		"flagID":     flagID,
		"entityID":   entityID,
		"entityType": "user",
		"entityContext": map[string]any{
			"tier": "premium",
		},
	}
}

func TestIntegration_DatarSummary(t *testing.T) {
	if len(seedFlagIDs) < 2 {
		t.Fatal("need at least 2 seeded flags for datar test")
	}

	flagID := seedFlagIDs[1]

	// Datar may not be enabled on all servers (e.g. checkr_flagr_with_sqlite).
	// Probe the endpoint first; skip if unavailable.
	resp, err := doReq("GET", "/api/v1/datar/summary", nil)
	if err != nil {
		t.Skipf("datar not available: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("datar summary returned %d, skipping", resp.StatusCode)
	}
	// Enable dataRecordsEnabled so evaluations are recorded into the datar engine.
	doReqOK(t, "PUT", fmt.Sprintf("/api/v1/flags/%d", flagID), map[string]any{
		"dataRecordsEnabled": true,
	})

	// Poll until datar summary shows eval counts. Evaluations are performed
	// inside the loop because the eval cache refreshes every ~3s — the
	// DataRecordsEnabled change won't take effect until the next reload.
	var summary datarSummaryResponse
	err = pollUntil("datar/summary", "/api/v1/datar/summary", datarPollTimeout, func() bool {
		postDatarEvalBurst(t, flagID, "datar-entity")
		waitForDatarFlushAfterEvals()
		resp, err := doReq("GET", "/api/v1/datar/summary", nil)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return false
		}
		json.NewDecoder(resp.Body).Decode(&summary)
		// Check that our flag has eval counts.
		for _, f := range summary.Flags {
			if f.FlagID == flagID && f.TotalEvalCount > 0 {
				return true
			}
		}
		return false
	})
	if err != nil {
		t.Fatalf("datar summary never showed eval counts for flag %d: %v", flagID, err)
	}
}

func TestIntegration_DatarFlagSummary(t *testing.T) {
	if len(seedFlagIDs) < 2 {
		t.Fatal("need at least 2 seeded flags for datar test")
	}

	flagID := seedFlagIDs[1]

	// Skip if datar is not available.
	resp, err := doReq("GET", fmt.Sprintf("/api/v1/datar/flags/%d/summary", flagID), nil)
	if err != nil {
		t.Skipf("datar not available: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("datar flag summary returned %d, skipping", resp.StatusCode)
	}
	// Enable dataRecordsEnabled so evaluations are recorded into the datar engine.
	doReqOK(t, "PUT", fmt.Sprintf("/api/v1/flags/%d", flagID), map[string]any{
		"dataRecordsEnabled": true,
	})

	// Poll until flag summary has traffic. Evaluations are performed
	// inside the loop because the eval cache refreshes every ~3s.
	var flagSummary datarFlagSummaryResponse
	err = pollUntil("datar/flags/summary", fmt.Sprintf("/api/v1/datar/flags/%d/summary", flagID), datarPollTimeout, func() bool {
		postDatarEvalBurst(t, flagID, "datar-flag-entity")
		waitForDatarFlushAfterEvals()
		resp, err := doReq("GET", fmt.Sprintf("/api/v1/datar/flags/%d/summary", flagID), nil)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return false
		}
		json.NewDecoder(resp.Body).Decode(&flagSummary)
		return len(flagSummary.TrafficByVariant) > 0
	})
	if err != nil {
		t.Fatalf("datar flag summary never populated for flag %d: %v", flagID, err)
	}

	if flagSummary.FlagID != flagID {
		t.Errorf("expected flagID %d, got %d", flagID, flagSummary.FlagID)
	}

	for _, v := range flagSummary.TrafficByVariant {
		if v.Count <= 0 {
			t.Errorf("variant %d has non-positive count %d", v.VariantID, v.Count)
		}
	}

	for _, s := range flagSummary.TrafficBySegment {
		if s.Count <= 0 {
			t.Errorf("segment %d has non-positive count %d", s.SegmentID, s.Count)
		}
	}

	if len(flagSummary.TrafficByDay) == 0 {
		t.Error("expected non-empty trafficByDay")
	}
	for _, d := range flagSummary.TrafficByDay {
		if d.Count <= 0 {
			t.Errorf("day %s has non-positive count %d", d.Date, d.Count)
		}
	}
}
