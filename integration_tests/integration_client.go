//go:build integration

package flagr_integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

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

// doReqAndDecode performs an HTTP request, checks for 2xx, and optionally decodes JSON dst.
// errorf is t.Fatalf for tests, log.Fatalf for seed.
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

func putJSON(t *testing.T, path string, body, dst any) {
	t.Helper()
	doReqAndDecode("PUT", path, body, dst, t.Fatalf)
}

func deleteResource(t *testing.T, path string) {
	t.Helper()
	doReqAndDecode("DELETE", path, nil, nil, t.Fatalf)
}

// doReqOrDie is identical to doReqAndDecode but uses log.Fatalf for use outside test context.
func doReqOrDie(method, path string, body, dst any) {
	doReqAndDecode(method, path, body, dst, log.Fatalf)
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
