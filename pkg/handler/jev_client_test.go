package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubJevRetry shrinks the retry knobs so tests run fast.
func stubJevRetry(t *testing.T, maxRetries int, base, max time.Duration) {
	t.Helper()
	sbMax := gostub.Stub(&config.Config.JevMaxRetries, maxRetries)
	sbBase := gostub.Stub(&config.Config.JevRetryBase, base)
	sbMaxDelay := gostub.Stub(&config.Config.JevRetryMax, max)
	t.Cleanup(func() {
		sbMax.Reset()
		sbBase.Reset()
		sbMaxDelay.Reset()
	})
}

func TestJevClientSystemOne(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotBody jevRequestPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"jev-1.13.0","answers":{"intent":{"type":"noul","noul":0.82}},"usage":{"input_tokens":10,"output_tokens":0},"latency_ms":12.5}`))
	}))
	defer server.Close()

	sbURL := gostub.Stub(&config.Config.JevBaseURL, server.URL+"/")
	defer sbURL.Reset()
	sbKey := gostub.Stub(&config.Config.JevAPIKey, "secret")
	defer sbKey.Reset()
	sbModel := gostub.Stub(&config.Config.JevModel, "jev-latest")
	defer sbModel.Reset()

	client := NewJevClient()
	resp, err := client.SystemOne(context.Background(), map[string]any{"plan": "pro"}, map[string]entity.JevQuestion{
		"intent": {Type: entity.JevTypeNoul, Instructions: "Is this about billing?"},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "/v1/systemone", gotPath)
	assert.Equal(t, "Bearer secret", gotAuth)
	assert.Equal(t, "jev-latest", gotBody.Model)
	assert.Equal(t, "Is this about billing?", gotBody.Questions["intent"].Instructions)
	require.NotNil(t, resp.Answers["intent"].Noul)
	assert.Equal(t, 0.82, *resp.Answers["intent"].Noul)
	require.NotNil(t, resp.LatencyMs)
	assert.Equal(t, 12.5, *resp.LatencyMs)
	assert.GreaterOrEqual(t, resp.ClientLatencyMs, 0.0)
	assert.Zero(t, resp.Retries)
}

func TestJevClientSystemOneErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"detail":[{"loc":["body","state"],"msg":"invalid"}]}`))
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()

	_, err := NewJevClient().SystemOne(context.Background(), "state", map[string]entity.JevQuestion{
		"intent": {Type: entity.JevTypeNoul, Instructions: "x"},
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
	defer gostub.Stub(&config.Config.JevMaxRetries, 0).Reset()

	_, err := NewJevClient().SystemOne(context.Background(), "state", map[string]entity.JevQuestion{
		"intent": {Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.Error(t, err)
}

func TestJevClientSystemOneRetriesTransientFailures(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"warming up"}`))
			return
		}
		_, _ = w.Write([]byte(`{"answers":{"intent":{"type":"noul","noul":0.9}}}`))
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()
	stubJevRetry(t, 3, time.Millisecond, 5*time.Millisecond)

	resp, err := NewJevClient().SystemOne(context.Background(), "state", map[string]entity.JevQuestion{
		"intent": {Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(3), calls.Load())
	assert.Equal(t, 2, resp.Retries)
}

func TestJevClientSystemOneRetriesRateLimit(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"answers":{"intent":{"type":"noul","noul":0.9}}}`))
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()
	stubJevRetry(t, 2, time.Millisecond, 5*time.Millisecond)

	resp, err := NewJevClient().SystemOne(context.Background(), "state", map[string]entity.JevQuestion{
		"intent": {Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(2), calls.Load())
	assert.Equal(t, 1, resp.Retries)
}

func TestJevClientSystemOneDoesNotRetryClientError(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()
	stubJevRetry(t, 3, time.Millisecond, 5*time.Millisecond)

	_, err := NewJevClient().SystemOne(context.Background(), "state", map[string]entity.JevQuestion{
		"intent": {Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.Error(t, err)
	assert.Equal(t, int32(1), calls.Load(), "4xx must not be retried")
}

func TestJevClientSystemOneRetriesExhausted(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()
	stubJevRetry(t, 2, time.Millisecond, 5*time.Millisecond)

	_, err := NewJevClient().SystemOne(context.Background(), "state", map[string]entity.JevQuestion{
		"intent": {Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.Error(t, err)
	assert.Equal(t, int32(3), calls.Load(), "maxRetries+1 attempts")
}

func TestJevClientSystemOneRetryStopsAtContextDeadline(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()
	// A long backoff with a short deadline must stop after the first attempt.
	stubJevRetry(t, 5, 200*time.Millisecond, 200*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := NewJevClient().SystemOne(ctx, "state", map[string]entity.JevQuestion{
		"intent": {Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.Error(t, err)
	assert.Equal(t, int32(1), calls.Load(), "no retry once the deadline is reached")
}

type jevErrorTransport struct{ calls *atomic.Int32 }

func (t jevErrorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	t.calls.Add(1)
	return nil, errors.New("connection reset")
}

func TestJevClientSystemOneRetriesNetworkErrors(t *testing.T) {
	stubJevRetry(t, 2, time.Millisecond, 5*time.Millisecond)
	var calls atomic.Int32
	client := &jevHTTPClient{
		baseURL: "http://jev.invalid",
		http:    &http.Client{Transport: jevErrorTransport{calls: &calls}},
	}

	_, err := client.SystemOne(context.Background(), "state", map[string]entity.JevQuestion{
		"intent": {Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.Error(t, err)
	assert.Equal(t, int32(3), calls.Load(), "network errors are retried")
}

func TestJevClientSystemOneEmptyAnswers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"model":"jev-latest"}`))
	}))
	defer server.Close()

	defer gostub.Stub(&config.Config.JevBaseURL, server.URL).Reset()

	_, err := NewJevClient().SystemOne(context.Background(), "state", map[string]entity.JevQuestion{
		"intent": {Type: entity.JevTypeNoul, Instructions: "x"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no answers")
}
