//go:build integration

package flagr_integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// mcpRequest sends a JSON-RPC request to /mcp and returns the decoded response.
// For notifications (method starts with "notifications/"), no id is sent and the
// response body may be empty — use mcpNotify for those.
func mcpRequest(t *testing.T, method string, params any, sessionID string) (map[string]any, string) {
	t.Helper()

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
	}
	if params != nil {
		msg["params"] = params
	}

	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal mcp request: %v", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/mcp", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("POST /mcp: %v", err)
	}
	defer resp.Body.Close()

	newSessionID := resp.Header.Get("Mcp-Session-Id")

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST /mcp %s: expected 200, got %d: %s", method, resp.StatusCode, string(b))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode mcp response: %v", err)
	}

	return result, newSessionID
}

// mcpNotify sends a JSON-RPC notification (no id, no response body expected).
func mcpNotify(t *testing.T, method string, sessionID string) {
	t.Helper()

	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal mcp notify: %v", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/mcp", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("POST /mcp: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /mcp %s: expected 200/202, got %d", method, resp.StatusCode)
	}
}

func mcpInit(t *testing.T) string {
	t.Helper()
	_, sessionID := mcpRequest(t, "initialize", map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "integration-test", "version": "0.0.1"},
	}, "")
	if sessionID == "" {
		t.Fatal("expected Mcp-Session-Id header in initialize response")
	}
	mcpNotify(t, "notifications/initialized", sessionID)
	return sessionID
}

func mcpCallTool(t *testing.T, sessionID, name string, args any) map[string]any {
	t.Helper()
	resp, _ := mcpRequest(t, "tools/call", map[string]any{
		"name":      name,
		"arguments": args,
	}, sessionID)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("tools/call %s: missing result: %v", name, resp)
	}
	return result
}

func TestIntegration_MCP_Initialize(t *testing.T) {
	requireMCPBackend(t)
	result, _ := mcpRequest(t, "initialize", map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "test", "version": "0.0.1"},
	}, "")

	serverInfo, ok := result["result"].(map[string]any)
	if !ok {
		t.Fatalf("initialize: missing result: %v", result)
	}
	info := serverInfo["serverInfo"].(map[string]any)
	if info["name"] != "flagr-mcp" {
		t.Fatalf("expected server name flagr-mcp, got %v", info["name"])
	}
}

func TestIntegration_MCP_ToolsList(t *testing.T) {
	requireMCPBackend(t)
	sessionID := mcpInit(t)

	resp, _ := mcpRequest(t, "tools/list", map[string]any{}, sessionID)
	result := resp["result"].(map[string]any)
	tools := result["tools"].([]any)

	if len(tools) < 20 {
		t.Fatalf("expected at least 20 tools, got %d", len(tools))
	}

	// Verify key tools exist.
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.(map[string]any)["name"].(string)] = true
	}
	for _, want := range []string{"create_flag", "get_flag", "list_flags", "evaluate_flag", "health"} {
		if !names[want] {
			t.Errorf("missing tool: %s", want)
		}
	}
}

func TestIntegration_MCP_FlagCRUD(t *testing.T) {
	requireMCPBackend(t)
	sessionID := mcpInit(t)

	// Create a flag via MCP.
	createResult := mcpCallTool(t, sessionID, "create_flag", map[string]any{
		"key":         "mcp-integration-flag",
		"description": "created by MCP integration test",
	})
	createText := createResult["content"].([]any)[0].(map[string]any)["text"].(string)
	var created map[string]any
	if err := json.Unmarshal([]byte(createText), &created); err != nil {
		t.Fatalf("decode create_flag result: %v", err)
	}
	flagID := int(created["id"].(float64))
	if flagID == 0 {
		t.Fatal("expected non-zero flag ID")
	}

	// Get the flag via MCP.
	getResult := mcpCallTool(t, sessionID, "get_flag", map[string]any{"flag_id": flagID})
	getText := getResult["content"].([]any)[0].(map[string]any)["text"].(string)
	var got map[string]any
	if err := json.Unmarshal([]byte(getText), &got); err != nil {
		t.Fatalf("decode get_flag result: %v", err)
	}
	if got["key"] != "mcp-integration-flag" {
		t.Fatalf("expected key mcp-integration-flag, got %v", got["key"])
	}

	// Verify the flag also appears in the HTTP API.
	var flags []flagResponse
	getJSON(t, "/api/v1/flags", &flags)
	found := false
	for _, f := range flags {
		if f.Key == "mcp-integration-flag" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("flag created via MCP not found in HTTP API")
	}

	// Clean up via MCP.
	mcpCallTool(t, sessionID, "delete_flag", map[string]any{"flag_id": flagID})
}

func TestIntegration_MCP_Disabled(t *testing.T) {
	// Start a server with MCP disabled (the default).
	projectRoot := findProjectRoot()
	binPath := filepath.Join(t.TempDir(), "flagr")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/flagr-server/")
	build.Dir = projectRoot
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("build: %v", err)
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	cmd := exec.Command(binPath, "--port", strconv.Itoa(port))
	cmd.Env = append(os.Environ(),
		"FLAGR_DB_DBDRIVER=sqlite3",
		"FLAGR_DB_DBCONNECTIONSTR=file::memory:?cache=shared",
		"FLAGR_LOGRUS_LEVEL=warn",
	)
	logFile, _ := os.CreateTemp("", "flagr-mcp-disabled-*.log")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	if err := pollUntil("server", base, 10*time.Second, func() bool {
		resp, err := http.Get(base + "/api/v1/health")
		if err != nil {
			return false
		}
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}); err != nil {
		t.Fatalf("server not ready: %v", err)
	}

	// POST to /mcp should return 404 — route does not exist.
	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]any{"protocolVersion": "2025-03-26"},
	})
	resp, err := http.Post(base+"/mcp", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /mcp: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected /mcp to be unavailable when FLAGR_MCP_ENABLED=false, got 200: %s", string(b))
	}
	t.Logf("/mcp returned %d when MCP disabled — correct", resp.StatusCode)
}
