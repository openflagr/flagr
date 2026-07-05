//go:build integration

package flagr_integration

import (
	"testing"
)

func TestResponseIndicatesRouteNotRegistered(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		status int
		body   string
		want   bool
	}{
		{
			name:   "swagger path missing",
			status: 404,
			body:   `{"code":404,"message":"path /api/v1/flags/1/duplicate was not found"}`,
			want:   true,
		},
		{
			name:   "application flag missing",
			status: 404,
			body:   `{"code":404,"message":"unable to find flag 999 in the database"}`,
			want:   false,
		},
		{
			name:   "current duplicate probe",
			status: 404,
			body:   `{"code":404,"message":"unable to find flag 999999999 in the database"}`,
			want:   false,
		},
		{
			name:   "non-404",
			status: 500,
			body:   `{"message":"path /x was not found"}`,
			want:   false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := responseIndicatesRouteNotRegistered(tc.status, []byte(tc.body))
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestResponseIndicatesOptionalRouteUnavailable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		status int
		body   string
		want   bool
	}{
		{
			name:   "swagger path missing",
			status: 404,
			body:   `{"code":404,"message":"path /api/v1/evaluation was not found"}`,
			want:   true,
		},
		{
			name:   "legacy GET eval unsupported",
			status: 405,
			body:   `{"code":405,"message":"method GET is not allowed, but [POST] are"}`,
			want:   true,
		},
		{
			name:   "application flag missing",
			status: 404,
			body:   `{"code":404,"message":"unable to find flag 999 in the database"}`,
			want:   false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := responseIndicatesOptionalRouteUnavailable(tc.status, []byte(tc.body))
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestIsLegacyIntegrationBaseline(t *testing.T) {
	prev := baseURL
	t.Cleanup(func() { baseURL = prev })

	baseURL = "http://localhost:18001"
	if isLegacyIntegrationBaseline() {
		t.Fatal("18001 should not be legacy")
	}
	baseURL = "http://localhost:18006"
	if !isLegacyIntegrationBaseline() {
		t.Fatal("18006 should be legacy baseline")
	}
}
