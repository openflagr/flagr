package util

import (
	"fmt"
	"math"
	"path"
	"regexp"
	"strings"
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

// HasSafePrefix checks if the given string is a safe URL path prefix
func HasSafePrefix(s string, prefix string) bool {
	if prefix == "" {
		return true
	}

	// Check for path traversal attempts or suspicious patterns
	if s == "." || s == ".." || strings.Contains(s, "..") {
		return false
	}

	// First normalize the path (prefix is controlled by us, no need to clean it)
	cleanedS := path.Clean(s)

	// Check if the normalized path starts with the prefix
	return strings.HasPrefix(cleanedS, prefix)
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

// TimeNow follows RFC3339 time format
func TimeNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}
