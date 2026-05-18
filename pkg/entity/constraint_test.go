package entity

import (
	"testing"

	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/stretchr/testify/assert"
	"github.com/zhouzhuojie/conditions"
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

func TestConstraintToExpr_NestedField(t *testing.T) {
	t.Run("dotted path property - EQ", func(t *testing.T) {
		c := Constraint{
			Property: "user.name",
			Operator: models.ConstraintOperatorEQ,
			Value:    `"Alice"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		assert.NotNil(t, expr)

		// Should produce: ({user.name} == "Alice")
		// Evaluated against nested context: {"user": {"name": "Alice"}}
		match, err := conditions.Evaluate(expr, map[string]any{"user": map[string]any{"name": "Alice"}})
		assert.NoError(t, err)
		assert.True(t, match)

		match, err = conditions.Evaluate(expr, map[string]any{"user": map[string]any{"name": "Bob"}})
		assert.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("dotted path property - GT", func(t *testing.T) {
		c := Constraint{
			Property: "user.age",
			Operator: models.ConstraintOperatorGT,
			Value:    "18",
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		assert.NotNil(t, expr)

		match, err := conditions.Evaluate(expr, map[string]any{"user": map[string]any{"age": float64(25)}})
		assert.NoError(t, err)
		assert.True(t, match)

		match, err = conditions.Evaluate(expr, map[string]any{"user": map[string]any{"age": float64(15)}})
		assert.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("array index property", func(t *testing.T) {
		c := Constraint{
			Property: "users[0]",
			Operator: models.ConstraintOperatorEQ,
			Value:    `"admin"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		assert.NotNil(t, expr)

		match, err := conditions.Evaluate(expr, map[string]any{
			"users": []any{"admin", "viewer"},
		})
		assert.NoError(t, err)
		assert.True(t, match)

		match, err = conditions.Evaluate(expr, map[string]any{
			"users": []any{"viewer", "admin"},
		})
		assert.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("array index with dot chained", func(t *testing.T) {
		c := Constraint{
			Property: "users[0].name",
			Operator: models.ConstraintOperatorEQ,
			Value:    `"Alice"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		assert.NotNil(t, expr)

		match, err := conditions.Evaluate(expr, map[string]any{
			"users": []any{
				map[string]any{"name": "Alice"},
				map[string]any{"name": "Bob"},
			},
		})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	t.Run("nested path with IN operator", func(t *testing.T) {
		c := Constraint{
			Property: "user.role",
			Operator: models.ConstraintOperatorIN,
			Value:    `["admin", "moderator"]`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		assert.NotNil(t, expr)

		match, err := conditions.Evaluate(expr, map[string]any{"user": map[string]any{"role": "admin"}})
		assert.NoError(t, err)
		assert.True(t, match)

		match, err = conditions.Evaluate(expr, map[string]any{"user": map[string]any{"role": "viewer"}})
		assert.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("nested path with CONTAINS", func(t *testing.T) {
		c := Constraint{
			Property: "user.tags",
			Operator: models.ConstraintOperatorCONTAINS,
			Value:    `"urgent"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		assert.NotNil(t, expr)

		match, err := conditions.Evaluate(expr, map[string]any{
			"user": map[string]any{"tags": []string{"urgent", "billing"}},
		})
		assert.NoError(t, err)
		assert.True(t, match)

		match, err = conditions.Evaluate(expr, map[string]any{
			"user": map[string]any{"tags": []string{"normal", "billing"}},
		})
		assert.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("nested path with regex (string pattern)", func(t *testing.T) {
		c := Constraint{
			Property: "user.status",
			Operator: models.ConstraintOperatorEREG,
			Value:    `"^5[0-9][0-9]$"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		assert.NotNil(t, expr)

		match, err := conditions.Evaluate(expr, map[string]any{
			"user": map[string]any{"status": "500"},
		})
		assert.NoError(t, err)
		assert.True(t, match)

		match, err = conditions.Evaluate(expr, map[string]any{
			"user": map[string]any{"status": "400"},
		})
		assert.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("nested path with regex (literal pattern)", func(t *testing.T) {
		c := Constraint{
			Property: "user.status",
			Operator: models.ConstraintOperatorEREG,
			Value:    `/^5[0-9][0-9]$/`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		assert.NotNil(t, expr)

		match, err := conditions.Evaluate(expr, map[string]any{
			"user": map[string]any{"status": "500"},
		})
		assert.NoError(t, err)
		assert.True(t, match)

		match, err = conditions.Evaluate(expr, map[string]any{
			"user": map[string]any{"status": "400"},
		})
		assert.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("nested path with NOT CONTAINS", func(t *testing.T) {
		c := Constraint{
			Property: "user.exclusions",
			Operator: models.ConstraintOperatorNOTCONTAINS,
			Value:    `"banned"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		assert.NotNil(t, expr)

		match, err := conditions.Evaluate(expr, map[string]any{
			"user": map[string]any{"exclusions": []string{"active"}},
		})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	t.Run("nested path validates correctly", func(t *testing.T) {
		c := Constraint{
			Property: "user.name",
			Operator: models.ConstraintOperatorEQ,
			Value:    `"Alice"`,
		}
		assert.NoError(t, c.Validate())
	})

	t.Run("nested path with missing key errors", func(t *testing.T) {
		c := Constraint{
			Property: "user.missing",
			Operator: models.ConstraintOperatorEQ,
			Value:    `"Alice"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)

		_, err = conditions.Evaluate(expr, map[string]any{"user": map[string]any{"name": "Alice"}})
		assert.Error(t, err)
	})

	t.Run("deeply nested path", func(t *testing.T) {
		c := Constraint{
			Property: "a.b.c.d",
			Operator: models.ConstraintOperatorEQ,
			Value:    "42",
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)

		match, err := conditions.Evaluate(expr, map[string]any{
			"a": map[string]any{
				"b": map[string]any{
					"c": map[string]any{"d": float64(42)},
				},
			},
		})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	t.Run("nested path in ConstraintArray", func(t *testing.T) {
		cs := ConstraintArray{
			{
				Property: "user.tier",
				Operator: models.ConstraintOperatorIN,
				Value:    `["premium", "enterprise"]`,
			},
			{
				Property: "user.age",
				Operator: models.ConstraintOperatorGT,
				Value:    "18",
			},
		}
		expr, err := cs.ToExpr()
		assert.NoError(t, err)

		match, err := conditions.Evaluate(expr, map[string]any{
			"user": map[string]any{"tier": "premium", "age": float64(25)},
		})
		assert.NoError(t, err)
		assert.True(t, match)

		match, err = conditions.Evaluate(expr, map[string]any{
			"user": map[string]any{"tier": "free", "age": float64(25)},
		})
		assert.NoError(t, err)
		assert.False(t, match)
	})
}
