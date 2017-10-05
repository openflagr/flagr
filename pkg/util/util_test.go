package util

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestSafeString(t *testing.T) {
	assert.Equal(t, SafeString("123"), "123")
	assert.Equal(t, SafeString(StringPtr("123")), "123")
	assert.Equal(t, SafeString(123), "")
	assert.Equal(t, SafeString(nil), "")
}
