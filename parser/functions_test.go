package parser

import (
	"gloss/ast"
	"testing"
)

func TestMainFunction(t *testing.T) {
	input := `fn main() {}`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{
				Name: &ast.Identifier{Name: "main"},
				Body: &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}

func TestVoidFunction(t *testing.T) {
	input := `fn main() void {}`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{
				Name:       &ast.Identifier{Name: "main"},
				ReturnType: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "void"}},
				Body:       &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}

func TestGenericResultFunction(t *testing.T) {
	input := `fn main() Future<int> {}`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{
				Name: &ast.Identifier{Name: "main"},
				ReturnType: &ast.TypeIdentifier{
					Name: &ast.Identifier{Name: "Future"},
					Parameters: []ast.Type{
						&ast.TypeIdentifier{Name: &ast.Identifier{Name: "int"}},
					},
				},
				Body: &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}

func TestParameterFunction(t *testing.T) {
	input := `fn sum(a: int, b: int) {}`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{
				Name: &ast.Identifier{Name: "sum"},
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{Name: "a"},
						Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "int"}},
					},
					{
						Name: &ast.Identifier{Name: "b"},
						Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "int"}},
					},
				},
				Body: &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}

func TestGenericFunction(t *testing.T) {
	input := `fn sum<T>(a: T, b: T) T {}`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{
				Name: &ast.Identifier{Name: "sum"},
				TypeParameters: []ast.Type{
					&ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
				},
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{Name: "a"},
						Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
					},
					{
						Name: &ast.Identifier{Name: "b"},
						Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
					},
				},

				ReturnType: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
				Body:       &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}

func TestComplexGenericFunction(t *testing.T) {
	input := `fn cmp<A, B>(a: A, b: B) Either<A, B> {}`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{
				Name: &ast.Identifier{Name: "cmp"},
				TypeParameters: []ast.Type{
					&ast.TypeIdentifier{Name: &ast.Identifier{Name: "A"}},
					&ast.TypeIdentifier{Name: &ast.Identifier{Name: "B"}},
				},
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{Name: "a"},
						Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "A"}},
					},
					{
						Name: &ast.Identifier{Name: "b"},
						Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "B"}},
					},
				},

				ReturnType: &ast.TypeIdentifier{
					Name: &ast.Identifier{Name: "Either"},
					Parameters: []ast.Type{
						&ast.TypeIdentifier{Name: &ast.Identifier{Name: "A"}},
						&ast.TypeIdentifier{Name: &ast.Identifier{Name: "B"}},
					},
				},
				Body: &ast.BlockStatement{},
			},
		},
	}
	assertParse(t, input, want)
}
