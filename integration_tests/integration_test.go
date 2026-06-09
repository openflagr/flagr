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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Package-level state — set by TestMain, read by tests and benchmarks
// ---------------------------------------------------------------------------

var (
	baseURL      string
	seedFlagIDs  []int64
	seedFlagKeys []string
	httpClient   = &http.Client{Timeout: 10 * time.Second}
)

// ---------------------------------------------------------------------------
// TestMain — entry point
// ---------------------------------------------------------------------------

func TestMain(m *testing.M) {
	url := os.Getenv("FLAGR_SERVER_URL")
	if url == "" {
		url = startLocalServer()
	}
	baseURL = url

	waitForServer(baseURL, 30*time.Second)
	seedFlags(baseURL)
	waitForEvalCache(baseURL, 15*time.Second)

	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Local server lifecycle
// ---------------------------------------------------------------------------

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("cannot getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatal("cannot find project root (go.mod)")
		}
		dir = parent
	}
}

func startLocalServer() string {
	projectRoot := findProjectRoot()

	// Check for pre-built binary
	binPath := filepath.Join(projectRoot, "flagr")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		tmpDir, err := os.MkdirTemp("", "flagr-integration-*")
		if err != nil {
			log.Fatalf("failed to create temp dir: %v", err)
		}
		binPath = filepath.Join(tmpDir, "flagr")
		cmd := exec.Command("go", "build", "-o", binPath, "./swagger_gen/cmd/flagr-server/")
		cmd.Dir = projectRoot
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("failed to build server binary: %v", err)
		}
	}

	// Find a free port
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("cannot find free port: %v", err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	// Start server subprocess
	cmd := exec.Command(binPath,
		"--port", strconv.Itoa(port),
	)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"FLAGR_DB_DBDRIVER=sqlite3",
		"FLAGR_DB_DBCONNECTIONSTR=file::memory:?cache=shared",
	)
	serverLog, err := os.Create(filepath.Join(os.TempDir(), "flagr-integration-server.log"))
	if err != nil {
		log.Fatalf("cannot create server log: %v", err)
	}
	cmd.Stdout = serverLog
	cmd.Stderr = serverLog
	if err := cmd.Start(); err != nil {
		log.Fatalf("cannot start server: %v", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	return url
}

func waitForServer(url string, timeout time.Duration) {
	deadline := time.After(timeout)
	healthURL := url + "/api/v1/health"
	for {
		resp, err := http.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Println("Server is healthy at", url)
			return
		}
		select {
		case <-deadline:
			log.Fatalf("server at %s not healthy after %v", url, timeout)
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func waitForEvalCache(url string, timeout time.Duration) {
	deadline := time.After(timeout)
	evalURL := url + "/api/v1/export/eval_cache/json"
	for {
		resp, err := http.Get(evalURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				fmt.Println("Eval cache ready")
				return
			}
		}
		select {
		case <-deadline:
			fmt.Println("Warning: eval cache not ready within timeout, continuing anyway")
			return
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func doReq(method, path string, body any) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("json marshal: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return httpClient.Do(req)
}

func requireOK(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("expected 2xx, got %d: %s", resp.StatusCode, string(body))
	}
}

func requireStatus(t *testing.T, resp *http.Response, code int) {
	t.Helper()
	if resp.StatusCode != code {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("expected %d, got %d: %s", code, resp.StatusCode, string(body))
	}
}

func getJSON(t *testing.T, path string, dst any) {
	t.Helper()
	resp, err := doReq("GET", path, nil)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	requireOK(t, resp)
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode GET %s response: %v", path, err)
	}
}

func postJSON(t *testing.T, path string, body, dst any) {
	t.Helper()
	resp, err := doReq("POST", path, body)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	requireOK(t, resp)
	if dst != nil {
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			t.Fatalf("decode POST %s response: %v", path, err)
		}
	}
}

func putJSON(t *testing.T, path string, body, dst any) {
	t.Helper()
	resp, err := doReq("PUT", path, body)
	if err != nil {
		t.Fatalf("PUT %s: %v", path, err)
	}
	defer resp.Body.Close()
	requireOK(t, resp)
	if dst != nil {
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			t.Fatalf("decode PUT %s response: %v", path, err)
		}
	}
}

func deleteResource(t *testing.T, path string) {
	t.Helper()
	resp, err := doReq("DELETE", path, nil)
	if err != nil {
		t.Fatalf("DELETE %s: %v", path, err)
	}
	resp.Body.Close()
	requireOK(t, resp)
}

// seed helpers (non-test, use log.Fatal)

func doReqOrDie(method, path string, body, dst any) {
	resp, err := doReq(method, path, body)
	if err != nil {
		log.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		log.Fatalf("%s %s: expected 2xx, got %d: %s", method, path, resp.StatusCode, string(b))
	}
	if dst != nil {
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			log.Fatalf("decode %s %s response: %v", method, path, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Flag seeding
// ---------------------------------------------------------------------------

func seedFlags(serverURL string) {
	for _, f := range allFlagDefs {
		// Create flag
		desc := f.Description
		var flag struct {
			ID         int64  `json:"id"`
			Key        string `json:"key"`
			EntityType string `json:"entityType"`
			Enabled    bool   `json:"enabled"`
		}
		doReqOrDie("POST", "/api/v1/flags", map[string]any{
			"description": desc,
			"key":         f.Key,
			"enabled":     f.Enabled,
		}, &flag)
		seedFlagIDs = append(seedFlagIDs, flag.ID)
		seedFlagKeys = append(seedFlagKeys, flag.Key)

		// Set entity type if custom
		if f.EntityType != "" && f.EntityType != "user" {
			doReqOrDie("PUT", fmt.Sprintf("/api/v1/flags/%d", flag.ID), map[string]any{
				"entityType": f.EntityType,
			}, nil)
		}

		// Create segment (100% rollout)
		var seg struct {
			ID int64 `json:"id"`
		}
		doReqOrDie("POST", fmt.Sprintf("/api/v1/flags/%d/segments", flag.ID), map[string]any{
			"description":    "default segment",
			"rolloutPercent": 100,
		}, &seg)

		// Create constraints
		for _, c := range f.Constraints {
			doReqOrDie("POST", fmt.Sprintf("/api/v1/flags/%d/segments/%d/constraints", flag.ID, seg.ID), map[string]any{
				"property": c.Property,
				"operator": c.Operator,
				"value":    c.Value,
			}, nil)
		}

		// Create variants
		var v1, v2 struct {
			ID  int64  `json:"id"`
			Key string `json:"key"`
		}
		doReqOrDie("POST", fmt.Sprintf("/api/v1/flags/%d/variants", flag.ID), map[string]any{
			"key": "variant_control",
		}, &v1)
		doReqOrDie("POST", fmt.Sprintf("/api/v1/flags/%d/variants", flag.ID), map[string]any{
			"key": "variant_treatment",
		}, &v2)

		// Set distributions (100% to variant_control)
		doReqOrDie("PUT", fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", flag.ID, seg.ID), map[string]any{
			"distributions": []map[string]any{
				{"percent": 100, "variantID": v1.ID, "variantKey": v1.Key},
				{"percent": 0, "variantID": v2.ID, "variantKey": v2.Key},
			},
		}, nil)

		// Create tags
		for _, tag := range f.Tags {
			doReqOrDie("POST", fmt.Sprintf("/api/v1/flags/%d/tags", flag.ID), map[string]any{
				"value": tag,
			}, nil)
		}
	}

	fmt.Printf("Seeded %d flags\n", len(seedFlagIDs))
}

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

	// Put flag
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), map[string]any{
		"description": "updated description",
		"key":         key,
	}, &fetched)

	// Set enabled (PUT)
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/enabled", flagID), map[string]any{
		"enabled": true,
	}, &fetched)

	// Get flag entity types
	var types []any
	getJSON(t, "/api/v1/flags/entity_types", &types)

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

	// Put segment
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d", flagID, segID), map[string]any{
		"description":    "updated segment",
		"rolloutPercent": 100,
	}, nil)
	// Reorder segments
	var flagObj map[string]any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flagObj)
	_ = flagObj

	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/reorder", flagID), map[string]any{
		"segmentIDs": []int64{segID},
	}, nil)
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

	// Update variant
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/variants/%d", flagID, variantID), map[string]any{
		"key": "test_variant_updated",
	}, &v)

	// Delete variant
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/variants/%d", flagID, variantID))
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

	// Put distributions
	putJSON(t, fmt.Sprintf("/api/v1/flags/%d/segments/%d/distributions", flagID, segID), map[string]any{
		"distributions": []map[string]any{
			{"percent": 100, "variantID": variantID, "variantKey": variantKey},
		},
	}, nil)

	// Verify by getting the flag (includes segments with distributions)
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flag)
}

func TestIntegration_Evaluation(t *testing.T) {
	if len(seedFlagIDs) == 0 || len(seedFlagKeys) == 0 {
		t.Fatal("no seeded flags available")
	}
	flagID := seedFlagIDs[0]
	flagKey := seedFlagKeys[0]

	eval := func(body map[string]any) map[string]any {
		var result map[string]any
		postJSON(t, "/api/v1/evaluation", body, &result)
		return result
	}

	// Eval by flagID
	result := eval(map[string]any{
		"flagID":    flagID,
		"entityID":  "test-entity",
		"entityType": "user",
		"entityContext": map[string]any{
			"region": "us-west",
			"age":    30,
		},
	})
	if result["evalContext"] == nil {
		t.Fatal("eval response missing evalContext")
	}

	// Eval by flagKey
	result = eval(map[string]any{
		"flagKey":   flagKey,
		"entityID":  "test-entity",
		"entityType": "user",
		"entityContext": map[string]any{
			"region": "us-west",
		},
	})
	if result["evalContext"] == nil {
		t.Fatal("eval response missing evalContext")
	}
	// Eval with entity type override
	var flag map[string]any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d", flagID), &flag)
	entityType := "user"
	if et, ok := flag["entityType"].(string); ok && et != "" {
		entityType = et
	}
	result = eval(map[string]any{
		"flagID":    flagID,
		"entityID":  "test-entity",
		"entityType": entityType,
	})
	if result["evalContext"] == nil {
		t.Fatal("eval response missing evalContext")
	}
}

func TestIntegration_Preload(t *testing.T) {
	// Get flags without preload
	var without []map[string]any
	getJSON(t, "/api/v1/flags", &without)

	// Get flags with preload
	var with []map[string]any
	getJSON(t, "/api/v1/flags?preload=true", &with)

	if len(with) == 0 {
		t.Fatal("expected at least one flag")
	}
	_ = without
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

	// Export eval cache json (returns {"flags": [...]})
	var cache map[string]any
	getJSON(t, "/api/v1/export/eval_cache/json", &cache)
	if cache["flags"] == nil {
		t.Log("eval cache json has no flags key")
	}
}

func TestIntegration_TagCRUD(t *testing.T) {
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	flagID := seedFlagIDs[0]

	// Create a tag
	var tag map[string]any
	postJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), map[string]any{
		"value": fmt.Sprintf("tag_crud_%d", time.Now().UnixNano()),
	}, &tag)
	tagID := int64(tag["id"].(float64))

	// List tags on flag
	var tags []any
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/tags", flagID), &tags)

	// Delete tag
	deleteResource(t, fmt.Sprintf("/api/v1/flags/%d/tags/%d", flagID, tagID))

	// List all tags
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
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkEvalByFlagID(b *testing.B) {
	if len(seedFlagIDs) == 0 {
		b.Skip("no seeded flags")
	}
	flagID := seedFlagIDs[0]
	body := fmt.Sprintf(`{"flagID":%d,"entityID":"bench-user","entityType":"user","entityContext":{"region":"us-west","age":30}}`, flagID)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalByFlagKey(b *testing.B) {
	if len(seedFlagKeys) == 0 {
		b.Skip("no seeded flags")
	}
	flagKey := seedFlagKeys[0]
	body := fmt.Sprintf(`{"flagKey":"%s","entityID":"bench-user","entityType":"user","entityContext":{"region":"us-west","age":30}}`, flagKey)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalNoMatch(b *testing.B) {
	if len(seedFlagIDs) == 0 {
		b.Skip("no seeded flags")
	}
	flagID := seedFlagIDs[0]
	body := fmt.Sprintf(`{"flagID":%d,"entityID":"bench-user","entityType":"user","entityContext":{"region":"nonexistent"}}`, flagID)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalDisabledFlag(b *testing.B) {
	id := seedFlagIDs[0] // fallback to first flag
	body := fmt.Sprintf(`{"flagID":%d,"entityID":"bench-user","entityType":"user"}`, id)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalBatchByIDs(b *testing.B) {
	if len(seedFlagIDs) < 5 {
		b.Skip("need at least 5 seeded flags")
	}
	ids := seedFlagIDs[:5]
	body := fmt.Sprintf(
		`{"entities":[{"entityID":"bench","entityType":"user","entityContext":{"region":"us-west","age":30}}],"flagIDs":%s}`,
		jsonInts(ids),
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation/batch", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalBatchByTags(b *testing.B) {
	body := `{"entities":[{"entityID":"bench","entityType":"user","entityContext":{"region":"us-west"}}],"flagTags":["int_test"],"flagTagsOperator":"ANY"}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation/batch", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalBatchLarge(b *testing.B) {
	if len(seedFlagIDs) < 10 {
		b.Skip("need at least 10 seeded flags")
	}
	body := fmt.Sprintf(
		`{"entities":[{"entityID":"bench0","entityType":"user","entityContext":{"region":"us-west"}},{"entityID":"bench1","entityType":"user","entityContext":{"region":"us-east"}},{"entityID":"bench2","entityType":"user","entityContext":{"region":"eu-west"}},{"entityID":"bench3","entityType":"user","entityContext":{"region":"ap-northeast"}},{"entityID":"bench4","entityType":"user","entityContext":{"region":"us-west"}}],"flagIDs":%s}`,
		jsonInts(seedFlagIDs[:10]),
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation/batch", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalEQ(b *testing.B) {
	body := `{"flagKey":"int_flag_EQ_01","entityID":"bench","entityType":"user","entityContext":{"region":"us-west"}}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalIN(b *testing.B) {
	body := `{"flagKey":"int_flag_IN_01","entityID":"bench","entityType":"user","entityContext":{"region":"us-west"}}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalRegex(b *testing.B) {
	body := `{"flagKey":"int_flag_EREG_01","entityID":"bench","entityType":"user","entityContext":{"email":"user@company.com"}}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalMultiConstraint(b *testing.B) {
	body := `{"flagKey":"int_flag_multi_segment","entityID":"bench","entityType":"user","entityContext":{"region":"us-west","age":25,"tier":"premium"}}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEvalNestedContext(b *testing.B) {
	body := `{"flagKey":"int_flag_complex_and","entityID":"bench","entityType":"user","entityContext":{"user":{"name":"Alice","age":30,"tier":"premium"},"region":"us-west"}}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/api/v1/evaluation", "application/json", strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// ---------------------------------------------------------------------------
// Utility helpers
// ---------------------------------------------------------------------------

func jsonInts(ids []int64) string {
	if len(ids) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteByte('[')
	for i, id := range ids {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatInt(id, 10))
	}
	b.WriteByte(']')
	return b.String()
}
