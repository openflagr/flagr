package utilization

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/newrelic/go-agent/internal/crossagent"
)

type maybeResponse struct {
	Response string `json:"response"`
	Timeout  bool   `json:"timeout"`
}

type mockTransport struct {
	ID, Type, Zone maybeResponse
	reader         *strings.Reader
	closed         int
}

func (m *mockTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var response maybeResponse

	if r.URL.Host != awsHost {
		return nil, fmt.Errorf("invalid endpoint host %s", r.URL.Host)
	}

	switch r.URL.Path {
	case typeEndpointPath:
		response = m.Type
	case zoneEndpointPath:
		response = m.Zone
	case idEndpointPath:
		response = m.ID
	default:
		return nil, fmt.Errorf("invalid endpoint %s", r.URL.Path)
	}

	if response.Timeout {
		return nil, errors.New("timed out")
	}

	m.reader = strings.NewReader(response.Response)

	return &http.Response{
		StatusCode: 200,
		Body:       m,
	}, nil
}

func (m *mockTransport) CancelRequest(req *http.Request) {
}

func (m *mockTransport) Read(b []byte) (int, error) {
	return m.reader.Read(b)
}

func (m *mockTransport) Close() error {
	m.closed++
	m.reader = nil
	return nil
}

func TestGetAndValidateStatus(t *testing.T) {
	transport := &mockTransport{Type: maybeResponse{Response: "e2-micro"}}
	client := &http.Client{Transport: transport}
	resp, err := getAndValidate(client, typeEndpoint)
	if err != nil || resp != "e2-micro" {
		t.Error(err, resp)
	}
	if 1 != transport.closed {
		t.Error("response body not closed")
	}

	transport = &mockTransport{Type: maybeResponse{Response: "e2,micro"}}
	client = &http.Client{Transport: transport}
	_, err = getAndValidate(client, typeEndpoint)
	if err == nil || !isAWSValidationError(err) {
		t.Error(err)
	}
	if 1 != transport.closed {
		t.Error("response body not closed")
	}
}

func TestCrossagentAWS(t *testing.T) {
	var testCases []struct {
		Name string `json:"testname"`
		URIs struct {
			Type maybeResponse `json:"http://169.254.169.254/2008-02-01/meta-data/instance-type"`
			ID   maybeResponse `json:"http://169.254.169.254/2008-02-01/meta-data/instance-id"`
			Zone maybeResponse `json:"http://169.254.169.254/2008-02-01/meta-data/placement/availability-zone"`
		} `json:"uris"`
		Vendors vendors `json:"expected_vendors_hash"`
		Metrics struct {
			Supportability struct {
				CallCount int `json:"call_count"`
			} `json:"Supportability/utilization/aws/error"`
		} `json:"expected_metrics"`
	}

	err := crossagent.ReadJSON("aws.json", &testCases)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		client := &http.Client{
			Transport: &mockTransport{
				ID:   tc.URIs.ID,
				Type: tc.URIs.Type,
				Zone: tc.URIs.Zone,
			},
		}

		v, err := getEndpoints(client)

		expectInvalid := tc.Metrics.Supportability.CallCount > 0
		if expectInvalid != isAWSValidationError(err) {
			t.Error(tc.Name, err, expectInvalid, isAWSValidationError(err))
		}

		expectTimeout := tc.URIs.Type.Timeout || tc.URIs.ID.Timeout || tc.URIs.Zone.Timeout
		if expectTimeout && nil == err {
			t.Error(tc.Name, err)
		}

		if tc.Vendors.AWS != nil {
			if nil == v {
				t.Error(tc.Name, "missing vendor")
			} else if v.ID != tc.Vendors.AWS.ID {
				t.Error(tc.Name, "Id mismatch", v.ID, tc.Vendors.AWS.ID)
			} else if v.Type != tc.Vendors.AWS.Type {
				t.Error(tc.Name, "Type mismatch", v.Type, tc.Vendors.AWS.Type)
			} else if v.Zone != tc.Vendors.AWS.Zone {
				t.Error(tc.Name, "Zone mismatch", v.Zone, tc.Vendors.AWS.Zone)
			}
		} else if nil != v {
			t.Error(tc.Name, "unexpected vendor")
		}
	}
}
