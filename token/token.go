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
	IDENT  = "IDENT"
	INT    = "INT"
	STRING = "STR"
	BOOL   = "BOOL"

	// Operators
	ASSIGN = "ASSIGN"
	PLUS   = "PLUS"
	MINUS  = "MINUS"
	MUL    = "MUL"
	DIV    = "DIV"
	MOD    = "MOD"
	AND    = "AND"
	EQ     = "EQ"
	NOT_EQ = "NOT_EQ"
	LT     = "LESS_THAN"
	GT     = "GREATER_THAN"

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
	FUNC                = "FUNC"
	IMPORT              = "IMPORT"
	END                 = "END"
	ENUM                = "ENUM"
	STRUCT              = "STRUCT"
	IFACE               = "IFACE"
	EXTERN              = "EXTERN"
	IF                  = "IF"
	ELSE                = "ELSE"
	SWITCH              = "SWITCH"
	CASE                = "CASE"
	DEFAULT             = "DEFAULT"
	FOR                 = "FOR"
	CONTINUE            = "CONTINUE"
	BREAK               = "BREAK"
	RETURN              = "RETURN"
	ELEMENT_OPEN_START  = "EL_OPEN_START"
	ELEMENT_OPEN_END    = "EL_OPEN_END"
	ELEMENT_CLOSE_START = "EL_CLOSE_START"
	ELEMENT_CLOSE_END   = "EL_CLOSE_END"
	ELEMENT_VOID_END    = "EL_VOID_END"
	ELEMENT_IDENT       = "EL_IDENT"
	ELEMENT_ATTR        = "EL_ATTR"
	ELEMENT_TEXT        = "EL_TEXT"
)

const (
	TypeString = "string"
	TypeInt    = "int"
	TypeFloat  = "float"
	TypeBool   = "bool"
)
