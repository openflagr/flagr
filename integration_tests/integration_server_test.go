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

// prepareServer waits for a server to be healthy, seeds flags, and waits for eval cache.
func prepareServer(url string) {
	waitForServer(url, 30*time.Second)
	seedFlags(log.Fatalf)
	waitForEvalCache(url, 5*time.Second)
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

func waitForEvalCache(url string, timeout time.Duration) {
	if err := pollUntil("eval cache", url, timeout, func() bool {
		resp, err := doReq("GET", "/api/v1/export/eval_cache/json", nil)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return false
		}
		var cache struct {
			Flags []any `json:"Flags"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&cache); err != nil {
			return false
		}
		return len(cache.Flags) > 0
	}); err != nil {
		log.Fatalf("%v", err)
	}
}
