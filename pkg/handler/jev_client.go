package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/sirupsen/logrus"
)

const (
	// jevSystemOnePath is the System One evaluation endpoint, shared by the
	// hosted TypeSafe API and open-source drop-in servers.
	jevSystemOnePath = "/v1/systemone"
	// jevMaxErrorBodyBytes bounds how much of an error response we read.
	jevMaxErrorBodyBytes = 4096
)

// JevClient evaluates typed questions against a state on a System One endpoint.
type JevClient interface {
	SystemOne(ctx context.Context, state any, questions map[string]entity.JevQuestion) (*JevResponse, error)
}

// NewJevClient builds the client from config. Overridable in tests.
var NewJevClient = func() JevClient {
	return &jevHTTPClient{
		baseURL: strings.TrimRight(config.Config.JevBaseURL, "/"),
		apiKey:  config.Config.JevAPIKey,
		model:   config.Config.JevModel,
		http:    &http.Client{Timeout: config.Config.JevTimeout},
	}
}

type jevHTTPClient struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

type jevQuestionPayload struct {
	Type         string `json:"type"`
	Instructions any    `json:"instructions,omitempty"`
	Criteria     any    `json:"criteria,omitempty"`
}

type jevRequestPayload struct {
	State     any                           `json:"state"`
	Model     string                        `json:"model"`
	Questions map[string]jevQuestionPayload `json:"questions"`
}

// JevAnswer is a single System One answer. Only the fields relevant to the
// answer's type are populated.
type JevAnswer struct {
	Type          string             `json:"type,omitempty"`
	Noul          *float64           `json:"noul,omitempty"`
	Choice        string             `json:"choice,omitempty"`
	Score         *float64           `json:"score,omitempty"`
	Confidence    *float64           `json:"confidence,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Legend        map[string]string  `json:"legend,omitempty"`
}

// JevResponse is a System One evaluation response.
type JevResponse struct {
	Model   string               `json:"model,omitempty"`
	Answers map[string]JevAnswer `json:"answers"`
	Usage   *JevUsage            `json:"usage,omitempty"`
	// LatencyMs is server-reported inference latency, when the endpoint provides
	// it (Kev and oido-systemone do; the hosted API may not).
	LatencyMs *float64 `json:"latency_ms,omitempty"`

	// ClientLatencyMs is the client-measured wall clock (including retries) and
	// Retries is how many extra attempts were made. Neither is part of the API.
	ClientLatencyMs float64 `json:"-"`
	Retries         int     `json:"-"`
}

// JevUsage is the token usage of a System One call.
type JevUsage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

func (c *jevHTTPClient) SystemOne(ctx context.Context, state any, questions map[string]entity.JevQuestion) (*JevResponse, error) {
	payload := jevRequestPayload{
		State:     state,
		Model:     c.model,
		Questions: make(map[string]jevQuestionPayload, len(questions)),
	}
	for name, q := range questions {
		payload.Questions[name] = jevQuestionPayload{
			Type:         q.Type,
			Instructions: q.Instructions,
			Criteria:     q.Criteria,
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding jev request: %w", err)
	}

	start := time.Now()
	resp, retries, err := c.doWithRetry(ctx, body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var out JevResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding jev response: %w", err)
	}
	if out.Answers == nil {
		return nil, fmt.Errorf("jev response contained no answers")
	}
	out.Retries = retries
	out.ClientLatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
	return &out, nil
}

// doWithRetry sends the request body, retrying transient failures (network
// errors, 5xx, 429) with exponential backoff and jitter. It stops retrying once
// the context is done, so retries never extend the call past JevTimeout.
//
// It returns the successful response, the number of retries used, and an error.
func (c *jevHTTPClient) doWithRetry(ctx context.Context, body []byte) (*http.Response, int, error) {
	var lastErr error
	delay := config.Config.JevRetryBase

	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			if err := sleepContext(ctx, delay); err != nil {
				return nil, attempt - 1, fmt.Errorf("jev retry canceled: %w", err)
			}
		}

		// A fresh request per attempt: the previous body reader is consumed.
		req, err := c.newRequest(ctx, body)
		if err != nil {
			return nil, attempt, err
		}

		resp, err := c.http.Do(req)
		switch {
		case err != nil:
			lastErr = fmt.Errorf("calling jev: %w", err)
		case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
			return resp, attempt, nil
		default:
			msg, _ := io.ReadAll(io.LimitReader(resp.Body, jevMaxErrorBodyBytes))
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("jev systemone returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
			if !jevRetryableStatus(resp.StatusCode) {
				return nil, attempt, lastErr // 4xx (except 429): retrying will not help
			}
		}

		if attempt >= config.Config.JevMaxRetries {
			return nil, attempt, lastErr
		}
		logrus.WithError(lastErr).WithField("attempt", attempt+1).Debug("retrying jev systemone call")
		delay = nextJevDelay(delay)
	}
}

func (c *jevHTTPClient) newRequest(ctx context.Context, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+jevSystemOnePath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building jev request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	return req, nil
}

// jevRetryableStatus reports whether an HTTP status is worth retrying.
func jevRetryableStatus(code int) bool {
	return code >= http.StatusInternalServerError || code == http.StatusTooManyRequests
}

// nextJevDelay returns the next retry delay with up to 50% jitter, capped at
// JevRetryMax, to avoid a thundering herd against a recovering endpoint.
func nextJevDelay(prev time.Duration) time.Duration {
	next := min(2*prev, config.Config.JevRetryMax)
	if next <= 0 {
		return 0
	}
	jitter := time.Duration(rand.Int64N(int64(next)/2 + 1))
	return next + jitter
}

// sleepContext waits for d or until the context is done, whichever comes first.
func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
