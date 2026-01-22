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
								Type: &ast.TypeLiteral{Type: "string"},
							},
						},
						ReturnType: nil,
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
								Type: &ast.TypeLiteral{Type: "int"},
							},
							{
								Name: "b",
								Type: &ast.TypeLiteral{Type: "int"},
							},
						},
						ReturnType: &ast.TypeLiteral{Type: "int"},
					},
				},
			},
		},
		{
			name:  "returns void",
			input: `fn withreturn() { return }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Func{
						Name: "withreturn",
						Body: &ast.BlockStatement{
							Statements: []ast.Node{
								&ast.ReturnStatement{
									Value: nil,
								},
							},
						},
					},
				},
			},
		},

		{
			name:  "returns string",
			input: `fn withreturn() { return "a" }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Func{
						Name: "withreturn",
						Body: &ast.BlockStatement{
							Statements: []ast.Node{
								&ast.ReturnStatement{
									Value: &ast.StringLiteral{
										Value: "a",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "returns int",
			input: `fn withreturn() { return 0 }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Func{
						Name: "withreturn",
						Body: &ast.BlockStatement{
							Statements: []ast.Node{
								&ast.ReturnStatement{
									Value: &ast.IntegerLiteral{
										Value: 0,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "returns expression",
			input: `fn withreturn() { return 2 + 3 }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Func{
						Name: "withreturn",
						Body: &ast.BlockStatement{
							Statements: []ast.Node{
								&ast.ReturnStatement{
									Value: &ast.BinaryExpression{
										Left:     &ast.IntegerLiteral{Value: 2},
										Right:    &ast.IntegerLiteral{Value: 3},
										Operator: "+",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "returns boolean",
			input: `fn withreturn() { return true }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Func{
						Name: "withreturn",
						Body: &ast.BlockStatement{
							Statements: []ast.Node{
								&ast.ReturnStatement{
									Value: &ast.Boolean{Value: true},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "returns paren expression",
			input: `fn withreturn() { return (2) }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Func{
						Name: "withreturn",
						Body: &ast.BlockStatement{
							Statements: []ast.Node{
								&ast.ReturnStatement{
									Value: &ast.ParenExpression{
										Expression: &ast.IntegerLiteral{Value: 2},
									},
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

func TestParseLetStatement(t *testing.T) {
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
			input: `let msg = "hello world"`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.LetStatement{
						Name: &ast.Identifier{
							Name: "msg",
						},
						Value: &ast.StringLiteral{
							Value: "hello world",
						},
					},
				},
			},
		},
		{
			name:  "Assign number",
			input: `let duration = 1_000`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.LetStatement{
						Name: &ast.Identifier{
							Name: "duration",
						},
						Value: &ast.IntegerLiteral{
							Value: 1000,
						},
					},
				},
			},
		},
		{
			name:  "Assign boolean",
			input: `let enabled = true`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.LetStatement{
						Name: &ast.Identifier{
							Name: "enabled",
						},
						Value: &ast.Boolean{
							Value: true,
						},
					},
				},
			},
		},

		{
			name: "Assign expression",
			input: `
				let five = 2 + 3
				let ten = (5 + 5)
				let zero = (10-5)*0
				let result = calc()
			`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.LetStatement{
						Name: &ast.Identifier{
							Name: "five",
						},
						Value: &ast.BinaryExpression{
							Left:     &ast.IntegerLiteral{Value: 2},
							Right:    &ast.IntegerLiteral{Value: 3},
							Operator: "+",
						},
					},
					&ast.LetStatement{
						Name: &ast.Identifier{
							Name: "ten",
						},
						Value: &ast.ParenExpression{
							Expression: &ast.BinaryExpression{
								Left:     &ast.IntegerLiteral{Value: 5},
								Right:    &ast.IntegerLiteral{Value: 5},
								Operator: "+",
							},
						},
					},
					&ast.LetStatement{
						Name: &ast.Identifier{
							Name: "zero",
						},
						Value: &ast.BinaryExpression{
							Left: &ast.ParenExpression{
								Expression: &ast.BinaryExpression{
									Left:     &ast.IntegerLiteral{Value: 10},
									Right:    &ast.IntegerLiteral{Value: 5},
									Operator: "-",
								},
							},
							Right: &ast.IntegerLiteral{
								Value: 0,
							},
							Operator: "*",
						},
					},
					&ast.LetStatement{
						Name: &ast.Identifier{
							Name: "result",
						},
						Value: &ast.CallExpression{
							Function:  &ast.Identifier{Name: "calc"},
							Arguments: []ast.Expression{},
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

func TestParseEnumStatement(t *testing.T) {
	cmpOpts := []cmp.Option{
		// cmpopts.IgnoreFields(token.Token{}, "Line", "Column"),
	}

	tests := []struct {
		name  string
		input string
		want  ast.SourceFile
	}{
		{
			name:  "Naked Enum",
			input: `enum Message { Increment, Decrement, }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Enum{
						Name: "Message",
						Members: []*ast.EnumMember{
							{Name: "Increment", IntValue: 0, Value: nil},
							{Name: "Decrement", IntValue: 1, Value: nil},
						},
					},
				},
			},
		},
		{
			name:  "Backed Enum",
			input: `enum Message { Increment = 0, Decrement = 1, }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Enum{
						Name: "Message",
						Members: []*ast.EnumMember{
							{Name: "Increment", IntValue: 0, Value: &ast.IntegerLiteral{Value: 0}},
							{Name: "Decrement", IntValue: 1, Value: &ast.IntegerLiteral{Value: 1}},
						},
					},
				},
			},
		},
		{
			name:  "Backed Enum with inferred values",
			input: `enum Message { Increment = 1, Decrement, }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Enum{
						Name: "Message",
						Members: []*ast.EnumMember{
							{Name: "Increment", IntValue: 1, Value: &ast.IntegerLiteral{Value: 1}},
							{Name: "Decrement", IntValue: 2, Value: nil},
						},
					},
				},
			},
		},
		{
			name:  "Backed Enum with mixed values",
			input: `enum Message { Increment = 1, Decrement = "down", Clear, }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Enum{
						Name: "Message",
						Members: []*ast.EnumMember{
							{Name: "Increment", IntValue: 1, Value: &ast.IntegerLiteral{Value: 1}},
							{Name: "Decrement", IntValue: 2, Value: &ast.StringLiteral{Value: "down"}},
							{Name: "Clear", IntValue: 3, Value: nil},
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

func TestParseUnionStatement(t *testing.T) {
	cmpOpts := []cmp.Option{
		// cmpopts.IgnoreFields(token.Token{}, "Line", "Column"),
	}

	tests := []struct {
		name  string
		input string
		want  ast.SourceFile
	}{
		{
			name:  "Naked Union",
			input: `union Message { Increment, Decrement, }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Union{
						Name: "Message",
						Fields: []*ast.UnionField{
							{Name: "Increment", Type: nil},
							{Name: "Decrement", Type: nil},
						},
					},
				},
			},
		},
		{
			name: "Union with literal types",
			input: `
				union Message {
					Increment(int),
					Decrement(int),
					Reset(string),
					Done(bool),
				}
			`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Union{
						Name: "Message",
						Fields: []*ast.UnionField{
							{Name: "Increment", Type: &ast.TypeLiteral{Type: "int"}},
							{Name: "Decrement", Type: &ast.TypeLiteral{Type: "int"}},
							{Name: "Reset", Type: &ast.TypeLiteral{Type: "string"}},
							{Name: "Done", Type: &ast.TypeLiteral{Type: "bool"}},
						},
					},
				},
			},
		},
		{
			name: "Union with struct types",
			input: `
				union Shape {
					Square({ size: int }),
					Rectangle({ width: int, height: int, }),
					Circle({ radius: int, }),
				}
			`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Union{
						Name: "Shape",
						Fields: []*ast.UnionField{
							{Name: "Square", Type: &ast.StructBody{
								Fields: []*ast.StructField{
									{Name: "size", Type: &ast.TypeLiteral{Type: "int"}},
								},
							}},
							{Name: "Rectangle", Type: &ast.StructBody{
								Fields: []*ast.StructField{
									{Name: "width", Type: &ast.TypeLiteral{Type: "int"}},
									{Name: "height", Type: &ast.TypeLiteral{Type: "int"}},
								},
							}},
							{Name: "Circle", Type: &ast.StructBody{
								Fields: []*ast.StructField{
									{Name: "radius", Type: &ast.TypeLiteral{Type: "int"}},
								},
							}},
						},
					},
				},
			},
		},
		{
			name:  "Union with tuple types",
			input: `union Message { Increment(int), Decrement(int), }`,
			want: ast.SourceFile{
				Declarations: []ast.Node{
					&ast.Union{
						Name: "Message",
						Fields: []*ast.UnionField{
							{Name: "Increment", Type: &ast.TypeLiteral{Type: "int"}},
							{Name: "Decrement", Type: &ast.TypeLiteral{Type: "int"}},
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
