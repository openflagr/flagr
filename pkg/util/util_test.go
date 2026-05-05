package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeString(t *testing.T) {
	assert.Equal(t, SafeString("123"), "123")
	assert.Equal(t, SafeString(new("123")), "123")
	assert.Equal(t, SafeString(123), "123")
	assert.Equal(t, SafeString(nil), "")

	var ptr *string
	assert.Equal(t, SafeString(ptr), "")
}

func TestSafeStringWithDefault(t *testing.T) {
	assert.Equal(t, SafeStringWithDefault("123", ""), "123")
	assert.Equal(t, SafeStringWithDefault(new("123"), ""), "123")
	assert.Equal(t, SafeStringWithDefault(123, ""), "123")
	assert.Equal(t, SafeStringWithDefault(nil, ""), "")

	assert.Equal(t, SafeStringWithDefault("123", "<nil>"), "123")
	assert.Equal(t, SafeStringWithDefault(new("123"), "<nil>"), "123")
	assert.Equal(t, SafeStringWithDefault(123, "<nil>"), "123")
	assert.Equal(t, SafeStringWithDefault(nil, "<nil>"), "<nil>")
}

func TestSafeUint(t *testing.T) {
	assert.Equal(t, SafeUint(nil), uint(0))
	assert.Equal(t, SafeUint("123"), uint(123))

	assert.Equal(t, SafeUint(int(1)), uint(1))
	assert.Equal(t, SafeUint(new(1)), uint(1))
	assert.Equal(t, SafeUint(int32(1)), uint(1))
	assert.Equal(t, SafeUint(new(int32(1))), uint(1))
	assert.Equal(t, SafeUint(int64(1)), uint(1))
	assert.Equal(t, SafeUint(new(int64(1))), uint(1))
	assert.Equal(t, SafeUint(uint(1)), uint(1))
	assert.Equal(t, SafeUint(new(uint(1))), uint(1))
	assert.Equal(t, SafeUint(uint32(1)), uint(1))
	assert.Equal(t, SafeUint(new(uint32(1))), uint(1))
	assert.Equal(t, SafeUint(uint64(1)), uint(1))
	assert.Equal(t, SafeUint(new(uint64(1))), uint(1))
	assert.Equal(t, SafeUint(int32(1)), uint(1))
	assert.Equal(t, SafeUint(int64(1)), uint(1))

	assert.Equal(t, SafeUint(float32(123)), uint(123))
	assert.Equal(t, SafeUint(new(float32(123))), uint(123))
	assert.Equal(t, SafeUint(float64(123)), uint(123))
	assert.Equal(t, SafeUint(new(float64(123))), uint(123))

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

func TestIsSafeValue(t *testing.T) {
	var b bool
	var msg string

	b, msg = IsSafeValue("a")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeValue("a1")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeValue("a_1")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeValue(strings.Repeat("a", 63))
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeValue("1a")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeValue(" spaces in value are allowed ")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeValue("a@$")
	assert.False(t, b)
	assert.NotEmpty(t, msg)

	b, msg = IsSafeValue("_a")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeValue(strings.Repeat("a", 64))
	assert.False(t, b)
	assert.NotEmpty(t, msg)

	b, msg = IsSafeValue("slashes/are/valid")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeValue("dots.are.valid")
	assert.True(t, b)
	assert.Empty(t, msg)

	b, msg = IsSafeValue("colons:are:valid")
	assert.True(t, b)
	assert.Empty(t, msg)
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

func TestHasSafePrefix(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		prefix string
		want   bool
	}{
		{
			name:   "empty prefix always matches",
			s:      "any/path/here",
			prefix: "",
			want:   true,
		},
		{
			name:   "exact prefix match",
			s:      "api/v1/flags",
			prefix: "api/v1",
			want:   true,
		},
		{
			name:   "non-matching prefix",
			s:      "api/v1/flags",
			prefix: "api/v2",
			want:   false,
		},
		{
			name:   "prefix with trailing slash",
			s:      "api/v1/flags",
			prefix: "api/v1/",
			want:   true,
		},
		{
			name:   "path traversal attempt should fail",
			s:      "../api/v1/flags",
			prefix: "api",
			want:   false,
		},
		{
			name:   "good match",
			s:      "api",
			prefix: "api",
			want:   true,
		},
		{
			name:   "path traversal attempt should fail",
			s:      "..",
			prefix: "api",
			want:   false,
		},
		{
			name:   "path traversal attempt should fail",
			s:      ".",
			prefix: "api",
			want:   false,
		},
		{
			name:   "sneaky path traversal should fail",
			s:      "api/v1/../../secrets",
			prefix: "api",
			want:   false,
		},
		{
			name:   "path with dot should be cleaned",
			s:      "api/./v1/flags",
			prefix: "api/v1",
			want:   true,
		},
		{
			name:   "prefix with dot should be cleaned",
			s:      "api/v1/flags",
			prefix: "api/./v1",
			want:   false,
		},
		{
			name:   "multiple slashes should be cleaned",
			s:      "api///v1////flags",
			prefix: "api/v1",
			want:   true,
		},
		{
			name:   "complex traversal attempt should fail",
			s:      "api/v1/flags/../../../etc/passwd",
			prefix: "api",
			want:   false,
		},
		{
			name:   "complex traversal attempt should fail",
			s:      "api/v1/health/../flags",
			prefix: "api/v1/health",
			want:   false,
		},
		{
			name:   "longer path with valid prefix",
			s:      "api/v1/flags/123/settings",
			prefix: "api/v1/flags",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasSafePrefix(tt.s, tt.prefix)
			assert.Equal(t, tt.want, got, "HasSafePrefix(%v, %v)", tt.s, tt.prefix)
		})
	}
}

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "single valid header",
			input: "Authorization: Bearer token",
			expected: map[string]string{
				"Authorization": "Bearer token",
			},
		},
		{
			name:  "multiple valid headers",
			input: "Authorization: Bearer token, X-Custom-Header: value",
			expected: map[string]string{
				"Authorization":   "Bearer token",
				"X-Custom-Header": "value",
			},
		},
		{
			name:  "messy spacing around colons and commas",
			input: "  Auth :  Token  ,  Another : Value  ",
			expected: map[string]string{
				"Auth":    "Token",
				"Another": "Value",
			},
		},
		{
			name:  "missing value formatting",
			input: "Authorization:,",
			expected: map[string]string{
				"Authorization": "",
			},
		},
		{
			name:     "missing colon format is ignored",
			input:    "InvalidFormat",
			expected: map[string]string{},
		},
		{
			name:  "extra colons in the value are kept",
			input: "Trace-Id: 123:456:789",
			expected: map[string]string{
				"Trace-Id": "123:456:789",
			},
		},
		{
			name:     "spaces only",
			input:    "     ",
			expected: map[string]string{},
		},
		{
			name:     "colons only",
			input:    ":::",
			expected: map[string]string{},
		},
		{
			name:  "trailing and leading commas",
			input: ",Authorization: Bearer token,",
			expected: map[string]string{
				"Authorization": "Bearer token",
			},
		},
		{
			name:  "valid headers mixed with invalid garbage",
			input: "InvalidFormat, Authorization: Bearer token, , :valueOnly",
			expected: map[string]string{
				"Authorization": "Bearer token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseHeaders(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
