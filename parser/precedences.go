package parser

import "gloss/token"

const (
	_ int = iota
	LOWEST
	OR          // or
	AND         // and
	BITWISE_OR  // | or ^
	BITWISE_AND // &
	EQUALS      // == or !=
	LESSGREATER // > or < or >= or <=
	BITSHIFT    // << or >>
	SUM         // + or -
	PRODUCT     // * or / or %
	PREFIX      // -X or !X or ~X
	CALL        // fn(X)
)

var precedences = map[token.TokenType]int{
	// Equality and Logic
	token.EQ:     EQUALS,
	token.NOT_EQ: EQUALS,
	token.AND:    AND,
	token.OR:     OR,

	// Comparisons
	token.LANGLE: LESSGREATER,
	token.RANGLE: LESSGREATER,

	token.LT:    LESSGREATER,
	token.GT:    LESSGREATER,
	token.LT_EQ: LESSGREATER,
	token.GT_EQ: LESSGREATER,

	// Bitwise
	token.BITWISE_OR:  BITWISE_OR,
	token.BITWISE_XOR: BITWISE_OR,
	token.BITWISE_AND: BITWISE_AND,
	token.BITSHIFTL:   BITSHIFT,
	token.BITSHIFTR:   BITSHIFT,

	// Math
	token.PLUS:  SUM,
	token.MINUS: SUM,
	token.MUL:   PRODUCT,
	token.DIV:   PRODUCT,
	token.MOD:   PRODUCT,

	// Access / Calls
	token.LPAREN: CALL,
}
