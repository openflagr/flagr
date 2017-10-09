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

	SegmentID uint `gorm:"index:idx_segmentid"`
	Property  string
	Operator  string
	Value     string `sql:"type:text"`
}

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
	if c.Property == "" || c.Operator == "" || c.Value == "" {
		return nil, fmt.Errorf(
			"empty Property/Operator/Value: %s/%s/%s",
			c.Property,
			c.Operator,
			c.Value,
		)
	}
	o, ok := OperatorToExprMap[c.Operator]
	if !ok {
		return nil, fmt.Errorf("not supported operator: %s", c.Operator)
	}
	s := fmt.Sprintf("[%s] %s %s", c.Property, o, c.Value)
	p := conditions.NewParser(strings.NewReader(s))
	expr, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return expr, nil
}

// Validate validates Constraint
func (c *Constraint) Validate() error {
	_, err := c.ToExpr()
	if err != nil {
		return err
	}
	return nil
}
