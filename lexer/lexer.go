package lexer

import (
	"gloss/token"
	"unicode"
)

var keywords = map[string]token.TokenType{
	"if":       token.IF,
	"else":     token.ELSE,
	"for":      token.FOR,
	"break":    token.BREAK,
	"continue": token.CONTINUE,
	"switch":   token.SWITCH,
	"case":     token.CASE,
	"default":  token.DEFAULT,
	"return":   token.RETURN,
	"let":      token.LET,
	"fn":       token.FUNC,
	"enum":     token.ENUM,
	"union":    token.UNION,
	"struct":   token.STRUCT,
	"true":     token.BOOL,
	"false":    token.BOOL,
}

var builtins = map[string]token.TokenType{
	"bool":   token.TYPE_BOOL,
	"string": token.TYPE_STRING,
	"int":    token.TYPE_INT,
}

type Lexer struct {
	input []byte

	// Current read positions
	pos  int
	line int
	col  int
	char rune

	braceDepth    int // Tracks nested expressions when inside elements
	tagBraceDepth int
	elementDepth  int  // Tracks nested elements
	insideOpenTag bool // Tracks if we are lexing an element tag which has not yet been terminated by > or />

	tokenBuffer []token.Token
	lastToken   *token.Token
}

func New(input []byte) *Lexer {
	lex := &Lexer{
		input: input,
		char:  rune(input[0]),
	}

	if len(input) > 0 {
		lex.char = rune(input[0])
	} else {
		lex.char = 0
	}

	return lex
}

func (l *Lexer) advance() {
	l.pos++
	l.col++
	if l.pos < len(l.input) {
		l.char = rune(l.input[l.pos])
	} else {
		l.char = 0 // EOF
	}
}

func (l *Lexer) peek() (rune, bool) {
	if l.pos+1 < len(l.input) {
		return rune(l.input[l.pos+1]), true
	}
	return 0, false
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		char := l.input[l.col]
		if unicode.IsSpace(rune(char)) {
			if char == '\n' {
				l.line++
				l.col = 0
			}
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) readDigits() token.Token {
	startCol := l.col
	startPos := l.pos
	startRow := l.line

	for l.pos < len(l.input) {
		if unicode.IsDigit(l.char) || l.char == '_' {
			l.advance()
		} else {
			break
		}
	}

	return token.Token{
		Type:    token.INT,
		Literal: string(l.input[startPos:l.pos]),
		Line:    startRow,
		Column:  startCol,
	}
}

func (l *Lexer) readString() token.Token {
	start := l.pos
	startCol := l.col
	startLine := l.line

	// Eat opening "
	l.advance()

	for {
		if l.char == '"' {
			break
		}

		if l.char == 0 {
			// Error: Reached EOF without closing quote
			break
		}

		// Handle Escape Characters
		if l.char == '\\' {
			l.advance() // Skip the backslash

			if l.char != 0 {
				// Skip the next character (the escaped char), preventing
				// the loop from breaking if that char happens to be '"'
				l.advance()
			}
			continue
		}

		l.advance()
	}

	// Eat closing "
	l.advance()

	return token.Token{
		Type:    token.STRING,
		Literal: string(l.input[start:l.pos]),
		Line:    startLine,
		Column:  startCol,
	}
}

func (l *Lexer) readElementIdentifier() token.Token {
	startCol := l.col
	startPos := l.pos
	startRow := l.line

	for l.pos < len(l.input) {
		if unicode.IsLetter(l.char) || unicode.IsDigit(l.char) || l.char == '.' {
			l.advance()
		} else {
			break
		}
	}

	return token.Token{
		Type:    token.ELEMENT_IDENT,
		Literal: string(l.input[startPos:l.pos]),
		Line:    startRow,
		Column:  startCol,
	}
}

func (l *Lexer) readElementText() token.Token {
	startPos := l.pos
	startCol := l.col
	startLine := l.line

	// Consume until we hit the start of a tag '<' or an expression '{'
	for l.pos < len(l.input) {
		if l.char == '<' || l.char == '{' {
			break
		}

		// Handle newlines within text
		// TODO: Move this to l.advance()
		if l.char == '\n' {
			l.line++
			l.col = 0
			l.advance()
			continue
		}

		l.advance()
	}

	return token.Token{
		Type:    token.ELEMENT_TEXT,
		Literal: string(l.input[startPos:l.pos]),
		Line:    startLine,
		Column:  startCol,
	}
}

func (l *Lexer) readAttributeName() token.Token {
	start := l.pos
	startCol := l.col

	for l.pos < len(l.input) && (unicode.IsLetter(l.char) || unicode.IsDigit(l.char) || l.char == '-') {
		l.advance()
	}

	return token.Token{
		Type:    token.ELEMENT_ATTR,
		Literal: string(l.input[start:l.pos]),
		Line:    l.line,
		Column:  startCol,
	}
}

func (l *Lexer) readIdentifer() token.Token {
	start := l.pos
	startCol := l.col

	for l.pos < len(l.input) && (isLetter(l.char) || isDigit(l.char) || l.char == '_') {
		l.advance()
	}

	return token.Token{
		Type:    token.IDENT,
		Literal: string(l.input[start:l.pos]),
		Line:    l.line,
		Column:  startCol,
	}
}

func (l *Lexer) readTagStart() []token.Token {
	tokens := []token.Token{
		{Type: token.ELEMENT_OPEN_START, Literal: "<", Line: l.line, Column: l.col},
	}

	l.advance() // Eat '<'

	// Read Tag Name
	tokens = append(tokens, l.readElementIdentifier())

	// Update State
	l.elementDepth++
	l.tagBraceDepth = l.braceDepth
	l.insideOpenTag = true

	return tokens
}

func (l *Lexer) tryReadTagEnd() ([]token.Token, bool) {
	tagStartCol := l.col
	tagStartRow := l.line

	if nextChar, ok := l.peek(); !ok || nextChar != '/' {
		return nil, false
	}

	// 1. Consume the "</"
	l.advance()
	l.advance()
	l.skipWhitespace()

	// 2. Read Element Name
	if !unicode.IsLetter(l.char) {
		return nil, false
	}

	tokens := []token.Token{
		{
			Type:    token.ELEMENT_CLOSE_START,
			Literal: "</",
			Line:    tagStartRow,
			Column:  tagStartCol,
		},
		l.readElementIdentifier(),
	}

	l.skipWhitespace()

	if l.char != '>' {
		return nil, false
	}

	tokens = append(tokens,
		token.Token{
			Type:    token.ELEMENT_CLOSE_END,
			Literal: ">",
			Line:    l.line,
			Column:  l.col,
		},
	)

	l.elementDepth-- // Decrement depth
	l.advance()      // Eat '>'

	return tokens, true
}

func (l *Lexer) NextToken() token.Token {
	// Drain buffer
	if len(l.tokenBuffer) > 0 {
		t := l.tokenBuffer[0]
		l.tokenBuffer = l.tokenBuffer[1:]
		l.lastToken = &t
		return t
	}

	for l.pos < len(l.input) {
		startCol := l.col

		// ---------------------------------------------------------
		//  MODE 1: ELEMENT CONTENT
		// ---------------------------------------------------------
		// If we are inside an element, but NOT inside an expression block ({...}),
		// we treat content as raw text.
		if l.elementDepth > 0 && l.braceDepth == 0 && !l.insideOpenTag {
			// If we hit '<', check if it's a valid tag (start or close)
			// If we hit '{', we switch to Code Mode (handled below in standard switch)
			// Otherwise, it is text.
			if l.char != '<' && l.char != '{' {
				t := l.readElementText()
				l.lastToken = &t
				return t
			}
		}

		// ----------------------------------------------------------------
		// MODE 2: ATTRIBUTE SCANNING
		// Inside an opening tag definition <div ... >
		// But NOT inside an attribute expression like prop={...}
		// ----------------------------------------------------------------
		if l.insideOpenTag && l.braceDepth == l.tagBraceDepth {
			// 1. Ignore whitespace
			if unicode.IsSpace(l.char) {
				if l.char == '\n' {
					l.line++
					l.col = 0
				}
				l.advance()
				continue
			}

			// 2.a Open Tag Endings
			if l.char == '>' {
				t := token.Token{Type: token.ELEMENT_OPEN_END, Literal: ">", Line: l.line, Column: startCol}
				l.insideOpenTag = false
				l.advance()
				l.lastToken = &t
				return t
			}

			// 2.b Void Tag Endings
			if l.char == '/' {
				if next, ok := l.peek(); ok && next == '>' {
					t := token.Token{Type: token.ELEMENT_VOID_END, Literal: "/>", Line: l.line, Column: startCol}
					l.insideOpenTag = false
					l.elementDepth--
					l.advance()
					l.advance()
					l.lastToken = &t
					return t
				}
			}

			// 3. Expressions start
			// We emit the brace, increment depth, and let the NEXT loop iteration handle the inside as Standard Code.
			if l.char == '{' {
				t := token.Token{Type: token.LBRACE, Literal: "{", Line: l.line, Column: startCol}
				l.braceDepth++
				l.advance()
				l.lastToken = &t
				return t
			}

			// 4. Assignments
			if l.char == '=' {
				t := token.Token{Type: token.ASSIGN, Literal: "=", Line: l.line, Column: startCol}
				l.advance()
				l.lastToken = &t
				return t
			}

			// 5. String Literal Attribute Values (e.g. class="foo")
			if l.char == '"' {
				t := l.readString()
				l.lastToken = &t
				return t
			}

			// 6. Attribute Names (Ident)
			if unicode.IsLetter(l.char) {
				t := l.readAttributeName()
				l.lastToken = &t
				return t
			}
		}

		// ---------------------------------------------------------
		// 3. STANDARD CODE MODE
		// ---------------------------------------------------------

		// Skip whitespace
		if unicode.IsSpace(l.char) {
			if l.char == '\n' {
				l.line++
				l.col = 0
			}
			l.advance()
			continue
		}

		// Identifiers, Keywords, and Builtins
		if isLetter(l.char) || l.char == '_' {
			identToken := l.readIdentifer()

			// Text is a keyword?
			if tokenType, ok := keywords[identToken.Literal]; ok {
				identToken.Type = tokenType
				l.lastToken = &identToken
				return identToken
			}

			// Text is a builtin?
			if tokenType, ok := builtins[identToken.Literal]; ok {
				identToken.Type = tokenType
				l.lastToken = &identToken
				return identToken
			}

			l.lastToken = &identToken
			return identToken
		}

		// Numbers
		if isDigit(l.char) {
			t := l.readDigits()
			l.lastToken = &t
			return t
		}

		// Strings
		if l.char == '"' {
			t := l.readString()
			l.lastToken = &t
			return t
		}

		// Check for Tag Start

		// These patterns should not be treated as elements
		// union Option<T> {}
		// struct Point<T> { x: T, y: T }
		// fn join<T>(a: T, b: T) T {}

		if l.char == '<' && !maybeExpectingGenericParameters(l.lastToken) {

			// Check if it is a closing tag </...
			next, hasNext := l.peek()

			if hasNext && next == '/' {
				if toks, ok := l.tryReadTagEnd(); ok {
					t := toks[0]
					l.tokenBuffer = append(l.tokenBuffer, toks[1:]...)
					l.lastToken = &t
					return t
				}
			}
			// Check if it is an opening tag <div...
			// Must ensure it's not a Less Than operator (e.g. "if a < b")
			if hasNext && unicode.IsLetter(next) {
				toks := l.readTagStart()
				t := toks[0]
				l.tokenBuffer = append(l.tokenBuffer, toks[1:]...)
				l.lastToken = &t
				return t
			}
		}

		// Symbols
		var tt token.TokenType
		tl := string(l.char)

		switch l.char {
		case ':':
			tt = token.COLON
		case '\'':
			tt = token.TICK
		case '`':
			tt = token.BACKTICK
		case '.':
			tt = token.PERIOD
		case ',':
			tt = token.COMMA
		case '+':
			tt = token.PLUS
		case '-':
			tt = token.MINUS
		case '*':
			tt = token.MUL
		case '/':
			tt = token.DIV
		case '%':
			tt = token.MOD
		case '[':
			tt = token.LBRACKET
		case ']':
			tt = token.RBRACKET
		case '(':
			tt = token.LPAREN
		case ')':
			tt = token.RPAREN
		case '{':
			tt = token.LBRACE
			l.braceDepth++
		case '}':
			tt = token.RBRACE
			if l.braceDepth > 0 {
				l.braceDepth--
			}

		// Possible operators
		case '^':
			tt = token.BITWISE_XOR

		case '~':
			tt = token.BITWISE_NOT

		case '|':
			if next, ok := l.peek(); ok && next == '|' {
				tt = token.OR
				tl = "||"
				l.advance()
			} else {
				tt = token.BITWISE_OR
			}

		case '&':
			if next, ok := l.peek(); ok && next == '&' {
				tt = token.AND
				tl = "&&"
				l.advance()
			} else {
				tt = token.BITWISE_AND
			}

		case '=':
			if next, ok := l.peek(); ok && next == '=' {
				tt = token.EQ
				tl = "=="
				l.advance()
			} else {
				tt = token.ASSIGN
			}

		case '<':
			if next, ok := l.peek(); ok && next == '=' {
				tt = token.LT_EQ
				tl = "<="
				l.advance()
			} else if next, ok := l.peek(); ok && next == '<' {
				tt = token.BITSHIFTL
				tl = "<<"
				l.advance()
			} else {
				tt = token.LANGLE
			}

		case '>':
			if next, ok := l.peek(); ok && next == '=' {
				tt = token.GT_EQ
				tl = ">="
				l.advance()
			} else if next, ok := l.peek(); ok && next == '>' {
				tt = token.BITSHIFTR
				tl = ">>"
				l.advance()
			} else {
				tt = token.RANGLE
			}

		case '!':
			if next, ok := l.peek(); ok && next == '=' {
				tt = token.NOT_EQ
				tl = "!="
				l.advance()
			} else {
				tt = token.BANG
			}
		default:
			tt = token.ILLEGAL
		}

		t := token.Token{Type: tt, Literal: tl, Line: l.line, Column: startCol}
		l.advance()
		l.lastToken = &t
		return t
	}

	return token.Token{Type: token.EOF, Line: l.line, Column: l.col}
}

// helpers

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch rune) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func maybeExpectingGenericParameters(tok *token.Token) bool {
	if tok != nil {
		switch tok.Type {
		case token.IDENT:
			return true
		}
	}
	return false
}
