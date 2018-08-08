//go:generate goqueryset -in constraint.go

package entity

import (
	"fmt"
	"strings"

	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/jinzhu/gorm"
	"github.com/zhouzhuojie/conditions"
)

// Constraint is the unit of constraints
// gen:qs
type Constraint struct {
	gorm.Model

	SegmentID uint `gorm:"index:idx_constraint_segmentid"`
	Property  string
	Operator  string
	Value     string `sql:"type:text"`
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
		return nil, fmt.Errorf("%s. Note: if it's string or array of string, warp it with quotes \"...\"", err)
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

	return fmt.Sprintf("({%s} %s %s)", c.Property, o, c.Value), nil
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
		return nil, fmt.Errorf("%s. Note: if it's string or array of string, wrap it with quotes \"...\"", err)
	}
	return expr, nil
}
