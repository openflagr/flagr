package entity

import (
	"testing"

	"github.com/openflagr/flagr/swagger_gen/models"
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

	t.Run("parse error - invalid ]", func(t *testing.T) {
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

	t.Run("parse error - no quotes", func(t *testing.T) {
		c := Constraint{
			SegmentID: 0,
			Property:  "dl_state",
			Operator:  models.ConstraintOperatorEQ,
			Value:     "NY", // Invalid string b/c no ""
		}
		expr, err := c.ToExpr()
		assert.Error(t, err)
		assert.Nil(t, expr)
	})

	t.Run("parse error - no quotes in array", func(t *testing.T) {
		c := Constraint{
			SegmentID: 0,
			Property:  "dl_state",
			Operator:  models.ConstraintOperatorIN,
			Value:     "[NY]", // Invalid string b/c no ""
		}
		expr, err := c.ToExpr()
		assert.Error(t, err)
		assert.Nil(t, expr)
	})

	t.Run("happy code path - single EQ", func(t *testing.T) {
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

	t.Run("happy code path - IN", func(t *testing.T) {
		c := Constraint{
			SegmentID: 0,
			Property:  "dl_state",
			Operator:  models.ConstraintOperatorIN,
			Value:     `["CA", "NY"]`,
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

func TestConstraintArray(t *testing.T) {
	cs := ConstraintArray{
		{
			SegmentID: 0,
			Property:  "dl_state",
			Operator:  models.ConstraintOperatorIN,
			Value:     `["CA", "NY"]`,
		},
		{
			SegmentID: 0,
			Property:  "state",
			Operator:  models.ConstraintOperatorEQ,
			Value:     `{dl_state}`,
		},
	}
	expr, err := cs.ToExpr()
	assert.NoError(t, err)
	assert.NotNil(t, expr)
}
