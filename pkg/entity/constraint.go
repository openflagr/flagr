package entity

import (
	"fmt"
	"strings"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/zhouzhuojie/conditions"
	"gorm.io/gorm"
)

// Constraint is the unit of constraints
type Constraint struct {
	gorm.Model

	SegmentID uint `gorm:"index:idx_constraint_segmentid"`
	Property  string
	Operator  string
	Value     string `gorm:"type:text"`
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
	_, err := c.ToExpr()
	return err
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
