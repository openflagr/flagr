package entity

import (
	"testing"

	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/stretchr/testify/assert"
)

func TestConstraintToExpr(t *testing.T) {
	t.Run("empty case", func(t *testing.T) {
		c := Constraint{}
		expr, err := c.ToExpr()
		assert.Error(t, err)
		assert.Nil(t, expr)
	})

	t.Run("not supported operator case", func(t *testing.T) {
		c := Constraint{
			SegmentID: 0,
			Property:  "dl_state",
			Operator:  "===",
			Value:     "\"CA\"",
		}
		expr, err := c.ToExpr()
		assert.Error(t, err)
		assert.Nil(t, expr)
	})

	t.Run("parse error", func(t *testing.T) {
		c := Constraint{
			SegmentID: 0,
			Property:  "dl_state",
			Operator:  models.ConstraintOperatorEQ,
			Value:     "\"CA\"]", // Invalid "]"
		}
		expr, err := c.ToExpr()
		assert.Error(t, err)
		assert.Nil(t, expr)
	})

	t.Run("happy code path", func(t *testing.T) {
		c := Constraint{
			SegmentID: 0,
			Property:  "dl_state",
			Operator:  models.ConstraintOperatorEQ,
			Value:     "\"CA\"",
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		assert.NotNil(t, expr)
	})
}

func TestConstraintValidate(t *testing.T) {
	t.Run("empty case", func(t *testing.T) {
		c := Constraint{}
		assert.Error(t, c.Validate())
	})

	t.Run("happy code path", func(t *testing.T) {
		c := Constraint{
			SegmentID: 0,
			Property:  "dl_state",
			Operator:  models.ConstraintOperatorEQ,
			Value:     "\"CA\"",
		}
		assert.NoError(t, c.Validate())
	})
}
