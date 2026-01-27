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

// --- If/Else Tests ---

func TestParseIf_ConstantCondition(t *testing.T) {
	input := `if true { }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.If{
				Condition: &ast.Boolean{Value: true},
				Then:      &ast.BlockStatement{},
				Else:      nil,
			},
		},
	}
	assertParse(t, input, want)
}

func TestParseIf_WithElse(t *testing.T) {
	input := `if true { } else { }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.If{
				Condition: &ast.Boolean{Value: true},
				Then:      &ast.BlockStatement{},
				Else:      &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}

func TestParseIf_WithElseIf(t *testing.T) {
	input := `if true { } else if false { }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.If{
				Condition: &ast.Boolean{Value: true},
				Then:      &ast.BlockStatement{},
				Else: &ast.If{
					Condition: &ast.Boolean{Value: false},
					Then:      &ast.BlockStatement{},
					Else:      nil,
				},
			},
		},
	}
	assertParse(t, input, want)
}

func TestParseIf_WithBinaryExpression(t *testing.T) {
	input := `if 10 > 0 { }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.If{
				Condition: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 10},
					Right:    &ast.IntegerLiteral{Value: 0},
					Operator: ">",
				},
				Then: &ast.BlockStatement{},
				Else: nil,
			},
		},
	}
	assertParse(t, input, want)
}

func TestParseIf_WithBinaryAndExpression(t *testing.T) {
	input := `if 1 && 2 { }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.If{
				Condition: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 1},
					Right:    &ast.IntegerLiteral{Value: 2},
					Operator: "&&",
				},
				Then: &ast.BlockStatement{},
				Else: nil,
			},
		},
	}
	assertParse(t, input, want)
}

func TestParseIf_WithUnaryExpression(t *testing.T) {
	input := `if !false { }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.If{
				Condition: &ast.UnaryExpression{
					Right:    &ast.Boolean{Value: false},
					Operator: "!",
				},
				Then: &ast.BlockStatement{},
				Else: nil,
			},
		},
	}
	assertParse(t, input, want)
}

// Loops

func TestParseLoop_Loop(t *testing.T) {
	input := `loop { break }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Loop{
				Body: &ast.BlockStatement{
					Statements: []ast.Node{
						&ast.BreakStatement{},
					},
				},
			},
		},
	}
	assertParse(t, input, want)
}

func TestParseLoop_For(t *testing.T) {
	input := `for a > 0 {  }`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.For{
				Condition: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Right:    &ast.IntegerLiteral{Value: 0},
					Operator: ">",
				},
				Body: &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}
