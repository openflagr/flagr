package conditions

// Token represents a lexical token.
type Token int

const (
	// ILLEGAL token represent illegal token found in the statement
	ILLEGAL Token = iota
	// EOF token represents end of statement
	EOF

	// Literals
	literalBegin
	IDENT  // Variable references $0, $5, etc
	NUMBER // 12345.67
	STRING // "abc"
	ARRAY  // array of values (string or number) ["a","b","c"]  [342,4325,6,4]
	TRUE   // true
	FALSE  // false
	literalEnd

	operatorBegin
	AND         // AND
	OR          // OR
	EQ          // =
	NEQ         // !=
	LT          // <
	LTE         // <=
	GT          // >
	GTE         // >=
	NAND        // NAND
	XOR         // XOR
	EREG        // =~
	NEREG       // !~
	IN          // IN
	NOTIN       // NOT IN
	CONTAINS    // CONTAINS
	NOTCONTAINS // NOT CONTAINS
	operatorEnd

	LPAREN // (
	RPAREN // )
)

var tokens = []string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	IDENT:  "IDENT",
	NUMBER: "NUMBER",
	STRING: "STRING",
	ARRAY:  "ARRAY",
	TRUE:   "TRUE",
	FALSE:  "FALSE",

	AND: "AND",
	OR:  "OR",
	EQ:  "==",
	NEQ: "!=",
	LT:  "<",
	LTE: "<=",
	GT:  ">",
	GTE: ">=",

	NAND:        "NAND",
	XOR:         "XOR",
	EREG:        "=~",
	NEREG:       "!~",
	IN:          "IN",
	NOTIN:       "NOT IN",
	CONTAINS:    "CONTAINS",
	NOTCONTAINS: "NOT CONTAINS",

	LPAREN: "(",
	RPAREN: ")",
}

// String returns the string representation of the token.
func (tok Token) String() string {
	if tok >= 0 && tok < Token(len(tokens)) {
		return tokens[tok]
	}
	return ""
}

// Precedence returns the operator precedence of the binary operator token.
func (tok Token) Precedence() int {
	switch tok {
	case OR, XOR:
		return 1
	case AND, NAND:
		return 2
	case EQ, NEQ, LT, LTE, GT, GTE, IN, NOTIN, EREG, NEREG, CONTAINS, NOTCONTAINS:
		return 3
	}
	return 0
}

// isOperator returns true for operator tokens.
func (tok Token) isOperator() bool { return tok > operatorBegin && tok < operatorEnd }

// tokstr returns a literal if provided, otherwise returns the token string.
func tokstr(tok Token, lit string) string {
	if lit != "" {
		return lit
	}
	return tok.String()
}
