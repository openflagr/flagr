package cat

import (
	"testing"
)

func TestIDHeaderUnmarshal(t *testing.T) {
	// Test error cases where the output is errUnexpectedArraySize.
	for _, input := range []string{
		``,
		`1234`,
		`1234#5678#90`,
		`foo`,
	} {
		_, err := NewIDHeader([]byte(input))
		if _, ok := err.(errUnexpectedArraySize); !ok {
			t.Errorf("given %s: error expected to be errUnexpectedArraySize; got %v", input, err)
		}
	}

	// Test error cases where the output is errInvalidAccountID.
	for _, input := range []string{
		`#1234`,
		`foo#bar`,
	} {
		if _, err := NewIDHeader([]byte(input)); err != errInvalidAccountID {
			t.Errorf("given %s: error expected to be %v; got %v", input, errInvalidAccountID, err)
		}
	}

	// Test success cases.
	for _, test := range []struct {
		input    string
		expected IDHeader
	}{
		{`1234#`, IDHeader{1234, ""}},
		{`1234#5678`, IDHeader{1234, "5678"}},
		{`1234#blob`, IDHeader{1234, "blob"}},
		{`0#5678`, IDHeader{0, "5678"}},
	} {
		id, err := NewIDHeader([]byte(test.input))

		if err != nil {
			t.Errorf("given %s: error expected to be nil; got %v", test.input, err)
		}
		if test.expected.AccountID != id.AccountID {
			t.Errorf("given %s: account ID expected to be %d; got %d", test.input, test.expected.AccountID, id.AccountID)
		}
		if test.expected.Blob != id.Blob {
			t.Errorf("given %s: account ID expected to be %s; got %s", test.input, test.expected.Blob, id.Blob)
		}
	}
}
