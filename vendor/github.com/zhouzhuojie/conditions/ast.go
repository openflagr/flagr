package conditions

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DataType represents the primitive data types available in InfluxQL.
type DataType string

const (
	Unknown  = DataType("")
	Number   = DataType("number")
	Boolean  = DataType("boolean")
	String   = DataType("string")
	Time     = DataType("time")
	Duration = DataType("duration")
)

// InspectDataType returns the data type of a given value.
func InspectDataType(v interface{}) DataType {
	switch v.(type) {
	case float64:
		return Number
	case bool:
		return Boolean
	case string:
		return String
	case time.Time:
		return Time
	case time.Duration:
		return Duration
	default:
		return Unknown
	}
}

// Node represents a node in the conditions abstract syntax tree.
type Node interface {
	node()
	String() string
}

func (_ *VarRef) node()             {}
func (_ *NumberLiteral) node()      {}
func (_ *StringLiteral) node()      {}
func (_ *BooleanLiteral) node()     {}
func (_ *TimeLiteral) node()        {}
func (_ *DurationLiteral) node()    {}
func (_ *BinaryExpr) node()         {}
func (_ *ParenExpr) node()          {}
func (_ *SliceStringLiteral) node() {}
func (_ *SliceNumberLiteral) node() {}

// Expr represents an expression that can be evaluated to a value.
type Expr interface {
	Node
	expr()
	Args() []string
}

func (_ *VarRef) expr()             {}
func (_ *NumberLiteral) expr()      {}
func (_ *StringLiteral) expr()      {}
func (_ *BooleanLiteral) expr()     {}
func (_ *TimeLiteral) expr()        {}
func (_ *DurationLiteral) expr()    {}
func (_ *BinaryExpr) expr()         {}
func (_ *ParenExpr) expr()          {}
func (_ *SliceStringLiteral) expr() {}
func (_ *SliceNumberLiteral) expr() {}

// VarRef represents a reference to a variable.
type VarRef struct {
	Val string
}

// String returns a string representation of the variable reference.
func (r *VarRef) String() string { return QuoteIdent(r.Val) }

func (r *VarRef) Args() []string {
	return []string{r.Val}
}

// NumberLiteral represents a numeric literal.
type NumberLiteral struct {
	Val float64
}

// String returns a string representation of the literal.
func (l *NumberLiteral) String() string { return strconv.FormatFloat(l.Val, 'f', 3, 64) }

func (n *NumberLiteral) Args() []string {
	args := []string{}
	return args
}

type SliceStringLiteral struct {
	Val []string
}

// String returns a string representation of the literal.
func (l *SliceStringLiteral) String() string {
	return fmt.Sprintf("%s", l.Val)
}

func (l *SliceStringLiteral) Args() []string {
	args := []string{}
	return args
}

type SliceNumberLiteral struct {
	Val []float64
}

// String returns a string representation of the literal.
func (l *SliceNumberLiteral) String() string {
	return fmt.Sprintf("%v", l.Val)
}

func (l *SliceNumberLiteral) Args() []string {
	args := []string{}
	return args
}

// BooleanLiteral represents a boolean literal.
type BooleanLiteral struct {
	Val bool
}

// String returns a string representation of the literal.
func (l *BooleanLiteral) String() string {
	if l.Val {
		return "true"
	}
	return "false"
}

func (l *BooleanLiteral) Args() []string {
	args := []string{}
	return args
}

// StringLiteral represents a string literal.
type StringLiteral struct {
	Val string
}

// String returns a string representation of the literal.
func (l *StringLiteral) String() string { return Quote(l.Val) }

// TimeLiteral represents a point-in-time literal.
type TimeLiteral struct {
	Val time.Time
}

func (l *StringLiteral) Args() []string {
	args := []string{}
	return args
}

// String returns a string representation of the literal.
func (l *TimeLiteral) String() string { return l.Val.UTC().Format("2006-01-02 15:04:05.999") }

// DurationLiteral represents a duration literal.
type DurationLiteral struct {
	Val time.Duration
}

// String returns a string representation of the literal.
func (l *DurationLiteral) String() string { return FormatDuration(l.Val) }

// BinaryExpr represents an operation between two expressions.
type BinaryExpr struct {
	Op  Token
	LHS Expr
	RHS Expr
}

// String returns a string representation of the binary expression.
func (e *BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.LHS.String(), e.Op, e.RHS.String())
}

func (e *BinaryExpr) Args() []string {
	args := []string{}

	args = append(e.LHS.Args(), args...)
	args = append(e.RHS.Args(), args...)

	return args
}

// ParenExpr represents a parenthesized expression.
type ParenExpr struct {
	Expr Expr
}

// String returns a string representation of the parenthesized expression.
func (e *ParenExpr) String() string { return fmt.Sprintf("(%s)", e.Expr.String()) }

func (p *ParenExpr) Args() []string {
	args := []string{}
	args = append(p.Expr.Args(), args...)

	return args
}

// Visitor can be called by Walk to traverse an AST hierarchy.
// The Visit() function is called once per node.
type Visitor interface {
	Visit(Node) Visitor
}

// Walk traverses a node hierarchy in depth-first order.
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *BinaryExpr:
		Walk(v, n.LHS)
		Walk(v, n.RHS)

	case *ParenExpr:
		Walk(v, n.Expr)
	}
}

// WalkFunc traverses a node hierarchy in depth-first order.
func WalkFunc(node Node, fn func(Node)) {
	Walk(walkFuncVisitor(fn), node)
}

type walkFuncVisitor func(Node)

func (fn walkFuncVisitor) Visit(n Node) Visitor { fn(n); return fn }

// Quote returns a quoted string.
func Quote(s string) string {
	return `"` + strings.NewReplacer("\n", `\n`, `\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

// QuoteIdent returns a quoted identifier if the identifier requires quoting.
// Otherwise returns the original string passed in.
func QuoteIdent(s string) string {
	if s == "" || regexp.MustCompile(`[^a-zA-Z_.]`).MatchString(s) {
		return Quote(s)
	}
	return s
}

// FormatDuration formats a duration to a string.
func FormatDuration(d time.Duration) string {
	if d%(7*24*time.Hour) == 0 {
		return fmt.Sprintf("%dw", d/(7*24*time.Hour))
	} else if d%(24*time.Hour) == 0 {
		return fmt.Sprintf("%dd", d/(24*time.Hour))
	} else if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", d/time.Hour)
	} else if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", d/time.Minute)
	} else if d%time.Second == 0 {
		return fmt.Sprintf("%ds", d/time.Second)
	} else if d%time.Millisecond == 0 {
		return fmt.Sprintf("%dms", d/time.Millisecond)
	} else {
		return fmt.Sprintf("%d", d/time.Microsecond)
	}
}
