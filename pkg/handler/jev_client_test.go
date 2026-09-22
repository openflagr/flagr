package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJevClientSystemOne(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotBody jevRequestPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"jev-1.13.0","answers":{"intent":{"type":"noul","noul":0.82}},"usage":{"input_tokens":10,"output_tokens":0}}`))
	}))
	defer server.Close()

	sbURL := gostub.Stub(&config.Config.JevBaseURL, server.URL+"/")
	defer sbURL.Reset()
	sbKey := gostub.Stub(&config.Config.JevAPIKey, "secret")
	defer sbKey.Reset()
	sbModel := gostub.Stub(&config.Config.JevModel, "jev-latest")
	defer sbModel.Reset()

	client := NewJevClient()
	resp, err := client.SystemOne(context.Background(), map[string]any{"plan": "pro"}, map[string]entity.JevConstraintSpec{
		"intent": {Name: "intent", Type: entity.JevTypeNoul, Instructions: "Is this about billing?"},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "/v1/systemone", gotPath)
	assert.Equal(t, "Bearer secret", gotAuth)
	assert.Equal(t, "jev-latest", gotBody.Model)
	assert.Equal(t, "Is this about billing?", gotBody.Questions["intent"].Instructions)
	require.NotNil(t, resp.Answers["intent"].Noul)
	assert.Equal(t, 0.82, *resp.Answers["intent"].Noul)
}

func TestJevClientSystemOneErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"detail":[{"loc":["body","state"],"msg":"invalid"}]}`))
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()

	_, err := NewJevClient().SystemOne(context.Background(), "state", map[string]entity.JevConstraintSpec{
		"intent": {Name: "intent", Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "422")
}

func TestJevClientSystemOneTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{"answers":{}}`))
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()
	defer gostub.Stub(&config.Config.JevTimeout, 20*time.Millisecond).Reset()

	_, err := NewJevClient().SystemOne(context.Background(), "state", map[string]entity.JevConstraintSpec{
		"intent": {Name: "intent", Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.Error(t, err)
}

func TestJevClientSystemOneEmptyAnswers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"model":"jev-latest"}`))
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()

	_, err := NewJevClient().SystemOne(context.Background(), "state", map[string]entity.JevConstraintSpec{
		"intent": {Name: "intent", Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no answers")
}
