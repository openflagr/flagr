package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
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
	SystemOne(ctx context.Context, state any, questions map[string]entity.JevConstraintSpec) (*JevResponse, error)
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
}

// JevUsage is the token usage of a System One call.
type JevUsage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

func (c *jevHTTPClient) SystemOne(ctx context.Context, state any, questions map[string]entity.JevConstraintSpec) (*JevResponse, error) {
	payload := jevRequestPayload{
		State:     state,
		Model:     c.model,
		Questions: make(map[string]jevQuestionPayload, len(questions)),
	}
	for name, spec := range questions {
		payload.Questions[name] = jevQuestionPayload{
			Type:         spec.Type,
			Instructions: spec.Instructions,
			Criteria:     spec.Criteria,
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding jev request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+jevSystemOnePath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building jev request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling jev: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, jevMaxErrorBodyBytes))
		return nil, fmt.Errorf("jev systemone returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var out JevResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding jev response: %w", err)
	}
	if out.Answers == nil {
		return nil, fmt.Errorf("jev response contained no answers")
	}
	return &out, nil
}
