package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// TODO: Convert to iota

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT      = "IDENT"
	TYPE_IDENT = "TYPE_IDENT"
	INT        = "INT"
	STRING     = "STR"
	BOOL   = "BOOL"

	// Operators
	ASSIGN = "ASSIGN"
	PLUS   = "PLUS"
	MINUS  = "MINUS"
	MUL    = "MUL"
	DIV    = "DIV"
	MOD    = "MOD"
	EQ     = "EQ"
	NOT_EQ = "NOT_EQ"
	LT     = "LESS_THAN"
	LT_EQ  = "LESS_THAN_OR_EQUAL"
	GT     = "GREATER_THAN"
	GT_EQ  = "GREATER_THAN_OR_EQUAL"
	AND    = "AND"
	OR     = "OR"

	BITWISE_OR  = "B_OR"
	BITWISE_XOR = "B_XOR"
	BITWISE_NOT = "B_NOT"
	BITWISE_AND = "B_AND"
	BITSHIFTL   = "B_SHIFT_L"
	BITSHIFTR   = "B_SHIFT_R"

	// New Operators
	ARROW        = "ARROW"        // =>
	QUESTION_DOT = "QUESTION_DOT" // ?.
	INC          = "INC"          // ++
	DEC          = "DEC"          // --

	// Compound Assignments
	PLUS_ASSIGN  = "PLUS_ASSIGN"  // +=
	MINUS_ASSIGN = "MINUS_ASSIGN" // -=
	MUL_ASSIGN   = "MUL_ASSIGN"   // *=
	DIV_ASSIGN   = "DIV_ASSIGN"   // /=
	MOD_ASSIGN   = "MOD_ASSIGN"   // %=
	AND_ASSIGN   = "AND_ASSIGN"   // &=
	OR_ASSIGN    = "OR_ASSIGN"    // |=
	XOR_ASSIGN   = "XOR_ASSIGN"   // ^=
	SHL_ASSIGN   = "SHL_ASSIGN"   // <<=
	SHR_ASSIGN   = "SHR_ASSIGN"   // >>=

	// Delimiters
	PERIOD    = "PERIOD"
	CARET     = "CARET"
	BANG      = "BANG"
	QUOTE     = "QUOTE"
	TICK      = "TICK"
	BACKTICK  = "BACKTICK"
	COMMA     = "COMMA"
	COLON     = "COLON"
	SEMICOLON = "SEMICOLON"
	LPAREN    = "LPAREN"
	RPAREN    = "RPAREN"
	LBRACE    = "LBRACE"
	RBRACE    = "RBRACE"
	LBRACKET  = "LBRACKET"
	RBRACKET  = "RBRACKET"
	LANGLE    = "LANGLE"
	RANGLE    = "RANGLE"

	// Keywords
	LET                 = "LET"
	CONST               = "CONST"
	FUNC                = "FUNC"
	IMPORT              = "IMPORT" // Keeping IMPORT if needed, but USE is the new one
	USE                 = "USE"
	TYPE                = "TYPE"
	ENUM                = "ENUM"
	UNION               = "UNION"
	STRUCT              = "STRUCT"
	EXTERN              = "EXTERN"
	PUB                 = "PUB"
	IF                  = "IF"
	ELSE                = "ELSE"
	SWITCH              = "SWITCH"
	MATCH               = "MATCH"
	CASE                = "CASE"
	DEFAULT             = "DEFAULT"
	FOR                 = "FOR"
	WHILE               = "WHILE"
	LOOP                = "LOOP"
	CONTINUE            = "CONTINUE"
	BREAK               = "BREAK"
	RETURN              = "RETURN"
	IN                  = "IN"
	ELEMENT_OPEN_START  = "EL_OPEN_START"
	ELEMENT_OPEN_END    = "EL_OPEN_END"
	ELEMENT_CLOSE_START = "EL_CLOSE_START"
	ELEMENT_CLOSE_END   = "EL_CLOSE_END"
	ELEMENT_VOID_END    = "EL_VOID_END"
	ELEMENT_IDENT       = "EL_IDENT"
	ELEMENT_ATTR        = "EL_ATTR"
	ELEMENT_TEXT        = "EL_TEXT"

	TYPE_STRING = "T_STRING"
	TYPE_INT    = "T_INT"
	TYPE_BOOL   = "T_BOOL"
	TYPE_VOID   = "T_VOID"
	NIL         = "NIL"
)
