package notification

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// doRequestWithRetry performs an HTTP request with exponential backoff and jitter.
// On success (status < 500), it returns (resp, nil).
// On failure after retries, it returns the last response (if any) and an error.
func doRequestWithRetry(ctx context.Context, client *http.Client, req *http.Request, maxRetries int, baseDelay, maxDelay time.Duration) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error
	delay := baseDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return lastResp, fmt.Errorf("retry canceled: %w", ctx.Err())
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			if attempt < maxRetries {
				delay = nextDelayWithJitter(delay, maxDelay)
				continue
			}
			return nil, lastErr
		}
		// Don't close body here; caller will handle it if resp is returned

		if resp.StatusCode < 500 {
			// Close any previous failed response body before returning success
			if lastResp != nil {
				lastResp.Body.Close()
			}
			return resp, nil // Success or client error (4xx) is considered final; no retry on 4xx
		}

		// 5xx - retryable
		// Close previous lastResp.Body before overwriting to prevent resource leak
		if lastResp != nil {
			lastResp.Body.Close()
		}
		lastResp = resp
		lastErr = fmt.Errorf("HTTP %d error", resp.StatusCode)
		if attempt < maxRetries {
			delay = nextDelayWithJitter(delay, maxDelay)
			continue
		}
		// Final attempt failed with 5xx - caller is responsible for closing body
		return resp, lastErr
	}

	return lastResp, lastErr
}

// nextDelayWithJitter returns the next retry delay using exponential backoff
// with up to 50% jitter to prevent thundering herd.
func nextDelayWithJitter(prevDelay, maxDelay time.Duration) time.Duration {
	next := minDuration(2*prevDelay, maxDelay)
	jitter := time.Duration(rand.Int64N(int64(next) / 2))
	return next + jitter
}
