package util

import (
	"fmt"
	"math"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dchest/uniuri"
	"github.com/spf13/cast"
)

var (
	keyLengthLimit = 63
	keyRegex       = regexp.MustCompile(`^[\w\d-/\.\:]+$`)

	valueLengthLimit = 63
	valueRegex       = regexp.MustCompile(`^[ \w\d-/\.\:]+$`)

	randomKeyCharset = []byte("123456789abcdefghijkmnopqrstuvwxyz")
	randomKeyPrefix  = "k"
)

// IsSafeKey return if the key is safe to store
func IsSafeKey(s string) (bool, string) {
	if !keyRegex.MatchString(s) {
		return false, fmt.Sprintf("key:%s should have the format %v", s, keyRegex)
	}
	if len(s) > keyLengthLimit {
		return false, fmt.Sprintf("key:%s cannot be longer than %d", s, keyLengthLimit)
	}
	return true, ""
}

// IsSafeValue return if the value is safe to store
func IsSafeValue(s string) (bool, string) {
	if !valueRegex.MatchString(s) {
		return false, fmt.Sprintf("value:%s should have the format %v", s, valueRegex)
	}
	if len(s) > valueLengthLimit {
		return false, fmt.Sprintf("value:%s cannot be longer than %d", s, valueLengthLimit)
	}
	return true, ""
}

// maxPathUnescape is how many times HasDotDot will PathUnescape a segment
// string. One catch is Go's URL parser; two-three catch %252e double-encoding
// (OWASP path-traversal bypass). Bounded so a malformed % chain cannot loop.
const maxPathUnescape = 3

// HasDotDot reports whether p contains a parent-directory path segment ("..").
// Naive strings.Contains(p, "..") misses URL-encoded forms (%2e%2e, %252e%252e)
// and backslash separators, and false-positives on names like "foo..bar".
// This checks slash/backslash-separated segments after a bounded unescape,
// which is the same class of prefix-escape Traefik GHSA-vrch-868g-9jx5 hit.
func HasDotDot(p string) bool {
	return p != "" && hasDotDotSegment(decodePath(p))
}

func decodePath(p string) string {
	u := p
	for range maxPathUnescape {
		u = strings.ReplaceAll(u, `\`, "/")
		next, err := url.PathUnescape(u)
		if err != nil || next == u {
			break
		}
		u = next
	}
	return strings.ReplaceAll(u, `\`, "/")
}

func hasDotDotSegment(p string) bool {
	for p != "" {
		var seg string
		if i := strings.IndexByte(p, '/'); i >= 0 {
			seg, p = p[:i], p[i+1:]
		} else {
			seg, p = p, ""
		}
		if seg == ".." {
			return true
		}
	}
	return false
}

// HasSafePrefix checks if the given string is a safe URL path prefix.
// A path with a ".." segment is never a safe prefix match (callers must
// not skip JWT/basic whitelist or treat it as /api/v1/health, etc.).
func HasSafePrefix(s string, prefix string) bool {
	if prefix == "" {
		return true
	}

	if s == "." || HasDotDot(s) {
		return false
	}

	// Decode then Clean so /api/v1/./flags and /api/v1/%2e/flags match
	// /api/v1/flags. "." cannot leave a directory; only ".." can, and
	// HasDotDot already rejected it. Prefix is controlled by us.
	return strings.HasPrefix(path.Clean(decodePath(s)), prefix)
}

// NewSecureRandomKey creates a new secure random key
func NewSecureRandomKey() string {
	return randomKeyPrefix + uniuri.NewLenChars(uniuri.StdLen, randomKeyCharset)
}

// SafeStringWithDefault parse an any to string
// and set it to default value if it's empty
func SafeStringWithDefault(s any, deft string) (ret string) {
	ret = SafeString(s)
	if ret == "" {
		ret = deft
	}
	return ret
}

// SafeString safely cast to string
func SafeString(s any) (ret string) {
	return cast.ToString(s)
}

// SafeUint returns the uint of the value
func SafeUint(s any) (ret uint) {
	return cast.ToUint(s)
}

// Round makes the float to int conversion with rounding
func Round(f float64) int {
	return int(f + math.Copysign(0.5, f))
}

var (
	timeNowMu     sync.RWMutex
	timeNowSec    int64
	timeNowCached string
)

// TimeNow follows RFC3339 time format. Timestamps are cached per UTC second so
// high-volume evaluation does not allocate a new RFC3339 string on every result.
func TimeNow() string {
	now := time.Now().UTC()
	sec := now.Unix()
	timeNowMu.RLock()
	if sec == timeNowSec {
		s := timeNowCached
		timeNowMu.RUnlock()
		return s
	}
	timeNowMu.RUnlock()

	timeNowMu.Lock()
	defer timeNowMu.Unlock()
	if sec == timeNowSec {
		return timeNowCached
	}
	timeNowSec = sec
	timeNowCached = now.Format(time.RFC3339)
	return timeNowCached
}

// ParseHeaders converts a comma-separated list of key-value pairs separated by colons into a map of strings.
// It gracefully handles edge cases such as empty headers, missing values, spaces around keys and values,
// and malformed chunks by filtering them out.
// Example: "Authorization: Bearer token, X-Custom-Header: value" will be parsed correctly.
func ParseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if key != "" {
				headers[key] = val
			}
		}
	}
	return headers
}
