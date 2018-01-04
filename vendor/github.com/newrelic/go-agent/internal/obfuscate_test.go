package internal

import (
	"testing"
)

func TestDeobfuscate(t *testing.T) {
	var out []byte
	var err error

	for _, in := range []string{"", "foo"} {
		out, err = deobfuscate(in, []byte(""))
		if err == nil {
			t.Error("error is nil for an empty key")
		}
		if out != nil {
			t.Errorf("out is not nil; got: %s", out)
		}
	}

	for _, in := range []string{"invalid_base64", "=moreinvalidbase64", "xx"} {
		out, err = deobfuscate(in, []byte(""))
		if err == nil {
			t.Error("error is nil for invalid base64")
		}
		if out != nil {
			t.Errorf("out is not nil; got: %s", out)
		}
	}

	for _, test := range []struct {
		input    string
		key      string
		expected string
	}{
		{"", "BLAHHHH", ""},
		{"NikyPBs8OisiJg==", "BLAHHHH", "testString"},
	} {
		out, err = deobfuscate(test.input, []byte(test.key))
		if err != nil {
			t.Errorf("error expected to be nil; got: %v", err)
		}
		if string(out) != test.expected {
			t.Errorf("output mismatch; expected: %s; got: %s", test.expected, out)
		}
	}
}

func TestObfuscate(t *testing.T) {
	var out string
	var err error

	for _, in := range []string{"", "foo"} {
		out, err = obfuscate([]byte(in), []byte(""))
		if err == nil {
			t.Error("error is nil for an empty key")
		}
		if out != "" {
			t.Errorf("out is not an empty string; got: %s", out)
		}
	}

	for _, test := range []struct {
		input    string
		key      string
		expected string
	}{
		{"", "BLAHHHH", ""},
		{"testString", "BLAHHHH", "NikyPBs8OisiJg=="},
	} {
		out, err = obfuscate([]byte(test.input), []byte(test.key))
		if err != nil {
			t.Errorf("error expected to be nil; got: %v", err)
		}
		if out != test.expected {
			t.Errorf("output mismatch; expected: %s; got: %s", test.expected, out)
		}
	}
}
