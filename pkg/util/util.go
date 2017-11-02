package util

import (
	"fmt"
	"regexp"
	"time"
)

var keyRegex = regexp.MustCompile("^[a-z]+[a-z0-9_]*$")

// IsSafeKey return if the key is safe to store
func IsSafeKey(s string) (bool, string) {
	if !keyRegex.MatchString(s) {
		return false, fmt.Sprintf("key should have the format %v", keyRegex)
	}
	if len(s) > 63 {
		return false, fmt.Sprintf("the key is too long")
	}
	return true, ""
}

// SafeString safely dereference the string or string ptr
func SafeString(s interface{}) (ret string) {
	defer func() {
		if r := recover(); r != nil {
			ret = ""
		}
	}()

	switch s.(type) {
	case string:
		return s.(string)
	case *string:
		return *s.(*string)
	}
	return ""
}

// SafeUint returns the uint of the value
func SafeUint(s interface{}) (ret uint) {
	defer func() {
		if r := recover(); r != nil {
			ret = 0
		}
	}()

	switch s.(type) {
	case int64:
		return uint(s.(int64))
	case *int64:
		return uint(*s.(*int64))
	case int32:
		return uint(s.(int32))
	case *int32:
		return uint(*s.(*int32))
	case int:
		return uint(s.(int))
	case *int:
		return uint(*s.(*int))
	case uint64:
		return uint(s.(uint64))
	case *uint64:
		return uint(*s.(*uint64))
	case uint32:
		return uint(s.(uint32))
	case *uint32:
		return uint(*s.(*uint32))
	case uint:
		return s.(uint)
	case *uint:
		return *s.(*uint)
	}
	return 0
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
