package lexer

import (
	"gloss/token"
	"unicode"
)

type Lexer struct {
	input []byte
	pos   int
	line  int
	col   int
	char  rune

	// TODO: track open element tokens count to aid with inner expressions
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

func (l *Lexer) consumeDigits() {
	for l.pos < len(l.input) {
		ch := rune(l.input[l.pos])
		if unicode.IsDigit(ch) || ch == '_' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) consumeStringLiteral() {
	// record position of the opening quote
	// startPos := l.pos

	// Loop until we hit the closing quote or EOF
	for {
		l.advance()

		// 1. Handle EOF (Unterminated string)
		if l.char == 0 {
			break
			// return token.Token{
			// 	Type:    token.ILLEGAL,
			// 	Literal: string(l.input[]),
			// 	Line:    startLine,
			// 	Column:  startCol,
			// }
		}

		// 2. Handle Escape Sequences
		if l.char == '\\' {
			l.advance()

			// If we hit EOF immediately after a backslash (e.g. "abc \ )
			if l.char == 0 {
				break
				// return token.Token{
				// 	Type:    token.ILLEGAL,
				// 	Literal: "Unterminated string literal (trailing escape)",
				// 	Line:    startLine,
				// 	Column:  startCol,
				// }
			}

			continue
		}

		if l.char == '"' {
			break
		}
	}

	// Capture the full text including the opening and closing quotes
	// l.pos is currently at the closing quote, so we need l.pos+1 to include it
	// text := string(l.input[startPos : l.pos+1])

	// Advance past the closing quote so the main loop doesn't re-process it
	l.advance()
}

func (l *Lexer) consumeElementIdentifier() {
	for l.pos < len(l.input) {
		ch := rune(l.input[l.pos])
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '.' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) consumeWhitespace() {
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

func (l *Lexer) tryConsumeOpenTag() ([]token.Token, bool) {
	tokens := []token.Token{
		{Type: token.ELEMENT_OPEN_START, Literal: string(l.char), Line: l.line, Column: l.col},
	}

	l.advance()
	if !unicode.IsLetter(l.char) {
		return nil, false
	}

	startCol := l.col
	startPos := l.pos
	startRow := l.line
	l.consumeElementIdentifier()
	tokens = append(tokens, token.Token{Type: token.ELEMENT_IDENT, Literal: string(l.input[startPos:l.pos]), Line: startRow, Column: startCol})

	// Scan for attributes
	for l.pos < len(l.input) {

		if unicode.IsSpace(l.char) {
			if l.char == '\n' {
				l.line++
				l.col = 0
			}
			l.advance()
			continue
		}

		// Check for Tag Endings
		if l.char == '>' {
			tokens = append(tokens, token.Token{Type: token.ELEMENT_OPEN_END, Literal: ">", Line: l.line, Column: l.col})
			l.advance()
			return tokens, true
		}

		// Check for Void Tag End "/>"
		if l.char == '/' {
			if next, ok := l.peek(); ok && next == '>' {
				tokens = append(tokens, token.Token{Type: token.ELEMENT_VOID_END, Literal: "/>", Line: l.line, Column: l.col})
				l.advance() // eat /
				l.advance() // eat >
				return tokens, true
			}
		}

		// Assume Attribute Identifier
		if unicode.IsLetter(l.char) {
			attrStart := l.pos

			// eat attribute identifier
			for {
				if unicode.IsLetter(l.char) {
					l.advance()
				} else {
					break
				}
			}

			attrText := string(l.input[attrStart:l.pos])
			tokens = append(tokens, token.Token{Type: token.ELEMENT_ATTR, Literal: attrText, Line: l.line, Column: l.col})

			// Check for Assignment
			l.consumeWhitespace()
			if l.char == '=' {
				tokens = append(tokens, token.Token{Type: token.ASSIGN, Literal: "=", Line: l.line, Column: l.col})
				l.advance() // eat =

				// TODO: Handle Attribute Value (String "..." or Expression {...})
				// For now, let's just consume a string if present
				l.consumeWhitespace()
				switch l.char {
				case '"':
					strStartCol := l.col
					strStartPos := l.pos
					strStartRow := l.line
					l.consumeStringLiteral()
					tokens = append(tokens, token.Token{Type: token.STRING, Literal: string(l.input[strStartPos:l.pos]), Line: strStartRow, Column: strStartCol})
				case '{':
					// TODO: capture expression
				}
			}
			continue
		}

		// Only break if we hit something unexpected to prevent infinite loops
		break
	}

	return nil, false
}

func (l *Lexer) tryConsumeClosingTag() ([]token.Token, bool) {
	tagStartPos := l.pos
	tagStartCol := l.col
	tagStartRow := l.line

	if nextChar, ok := l.peek(); !ok || nextChar != '/' {
		return nil, false
	}

	l.advance()
	l.advance()

	tokens := []token.Token{
		{Type: token.ELEMENT_CLOSE_START, Literal: string(l.input[tagStartPos:l.pos]), Line: tagStartRow, Column: tagStartCol},
	}

	if nextChar, ok := l.peek(); !ok || !unicode.IsLetter(nextChar) {
		return nil, false
	}

	identStartPos := l.pos
	identStartCol := l.col
	identStartRow := l.line
	l.consumeElementIdentifier()
	identText := string(l.input[identStartPos:l.pos])

	tokens = append(tokens,
		token.Token{Type: token.ELEMENT_IDENT, Literal: identText, Line: identStartRow, Column: identStartCol},
	)

	if l.char != '>' {
		return nil, false
	}

	tokens = append(tokens,
		token.Token{Type: token.ELEMENT_CLOSE_END, Literal: string(l.char), Line: l.line, Column: l.col},
	)

	// Don't process < again
	l.advance()

	return tokens, true

}

func (l *Lexer) tryConsumeElement() ([]token.Token, bool) {
	nextChar, ok := l.peek()
	if !ok {
		return nil, false
	}

	if nextChar == '/' {
		return l.tryConsumeClosingTag()
	}

	if unicode.IsLetter(nextChar) {
		return l.tryConsumeOpenTag()
	}

	return nil, false
}

func (l *Lexer) Tokenize() []token.Token {
	var tokens []token.Token
	for l.pos < len(l.input) {
		char := l.input[l.pos]
		startCol := l.col

		// Skip whitespace
		if unicode.IsSpace(rune(char)) {
			if char == '\n' {
				l.line++
				l.col = 0
			}
			l.advance()
			continue
		}

		// Identifiers and Keywords
		if unicode.IsLetter(rune(char)) {
			start := l.pos
			for l.pos < len(l.input) && (unicode.IsLetter(rune(l.input[l.pos])) || unicode.IsDigit(rune(l.input[l.pos]))) {
				l.advance()
			}

			text := string(l.input[start:l.pos])
			switch text {
			case "use":
				tokens = append(tokens, token.Token{Type: token.IMPORT, Literal: text, Line: l.line, Column: startCol})
			case "enum":
				tokens = append(tokens, token.Token{Type: token.ENUM, Literal: text, Line: l.line, Column: startCol})
			case "struct":
				tokens = append(tokens, token.Token{Type: token.STRUCT, Literal: text, Line: l.line, Column: startCol})
			case "interface":
				tokens = append(tokens, token.Token{Type: token.IFACE, Literal: text, Line: l.line, Column: startCol})
			case "extern":
				tokens = append(tokens, token.Token{Type: token.EXTERN, Literal: text, Line: l.line, Column: startCol})
			case "end":
				tokens = append(tokens, token.Token{Type: token.END, Literal: text, Line: l.line, Column: startCol})
			case "if":
				tokens = append(tokens, token.Token{Type: token.IF, Literal: text, Line: l.line, Column: startCol})
			case "else":
				tokens = append(tokens, token.Token{Type: token.ELSE, Literal: text, Line: l.line, Column: startCol})
			case "switch":
				tokens = append(tokens, token.Token{Type: token.SWITCH, Literal: text, Line: l.line, Column: startCol})
			case "case":
				tokens = append(tokens, token.Token{Type: token.CASE, Literal: text, Line: l.line, Column: startCol})
			case "default":
				tokens = append(tokens, token.Token{Type: token.DEFAULT, Literal: text, Line: l.line, Column: startCol})
			case "for":
				tokens = append(tokens, token.Token{Type: token.FOR, Literal: text, Line: l.line, Column: startCol})
			case "continue":
				tokens = append(tokens, token.Token{Type: token.CONTINUE, Literal: text, Line: l.line, Column: startCol})
			case "break":
				tokens = append(tokens, token.Token{Type: token.BREAK, Literal: text, Line: l.line, Column: startCol})
			case "return":
				tokens = append(tokens, token.Token{Type: token.RETURN, Literal: text, Line: l.line, Column: startCol})
			case "let":
				tokens = append(tokens, token.Token{Type: token.LET, Literal: text, Line: l.line, Column: startCol})
			case "fn":
				tokens = append(tokens, token.Token{Type: token.FUNC, Literal: text, Line: l.line, Column: startCol})
			default:
				tokens = append(tokens, token.Token{Type: token.IDENT, Literal: text, Line: l.line, Column: startCol})
			}

			continue
		}

		// Numbers
		if unicode.IsDigit(rune(char)) {
			start := l.pos
			l.consumeDigits()
			tokens = append(tokens, token.Token{Type: token.INT, Literal: string(l.input[start:l.pos]), Line: l.line, Column: startCol})
			continue
		}

		// Symbols
		var tt token.TokenType
		switch l.char {
		case '"':
			// l.consumeStringLiteral()
			tt = token.QUOTE
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
		case '}':
			tt = token.RBRACE
		case '<':
			if toks, ok := l.tryConsumeElement(); ok {
				for _, tok := range toks {
					tokens = append(tokens, tok)
				}
				continue
			} else {
				tt = token.LANGLE
			}
		case '>':
			tt = token.RANGLE
		default:
			tt = token.ILLEGAL
		}

		tokens = append(tokens, token.Token{Type: tt, Literal: string(char), Line: l.line, Column: startCol})
		l.advance()
	}

	tokens = append(tokens, token.Token{Type: token.EOF, Line: l.line, Column: l.col})

	return tokens
}
