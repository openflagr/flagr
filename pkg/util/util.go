package util

import (
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/dchest/uniuri"
	"github.com/spf13/cast"
)

var (
	keyLengthLimit = 63
	keyRegex       = regexp.MustCompile(`^[\w\d-/\.\:]+$`)

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

// NewSecureRandomKey creates a new secure random key
func NewSecureRandomKey() string {
	return randomKeyPrefix + uniuri.NewLenChars(uniuri.StdLen, randomKeyCharset)
}

// SafeStringWithDefault parse an interface to string
// and set it to default value if it's empty
func SafeStringWithDefault(s interface{}, deft string) (ret string) {
	ret = SafeString(s)
	if ret == "" {
		ret = deft
	}
	return ret
}

// SafeString safely cast to string
func SafeString(s interface{}) (ret string) {
	return cast.ToString(s)
}

// SafeUint returns the uint of the value
func SafeUint(s interface{}) (ret uint) {
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

// Float32Ptr ...
func Float32Ptr(v float32) *float32 { return &v }

// Float64Ptr ...
func Float64Ptr(v float64) *float64 { return &v }

// IntPtr ...
func IntPtr(v int) *int { return &v }

// Int32Ptr ...
func Int32Ptr(v int32) *int32 { return &v }

// Int64Ptr ...
func Int64Ptr(v int64) *int64 { return &v }

// StringPtr ...
func StringPtr(v string) *string { return &v }

// UintPtr ...
func UintPtr(v uint) *uint { return &v }

// Uint32Ptr ...
func Uint32Ptr(v uint32) *uint32 { return &v }

// Uint64Ptr ...
func Uint64Ptr(v uint64) *uint64 { return &v }

// BoolPtr ...
func BoolPtr(v bool) *bool { return &v }

// ByteSlicePtr ...
func ByteSlicePtr(v []byte) *[]byte { return &v }
