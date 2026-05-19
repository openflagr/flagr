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

func TestConstraintToExpr_RegexEscaping(t *testing.T) {
	// EREG with backslash sequences (the main fix)
	// These previously caused Go scanner "invalid char escape" errors
	t.Run("EREG with \\d pattern", func(t *testing.T) {
		c := Constraint{
			Property: "status",
			Operator: models.ConstraintOperatorEREG,
			Value:    `"\d+"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"status": "123"})
		assert.NoError(t, err)
		assert.True(t, match)
		match, err = conditions.Evaluate(expr, map[string]any{"status": "abc"})
		assert.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("EREG with \\. pattern (literal dot)", func(t *testing.T) {
		c := Constraint{
			Property: "email",
			Operator: models.ConstraintOperatorEREG,
			Value:    `".+@example\.com"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"email": "user@example.com"})
		assert.NoError(t, err)
		assert.True(t, match)
		match, err = conditions.Evaluate(expr, map[string]any{"email": "user@exampleXcom"})
		assert.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("EREG with \\w pattern (word chars)", func(t *testing.T) {
		c := Constraint{
			Property: "name",
			Operator: models.ConstraintOperatorEREG,
			Value:    `"\w+"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"name": "hello"})
		assert.NoError(t, err)
		assert.True(t, match)
		match, err = conditions.Evaluate(expr, map[string]any{"name": "!!!"})
		assert.NoError(t, err)
		assert.False(t, match) // no word chars, \w+ has nothing to match
	})

	t.Run("EREG with \\s pattern (whitespace)", func(t *testing.T) {
		c := Constraint{
			Property: "text",
			Operator: models.ConstraintOperatorEREG,
			Value:    `"hello\sworld"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"text": "hello world"})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	t.Run("EREG with \\b pattern (word boundary)", func(t *testing.T) {
		c := Constraint{
			Property: "text",
			Operator: models.ConstraintOperatorEREG,
			Value:    `"\bword\b"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"text": "word"})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	// NEREG with backslash sequences
	t.Run("NEREG with \\d pattern", func(t *testing.T) {
		c := Constraint{
			Property: "status",
			Operator: models.ConstraintOperatorNEREG,
			Value:    `"\d+"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"status": "abc"})
		assert.NoError(t, err)
		assert.True(t, match)
		match, err = conditions.Evaluate(expr, map[string]any{"status": "123"})
		assert.NoError(t, err)
		assert.False(t, match)
	})

	// Simple regex without backslashes should still work (regression check)
	t.Run("EREG simple pattern without backslash", func(t *testing.T) {
		c := Constraint{
			Property: "status",
			Operator: models.ConstraintOperatorEREG,
			Value:    `"^5[0-9][0-9]$"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"status": "500"})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	// Regex literal form (starts with /) should be unchanged
	t.Run("EREG with regex literal form", func(t *testing.T) {
		c := Constraint{
			Property: "status",
			Operator: models.ConstraintOperatorEREG,
			Value:    `/^5[0-9][0-9]$/`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"status": "500"})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	// Pattern containing "/" falls back to string form
	t.Run("EREG with pattern containing slash", func(t *testing.T) {
		c := Constraint{
			Property: "path",
			Operator: models.ConstraintOperatorEREG,
			Value:    `"/api/v1/.+"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"path": "/api/v1/users"})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	// Variable reference as value (unchanged)
	t.Run("EREG with var ref value", func(t *testing.T) {
		c := Constraint{
			Property: "status",
			Operator: models.ConstraintOperatorEREG,
			Value:    `{pattern}`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{
			"status":  "123",
			"pattern": `\d+`,
		})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	// Trimmed value handling
	t.Run("EREG with trimmed value containing spaces", func(t *testing.T) {
		c := Constraint{
			Property: "status",
			Operator: models.ConstraintOperatorEREG,
			Value:    `  "\d+"  `,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"status": "123"})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	// Non-EREG operator should be unaffected
	t.Run("EQ operator unaffected by regex fix", func(t *testing.T) {
		c := Constraint{
			Property: "name",
			Operator: models.ConstraintOperatorEQ,
			Value:    `"Alice"`,
		}
		expr, err := c.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{"name": "Alice"})
		assert.NoError(t, err)
		assert.True(t, match)
	})

	// Multiple constraints with regex in ConstraintArray
	t.Run("ConstraintArray with regex patterns", func(t *testing.T) {
		cs := ConstraintArray{
			{
				Property: "email",
				Operator: models.ConstraintOperatorEREG,
				Value:    `".+@example\.com"`,
			},
			{
				Property: "status",
				Operator: models.ConstraintOperatorEREG,
				Value:    `"\d+"`,
			},
		}
		expr, err := cs.ToExpr()
		assert.NoError(t, err)
		match, err := conditions.Evaluate(expr, map[string]any{
			"email":  "user@example.com",
			"status": "200",
		})
		assert.NoError(t, err)
		assert.True(t, match)
		match, err = conditions.Evaluate(expr, map[string]any{
			"email":  "user@example.com",
			"status": "abc",
		})
		assert.NoError(t, err)
		assert.False(t, match)
	})
}
