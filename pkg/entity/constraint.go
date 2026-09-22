package entity

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/zhouzhuojie/conditions"
	"gorm.io/gorm"
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

// Constraint is the unit of constraints
type Constraint struct {
	gorm.Model

	SegmentID uint `gorm:"index:idx_constraint_segmentid"`
	Property  string
	Operator  string
	Value     string `gorm:"type:text"`

	// Jev / System One question backing this constraint. Empty JevType means
	// this is a plain entityContext constraint.
	JevType                string   `gorm:"type:varchar(16)"`
	JevInstructions        string   `gorm:"type:text"`
	JevCriteria            string   `gorm:"type:text"`
	JevConfidenceThreshold *float64 // nil means use the global default
}

// JevQuestion is the authored System One question behind a Jev constraint.
// Instructions and Criteria are free-form JSON (string | object | array), matching
// the System One `EntryType` contract.
type JevQuestion struct {
	Type                string   `json:"type"`
	Instructions        any      `json:"instructions,omitempty"`
	Criteria            any      `json:"criteria,omitempty"`
	ConfidenceThreshold *float64 `json:"confidenceThreshold,omitempty"`
}

// JevConstraintSpec is the evaluation-time representation of a Jev question.
type JevConstraintSpec struct {
	Name                string
	Type                string
	Instructions        any
	Criteria            any
	ConfidenceThreshold *float64
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

// ConstraintArray is an array of Constraint
type ConstraintArray []Constraint

// OperatorToExprMap maps from the swagger model operator to condition operator
var OperatorToExprMap = map[string]string{
	models.ConstraintOperatorEQ:          "==",
	models.ConstraintOperatorNEQ:         "!=",
	models.ConstraintOperatorLT:          "<",
	models.ConstraintOperatorLTE:         "<=",
	models.ConstraintOperatorGT:          ">",
	models.ConstraintOperatorGTE:         ">=",
	models.ConstraintOperatorEREG:        "=~",
	models.ConstraintOperatorNEREG:       "!~",
	models.ConstraintOperatorIN:          "IN",
	models.ConstraintOperatorNOTIN:       "NOT IN",
	models.ConstraintOperatorCONTAINS:    "CONTAINS",
	models.ConstraintOperatorNOTCONTAINS: "NOT CONTAINS",
}

// ToExpr transfer the constraint to conditions.Expr for evaluation
func (c *Constraint) ToExpr() (conditions.Expr, error) {
	s, err := c.toExprStr()
	if err != nil {
		return nil, err
	}
	p := conditions.NewParser(strings.NewReader(s))
	expr, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("%s. Note: if it's string or array of string, wrap it with quotes \"...\" For regex patterns, use a quoted string with Go regexp syntax (e.g. \"\\d+\" for digits)", err)
	}
	return expr, nil
}

func (c *Constraint) toExprStr() (string, error) {
	if c.Property == "" || c.Operator == "" || c.Value == "" {
		return "", fmt.Errorf(
			"empty Property/Operator/Value: %s/%s/%s",
			c.Property,
			c.Operator,
			c.Value,
		)
	}
	o, ok := OperatorToExprMap[c.Operator]
	if !ok {
		return "", fmt.Errorf("not supported operator: %s", c.Operator)
	}

	// Trim the value to be resilient against untrimmed values from API callers.
	val := strings.TrimSpace(c.Value)

	// For EREG/NEREG with quoted string values, use regex literal form /pattern/
	// to avoid Go text/scanner escape issues with sequences like \d, \., \s, etc.
	// The scanner interprets escape sequences inside quoted strings and rejects
	// unrecognized ones (like \., \d), but regex literals are read character-by-character
	// without escape processing, so patterns with backslashes work correctly.
	if (c.Operator == models.ConstraintOperatorEREG || c.Operator == models.ConstraintOperatorNEREG) &&
		isQuotedString(val) {
		pattern := val[1 : len(val)-1]
		// Only use regex literal form when the pattern doesn't contain "/",
		// because the conditions parser doesn't support escaping "/" inside //.
		if !strings.Contains(pattern, "/") {
			return fmt.Sprintf("({%s} %s /%s/)", c.Property, o, pattern), nil
		}
	}

	return fmt.Sprintf("({%s} %s %s)", c.Property, o, val), nil
}

// isQuotedString reports whether s is a double-quoted string like "foo".
func isQuotedString(s string) bool {
	return len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"'
}

// Validate validates Constraint
func (c *Constraint) Validate() error {
	if err := c.validateJev(); err != nil {
		return err
	}
	_, err := c.ToExpr()
	return err
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

// ToExpr maps ConstraintArray to expr by joining 'AND'
func (cs ConstraintArray) ToExpr() (conditions.Expr, error) {
	strs := make([]string, 0, len(cs))
	for _, c := range cs {
		s, err := c.toExprStr()
		if err != nil {
			return nil, err
		}
		strs = append(strs, s)
	}
	exprStr := strings.Join(strs, " AND ")
	p := conditions.NewParser(strings.NewReader(exprStr))
	expr, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("%s. Note: if it's string or array of string, wrap it with quotes \"...\" For regex patterns, use a quoted string with Go regexp syntax (e.g. \"\\d+\" for digits)", err)
	}
	return expr, nil
}
