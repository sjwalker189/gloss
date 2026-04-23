package parser

import "gloss/token"

const (
	_ int = iota
	LOWEST
	ASSIGN
	LOGICAL_OR
	LOGICAL_AND
	BITWISE_OR
	BITWISE_XOR
	BITWISE_AND
	EQUALITY
	RELATIONAL
	SHIFT
	ADD
	MULT
	UNARAY
	PREFIX
	CALL
	MEMBER
	POSTFIX
	INDEX
)

var precedences = map[token.TokenType]int{
	// Equality and Logic
	token.EQ:     EQUALITY,
	token.NOT_EQ: EQUALITY,
	token.AND:    LOGICAL_AND,
	token.OR:     LOGICAL_OR,

	// Comparisons
	token.LANGLE: RELATIONAL,
	token.RANGLE: RELATIONAL,

	token.LT:    RELATIONAL,
	token.GT:    RELATIONAL,
	token.LT_EQ: RELATIONAL,
	token.GT_EQ: RELATIONAL,

	// Bitwise
	token.BITWISE_OR:  BITWISE_OR,
	token.BITWISE_XOR: BITWISE_XOR,
	token.BITWISE_AND: BITWISE_AND,
	token.BITSHIFTL:   SHIFT,
	token.BITSHIFTR:   SHIFT,

	// Math
	token.PLUS:  ADD,
	token.MINUS: ADD,
	token.MUL:   MULT,
	token.DIV:   MULT,
	token.MOD:   MULT,

	// Access / Calls
	token.LPAREN: CALL,
}
