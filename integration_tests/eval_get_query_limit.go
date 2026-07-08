//go:build integration

package flagr_integration

import (
	"encoding/json"
	"net/url"
	"strings"
)

const integrationGetEvalMaxURLBytes = 8192

// buildGetEvalQuery builds the raw query string (json=...) for a GET /evaluation request.
// blobLen sets entityContext.blob length (ASCII) to grow encoded query size.
func buildGetEvalQuery(blobLen int) (rawQuery string, queryLen int, err error) {
	ctx := map[string]any{
		"entityID":   "get-eval-url-limit-probe",
		"entityType": "user",
		"flagID":     int64(1),
		"entityContext": map[string]any{
			"blob": strings.Repeat("a", blobLen),
		},
	}
	raw, err := json.Marshal(ctx)
	if err != nil {
		return "", 0, err
	}
	rawQuery = "json=" + url.QueryEscape(string(raw))
	return rawQuery, len(rawQuery), nil
}

// findBlobLenForQueryLen returns blobLen such that len("json="+escape(JSON)) == wantQueryLen,
// or the closest length at or below wantQueryLen if exact match is impossible.
func findBlobLenForQueryLen(wantQueryLen int) (blobLen int, gotQueryLen int, err error) {
	lo, hi := 0, 20000
	for lo < hi {
		mid := (lo + hi + 1) / 2
		_, n, err := buildGetEvalQuery(mid)
		if err != nil {
			return 0, 0, err
		}
		if n <= wantQueryLen {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	_, gotQueryLen, err = buildGetEvalQuery(lo)
	return lo, gotQueryLen, err
}
