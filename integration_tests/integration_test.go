// Tests run sequentially within this package (no t.Parallel) because they share
// a single server instance and mutate seeded state. Each CRUD test uses
// seedFlagIDs[0]; Evaluation uses seedFlagIDs[1]. Do not reorder or parallelize
// without isolating per-test state.

//go:build integration

// Package flagr_integration provides HTTP-based integration tests for Flagr.
//
// Execution modes:
//   - Local:   go test -tags=integration ./integration_tests/
//     (auto-starts server: SQLite :memory:, recorder on, Datar flush 500ms, eval cache 1s)
//   - BYO:     FLAGR_SERVER_URL=http://host:18000 go test -tags=integration ./integration_tests/
//   - Docker:  cd integration_tests && make test
//     (builds binary, runs against all compose instances — see README.md for which tests run on legacy checkr/flagr:1.1.12)
//
// TestIntegration_Exposures asserts POST /exposures recording via loggedCount (no test-process FLAGR_RECORDER_ENABLED).
//
// Built-in context tests (TestIntegration_BuiltInContext, TestIntegration_BuiltInContextHTTPHeader)
// use isLegacyIntegrationBaseline() to skip on checkr/flagr:1.1.12 — the /evaluation route exists
// on legacy but the server does not inject @ts/@http_* keys, so constraints would never match.
// Only the current Flagr image supports built-in context injection (FLAGR_INJECTED_CONTEXT_ENABLED).

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
	requireFlagSnapshotMaxIDAPI(t)
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
	requireDuplicateFlagAPI(t)
	requireFlagSnapshotMaxIDAPI(t)
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
	requireDuplicateFlagAPI(t)
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

func TestIntegration_EvalCacheExportQuery(t *testing.T) {
	if isLegacyIntegrationBaseline() {
		t.Skip("eval cache query params not available on legacy checkr/flagr:1.1.12")
	}

	// Create test flags with specific states
	var enabledFlag flagResponse
	postJSON(t, "/api/v1/flags", map[string]any{
		"key":         fmt.Sprintf("export_query_enabled_%d", time.Now().UnixNano()),
		"description": "enabled flag for export query",
		"enabled":     true,
	}, &enabledFlag)

	var disabledFlag flagResponse
	postJSON(t, "/api/v1/flags", map[string]any{
		"key":         fmt.Sprintf("export_query_disabled_%d", time.Now().UnixNano()),
		"description": "disabled flag for export query",
		"enabled":     false,
	}, &disabledFlag)

	// Add tags
	tagVal1 := fmt.Sprintf("export_tag_a_%d", time.Now().UnixNano())
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", enabledFlag.ID), map[string]any{"value": tagVal1}, nil)
	tagVal2 := fmt.Sprintf("export_tag_b_%d", time.Now().UnixNano())
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", enabledFlag.ID), map[string]any{"value": tagVal2}, nil)
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", disabledFlag.ID), map[string]any{"value": tagVal2}, nil)

	// Wait for eval cache to pick up the new flags by polling for the enabled flag by ID.
	// Can't just check len >= 2 because seeded flags already satisfy that.
	err := pollUntil("eval cache export", "/api/v1/export/eval_cache/json", 10*time.Second, func() bool {
		var cache struct{ Flags []flagResponse }
		getJSON(t, fmt.Sprintf("/api/v1/export/eval_cache/json?ids=%d", enabledFlag.ID), &cache)
		return len(cache.Flags) == 1
	})
	if err != nil {
		t.Fatalf("eval cache did not contain new flags: %v", err)
	}

	t.Run("no params returns all flags", func(t *testing.T) {
		var cache struct{ Flags []flagResponse }
		getJSON(t, "/api/v1/export/eval_cache/json", &cache)
		if len(cache.Flags) < 2 {
			t.Fatalf("expected at least 2 flags, got %d", len(cache.Flags))
		}
	})

	t.Run("filter by ids", func(t *testing.T) {
		var cache struct{ Flags []flagResponse }
		getJSON(t, fmt.Sprintf("/api/v1/export/eval_cache/json?ids=%d", enabledFlag.ID), &cache)
		if len(cache.Flags) != 1 {
			t.Fatalf("expected 1 flag, got %d", len(cache.Flags))
		}
		if cache.Flags[0].ID != enabledFlag.ID {
			t.Errorf("expected flag ID %d, got %d", enabledFlag.ID, cache.Flags[0].ID)
		}
	})

	t.Run("filter by multiple ids", func(t *testing.T) {
		var cache struct{ Flags []flagResponse }
		getJSON(t, fmt.Sprintf("/api/v1/export/eval_cache/json?ids=%d,%d", enabledFlag.ID, disabledFlag.ID), &cache)
		if len(cache.Flags) != 2 {
			t.Fatalf("expected 2 flags, got %d", len(cache.Flags))
		}
	})

	t.Run("filter by enabled true", func(t *testing.T) {
		var cache struct{ Flags []flagResponse }
		getJSON(t, "/api/v1/export/eval_cache/json?enabled=true", &cache)
		for _, f := range cache.Flags {
			if !f.Enabled {
				t.Errorf("expected flag %d to be enabled", f.ID)
			}
		}
	})

	t.Run("filter by enabled false", func(t *testing.T) {
		var cache struct{ Flags []flagResponse }
		getJSON(t, "/api/v1/export/eval_cache/json?enabled=false", &cache)
		found := false
		for _, f := range cache.Flags {
			if f.ID == disabledFlag.ID {
				found = true
			}
			if f.Enabled {
				t.Errorf("expected flag %d to be disabled", f.ID)
			}
		}
		if !found {
			t.Error("expected to find the disabled flag")
		}
	})

	t.Run("filter by tags ANY", func(t *testing.T) {
		var cache struct{ Flags []flagResponse }
		getJSON(t, fmt.Sprintf("/api/v1/export/eval_cache/json?tags=%s", tagVal1), &cache)
		if len(cache.Flags) != 1 {
			t.Fatalf("expected 1 flag with tag1, got %d", len(cache.Flags))
		}
	})

	t.Run("filter by tags ALL", func(t *testing.T) {
		var cache struct{ Flags []flagResponse }
		getJSON(t, fmt.Sprintf("/api/v1/export/eval_cache/json?tags=%s,%s&tagsOperator=ALL", tagVal1, tagVal2), &cache)
		if len(cache.Flags) != 1 {
			t.Fatalf("expected 1 flag with both tags, got %d", len(cache.Flags))
		}
		if cache.Flags[0].ID != enabledFlag.ID {
			t.Errorf("expected flag ID %d, got %d", enabledFlag.ID, cache.Flags[0].ID)
		}
	})

	t.Run("ids override other filters", func(t *testing.T) {
		var cache struct{ Flags []flagResponse }
		// disabled flag ID with enabled=true should still return it
		getJSON(t, fmt.Sprintf("/api/v1/export/eval_cache/json?ids=%d&enabled=true", disabledFlag.ID), &cache)
		if len(cache.Flags) != 1 {
			t.Fatalf("expected 1 flag (ids override), got %d", len(cache.Flags))
		}
	})

	t.Run("no match returns empty", func(t *testing.T) {
		var cache struct{ Flags []flagResponse }
		getJSON(t, "/api/v1/export/eval_cache/json?ids=999999", &cache)
		if len(cache.Flags) != 0 {
			t.Fatalf("expected 0 flags, got %d", len(cache.Flags))
		}
	})
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

	requireOptionalAPI(t, http.MethodPost, "/api/v1/exposures", map[string]any{
		"exposures": []map[string]any{{"flagID": flagID, "entityID": "exposure-probe"}},
	}, "POST /api/v1/exposures")

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

	// Datar is only enabled on flagr_integration_tests images (see README.md).
	requireOptionalAPI(t, http.MethodGet, "/api/v1/datar/summary", nil, "GET /api/v1/datar/summary")
	resp, err := doReq("GET", "/api/v1/datar/summary", nil)
	if err != nil {
		if isLegacyIntegrationBaseline() {
			t.Skipf("datar not available: %v", err)
		}
		t.Fatalf("GET /api/v1/datar/summary: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	requireRecorderEndpointOK(t, resp.StatusCode, body, "GET /api/v1/datar/summary")
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

	requireOptionalAPI(t, http.MethodGet, fmt.Sprintf("/api/v1/datar/flags/%d/summary", flagID), nil, "GET /api/v1/datar/flags/{flagID}/summary")
	resp, err := doReq("GET", fmt.Sprintf("/api/v1/datar/flags/%d/summary", flagID), nil)
	if err != nil {
		if isLegacyIntegrationBaseline() {
			t.Skipf("datar not available: %v", err)
		}
		t.Fatalf("GET /api/v1/datar/flags/%d/summary: %v", flagID, err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	requireRecorderEndpointOK(t, resp.StatusCode, body, "GET /api/v1/datar/flags/{flagID}/summary")
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

// Built-in context integration test constants.
const (
	// builtinCtxEvalCacheWait waits for the eval cache to pick up newly created
	// flags/constraints. The local auto-start server uses 1s refresh
	// (integrationEvalCacheRefresh), but docker-compose backends use the default
	// 3s (FLAGR_EVALCACHE_REFRESHINTERVAL envDefault). Use 2× the default to
	// account for DB latency on postgres/mysql/etc.
	builtinCtxEvalCacheWait = 6 * time.Second

	builtinCtxRolloutPercent = 100  // full rollout for constraint-only segments
	builtinCtxVariantOn      = "on"
	builtinCtxVariantEnabled = "enabled"
	builtinCtxEnvHeaderValue = "test"
)

func TestIntegration_BuiltInContext(t *testing.T) {
	// Built-in context injection (@ts, @ts_hour, @ts_weekday, @ts_month) is a
	// current-Flagr-only feature. The /evaluation route exists on legacy
	// checkr/flagr:1.1.12 but the server does not inject @ts keys, so the
	// constraint would never match. Skip on legacy baseline.
	if isLegacyIntegrationBaseline() {
		t.Skip("built-in context injection not available on legacy checkr/flagr:1.1.12")
	}
	// The server is started with FLAGR_INJECTED_CONTEXT_ENABLED=true.
	// Create a flag with a @ts constraint (always matches since @ts >= 0)
	key := fmt.Sprintf("builtin_ctx_%d", time.Now().UnixNano())
	var created flagResponse
	postJSON(t, "/api/v1/flags", map[string]any{
		"key":         key,
		"description": "built-in context test flag",
	}, &created)
	if created.ID == 0 {
		t.Fatal("expected non-zero id")
	}

	// Enable the flag (flags are disabled by default)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/enabled", created.ID), map[string]any{
		"enabled": true,
	}, nil)

	// Create a variant
	var variant variantResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants", created.ID), map[string]any{
		"key": builtinCtxVariantOn,
	}, &variant)
	if variant.ID == 0 {
		t.Fatal("expected non-zero variant id")
	}

	// Create a segment with @ts constraint (always matches)
	var segment segmentResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments", created.ID), map[string]any{
		"description":    "ts constraint segment",
		"rolloutPercent": builtinCtxRolloutPercent,
	}, &segment)
	if segment.ID == 0 {
		t.Fatal("expected non-zero segment id")
	}

	// Create constraint: @ts GTE 0 (always true)
	var constraint constraintResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints", created.ID, segment.ID), map[string]any{
		"property": "@ts",
		"operator": "GTE",
		"value":    "0",
	}, &constraint)
	if constraint.ID == 0 {
		t.Fatal("expected non-zero constraint id")
	}

	// Create distribution: 100% to "on"
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", created.ID, segment.ID), map[string]any{
		"distributions": []map[string]any{
			{
				"variantID":  variant.ID,
				"variantKey": builtinCtxVariantOn,
				"percent":    builtinCtxRolloutPercent,
			},
		},
	}, nil)

	// Wait for eval cache to pick up the new flag
	time.Sleep(builtinCtxEvalCacheWait)

	// Evaluate — should match because @ts >= 0 is always true
	var evalResult evalResponse
	postJSON(t, "/api/v1/evaluation", map[string]any{
		"flagID":     created.ID,
		"entityID":   "builtin-ctx-entity",
		"entityType": "user",
		"entityContext": map[string]any{
			"country": "US",
		},
	}, &evalResult)

	if evalResult.VariantKey != builtinCtxVariantOn {
		t.Fatalf("expected variantKey 'on', got '%s' — @ts constraint may not be injected", evalResult.VariantKey)
	}

	// Cleanup
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d", created.ID))
}

func TestIntegration_BuiltInContextHTTPHeader(t *testing.T) {
	// Built-in HTTP header injection (@http_*) is a current-Flagr-only feature.
	// The /evaluation route exists on legacy checkr/flagr:1.1.12 but the server
	// does not inject @http_* keys, so the constraint would never match.
	if isLegacyIntegrationBaseline() {
		t.Skip("built-in HTTP header injection not available on legacy checkr/flagr:1.1.12")
	}
	// Create a flag with @http_x_environment constraint
	key := fmt.Sprintf("builtin_http_%d", time.Now().UnixNano())
	var created flagResponse
	postJSON(t, "/api/v1/flags", map[string]any{
		"key":         key,
		"description": "built-in HTTP header context test flag",
	}, &created)
	if created.ID == 0 {
		t.Fatal("expected non-zero id")
	}

	// Enable the flag (flags are disabled by default)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/enabled", created.ID), map[string]any{
		"enabled": true,
	}, nil)

	// Create variant
	var variant variantResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants", created.ID), map[string]any{
		"key": builtinCtxVariantEnabled,
	}, &variant)
	if variant.ID == 0 {
		t.Fatal("expected non-zero variant id")
	}

	// Create segment with @http_x_environment EQ "test"
	var segment segmentResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments", created.ID), map[string]any{
		"description":    "environment constraint segment",
		"rolloutPercent": builtinCtxRolloutPercent,
	}, &segment)
	if segment.ID == 0 {
		t.Fatal("expected non-zero segment id")
	}

	// Create constraint
	var constraint constraintResponse
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints", created.ID, segment.ID), map[string]any{
		"property": "@http_x_environment",
		"operator": "EQ",
		"value":    `"` + builtinCtxEnvHeaderValue + `"`,
	}, &constraint)
	if constraint.ID == 0 {
		t.Fatal("expected non-zero constraint id")
	}

	// Create distribution
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", created.ID, segment.ID), map[string]any{
		"distributions": []map[string]any{
			{
				"variantID":  variant.ID,
				"variantKey": builtinCtxVariantEnabled,
				"percent":    builtinCtxRolloutPercent,
			},
		},
	}, nil)
	// Wait for eval cache to pick up the new flag
	time.Sleep(builtinCtxEvalCacheWait)
	// Evaluate WITHOUT X-Environment header — should NOT match
	var evalResult evalResponse
	postJSON(t, "/api/v1/evaluation", map[string]any{
		"flagID":     created.ID,
		"entityID":   "http-header-entity",
		"entityType": "user",
		"entityContext": map[string]any{},
	}, &evalResult)

	if evalResult.VariantKey != "" {
		t.Fatalf("expected empty variantKey (no match), got '%s' — constraint should not match without X-Environment header", evalResult.VariantKey)
	}

	// Positive test: evaluate WITH X-Environment header set to "test" — should match
	var positiveResult evalResponse
	resp, err := doReqWithHeaders("POST", "/api/v1/evaluation", map[string]any{
		"flagID":     created.ID,
		"entityID":   "http-header-positive-entity",
		"entityType": "user",
		"entityContext": map[string]any{},
	}, map[string]string{"X-Environment": builtinCtxEnvHeaderValue})
	if err != nil {
		t.Fatalf("evaluation request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}
	if err := json.NewDecoder(resp.Body).Decode(&positiveResult); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if positiveResult.VariantKey != builtinCtxVariantEnabled {
		t.Fatalf("expected variantKey 'enabled', got '%s' — @http_x_environment should match with X-Environment: test", positiveResult.VariantKey)
	}

	// Cleanup
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d", created.ID))
}
