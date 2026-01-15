package lexer

import (
	"gloss/token"
	"unicode"
)

type Lexer struct {
	input []byte

	// Current read positions
	pos  int
	line int
	col  int
	char rune

	braceDepth    int  // Tracks nested expressions when inside elements
	elementDepth  int  // Tracks nested elements
	insideOpenTag bool // Tracks if we are lexing an element tag which has not yet been terminated by > or />
}

func New(input []byte) *Lexer {
	lex := &Lexer{
		input: input,
		char:  rune(input[0]),
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

	for l.pos < len(l.input) && (unicode.IsLetter(l.char) || unicode.IsDigit(l.char)) {
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

func (l *Lexer) Tokenize() []token.Token {
	var tokens []token.Token
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
				tokens = append(tokens, l.readElementText())
				continue
			}
		}

		// ----------------------------------------------------------------
		// MODE 2: ATTRIBUTE SCANNING
		// Inside an opening tag definition <div ... >
		// But NOT inside an attribute expression like prop={...}
		// ----------------------------------------------------------------
		if l.insideOpenTag && l.braceDepth == 0 {
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
				tokens = append(tokens, token.Token{Type: token.ELEMENT_OPEN_END, Literal: ">", Line: l.line, Column: startCol})
				l.insideOpenTag = false
				l.advance()
				continue
			}

			// 2.b Void Tag Endings
			if l.char == '/' {
				if next, ok := l.peek(); ok && next == '>' {
					tokens = append(tokens, token.Token{Type: token.ELEMENT_VOID_END, Literal: "/>", Line: l.line, Column: startCol})
					l.insideOpenTag = false
					l.elementDepth--
					l.advance()
					l.advance()
					continue
				}
			}

			// 3. Expressions start
			// We emit the brace, increment depth, and let the NEXT loop iteration handle the inside as Standard Code.
			if l.char == '{' {
				tokens = append(tokens, token.Token{Type: token.LBRACE, Literal: "{", Line: l.line, Column: startCol})
				l.braceDepth++
				l.advance()
				continue
			}

			// 4. Assignments
			if l.char == '=' {
				tokens = append(tokens, token.Token{Type: token.ASSIGN, Literal: "=", Line: l.line, Column: startCol})
				l.advance()
				continue
			}

			// 5. String Literal Attribute Values (e.g. class="foo")
			if l.char == '"' {
				tokens = append(tokens, l.readString())
				continue
			}

			// 6. Attribute Names (Ident)
			if unicode.IsLetter(l.char) {
				tokens = append(tokens, l.readAttributeName())
				continue
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

		// Identifiers and Keywords
		if unicode.IsLetter(l.char) {
			identToken := l.readIdentifer()

			switch identToken.Literal {
			case "use":
				identToken.Type = token.IMPORT
			case "enum":
				identToken.Type = token.ENUM
			case "struct":
				identToken.Type = token.STRUCT
			case "interface":
				identToken.Type = token.IFACE
			case "extern":
				identToken.Type = token.EXTERN
			case "end":
				identToken.Type = token.END
			case "if":
				identToken.Type = token.IF
			case "else":
				identToken.Type = token.ELSE
			case "switch":
				identToken.Type = token.SWITCH
			case "case":
				identToken.Type = token.CASE
			case "default":
				identToken.Type = token.DEFAULT
			case "for":
				identToken.Type = token.FOR
			case "continue":
				identToken.Type = token.CONTINUE
			case "break":
				identToken.Type = token.BREAK
			case "return":
				identToken.Type = token.RETURN
			case "let":
				identToken.Type = token.LET
			case "fn":
				identToken.Type = token.FUNC
			}

			tokens = append(tokens, identToken)
			continue
		}

		// Numbers
		if unicode.IsDigit(l.char) {
			tokens = append(tokens, l.readDigits())
			continue
		}

		// Strings
		if l.char == '"' {
			tokens = append(tokens, l.readString())
			continue
		}

		// Check for Tag Start
		if l.char == '<' {
			// Check if it is a closing tag </...
			next, hasNext := l.peek()
			if hasNext && next == '/' {
				if toks, ok := l.tryReadTagEnd(); ok {
					tokens = append(tokens, toks...)
					continue
				}
			}
			// Check if it is an opening tag <div...
			// Must ensure it's not a Less Than operator (e.g. "if a < b")
			if hasNext && unicode.IsLetter(next) {
				tokens = append(tokens, l.readTagStart()...)
				continue
			}
		}

		// Symbols
		var tt token.TokenType
		switch l.char {
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
		case '=':
			tt = token.ASSIGN
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
		case '<':
			tt = token.LANGLE
		case '>':
			tt = token.RANGLE
		default:
			tt = token.ILLEGAL
		}

		tokens = append(tokens, token.Token{Type: tt, Literal: string(l.char), Line: l.line, Column: startCol})
		l.advance()
	}

	tokens = append(tokens, token.Token{Type: token.EOF, Line: l.line, Column: l.col})

	return tokens
}
