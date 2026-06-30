//go:build integration

package flagr_integration

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// Compatibility helpers for the docker-compose matrix (see README.md).
//
// Skips apply only to the legacy checkr/flagr:1.1.12 instance (:18006). The five
// flagr_integration_tests backends must pass gated tests — missing routes or broken
// Datar/recorder config fails the run instead of skipping.

// currentFlagrAPICapability labels a route required by “current Flagr only” tests.
type currentFlagrAPICapability string

const (
	capSnapshotsMaxID currentFlagrAPICapability = "GET /api/v1/flags/snapshots/max_id"
	capDuplicateFlag  currentFlagrAPICapability = "POST /api/v1/flags/{flagID}/duplicate"
)

// legacyComposePort is checkr_flagr_with_sqlite in docker-compose.yml (see README.md).
const legacyComposePort = ":18006"

func isLegacyIntegrationBaseline() bool {
	return strings.HasSuffix(baseURL, legacyComposePort)
}

// responseIndicatesRouteNotRegistered is true for swagger/router 404s ("path … was not found"),
// not for application 404s (e.g. missing flag id) where the handler is registered.
func responseIndicatesRouteNotRegistered(status int, body []byte) bool {
	if status != http.StatusNotFound {
		return false
	}
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	msg := strings.ToLower(payload.Message)
	return strings.Contains(msg, "path ") && strings.Contains(msg, "was not found")
}

func probeHTTP(t *testing.T, method, path string, body any) (status int, respBody []byte, err error) {
	t.Helper()
	resp, err := doReq(method, path, body)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, _ = io.ReadAll(resp.Body)
	return resp.StatusCode, respBody, nil
}

func skipOrFailRouteProbe(t *testing.T, method, path string, body any, label string) {
	t.Helper()
	status, respBody, err := probeHTTP(t, method, path, body)
	if err != nil {
		if isLegacyIntegrationBaseline() {
			t.Skipf("%s: unreachable on legacy baseline: %v", label, err)
		}
		t.Fatalf("%s: request failed on current Flagr (must be reachable): %v", label, err)
	}
	if responseIndicatesRouteNotRegistered(status, respBody) {
		if isLegacyIntegrationBaseline() {
			t.Skipf("%s not registered on legacy checkr/flagr:1.1.12", label)
		}
		t.Fatalf("%s not registered on current Flagr image (regression): %d %s", label, status, truncateProbeBody(respBody))
	}
}

func truncateProbeBody(b []byte) string {
	const max = 200
	s := strings.TrimSpace(string(b))
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func requireCurrentFlagrAPI(t *testing.T, method, path string, body any, cap currentFlagrAPICapability) {
	t.Helper()
	skipOrFailRouteProbe(t, method, path, body, string(cap))
}

// requireFlagSnapshotMaxIDAPI gates tests that use getSnapshotMaxID (global snapshot id).
func requireFlagSnapshotMaxIDAPI(t *testing.T) {
	t.Helper()
	requireCurrentFlagrAPI(t, http.MethodGet, "/api/v1/flags/snapshots/max_id", nil, capSnapshotsMaxID)
}

// requireDuplicateFlagAPI gates duplicate-flag tests. Probes with a non-existent flag id so
// current servers return application 404 (route exists) without creating a clone.
func requireDuplicateFlagAPI(t *testing.T) {
	t.Helper()
	requireCurrentFlagrAPI(t, http.MethodPost, "/api/v1/flags/999999999/duplicate", map[string]any{}, capDuplicateFlag)
}

// requireOptionalAPI skips only on legacy when the route is not registered; on current images
// an unregistered route fails the test.
func requireOptionalAPI(t *testing.T, method, path string, body any, label string) {
	t.Helper()
	skipOrFailRouteProbe(t, method, path, body, label)
}

// requireRecorderEndpointOK skips on legacy when Datar/recorder returns non-200; on current
// compose images (Datar enabled) non-200 is a test failure.
func requireRecorderEndpointOK(t *testing.T, status int, respBody []byte, label string) {
	t.Helper()
	if status == http.StatusOK {
		return
	}
	if isLegacyIntegrationBaseline() {
		t.Skipf("%s returned %d on legacy baseline (recorder not enabled)", label, status)
	}
	t.Fatalf("%s returned %d on current Flagr (expected 200 with Datar enabled): %s", label, status, truncateProbeBody(respBody))
}