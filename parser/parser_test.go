package parser

import (
	"fmt"
	"gloss/ast"
	"gloss/lexer"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func assertParse(t *testing.T, input string, want ast.SourceFile) {
	t.Helper()
	p := NewParser(lexer.New([]byte(input)))
	got := p.Parse()

	// TODO: should fail test if any error diagnostics raised
	for _, msg := range p.Diagnostics.Messages() {
		fmt.Println(msg.Text)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parser.Parse() mismatch (-want +got):\nInput:%s\n%s", input, diff)
	}
}

// --- Function Tests ---

func TestParseFunc_HelloWorld(t *testing.T) {
	input := `fn print() {}`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{Name: "print", Body: &ast.BlockStatement{}},
		},
	}
	assertParse(t, input, want)
}

func TestParseFunc_WithParams(t *testing.T) {
	input := `fn print(msg string) {}`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{
				Name: "print",
				Params: []*ast.Parameter{
					{Name: "msg", Type: &ast.TypeLiteral{Type: "string"}},
				},
				Body: &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}

func TestParseFunc_Adder(t *testing.T) {
	input := `fn add(a int, b int) int {}`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{
				Name: "add",
				Params: []*ast.Parameter{
					{Name: "a", Type: &ast.TypeLiteral{Type: "int"}},
					{Name: "b", Type: &ast.TypeLiteral{Type: "int"}},
				},
				ReturnType: &ast.TypeLiteral{Type: "int"},
				Body:       &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}

func TestParseFunc_ReturnBinaryExpression(t *testing.T) {
	input := `fn withreturn() { return 2 + 3 }`
	want := ast.SourceFile{
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
	}
	assertParse(t, input, want)
}

func TestParseFunc_Generic(t *testing.T) {
	input := `fn join<T>(a T, b T) T { }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{
				Name: "join",
				Params: []*ast.Parameter{
					{Name: "a", Type: &ast.TypeIdentifier{Name: "T"}},
					{Name: "b", Type: &ast.TypeIdentifier{Name: "T"}},
				},
				TypeParams: []*ast.TypeParameter{{Name: "T"}},
				ReturnType: &ast.TypeIdentifier{Name: "T"},
				Body:       &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}

// --- Let Statement Tests ---

func TestParseLet_String(t *testing.T) {
	input := `let msg = "hello world"`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.LetStatement{
				Name:  &ast.Identifier{Name: "msg"},
				Value: &ast.StringLiteral{Value: "hello world"},
			},
		},
	}
	assertParse(t, input, want)
}

func TestParseLet_ComplexExpression(t *testing.T) {
	input := `let zero = (10-5)*0`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.LetStatement{
				Name: &ast.Identifier{Name: "zero"},
				Value: &ast.BinaryExpression{
					Left: &ast.ParenExpression{
						Expression: &ast.BinaryExpression{
							Left:     &ast.IntegerLiteral{Value: 10},
							Right:    &ast.IntegerLiteral{Value: 5},
							Operator: "-",
						},
					},
					Right:    &ast.IntegerLiteral{Value: 0},
					Operator: "*",
				},
			},
		},
	}
	assertParse(t, input, want)
}

// --- Enum Tests ---

func TestParseEnum_MixedValues(t *testing.T) {
	input := `enum Message { Increment = 1, Decrement = "down", Clear, }`
	want := ast.SourceFile{
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
	}
	assertParse(t, input, want)
}

// --- Union Tests ---

func TestParseUnion_WithStructTypes(t *testing.T) {
	input := `
		union Shape {
			Square({ size: int }),
		}
	`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Union{
				Name: "Shape",
				Fields: []*ast.UnionField{
					{
						Name: "Square",
						Type: &ast.StructBody{
							Fields: []*ast.StructField{
								{Name: "size", Type: &ast.TypeLiteral{Type: "int"}},
							},
						},
					},
				},
			},
		},
	}
	assertParse(t, input, want)
}

func TestParseUnion_Generic(t *testing.T) {
	input := `union Option<T> { Some(T), None }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Union{
				Name:       "Option",
				Parameters: []*ast.TypeParameter{{Name: "T"}},
				Fields: []*ast.UnionField{
					{Name: "Some", Type: &ast.TypeIdentifier{Name: "T"}},
					{Name: "None", Type: nil},
				},
			},
		},
	}
	assertParse(t, input, want)
}

// --- Struct Tests ---

func TestParseStruct_Generic(t *testing.T) {
	input := `struct Point<T> { x: T, y: T }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Struct{
				Name:   "Point",
				Params: []*ast.TypeParameter{{Name: "T"}},
				Fields: []*ast.StructField{
					{Name: "x", Type: &ast.TypeIdentifier{Name: "T"}},
					{Name: "y", Type: &ast.TypeIdentifier{Name: "T"}},
				},
			},
		},
	}
	assertParse(t, input, want)
}
