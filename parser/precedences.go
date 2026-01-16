package parser

import "gloss/token"

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // + or -
	PRODUCT     // * or /
	PREFIX      // -X or !X
	CALL        // fn(X)
)

var precedences = map[token.TokenType]int{
	token.ASSIGN: EQUALS,
	// token.NOT_EQ:  EQUALS,
	token.LANGLE: LESSGREATER,
	token.RANGLE: LESSGREATER,
	token.PLUS:   SUM,
	token.MINUS:  SUM,
	token.DIV:    PRODUCT,
	// token.ASTERIX: PRODUCT,
	token.LPAREN: CALL,
}

