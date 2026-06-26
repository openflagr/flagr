//go:build integration

package flagr_integration

import (
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
	serverCmd    *exec.Cmd
	httpClient   = &http.Client{Timeout: 10 * time.Second}
)

// ---------------------------------------------------------------------------
// TestMain — entry point
// ---------------------------------------------------------------------------

func TestMain(m *testing.M) {
	// Multi-server mode: FLAGR_SERVER_URLS comma-separated
	if urlsStr := os.Getenv("FLAGR_SERVER_URLS"); urlsStr != "" {
		var urls []string
		for _, u := range strings.Split(urlsStr, ",") {
			u = strings.TrimSpace(u)
			if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
				u = "http://" + u
			}
			urls = append(urls, u)
		}
		exitCode := 0
		for _, u := range urls {
			fmt.Printf("=== Testing %s ===\n", u)
			baseURL = u
			seedFlagIDs = nil
			seedFlagKeys = nil
			prepareServer(u)
			if code := m.Run(); code != 0 {
				fmt.Printf("FAILED: %s\n", u)
				exitCode = code
			}
		}
		os.Exit(exitCode)
		return
	}

	// Single-server mode: FLAGR_SERVER_URL or auto-start local
	url := os.Getenv("FLAGR_SERVER_URL")
	if url == "" {
		url = startLocalServer()
		defer func() {
			if serverCmd != nil {
				serverCmd.Process.Kill()
				serverCmd.Wait()
			}
		}()
	}
	baseURL = url
	prepareServer(url)
	os.Exit(m.Run())
}

// prepareServer waits for a server to be healthy, seeds flags if needed, and waits for eval cache.
func prepareServer(url string) {
	waitForServer(url, 30*time.Second)
	// Check if flags already exist (idempotent for re-runs against same server).
	var existing []flagResponse
	getJSON := func(path string, dst any) {
		resp, err := doReq("GET", path, nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			json.NewDecoder(resp.Body).Decode(dst)
		}
	}
	getJSON("/api/v1/flags", &existing)
	if len(existing) == 0 {
		seedFlags(log.Fatalf)
	} else {
		fmt.Printf("Flags already exist at %s, skipping seed\n", url)
		for _, f := range existing {
			seedFlagIDs = append(seedFlagIDs, f.ID)
			seedFlagKeys = append(seedFlagKeys, f.Key)
		}
	}
	waitForEvalReady(url, 20*time.Second)
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
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/flagr-server/")
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
	tcpAddr, ok := lis.Addr().(*net.TCPAddr)
	if !ok {
		lis.Close()
		log.Fatalf("unexpected address type: %T", lis.Addr())
	}
	port := tcpAddr.Port
	lis.Close()

	// Start server subprocess
	serverCmd = exec.Command(binPath,
		"--port", strconv.Itoa(port),
	)
	serverCmd.Env = append(os.Environ(),
		"FLAGR_DB_DBDRIVER=sqlite3",
		"FLAGR_DB_DBCONNECTIONSTR=file::memory:?cache=shared",
		"FLAGR_RECORDER_ENABLED=true",
		"FLAGR_RECORDER_TYPE=datar",
	)
	// Redirect server output to a temp file to avoid "I/O incomplete" errors
	// when the test binary kills the server process.
	serverLog, err := os.CreateTemp("", "flagr-server-*.log")
	if err != nil {
		log.Fatalf("cannot create server log: %v", err)
	}
	serverCmd.Stdout = serverLog
	serverCmd.Stderr = serverLog
	serverCmd.WaitDelay = 5 * time.Second
	if err := serverCmd.Start(); err != nil {
		log.Fatalf("cannot start server: %v", err)
	}

	return fmt.Sprintf("http://127.0.0.1:%d", port)
}

func waitForServer(url string, timeout time.Duration) {
	if err := pollUntil("server", url, timeout, func() bool {
		resp, err := doReq("GET", "/api/v1/health", nil)
		if err != nil {
			return false
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}); err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Println("Server is healthy at", url)
}

// waitForEvalReady blocks until the eval cache has reloaded seeded data and
// evaluation endpoints return real results (not empty variantKey / batch results).
// Export-only checks are insufficient: the in-memory cache can lag behind DB
// commits on slower backends (e.g. postgres13 in CI).
func waitForEvalReady(url string, timeout time.Duration) {
	if len(seedFlagIDs) < 2 || len(seedFlagKeys) < 2 {
		log.Fatalf("eval readiness: need at least 2 seeded flags, got ids=%d keys=%d", len(seedFlagIDs), len(seedFlagKeys))
	}
	flagID := seedFlagIDs[1]
	flagKey := seedFlagKeys[1]

	if err := pollUntil("eval readiness", url, timeout, func() bool {
		var batch batchEvalResponse
		resp, err := doReq("POST", "/api/v1/evaluation/batch", map[string]any{
			"entities": []map[string]any{{
				"entityID":   "eval-ready-probe",
				"entityType": "user",
				"entityContext": map[string]any{
					"region": "us-west",
				},
			}},
			"flagTags":         []string{"int_test"},
			"flagTagsOperator": "ANY",
		})
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return false
		}
		if err := json.NewDecoder(resp.Body).Decode(&batch); err != nil {
			return false
		}
		if len(batch.EvaluationResults) == 0 {
			return false
		}
		batchHasVariant := false
		for _, r := range batch.EvaluationResults {
			if r.VariantKey != "" {
				batchHasVariant = true
				break
			}
		}
		if !batchHasVariant {
			return false
		}

		var single evalResponse
		resp2, err := doReq("POST", "/api/v1/evaluation", map[string]any{
			"flagID":     flagID,
			"entityID":   "eval-ready-probe",
			"entityType": "user",
			"entityContext": map[string]any{
				"tier": "premium",
			},
		})
		if err != nil {
			return false
		}
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusOK {
			return false
		}
		if err := json.NewDecoder(resp2.Body).Decode(&single); err != nil {
			return false
		}
		if single.VariantKey == "" {
			return false
		}

		resp3, err := doReq("POST", "/api/v1/evaluation", map[string]any{
			"flagKey":    flagKey,
			"entityID":   "eval-ready-probe",
			"entityType": "user",
			"entityContext": map[string]any{
				"tier": "premium",
			},
		})
		if err != nil {
			return false
		}
		defer resp3.Body.Close()
		if resp3.StatusCode != http.StatusOK {
			return false
		}
		if err := json.NewDecoder(resp3.Body).Decode(&single); err != nil {
			return false
		}
		return single.VariantKey != ""
	}); err != nil {
		log.Fatalf("%v", err)
	}
}