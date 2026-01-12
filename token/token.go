package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT = "IDENT"
	INT   = "INT"

	// Operators
	ASSIGN = "="
	PLUS   = "+"
	MINUS  = "-"
	MUL    = "*"
	DIV    = "/"
	MOD    = "%"
	AND    = "&"

	// Delimiters
	PERIOD    = "."
	CARET     = "^"
	BANG      = "!"
	QUOTE     = "\""
	TICK      = "'"
	BACKTICK  = "`"
	COMMA     = ","
	COLON     = ":"
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"
	LANGLE    = "<"
	RANGLE    = ">"

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
)

const (
	TypeString = "string"
	TypeInt    = "int"
	TypeFloat  = "float"
	TypeBool   = "bool"
)
