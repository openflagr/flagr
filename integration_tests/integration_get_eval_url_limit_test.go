//go:build integration

package flagr_integration

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
)

// TestIntegration_GetEvaluation_QueryURLBytesLimit exercises FLAGR_EVAL_GET_MAX_URL_BYTES
// (default 8192) with variable-length entityContext payloads. Numbers are stable for docs.
func TestIntegration_GetEvaluation_QueryURLBytesLimit(t *testing.T) {
	if os.Getenv("FLAGR_SERVER_URL") != "" || os.Getenv("FLAGR_SERVER_URLS") != "" {
		t.Skip("query URL byte limit test requires auto-started local server (default FLAGR_EVAL_GET_MAX_URL_BYTES=8192)")
	}
	requireOptionalAPI(t, http.MethodGet, "/api/v1/evaluation?json=%7B%7D", nil, "GET /api/v1/evaluation")

	max := integrationGetEvalMaxURLBytes

	underBlob, underQLen, err := findBlobLenForQueryLen(max)
	if err != nil {
		t.Fatalf("under limit probe: %v", err)
	}
	if underQLen != max {
		t.Fatalf("expected query len %d at limit, got %d (blobLen=%d)", max, underQLen, underBlob)
	}
	underQuery, _, err := buildGetEvalQuery(underBlob)
	if err != nil {
		t.Fatalf("build under query: %v", err)
	}
	t.Logf("doc example: at limit rawQueryLen=%d entityContext.blob len=%d", underQLen, underBlob)

	var ok evalResponse
	getJSON(t, "/api/v1/evaluation?"+underQuery, &ok)

	overBlob := underBlob + 1
	overQuery, overQLen, err := buildGetEvalQuery(overBlob)
	if err != nil {
		t.Fatalf("build over query: %v", err)
	}
	if overQLen != max+1 {
		t.Fatalf("expected query len %d over limit, got %d (blobLen=%d)", max+1, overQLen, overBlob)
	}
	t.Logf("doc example: over limit rawQueryLen=%d entityContext.blob len=%d", overQLen, overBlob)

	resp, err := doReq("GET", "/api/v1/evaluation?"+overQuery, nil)
	if err != nil {
		t.Fatalf("GET over limit: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 400, got %d body=%s", resp.StatusCode, body)
	}
	var errPayload struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errPayload); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if errPayload.Message == "" {
		t.Fatal("expected non-empty error message")
	}
	t.Logf("error message: %s", errPayload.Message)
}
