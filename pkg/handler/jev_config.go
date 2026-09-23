package handler

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Jev / System One question types, mirroring the System One API contract.
const (
	// JevTypeNoul is a true/false question: the answer is P(true), a number in
	// [0,1]. It carries no confidence value, so it has no confidence threshold.
	JevTypeNoul   = "noul"
	JevTypeChoice = "choice"
	JevTypeScore  = "score"
)

// JevPropertyPrefix is the entityContext property prefix for Jev answers. The
// answer for a question named "plan_tier" is injected as `@jev_plan_tier`, flat
// and underscore-joined like the ambient `@ts_*` / `@http_*` properties.
const JevPropertyPrefix = "@jev_"

// Score level bounds, mirroring the System One API contract.
const (
	JevMinScoreLevels = 2
	JevMaxScoreLevels = 10
	// JevMaxChoiceOptions is the System One cap on choice options.
	JevMaxChoiceOptions = 255
)

// jevNamePattern is the safe charset for a question name. Names become
// `@jev_<name>` properties, so they must be identifier-like.
var jevNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// JevQuestion is an authored System One question. Instructions and Criteria are
// free-form JSON (string | object | array), matching the System One `EntryType`
// contract.
type JevQuestion struct {
	Type                string   `json:"type"`
	Instructions        any      `json:"instructions,omitempty"`
	Criteria            any      `json:"criteria,omitempty"`
	ConfidenceThreshold *float64 `json:"confidenceThreshold,omitempty"`
}

// JevEnricherConfig is the config for a flag-scoped `jev` enricher: the set of
// questions it asks, keyed by question name. Each name becomes `@jev_<name>`.
type JevEnricherConfig struct {
	Questions map[string]JevQuestion `json:"questions"`
}

// JevProperty returns the constraint property for a question name.
func JevProperty(name string) string { return JevPropertyPrefix + name }

// IsJevProperty reports whether a property belongs to the Jev namespace.
func IsJevProperty(property string) bool {
	return strings.HasPrefix(property, JevPropertyPrefix)
}

// JevQuestionName returns the question name behind a `@jev_<name>` property, or
// "" for properties outside the namespace.
func JevQuestionName(property string) string {
	if !IsJevProperty(property) {
		return ""
	}
	return strings.TrimPrefix(property, JevPropertyPrefix)
}

// QuestionNames returns the question names, sorted, for stable output.
func (c *JevEnricherConfig) QuestionNames() []string {
	names := make([]string, 0, len(c.Questions))
	for name := range c.Questions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Properties returns the `@jev_<name>` properties this config contributes.
func (c *JevEnricherConfig) Properties() []string {
	props := make([]string, 0, len(c.Questions))
	for _, name := range c.QuestionNames() {
		props = append(props, JevProperty(name))
	}
	return props
}

// Validate checks every question name and definition. A declared Jev enricher
// must ask at least one question; deleting the last question should delete the
// enricher instead.
func (c *JevEnricherConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("jev enricher config is required")
	}
	if len(c.Questions) == 0 {
		return fmt.Errorf("jev enricher requires at least one question")
	}
	for _, name := range c.QuestionNames() {
		if err := validateJevName(name); err != nil {
			return err
		}
		q := c.Questions[name]
		if err := q.Validate(); err != nil {
			return fmt.Errorf("jev question %q: %w", name, err)
		}
	}
	return nil
}

// Validate checks a question in isolation: type, instructions, per-type
// criteria shape, and confidence threshold bounds.
func (q *JevQuestion) Validate() error {
	if q == nil {
		return fmt.Errorf("jev question is required")
	}
	switch q.Type {
	case JevTypeNoul, JevTypeChoice, JevTypeScore:
	default:
		return fmt.Errorf("invalid jev.type %q: must be one of %s, %s, %s",
			q.Type, JevTypeNoul, JevTypeChoice, JevTypeScore)
	}
	if q.ConfidenceThreshold != nil && (*q.ConfidenceThreshold < 0 || *q.ConfidenceThreshold > 1) {
		return fmt.Errorf("jev.confidenceThreshold must be within [0,1], got %v", *q.ConfidenceThreshold)
	}
	if q.Type == JevTypeNoul && q.ConfidenceThreshold != nil {
		return fmt.Errorf("jev.confidenceThreshold is not supported for noul; the answer is already a 0-1 probability, so compare it instead (e.g. @jev_<name> >= 0.7)")
	}
	if q.Instructions == nil {
		return fmt.Errorf("jev.instructions is required")
	}
	if s, ok := q.Instructions.(string); ok && strings.TrimSpace(s) == "" {
		return fmt.Errorf("jev.instructions must not be empty")
	}
	switch q.Type {
	case JevTypeNoul:
		return validateNoulCriteria(q.Criteria)
	case JevTypeChoice:
		return validateChoiceCriteria(q.Criteria)
	case JevTypeScore:
		return validateScoreCriteria(q.Criteria)
	}
	return nil
}

// DecodeJevConfig decodes and validates a stored Jev enricher config.
func DecodeJevConfig(configJSON string) (*JevEnricherConfig, error) {
	if strings.TrimSpace(configJSON) == "" {
		return nil, fmt.Errorf("jev enricher config is required")
	}
	cfg := &JevEnricherConfig{}
	if err := json.Unmarshal([]byte(configJSON), cfg); err != nil {
		return nil, fmt.Errorf("invalid jev enricher config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func validateJevName(name string) error {
	if !jevNamePattern.MatchString(name) {
		return fmt.Errorf("jev question name %q must start with a letter or underscore and contain only letters, digits, and underscores", name)
	}
	return nil
}

func validateNoulCriteria(criteria any) error {
	if criteria == nil {
		return nil
	}
	m, ok := criteria.(map[string]any)
	if !ok {
		return fmt.Errorf("jev noul criteria must be a JSON object with \"true\" and \"false\" descriptions")
	}
	for key := range m {
		if key != "true" && key != "false" {
			return fmt.Errorf("jev noul criteria keys must be \"true\" or \"false\", got %q", key)
		}
	}
	return nil
}

func validateChoiceCriteria(criteria any) error {
	m, ok := criteria.(map[string]any)
	if !ok || len(m) == 0 {
		return fmt.Errorf("jev choice criteria must be a non-empty JSON object of option to description")
	}
	if len(m) > JevMaxChoiceOptions {
		return fmt.Errorf("jev choice criteria must have at most %d options, got %d", JevMaxChoiceOptions, len(m))
	}
	for name := range m {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("jev choice criteria option names must be non-empty")
		}
	}
	return nil
}

func validateScoreCriteria(criteria any) error {
	arr, ok := criteria.([]any)
	if !ok {
		return fmt.Errorf("jev score criteria must be a JSON array of levels")
	}
	if len(arr) < JevMinScoreLevels || len(arr) > JevMaxScoreLevels {
		return fmt.Errorf("jev score criteria must have between %d and %d levels, got %d",
			JevMinScoreLevels, JevMaxScoreLevels, len(arr))
	}
	return nil
}
