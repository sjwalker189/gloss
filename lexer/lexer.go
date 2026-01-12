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
	lex := &Lexer{input: input}
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

		// TODO:
		// Need to track if we are in a state that is expecting an expression or regular text.
		// When in expression mode, capture string literal tokens like: Token{Type: token.STRING, text: "\"example\""}
		// Otherwise capture text literal tokens like: Token{Type: token.ElementText, text: "example"}

		// TODO:
		// Need to be aware of expressions and assignments, for example:
		// let foo = bar <= 1; // should not enter element mode
		// let foo = bar < baz; // should not enter element mode
		// let foo = <h1>...</h1> // should enter element mode
		// let foo = bar <= 1 ? <A /> :  <B />

		// Strings (simplified, assumes double quotes)
		if char == '"' {
			// l.advance() // skip opening quote
			// start := l.pos
			// for l.pos < len(l.input) && l.input[l.pos] != '"' {
			// 	l.advance()
			// }
			// text := string(l.input[start:l.pos])
			// if l.pos < len(l.input) {
			// 	l.advance() // skip closing quote
			// }
			// tokens = append(tokens, token.Token{Type: token.QUOTE, Literal: text, Line: l.line, Column: startCol})
		}

		// Symbols
		var tt token.TokenType
		switch char {
		case '"':
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
			if ch, ok := l.peek(); ok {
				// TODO: Check element or expression mode
				// check for explicit matmatical operators " < ", " <= ", " << "
				if ch == '=' || ch == '<' || unicode.IsSpace(ch) {
					tokens = append(tokens, token.Token{Type: token.LANGLE, Literal: string(char), Line: l.line, Column: startCol})
					continue
				}

				// We're likely to be closing a tag
				if ch == '/' {
					startPos := l.pos
					startCol := l.col
					startLine := l.line
					l.advance()
					endPos := l.pos

					l.advance()
					if unicode.IsLetter(l.char) {
						elIdentStartPos := l.pos
						elIdentStartCol := l.col
						elIdentStartLine := l.line
						l.consumeElementIdentifier()
						elIdentEndPos := l.pos

						if l.char == '>' {
							tokens = append(tokens,
								token.Token{Type: token.ELEMENT_CLOSE_START, Literal: string(l.input[startPos:endPos]), Line: startLine, Column: startCol},
								token.Token{Type: token.ELEMENT_IDENT, Literal: string(l.input[elIdentStartPos:elIdentEndPos]), Line: elIdentStartLine, Column: elIdentStartCol},
								token.Token{Type: token.ELEMENT_CLOSE_END, Literal: string(l.char), Line: l.line, Column: l.col},
							)
							l.advance()
							continue
						}
					}
				}
			}

			// TODO: Lift below blocks above and set a stateful field on lexer to switch on how to tokeniz characters such as < / >

			// We're either consuming an identifier or an element. Elements will be terminated by ">" or "/>"

			// try capture the following patterns which idicates an element
			// <div>
			// <ui.div>
			// <div attr>
			// <div attr="string_literal">
			// <div attr={expression}>
			// <div />
			// <div/>
			//
			// if none of the above cases match we are likely parsing a comparator
			if ch, ok := l.peek(); ok && unicode.IsLetter(ch) {
				elOpenStartPos := l.pos
				elOpenStartCol := l.col
				elOpenStartLine := l.line

				l.advance()

				elIdentStartPos := l.pos
				elIdentStartCol := l.col
				elIdentStartLine := l.line
				l.consumeElementIdentifier()
				elIdentEndPos := l.pos

				// TODO: handle attributes
				l.consumeWhitespace()

				ch := rune(l.input[l.pos])

				// Consume open tag terminator and capture element open
				// TODO: switch to element mode lexer
				if ch == '>' {
					tokens = append(tokens,
						token.Token{Type: token.ELEMENT_OPEN_START, Literal: string(l.input[elOpenStartPos]), Line: elOpenStartLine, Column: elOpenStartCol},
						token.Token{Type: token.ELEMENT_IDENT, Literal: string(l.input[elIdentStartPos:elIdentEndPos]), Line: elIdentStartLine, Column: elIdentStartCol},
						token.Token{Type: token.ELEMENT_OPEN_END, Literal: string(ch), Line: l.line, Column: l.col},
					)
					l.advance()
					continue
				}

				ch = rune(l.input[l.pos])

				// Consume void tag terminator
				if ch == '/' {
					startPos := l.pos
					startCol := l.col
					if ch, ok := l.peek(); ok && ch == '>' {
						l.advance()
						l.advance()
						// We found a valid tag ending
						tokens = append(tokens,
							token.Token{Type: token.ELEMENT_OPEN_START, Literal: string(l.input[elOpenStartCol]), Line: elOpenStartLine, Column: elOpenStartCol},
							token.Token{Type: token.ELEMENT_IDENT, Literal: string(l.input[elIdentStartPos:elIdentEndPos]), Line: elIdentStartLine, Column: elIdentStartCol},
							token.Token{Type: token.ELEMENT_VOID_END, Literal: string(l.input[startCol:l.col]), Line: l.line, Column: startCol},
						)
					} else {
						// We found an unexpected token after /, we must be lexing a division expression against an identifier
						// TODO: if: there are any attributes captured, this token is illegal
						//  	 else: there are no attributes so division must be legal
						tokens = append(tokens,
							token.Token{Type: token.IDENT, Literal: string(l.input[elIdentStartPos:elIdentEndPos]), Line: elIdentStartLine, Column: elIdentStartCol},
							token.Token{Type: token.DIV, Literal: string(l.input[startPos:l.pos]), Line: l.line, Column: l.col},
						)
						l.advance()
					}
					continue
				} else {
					panic("TODO: handle void tag close")
				}
			} else {
				tt = token.LANGLE
			}
		case '>':
			tt = token.RANGLE
		default:
			tt = token.ILLEGAL
		}

		tokens = append(tokens, token.Token{Type: tt, Literal: string(char), Line: l.line, Column: startCol})
		l.pos++
		l.col++
	}

	tokens = append(tokens, token.Token{Type: token.EOF, Line: l.line, Column: l.col})

	return tokens
}
