package util

// SafeString safely dereference the string or string ptr
func SafeString(s interface{}) string {
	if s == nil {
		return ""
	}
	switch s.(type) {
	case string:
		return s.(string)
	case *string:
		return *s.(*string)
	}
	return ""
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

// Uint32Ptr ...
func Uint32Ptr(v uint32) *uint32 { return &v }

// Uint64Ptr ...
func Uint64Ptr(v uint64) *uint64 { return &v }

// BoolPtr ...
func BoolPtr(v bool) *bool { return &v }

// ByteSlicePtr ...
func ByteSlicePtr(v []byte) *[]byte { return &v }
