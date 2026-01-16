package parser

import (
	"fmt"
	"gloss/ast"
	"gloss/lexer"

	"testing"

	"github.com/google/go-cmp/cmp"
	// "github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseFunction(t *testing.T) {
	cmpOpts := []cmp.Option{
		// cmpopts.IgnoreFields(token.Token{}, "Line", "Column"),
	}

	tests := []struct {
		name  string
		input string
		want  ast.SourceFile
	}{
		{
			name:  "Hello world",
			input: `fn print() {}`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Func{
						Name: "print",
					},
				},
			},
		},

		{
			name:  "Printer",
			input: `fn print(msg string) {}`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Func{
						Name: "print",
						Params: []*ast.Parameter{
							{
								Name: "msg",
								Type: "string",
							},
						},
					},
				},
			},
		},

		{
			name:  "adder",
			input: `fn add(a int, b int) int {}`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Func{
						Name: "add",
						Params: []*ast.Parameter{
							{
								Name: "a",
								Type: "int",
							},
							{
								Name: "b",
								Type: "int",
							},
						},
						ReturnType: &ast.Type{
							Name: "int",
						},
					},
				},
			},
		},

		{
			name:  "returns literal",
			input: `fn withreturn() { return }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Func{
						Name: "withreturn",
						Body: &ast.BlockStatement{
							Statements: []ast.Node{
								&ast.ReturnStatement{
									// Value: nil,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(lexer.New([]byte(tt.input)))
			got := p.Parse()

			for _, msg := range p.Diagnostics.Items {
				fmt.Println(msg.Text)
			}

			if diff := cmp.Diff(tt.want, got, cmpOpts...); diff != "" {
				t.Errorf("parser.Parse() mismatch (-want +got):\nInput:%s\n%s", tt.input, diff)
			}
		})
	}
}
