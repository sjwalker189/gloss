package parser

import (
	"gloss/token"
)

const (
	LOWEST      = 0
	ASSIGN      = 1
	LOGICAL_OR  = 2  // ||
	LOGICAL_AND = 3  // &&
	BITWISE_OR  = 4  // |
	BITWISE_XOR = 5  // ^
	BITWISE_AND = 6  // &
	EQUALITY    = 7  // ==, !=
	RELATIONAL  = 8  // <, <=, >, >=
	SHIFT       = 9  // <<, >>
	ADD         = 10 // +, -
	MULT        = 11 // *, /, %
	UNARY       = 12 // !, -, +, ~
	CALL        = 13
	MEMBER      = 14 // .
	POSTFIX     = 15 // ++, --
	INDEX       = 16 // [
)

var precedences = map[token.TokenType]int{
	token.EQ:            EQUALITY,
	token.NOT_EQ:        EQUALITY,
	token.LT:            RELATIONAL,
	token.LT_EQ:         RELATIONAL,
	token.GT:            RELATIONAL,
	token.GT_EQ:         RELATIONAL,
	token.PLUS:          ADD,
	token.MINUS:         ADD,
	token.MUL:           MULT,
	token.DIV:           MULT,
	token.MOD:           MULT,
	token.BITSHIFTL:     SHIFT,
	token.BITSHIFTR:     SHIFT,
	token.BITWISE_AND:   BITWISE_AND,
	token.BITWISE_OR:    BITWISE_OR,
	token.BITWISE_XOR:   BITWISE_XOR,
	token.AND:           LOGICAL_AND,
	token.OR:            LOGICAL_OR,
	token.LPAREN:        CALL,
	token.PERIOD:        MEMBER,
	token.QUESTION_DOT:  MEMBER, // Optional chaining
	token.LBRACKET:      INDEX,
	token.ASSIGN:        ASSIGN,
	token.PLUS_ASSIGN:   ASSIGN,
	token.MINUS_ASSIGN:  ASSIGN,
	token.MUL_ASSIGN:    ASSIGN,
	token.DIV_ASSIGN:    ASSIGN,
	token.MOD_ASSIGN:    ASSIGN,
	token.AND_ASSIGN:    ASSIGN,
	token.OR_ASSIGN:     ASSIGN,
	token.XOR_ASSIGN:    ASSIGN,
	token.SHL_ASSIGN:    ASSIGN,
	token.SHR_ASSIGN:    ASSIGN,
	token.INC:           POSTFIX,
	token.DEC:           POSTFIX,
}
