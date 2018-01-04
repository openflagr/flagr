package internal

import (
	"net/http"
	"testing"
	"time"
)

func TestRemoveFirstSegment(t *testing.T) {
	testcases := []struct {
		input    string
		expected string
	}{
		{input: "no_seperators", expected: "no_seperators"},
		{input: "heyo/zip/zap", expected: "zip/zap"},
		{input: "ends_in_slash/", expected: ""},
		{input: "☃☃☃/✓✓✓/heyo", expected: "✓✓✓/heyo"},
		{input: "☃☃☃/", expected: ""},
		{input: "/", expected: ""},
		{input: "", expected: ""},
	}

	for _, tc := range testcases {
		out := removeFirstSegment(tc.input)
		if out != tc.expected {
			t.Fatal(tc.input, out, tc.expected)
		}
	}
}

func TestFloatSecondsToDuration(t *testing.T) {
	if d := floatSecondsToDuration(0.123); d != 123*time.Millisecond {
		t.Error(d)
	}
	if d := floatSecondsToDuration(456.0); d != 456*time.Second {
		t.Error(d)
	}
}

func TestAbsTimeDiff(t *testing.T) {
	diff := 5 * time.Second
	before := time.Now()
	after := before.Add(5 * time.Second)

	if out := absTimeDiff(before, after); out != diff {
		t.Error(out, diff)
	}
	if out := absTimeDiff(after, before); out != diff {
		t.Error(out, diff)
	}
	if out := absTimeDiff(after, after); out != 0 {
		t.Error(out)
	}
}

func TestTimeToFloatMilliseconds(t *testing.T) {
	tm := time.Unix(123, 456789000)
	if ms := timeToFloatMilliseconds(tm); ms != 123456.789 {
		t.Error(ms)
	}
}

func TestCompactJSON(t *testing.T) {
	in := `
	{   "zip":	1}`
	out := CompactJSONString(in)
	if out != `{"zip":1}` {
		t.Fatal(in, out)
	}
}

func TestGetContentLengthFromHeader(t *testing.T) {
	// Nil header.
	if cl := GetContentLengthFromHeader(nil); cl != -1 {
		t.Errorf("unexpected content length: expected -1; got %d", cl)
	}

	// Empty header.
	header := make(http.Header)
	if cl := GetContentLengthFromHeader(header); cl != -1 {
		t.Errorf("unexpected content length: expected -1; got %d", cl)
	}

	// Invalid header.
	header.Set("Content-Length", "foo")
	if cl := GetContentLengthFromHeader(header); cl != -1 {
		t.Errorf("unexpected content length: expected -1; got %d", cl)
	}

	// Zero header.
	header.Set("Content-Length", "0")
	if cl := GetContentLengthFromHeader(header); cl != 0 {
		t.Errorf("unexpected content length: expected 0; got %d", cl)
	}

	// Valid, non-zero header.
	header.Set("Content-Length", "1024")
	if cl := GetContentLengthFromHeader(header); cl != 1024 {
		t.Errorf("unexpected content length: expected 1024; got %d", cl)
	}
}

func TestStringLengthByteLimit(t *testing.T) {
	testcases := []struct {
		input  string
		limit  int
		expect string
	}{
		{"", 255, ""},
		{"awesome", -1, ""},
		{"awesome", 0, ""},
		{"awesome", 1, "a"},
		{"awesome", 7, "awesome"},
		{"awesome", 20, "awesome"},
		{"日本\x80語", 10, "日本\x80語"}, // bad unicode
		{"日本", 1, ""},
		{"日本", 2, ""},
		{"日本", 3, "日"},
		{"日本", 4, "日"},
		{"日本", 5, "日"},
		{"日本", 6, "日本"},
		{"日本", 7, "日本"},
	}

	for _, tc := range testcases {
		out := StringLengthByteLimit(tc.input, tc.limit)
		if out != tc.expect {
			t.Error(tc.input, tc.limit, tc.expect, out)
		}
	}
}
