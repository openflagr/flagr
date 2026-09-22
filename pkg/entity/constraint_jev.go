package entity

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/sirupsen/logrus"
)

// Jev / System One question types.
const (
	JevTypeNoul   = "noul"
	JevTypeChoice = "choice"
	JevTypeScore  = "score"
)

// JevPropertyPrefix is the entityContext property prefix for Jev answers.
// The answer for a question named "foo" is injected under `@jev.foo`.
const JevPropertyPrefix = "@jev."

// JevContextKey is the top-level entityContext key holding all Jev answers.
const JevContextKey = "@jev"

// Score level bounds, mirroring the System One API contract.
const (
	JevMinScoreLevels = 2
	JevMaxScoreLevels = 10
)

// JevQuestion is the authored System One question behind a Jev constraint.
// Instructions and Criteria are free-form JSON (string | object | array), matching
// the System One `EntryType` contract.
type JevQuestion struct {
	Type                string   `json:"type"`
	Instructions        any      `json:"instructions,omitempty"`
	Criteria            any      `json:"criteria,omitempty"`
	ConfidenceThreshold *float64 `json:"confidenceThreshold,omitempty"`
}

// IsJev reports whether the constraint is backed by a Jev question.
func (c *Constraint) IsJev() bool { return c.JevType != "" }

// JevName returns the `@jev.<name>` question name, or "" for a plain constraint.
func (c *Constraint) JevName() string {
	if !c.IsJev() {
		return ""
	}
	return strings.TrimPrefix(c.Property, JevPropertyPrefix)
}

// JevQuestion decodes the stored question. Returns nil for a plain constraint.
func (c *Constraint) JevQuestion() (*JevQuestion, error) {
	if !c.IsJev() {
		return nil, nil
	}
	q := &JevQuestion{Type: c.JevType, ConfidenceThreshold: c.JevConfidenceThreshold}
	if c.JevInstructions != "" {
		if err := json.Unmarshal([]byte(c.JevInstructions), &q.Instructions); err != nil {
			return nil, fmt.Errorf("invalid jev instructions: %w", err)
		}
	}
	if c.JevCriteria != "" {
		if err := json.Unmarshal([]byte(c.JevCriteria), &q.Criteria); err != nil {
			return nil, fmt.Errorf("invalid jev criteria: %w", err)
		}
	}
	return q, nil
}

// SetJevQuestion encodes and stores a question, or clears the Jev fields when nil.
func (c *Constraint) SetJevQuestion(q *JevQuestion) error {
	if q == nil {
		c.JevType = ""
		c.JevInstructions = ""
		c.JevCriteria = ""
		c.JevConfidenceThreshold = nil
		return nil
	}
	instructions, err := marshalJevEntry(q.Instructions)
	if err != nil {
		return fmt.Errorf("invalid jev instructions: %w", err)
	}
	criteria, err := marshalJevEntry(q.Criteria)
	if err != nil {
		return fmt.Errorf("invalid jev criteria: %w", err)
	}
	c.JevType = q.Type
	c.JevInstructions = instructions
	c.JevCriteria = criteria
	c.JevConfidenceThreshold = q.ConfidenceThreshold
	return nil
}

// CopyJevFrom copies the Jev question fields from src, so a constraint can be
// cloned (flag template / duplicate) without silently dropping its question.
func (c *Constraint) CopyJevFrom(src *Constraint) {
	c.JevType = src.JevType
	c.JevInstructions = src.JevInstructions
	c.JevCriteria = src.JevCriteria
	c.JevConfidenceThreshold = src.JevConfidenceThreshold
}

func marshalJevEntry(v any) (string, error) {
	if v == nil {
		return "", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DuplicateJevQuestionProperties returns the `@jev.<name>` properties used by
// more than one constraint, sorted. A flag may define each Jev question only
// once, so the single batched System One call has one unambiguous definition.
func DuplicateJevQuestionProperties(constraints []Constraint) []string {
	counts := make(map[string]int)
	for i := range constraints {
		c := &constraints[i]
		if !c.IsJev() {
			continue
		}
		counts[c.Property]++
	}
	names := make([]string, 0)
	for name, count := range counts {
		if count > 1 {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// validateJev validates the Jev question fields. Plain constraints are a no-op.
func (c *Constraint) validateJev() error {
	if !c.IsJev() {
		return nil
	}
	switch c.JevType {
	case JevTypeNoul, JevTypeChoice, JevTypeScore:
	default:
		return fmt.Errorf("invalid jev.type %q: must be one of %s, %s, %s",
			c.JevType, JevTypeNoul, JevTypeChoice, JevTypeScore)
	}
	if !strings.HasPrefix(c.Property, JevPropertyPrefix) || c.JevName() == "" {
		return fmt.Errorf("jev constraints require property %s<name>, got %q", JevPropertyPrefix, c.Property)
	}
	if err := validateJevOperator(c.JevType, c.Operator); err != nil {
		return err
	}
	q, err := c.JevQuestion()
	if err != nil {
		return err
	}
	if q.ConfidenceThreshold != nil && (*q.ConfidenceThreshold < 0 || *q.ConfidenceThreshold > 1) {
		return fmt.Errorf("jev.confidenceThreshold must be within [0,1], got %v", *q.ConfidenceThreshold)
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

// validateJevOperator enforces that the match operator makes sense for the
// question type, so an incompatible operator/value cannot silently never match.
func validateJevOperator(jevType, operator string) error {
	switch jevType {
	case JevTypeNoul, JevTypeScore:
		switch operator {
		case models.ConstraintOperatorGTE, models.ConstraintOperatorGT,
			models.ConstraintOperatorLTE, models.ConstraintOperatorLT:
			return nil
		}
	case JevTypeChoice:
		switch operator {
		case models.ConstraintOperatorEQ, models.ConstraintOperatorNEQ,
			models.ConstraintOperatorIN, models.ConstraintOperatorNOTIN:
			return nil
		}
	default:
		return nil // unknown type is reported by validateJev
	}
	return fmt.Errorf("jev %s does not support operator %q", jevType, operator)
}

func validateNoulCriteria(criteria any) error {
	if criteria == nil {
		return nil
	}
	if _, ok := criteria.(map[string]any); !ok {
		return fmt.Errorf("jev noul criteria must be a JSON object with true/false descriptions")
	}
	return nil
}

func validateChoiceCriteria(criteria any) error {
	m, ok := criteria.(map[string]any)
	if !ok || len(m) == 0 {
		return fmt.Errorf("jev choice criteria must be a non-empty JSON object of option to description")
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

// collectJevQuestions records the segment's Jev constraints on the flag's
// evaluation state. Segments are visited in rank order, so the first
// definition of a question name wins if it appears in more than one segment.
func (f *Flag) collectJevQuestions(s *Segment) {
	for i := range s.Constraints {
		c := &s.Constraints[i]
		if !c.IsJev() {
			continue
		}
		name := c.JevName()
		q, err := c.JevQuestion()
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"flagID":       f.ID,
				"segmentID":    s.ID,
				"constraintID": c.ID,
			}).Warn("skipping invalid jev constraint")
			continue
		}
		if _, exists := f.FlagEvaluation.JevQuestions[name]; exists {
			logrus.WithFields(logrus.Fields{
				"flagID":    f.ID,
				"segmentID": s.ID,
				"question":  name,
			}).Warn("duplicate jev question name; keeping the higher-priority definition")
			continue
		}
		f.FlagEvaluation.JevQuestions[name] = *q
	}
}
