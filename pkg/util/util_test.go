package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeString(t *testing.T) {
	assert.Equal(t, SafeString("123"), "123")
	assert.Equal(t, SafeString(StringPtr("123")), "123")
	assert.Equal(t, SafeString(123), "123")
	assert.Equal(t, SafeString(nil), "")

	var ptr *string
	assert.Equal(t, SafeString(ptr), "")
}

func TestSafeStringWithDefault(t *testing.T) {
	assert.Equal(t, SafeStringWithDefault("123", ""), "123")
	assert.Equal(t, SafeStringWithDefault(StringPtr("123"), ""), "123")
	assert.Equal(t, SafeStringWithDefault(123, ""), "123")
	assert.Equal(t, SafeStringWithDefault(nil, ""), "")

	assert.Equal(t, SafeStringWithDefault("123", "<nil>"), "123")
	assert.Equal(t, SafeStringWithDefault(StringPtr("123"), "<nil>"), "123")
	assert.Equal(t, SafeStringWithDefault(123, "<nil>"), "123")
	assert.Equal(t, SafeStringWithDefault(nil, "<nil>"), "<nil>")
}

func TestSafeUint(t *testing.T) {
	assert.Equal(t, SafeUint(nil), uint(0))
	assert.Equal(t, SafeUint("123"), uint(123))

	assert.Equal(t, SafeUint(int(1)), uint(1))
	assert.Equal(t, SafeUint(IntPtr(1)), uint(1))
	assert.Equal(t, SafeUint(int32(1)), uint(1))
	assert.Equal(t, SafeUint(Int32Ptr(1)), uint(1))
	assert.Equal(t, SafeUint(int64(1)), uint(1))
	assert.Equal(t, SafeUint(Int64Ptr(1)), uint(1))
	assert.Equal(t, SafeUint(uint(1)), uint(1))
	assert.Equal(t, SafeUint(UintPtr(1)), uint(1))
	assert.Equal(t, SafeUint(uint32(1)), uint(1))
	assert.Equal(t, SafeUint(Uint32Ptr(1)), uint(1))
	assert.Equal(t, SafeUint(uint64(1)), uint(1))
	assert.Equal(t, SafeUint(Uint64Ptr(1)), uint(1))
	assert.Equal(t, SafeUint(int32(1)), uint(1))
	assert.Equal(t, SafeUint(int64(1)), uint(1))

	assert.Equal(t, SafeUint(float32(123)), uint(123))
	assert.Equal(t, SafeUint(Float32Ptr(float32(123))), uint(123))
	assert.Equal(t, SafeUint(float64(123)), uint(123))
	assert.Equal(t, SafeUint(Float64Ptr(float64(123))), uint(123))

	var ptr *int64
	assert.Equal(t, SafeUint(ptr), uint(0))
}

func TestTimeNow(t *testing.T) {
	assert.Len(t, TimeNow(), 20)
}

func TestIsSafeKey(t *testing.T) {
	var b bool
	var msg string

	b, msg = IsSafeKey("a")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeKey("a1")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeKey("a_1")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeKey(strings.Repeat("a", 63))
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeKey("1a")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeKey(" spaces in key are not allowed ")
	assert.False(t, b)
	assert.NotEmpty(t, msg)	

	b, msg = IsSafeKey("_a")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeKey(strings.Repeat("a", 64))
	assert.False(t, b)
	assert.NotEmpty(t, msg)

	b, msg = IsSafeKey("slashes/are/valid")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeKey("dots.are.valid")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeKey("colons:are:valid")
	assert.True(t, b)
	assert.Empty(t, msg)
}

func TestPtrs(t *testing.T) {
	assert.Equal(t, "a", *StringPtr("a"))
	assert.Equal(t, int(1), *IntPtr(int(1)))
	assert.Equal(t, int32(1), *Int32Ptr(int32(1)))
	assert.Equal(t, int64(1), *Int64Ptr(int64(1)))
	assert.Equal(t, float32(1.0), *Float32Ptr(float32(1.0)))
	assert.Equal(t, float64(1.0), *Float64Ptr(float64(1.0)))
	assert.Equal(t, uint(1), *UintPtr(uint(1)))
	assert.Equal(t, uint32(1), *Uint32Ptr(uint32(1)))
	assert.Equal(t, uint64(1), *Uint64Ptr(uint64(1)))
	assert.Equal(t, true, *BoolPtr(true))
	assert.Equal(t, []byte("abc"), *ByteSlicePtr([]byte("abc")))
}

func TestRound(t *testing.T) {
	assert.Equal(t, Round(float64(1.01)), 1)
	assert.Equal(t, Round(float64(1.99)), 2)
	assert.Equal(t, Round(float64(-0.5)), -1)
	assert.Equal(t, Round(float64(-1.5)), -2)
	assert.Equal(t, Round(float64(0.0)), 0)
}

func TestNewSecureRandomKey(t *testing.T) {
	assert.NotZero(t, NewSecureRandomKey())
	assert.Contains(t, NewSecureRandomKey(), randomKeyPrefix)

	ok, _ := IsSafeKey(NewSecureRandomKey())
	assert.True(t, ok)
}
