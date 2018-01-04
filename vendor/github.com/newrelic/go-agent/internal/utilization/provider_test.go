package utilization

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

// Cross agent test types common to each provider's set of test cases.
type testCase struct {
	TestName            string                  `json:"testname"`
	URIs                map[string]jsonResponse `json:"uri"`
	EnvVars             map[string]envResponse  `json:"env_vars"`
	ExpectedVendorsHash vendors                 `json:"expected_vendors_hash"`
	ExpectedMetrics     map[string]metric       `json:"expected_metrics"`
}

type envResponse struct {
	Response string `json:"response"`
	Timeout  bool   `json:"timeout"`
}

type jsonResponse struct {
	Response json.RawMessage `json:"response"`
	Timeout  bool            `json:"timeout"`
}

type metric struct {
	CallCount int `json:"call_count"`
}

var errTimeout = errors.New("timeout")

type mockTransport struct {
	t         *testing.T
	responses map[string]jsonResponse
}

type mockBody struct {
	bytes.Reader
	closed bool
	t      *testing.T
}

func (m *mockTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	for match, response := range m.responses {
		if r.URL.String() == match {
			return m.respond(response)
		}
	}

	m.t.Errorf("Unknown request URI: %s", r.URL.String())
	return nil, nil
}

func (m *mockTransport) respond(resp jsonResponse) (*http.Response, error) {
	if resp.Timeout {
		return nil, errTimeout
	}

	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body: &mockBody{
			t:      m.t,
			Reader: *bytes.NewReader(resp.Response),
		},
	}, nil
}

// This function is included simply so that http.Client doesn't complain.
func (m *mockTransport) CancelRequest(r *http.Request) {}

func (m *mockBody) Close() error {
	if m.closed {
		m.t.Error("Close of already closed connection!")
	}

	m.closed = true
	return nil
}

func (m *mockBody) ensureClosed() {
	if !m.closed {
		m.t.Error("Connection was not closed")
	}
}

func TestNormaliseValue(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		{
			name:     "Valid - empty",
			input:    "",
			expected: "",
			isError:  false,
		},
		{
			name:     "Valid - symbols",
			input:    ". /-_",
			expected: ". /-_",
			isError:  false,
		},
		{
			name:     "Valid - string",
			input:    "simplesentence",
			expected: "simplesentence",
			isError:  false,
		},
		{
			name: "Invalid - More than 255",
			input: `256256256256256256256256256256256256256256256256256256256256
			256256256256256256256256256256256256256256256256256256256256256256256256
			256256256256256256256256256256256256256256256256256256256256256256256256
			2562562562562562562562562562562562562562562562562562`,
			expected: "",
			isError:  true,
		},
	}

	for _, tc := range testCases {
		actual, err := normalizeValue(tc.input)

		if tc.isError && err == nil {
			t.Fatalf("%s: expected error; got nil", tc.name)
		} else if !tc.isError {
			if err != nil {
				t.Fatalf("%s: expected not error; got: %v", tc.name, err)
			}
			if tc.expected != actual {
				t.Fatalf("%s: expected: %s; got: %s", tc.name, tc.expected, actual)
			}
		}
	}
}
