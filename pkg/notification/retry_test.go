package notification

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDoRequestWithRetry(t *testing.T) {
	t.Run("returns immediately on 2xx success", func(t *testing.T) {
		callCount := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := &http.Client{}
		req, _ := http.NewRequest("GET", ts.URL, nil)
		ctx := context.Background()

		resp, err := doRequestWithRetry(ctx, client, req, 3, 10*time.Millisecond, 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
		resp.Body.Close()
	})

	t.Run("does not retry on 4xx client error", func(t *testing.T) {
		callCount := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		client := &http.Client{}
		req, _ := http.NewRequest("GET", ts.URL, nil)
		ctx := context.Background()

		resp, err := doRequestWithRetry(ctx, client, req, 3, 10*time.Millisecond, 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "Should not retry on 4xx")
		resp.Body.Close()
	})

	t.Run("retries on 5xx server error", func(t *testing.T) {
		callCount := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&callCount, 1)
			if count < 3 {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer ts.Close()

		client := &http.Client{}
		req, _ := http.NewRequest("GET", ts.URL, nil)
		ctx := context.Background()

		resp, err := doRequestWithRetry(ctx, client, req, 3, 10*time.Millisecond, 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int32(3), atomic.LoadInt32(&callCount), "Should retry on 5xx")
		resp.Body.Close()
	})

	t.Run("returns error after max retries exhausted", func(t *testing.T) {
		callCount := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer ts.Close()

		client := &http.Client{}
		req, _ := http.NewRequest("GET", ts.URL, nil)
		ctx := context.Background()

		resp, err := doRequestWithRetry(ctx, client, req, 2, 10*time.Millisecond, 100*time.Millisecond)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 503")
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		assert.Equal(t, int32(3), atomic.LoadInt32(&callCount), "Should make maxRetries+1 attempts")
		resp.Body.Close()
	})

	t.Run("properly closes response body on retries to prevent leak", func(t *testing.T) {
		var bodiesClosed int32

		callCount := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&callCount, 1)
			if count < 3 {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer ts.Close()

		// Create a custom transport that tracks body closures
		originalTransport := http.DefaultTransport
		transport := &mockTransport{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				resp, err := originalTransport.RoundTrip(req)
				if resp != nil && resp.Body != nil {
					// Wrap body to track closure
					wrapped := &trackingBody{
						ReadCloser: resp.Body,
						onClose: func() {
							atomic.AddInt32(&bodiesClosed, 1)
						},
					}
					resp.Body = wrapped
				}
				return resp, err
			},
		}

		client := &http.Client{Transport: transport}
		req, _ := http.NewRequest("GET", ts.URL, nil)
		ctx := context.Background()

		resp, err := doRequestWithRetry(ctx, client, req, 3, 10*time.Millisecond, 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Close the final response
		resp.Body.Close()

		// We made 3 requests, so we expect 3 bodies total, all should be closed
		assert.Equal(t, int32(3), atomic.LoadInt32(&callCount), "Should make 3 requests")
		assert.Equal(t, int32(3), atomic.LoadInt32(&bodiesClosed), "All 3 response bodies should be closed after retries")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		callCount := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		client := &http.Client{}
		req, _ := http.NewRequest("GET", ts.URL, nil)
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel after first request
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		resp, err := doRequestWithRetry(ctx, client, req, 10, 100*time.Millisecond, 1*time.Second)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "canceled")
		// Should not have made all retries
		assert.LessOrEqual(t, atomic.LoadInt32(&callCount), int32(2))
		if resp != nil {
			resp.Body.Close()
		}
	})
}

// mockTransport is a custom http.RoundTripper for testing
type mockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

// trackingBody wraps an io.ReadCloser and calls onClose when Close is called
type trackingBody struct {
	io.ReadCloser
	onClose func()
}

func (t *trackingBody) Close() error {
	if t.onClose != nil {
		t.onClose()
	}
	return t.ReadCloser.Close()
}
