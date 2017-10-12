package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeString(t *testing.T) {
	assert.Equal(t, SafeString("123"), "123")
	assert.Equal(t, SafeString(StringPtr("123")), "123")
	assert.Equal(t, SafeString(123), "")
	assert.Equal(t, SafeString(nil), "")

	var ptr *string
	assert.Equal(t, SafeString(ptr), "")
}

func TestSafeUint(t *testing.T) {
	assert.Equal(t, SafeUint("123"), uint(0))
	assert.Equal(t, SafeUint(Uint32Ptr(123)), uint(123))
	assert.Equal(t, SafeUint(123), uint(123))
	assert.Equal(t, SafeUint(nil), uint(0))

	var ptr *int64
	assert.Equal(t, SafeUint(ptr), uint(0))
}

func TestTimeNow(t *testing.T) {
	assert.Len(t, TimeNow(), 20)
}
