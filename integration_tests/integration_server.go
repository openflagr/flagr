//go:build integration

package flagr_integration

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	serverCmd    *exec.Cmd
	httpClient   = &http.Client{Timeout: 10 * time.Second}
)

// ---------------------------------------------------------------------------
// TestMain — entry point
// ---------------------------------------------------------------------------
func TestMain(m *testing.M) {
	url := os.Getenv("FLAGR_SERVER_URL")
	if url == "" {
		url = startLocalServer()
		defer func() {
			if serverCmd != nil {
				serverCmd.Process.Signal(os.Interrupt)
				serverCmd.Wait()
			}
		}()
	}
	baseURL = url

	waitForServer(baseURL, 30*time.Second)
	seedFlags(baseURL)
	waitForEvalCache(baseURL, 5*time.Second)

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

	// Always build fresh to a temp directory to avoid stale-binary bugs.
	tmpDir, err := os.MkdirTemp("", "flagr-integration-*")
	if err != nil {
		log.Fatalf("failed to create temp dir: %v", err)
	}
	binPath := filepath.Join(tmpDir, "flagr")
	cmd := exec.Command("go", "build", "-o", binPath, "./swagger_gen/cmd/flagr-server/")
	cmd.Dir = projectRoot
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to build server binary: %v", err)
	}

	// Find a free port
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("cannot find free port: %v", err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	// Start server subprocess
	serverCmd = exec.Command(binPath,
		"--port", strconv.Itoa(port),
	)
	serverCmd.Dir = projectRoot
	serverCmd.Env = append(os.Environ(),
		"FLAGR_DB_DBDRIVER=sqlite3",
		"FLAGR_DB_DBCONNECTIONSTR=file::memory:?cache=shared",
	)
	serverLog, err := os.Create(filepath.Join(os.TempDir(), "flagr-integration-server.log"))
	if err != nil {
		log.Fatalf("cannot create server log: %v", err)
	}
	serverCmd.Stdout = serverLog
	serverCmd.Stderr = serverLog
	if err := serverCmd.Start(); err != nil {
		log.Fatalf("cannot start server: %v", err)
	}

	return fmt.Sprintf("http://127.0.0.1:%d", port)
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
		if err == nil && resp.StatusCode == http.StatusOK {
			// Decode response to check for non-empty flags
			var cache struct {
				Flags []any `json:"Flags"`
			}
			err := json.NewDecoder(resp.Body).Decode(&cache)
			resp.Body.Close()
			if err == nil && len(cache.Flags) > 0 {
				fmt.Printf("Eval cache ready (%d flags)\n", len(cache.Flags))
				return
			}
		} else if err == nil {
			resp.Body.Close()
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
// Flag seeding
// ---------------------------------------------------------------------------

func seedFlags(serverURL string) {
	defs := initFlagDefs()
	for _, f := range defs {
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

		// Set enabled (POST creates flag as disabled by default)
		if f.Enabled {
			doReqOrDie("PUT", fmt.Sprintf("/api/v1/flags/%d/enabled", flag.ID), map[string]any{
				"enabled": true,
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
