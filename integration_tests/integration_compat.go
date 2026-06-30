//go:build integration

package flagr_integration

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

// Compatibility helpers for the docker-compose matrix (see README.md).
//
// Five backends run the image built from this repo; checkr_flagr_with_sqlite runs
// checkr/flagr:1.1.12 as a legacy baseline. Tests that depend on APIs added after
// that release must call requireCurrentFlagrAPI* so the legacy URL skips instead of failing CI.

// currentFlagrAPICapability labels a route required by “current Flagr only” tests.
type currentFlagrAPICapability string

const (
	capSnapshotsMaxID currentFlagrAPICapability = "GET /api/v1/flags/snapshots/max_id"
	capDuplicateFlag  currentFlagrAPICapability = "POST /api/v1/flags/{flagID}/duplicate"
)

func endpointAvailable(t *testing.T, method, path string, body any) bool {
	t.Helper()
	resp, err := doReq(method, path, body)
	if err != nil {
		return false
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode != http.StatusNotFound
}

func requireCurrentFlagrAPI(t *testing.T, method, path string, body any, cap currentFlagrAPICapability) {
	t.Helper()
	if endpointAvailable(t, method, path, body) {
		return
	}
	t.Skipf("%s not available on this server (legacy checkr/flagr:1.1.12 omits newer routes)", cap)
}

// requireFlagSnapshotMaxIDAPI gates tests that use getSnapshotMaxID (global snapshot id).
func requireFlagSnapshotMaxIDAPI(t *testing.T) {
	t.Helper()
	requireCurrentFlagrAPI(t, http.MethodGet, "/api/v1/flags/snapshots/max_id", nil, capSnapshotsMaxID)
}

// requireDuplicateFlagAPI gates duplicate-flag tests.
func requireDuplicateFlagAPI(t *testing.T) {
	t.Helper()
	if len(seedFlagIDs) == 0 {
		t.Fatal("no seeded flags available")
	}
	path := fmt.Sprintf("/api/v1/flags/%d/duplicate", seedFlagIDs[0])
	requireCurrentFlagrAPI(t, http.MethodPost, path, map[string]any{}, capDuplicateFlag)
}

// requireOptionalAPI skips when a route is missing (404). Use for recorder/Datar-style
// endpoints that are not part of the core CRUD matrix on every backend.
func requireOptionalAPI(t *testing.T, method, path string, body any, label string) {
	t.Helper()
	if endpointAvailable(t, method, path, body) {
		return
	}
	t.Skipf("%s not available on this server (e.g. checkr/flagr:1.1.12)", label)
}