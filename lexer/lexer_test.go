package lexer

import (
	"gloss/token"

	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNextToken(t *testing.T) {
	cmpOpts := []cmp.Option{
		cmpopts.IgnoreFields(token.Token{}, "Line", "Column"),
	}

	tests := []struct {
		name  string
		input string
		want  []token.Token
	}{
		{
			name:  "Math tokens",
			input: "+=-*/%.",
			want: []token.Token{
				{Type: token.PLUS, Literal: "+"},
				{Type: token.ASSIGN, Literal: "="},
				{Type: token.MINUS, Literal: "-"},
				{Type: token.MUL, Literal: "*"},
				{Type: token.DIV, Literal: "/"},
				{Type: token.MOD, Literal: "%"},
				{Type: token.PERIOD, Literal: "."},
				{Type: token.EOF},
			},
		},
		{
			name:  "Bracket tokens",
			input: "(){}<>[]",
			want: []token.Token{
				{Type: token.LPAREN, Literal: "("},
				{Type: token.RPAREN, Literal: ")"},
				{Type: token.LBRACE, Literal: "{"},
				{Type: token.RBRACE, Literal: "}"},
				{Type: token.LANGLE, Literal: "<"},
				{Type: token.RANGLE, Literal: ">"},
				{Type: token.LBRACKET, Literal: "["},
				{Type: token.RBRACKET, Literal: "]"},
				{Type: token.EOF},
			},
		},
		{
			name:  "Keyword tokens",
			input: "use enum struct interface extern if else switch case default continue for fn return let foo #",
			want: []token.Token{
				{Type: token.IMPORT, Literal: "use"},
				{Type: token.ENUM, Literal: "enum"},
				{Type: token.STRUCT, Literal: "struct"},
				{Type: token.IFACE, Literal: "interface"},
				{Type: token.EXTERN, Literal: "extern"},
				{Type: token.IF, Literal: "if"},
				{Type: token.ELSE, Literal: "else"},
				{Type: token.SWITCH, Literal: "switch"},
				{Type: token.CASE, Literal: "case"},
				{Type: token.DEFAULT, Literal: "default"},
				{Type: token.CONTINUE, Literal: "continue"},
				{Type: token.FOR, Literal: "for"},
				{Type: token.FUNC, Literal: "fn"},
				{Type: token.RETURN, Literal: "return"},
				{Type: token.LET, Literal: "let"},
				{Type: token.IDENT, Literal: "foo"},
				{Type: token.ILLEGAL, Literal: "#"},
				{Type: token.EOF},
			},
		},

		{
			name: "Literal type tokens",
			input: `
			0
			100
			1_000
			1_000_000
			1_
			1__0
			"hello world"
			"with \"quoted\""
			`,
			want: []token.Token{
				// Integers
				{Type: token.INT, Literal: "0"},
				{Type: token.INT, Literal: "100"},
				{Type: token.INT, Literal: "1_000"},
				{Type: token.INT, Literal: "1_000_000"},
				{Type: token.INT, Literal: "1_"},   // syntax error
				{Type: token.INT, Literal: "1__0"}, // syntax error
				// Strings
				{Type: token.STRING, Literal: `"hello world"`},
				{Type: token.STRING, Literal: `"with \"quoted\""`},
				{Type: token.EOF},
			},
		},

		{
			name:  "Example enum declaration",
			input: "enum Boolean {\n\tcase on,\n\tcase off,\n}",
			want: []token.Token{
				{Type: token.ENUM, Literal: "enum"},
				{Type: token.IDENT, Literal: "Boolean"},
				{Type: token.LBRACE, Literal: "{"},
				{Type: token.CASE, Literal: "case"},
				{Type: token.IDENT, Literal: "on"},
				{Type: token.COMMA, Literal: ","},
				{Type: token.CASE, Literal: "case"},
				{Type: token.IDENT, Literal: "off"},
				{Type: token.COMMA, Literal: ","},
				{Type: token.RBRACE, Literal: "}"},
				{Type: token.EOF},
			},
		},
		{
			name:  "Basic Elements",
			input: "<div><hr/><custom /></div>",
			want: []token.Token{
				{Type: token.ELEMENT_OPEN_START, Literal: "<"},
				{Type: token.ELEMENT_IDENT, Literal: "div"},
				{Type: token.ELEMENT_OPEN_END, Literal: ">"},
				{Type: token.ELEMENT_OPEN_START, Literal: "<"},
				{Type: token.ELEMENT_IDENT, Literal: "hr"},
				{Type: token.ELEMENT_VOID_END, Literal: "/>"},
				{Type: token.ELEMENT_OPEN_START, Literal: "<"},
				{Type: token.ELEMENT_IDENT, Literal: "custom"},
				{Type: token.ELEMENT_VOID_END, Literal: "/>"},
				{Type: token.ELEMENT_CLOSE_START, Literal: "</"},
				{Type: token.ELEMENT_IDENT, Literal: "div"},
				{Type: token.ELEMENT_CLOSE_END, Literal: ">"},
				{Type: token.EOF},
			},
		},
		{
			name:  "Elements with attributes",
			input: `<input disabled />`,
			want: []token.Token{
				{Type: token.ELEMENT_OPEN_START, Literal: "<"},
				{Type: token.ELEMENT_IDENT, Literal: "input"},
				{Type: token.ELEMENT_ATTR, Literal: "disabled"},
				{Type: token.ELEMENT_VOID_END, Literal: "/>"},
				{Type: token.EOF},
			},
		},
		{
			name:  "Elements with attributes",
			input: `<button type="submit"></button><button type="reset" disabled></button>`,
			want: []token.Token{
				{Type: token.ELEMENT_OPEN_START, Literal: "<"},
				{Type: token.ELEMENT_IDENT, Literal: "button"},
				{Type: token.ELEMENT_ATTR, Literal: "type"},
				{Type: token.ASSIGN, Literal: "="},
				{Type: token.STRING, Literal: "\"submit\""},
				{Type: token.ELEMENT_OPEN_END, Literal: ">"},
				{Type: token.ELEMENT_CLOSE_START, Literal: "</"},
				{Type: token.ELEMENT_IDENT, Literal: "button"},
				{Type: token.ELEMENT_CLOSE_END, Literal: ">"},

				{Type: token.ELEMENT_OPEN_START, Literal: "<"},
				{Type: token.ELEMENT_IDENT, Literal: "button"},
				{Type: token.ELEMENT_ATTR, Literal: "type"},
				{Type: token.ASSIGN, Literal: "="},
				{Type: token.STRING, Literal: "\"reset\""},
				{Type: token.ELEMENT_ATTR, Literal: "disabled"},
				{Type: token.ELEMENT_OPEN_END, Literal: ">"},
				{Type: token.ELEMENT_CLOSE_START, Literal: "</"},
				{Type: token.ELEMENT_IDENT, Literal: "button"},
				{Type: token.ELEMENT_CLOSE_END, Literal: ">"},

				{Type: token.EOF},
			},
		},
		{
			name:  "Elements with inner content",
			input: `<h1>Hello, <b>World</b>!</h1>`,
			want: []token.Token{
				{Type: token.ELEMENT_OPEN_START, Literal: "<"},
				{Type: token.ELEMENT_IDENT, Literal: "h1"},
				{Type: token.ELEMENT_OPEN_END, Literal: ">"},
				{Type: token.ELEMENT_TEXT, Literal: "Hello, "},
				{Type: token.ELEMENT_OPEN_START, Literal: "<"},
				{Type: token.ELEMENT_IDENT, Literal: "b"},
				{Type: token.ELEMENT_OPEN_END, Literal: ">"},
				{Type: token.ELEMENT_TEXT, Literal: "World"},
				{Type: token.ELEMENT_CLOSE_START, Literal: "</"},
				{Type: token.ELEMENT_IDENT, Literal: "b"},
				{Type: token.ELEMENT_CLOSE_END, Literal: ">"},
				{Type: token.ELEMENT_TEXT, Literal: "!"},
				{Type: token.ELEMENT_CLOSE_START, Literal: "</"},
				{Type: token.ELEMENT_IDENT, Literal: "h1"},
				{Type: token.ELEMENT_CLOSE_END, Literal: ">"},
				{Type: token.EOF},
			},
		},
		{
			name:  "Elements with inner expressions",
			input: `<h1>{message}</h1><button type="button" disabled={true}>Hello, {name}!</button>`,
			want: []token.Token{
				{Type: token.ELEMENT_OPEN_START, Literal: "<"},
				{Type: token.ELEMENT_IDENT, Literal: "h1"},
				{Type: token.ELEMENT_OPEN_END, Literal: ">"},
				{Type: token.LBRACE, Literal: "{"},
				{Type: token.IDENT, Literal: "message"},
				{Type: token.RBRACE, Literal: "}"},
				{Type: token.ELEMENT_CLOSE_START, Literal: "</"},
				{Type: token.ELEMENT_IDENT, Literal: "h1"},
				{Type: token.ELEMENT_CLOSE_END, Literal: ">"},

				{Type: token.ELEMENT_OPEN_START, Literal: "<"},
				{Type: token.ELEMENT_IDENT, Literal: "button"},
				{Type: token.ELEMENT_ATTR, Literal: "type"},
				{Type: token.ASSIGN, Literal: "="},
				{Type: token.STRING, Literal: `"button"`},
				{Type: token.ELEMENT_ATTR, Literal: "disabled"},
				{Type: token.ASSIGN, Literal: "="},
				{Type: token.LBRACE, Literal: "{"},
				{Type: token.IDENT, Literal: "true"},
				{Type: token.RBRACE, Literal: "}"},
				{Type: token.ELEMENT_OPEN_END, Literal: ">"},

				{Type: token.ELEMENT_TEXT, Literal: "Hello, "},
				{Type: token.LBRACE, Literal: "{"},
				{Type: token.IDENT, Literal: "name"},
				{Type: token.RBRACE, Literal: "}"},
				{Type: token.ELEMENT_TEXT, Literal: "!"},

				{Type: token.ELEMENT_CLOSE_START, Literal: "</"},
				{Type: token.ELEMENT_IDENT, Literal: "button"},
				{Type: token.ELEMENT_CLOSE_END, Literal: ">"},
				{Type: token.EOF},
			},
		},
		{
			name:  "Function declaration",
			input: `fn print() {}`,
			want: []token.Token{
				{Type: token.FUNC, Literal: "fn"},
				{Type: token.IDENT, Literal: "print"},
				{Type: token.LPAREN, Literal: "("},
				{Type: token.RPAREN, Literal: ")"},
				{Type: token.LBRACE, Literal: "{"},
				{Type: token.RBRACE, Literal: "}"},
				{Type: token.EOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := New([]byte(tt.input))

			var tokens []token.Token

			for {
				t := lex.NextToken()
				tokens = append(tokens, t)
				if t.Type == token.EOF {
					break
				}
			}

			if diff := cmp.Diff(tt.want, tokens, cmpOpts...); diff != "" {
				t.Errorf("Tokenize() mismatch (-want +got):\nInput:%s\n%s", tt.input, diff)
			}
		})
	}
}
