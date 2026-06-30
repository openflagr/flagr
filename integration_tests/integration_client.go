//go:build integration

package flagr_integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// endpointAvailable returns false when the route is missing (404), e.g. on checkr/flagr:1.1.12.
func endpointAvailable(t *testing.T, method, path string, body any) bool {
	t.Helper()
	resp, err := doReq(method, path, body)
	if err != nil {
		return false
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode != http.StatusNotFound
}

func requireFlagSnapshotMaxIDAPI(t *testing.T) {
	t.Helper()
	if endpointAvailable(t, "GET", "/api/v1/flags/snapshots/max_id", nil) {
		return
	}
	t.Skip("GET /flags/snapshots/max_id not available on this server (e.g. checkr/flagr:1.1.12)")
}

func requireDuplicateFlagAPI(t *testing.T) {
	t.Helper()
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	path := fmt.Sprintf("/api/v1/flags/%d/duplicate", seedFlagIDs[0])
	if endpointAvailable(t, "POST", path, map[string]any{}) {
		return
	}
	t.Skip("POST /flags/{flagID}/duplicate not available on this server (e.g. checkr/flagr:1.1.12)")
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

// doReqAndDecode performs an HTTP request, checks for 2xx, and optionally decodes JSON into dst.
func doReqAndDecode(method, path string, body, dst any, errorf func(string, ...any)) {
	resp, err := doReq(method, path, body)
	if err != nil {
		errorf("%s %s: %v", method, path, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		errorf("%s %s: expected 2xx, got %d: %s", method, path, resp.StatusCode, string(b))
		return
	}
	if dst != nil {
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			errorf("decode %s %s: %v", method, path, err)
			return
		}
	}
}

func getJSON(t *testing.T, path string, dst any) {
	t.Helper()
	doReqAndDecode("GET", path, nil, dst, t.Fatalf)
}

func postJSON(t *testing.T, path string, body, dst any) {
	t.Helper()
	doReqAndDecode("POST", path, body, dst, t.Fatalf)
}

// postJSONExpectStatus POSTs JSON and requires an exact HTTP status (dst may be nil).
func postJSONExpectStatus(t *testing.T, path string, body any, wantStatus int, dst any) {
	t.Helper()
	resp, err := doReq("POST", path, body)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST %s: expected status %d, got %d: %s", path, wantStatus, resp.StatusCode, string(b))
	}
	if dst != nil {
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			t.Fatalf("decode POST %s: %v", path, err)
		}
	}
}

func putJSON(t *testing.T, path string, body, dst any) {
	t.Helper()
	doReqAndDecode("PUT", path, body, dst, t.Fatalf)
}

func deleteResource(t *testing.T, path string) {
	t.Helper()
	doReqAndDecode("DELETE", path, nil, nil, t.Fatalf)
}

// doReqOK performs an HTTP request and verifies a 2xx status, discarding the body.
func doReqOK(t *testing.T, method, path string, body any) {
	t.Helper()
	resp, err := doReq(method, path, body)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("%s %s: expected 2xx, got %d", method, path, resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Polling helper
// ---------------------------------------------------------------------------

// pollUntil calls check every pollInterval until it returns true or timeout expires.
// Returns an error on timeout.
func pollUntil(name, url string, timeout time.Duration, check func() bool) error {
	deadline := time.After(timeout)
	for {
		if check() {
			return nil
		}
		select {
		case <-deadline:
			return fmt.Errorf("%s at %s not ready after %v", name, url, timeout)
		default:
			time.Sleep(pollInterval)
		}
	}
}

type snapshotMaxIDResponse struct {
	MaxID int64 `json:"maxID"`
}

func getSnapshotMaxID(t *testing.T) int64 {
	t.Helper()
	var out snapshotMaxIDResponse
	getJSON(t, "/api/v1/flags/snapshots/max_id", &out)
	return out.MaxID
}

func countFlagSnapshots(t *testing.T, flagID int64) int {
	t.Helper()
	var snaps []json.RawMessage
	getJSON(t, fmt.Sprintf("/api/v1/flags/%d/snapshots", flagID), &snaps)
	return len(snaps)
}

// ---------------------------------------------------------------------------
// API response types
// ---------------------------------------------------------------------------

type flagResponse struct {
	ID                 int64              `json:"id"`
	Key                string             `json:"key"`
	Description        string             `json:"description"`
	Enabled            bool               `json:"enabled"`
	EntityType         string             `json:"entityType"`
	DataRecordsEnabled bool               `json:"dataRecordsEnabled"`
	Segments           []segmentResponse  `json:"segments"`
	Variants           []variantResponse  `json:"variants"`
}

type segmentResponse struct {
	ID             int64                  `json:"id"`
	Description    string                 `json:"description"`
	RolloutPercent int                    `json:"rolloutPercent"`
	Rank           int                    `json:"rank"`
	Distributions  []distributionResponse `json:"distributions"`
}

type variantResponse struct {
	ID         int64          `json:"id"`
	Key        string         `json:"key"`
	Attachment map[string]any `json:"attachment"`
}

type constraintResponse struct {
	ID       int64  `json:"id"`
	Property string `json:"property"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type distributionResponse struct {
	VariantID  int64   `json:"variantID"`
	VariantKey string  `json:"variantKey"`
	Percent    float64 `json:"percent"`
}

type tagResponse struct {
	ID    int64  `json:"id"`
	Value string `json:"value"`
}

type evalResponse struct {
	FlagID      int64          `json:"flagID"`
	FlagKey     string         `json:"flagKey"`
	VariantKey  string         `json:"variantKey"`
	EvalContext map[string]any `json:"evalContext"`
}

type batchEvalResponse struct {
	EvaluationResults []evalResponse `json:"evaluationResults"`
}
