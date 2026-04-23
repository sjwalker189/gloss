package parser

import (
	"gloss/ast"
	"testing"
)

func TestVariables(t *testing.T) {
	input := `
		let five = 5
		let ok: bool = true
		let foo: Future<Either<bool, error>>
		let bar: Future<Either<bool, error>> = fetch()
		const five = 5
		const ok: bool = true
		const foo: Future<Either<bool, error>> = fetch()
	`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.LetStatement{
				Name:  &ast.Identifier{Name: "five"},
				Value: &ast.IntegerLiteral{Value: 5},
				Type:  nil,
			},
			&ast.LetStatement{
				Name:  &ast.Identifier{Name: "ok"},
				Value: &ast.Boolean{Value: true},
				Type:  &ast.TypeIdentifier{Name: &ast.Identifier{Name: "bool"}},
			},
			&ast.LetStatement{
				Name:  &ast.Identifier{Name: "foo"},
				Value: nil,
				Type: &ast.TypeIdentifier{
					Name: &ast.Identifier{Name: "Future"},
					Parameters: []ast.Type{
						&ast.TypeIdentifier{
							Name: &ast.Identifier{Name: "Either"},
							Parameters: []ast.Type{
								&ast.TypeIdentifier{Name: &ast.Identifier{Name: "bool"}},
								&ast.TypeIdentifier{Name: &ast.Identifier{Name: "error"}},
							},
						},
					},
				},
			},
			&ast.LetStatement{
				Name: &ast.Identifier{Name: "bar"},
				Type: &ast.TypeIdentifier{
					Name: &ast.Identifier{Name: "Future"},
					Parameters: []ast.Type{
						&ast.TypeIdentifier{
							Name: &ast.Identifier{Name: "Either"},
							Parameters: []ast.Type{
								&ast.TypeIdentifier{Name: &ast.Identifier{Name: "bool"}},
								&ast.TypeIdentifier{Name: &ast.Identifier{Name: "error"}},
							},
						},
					},
				},
				Value: &ast.CallExpression{
					Function:  &ast.Identifier{Name: "fetch"},
					Arguments: []ast.Expression{},
				},
			},
			&ast.ConstStatement{
				Name:  &ast.Identifier{Name: "five"},
				Value: &ast.IntegerLiteral{Value: 5},
				Type:  nil,
			},
			&ast.ConstStatement{
				Name:  &ast.Identifier{Name: "ok"},
				Value: &ast.Boolean{Value: true},
				Type:  &ast.TypeIdentifier{Name: &ast.Identifier{Name: "bool"}},
			},
			&ast.ConstStatement{
				Name: &ast.Identifier{Name: "foo"},
				Type: &ast.TypeIdentifier{
					Name: &ast.Identifier{Name: "Future"},
					Parameters: []ast.Type{
						&ast.TypeIdentifier{
							Name: &ast.Identifier{Name: "Either"},
							Parameters: []ast.Type{
								&ast.TypeIdentifier{Name: &ast.Identifier{Name: "bool"}},
								&ast.TypeIdentifier{Name: &ast.Identifier{Name: "error"}},
							},
						},
					},
				},
				Value: &ast.CallExpression{
					Function:  &ast.Identifier{Name: "fetch"},
					Arguments: []ast.Expression{},
				},
			},
		},
	}
	assertParse(t, input, want)
}

func TestCallVariables(t *testing.T) {
	input := `
		let five = sum()
	`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.LetStatement{
				Name: &ast.Identifier{Name: "five"},
				Value: &ast.CallExpression{
					Function:  &ast.Identifier{Name: "sum"},
					Arguments: []ast.Expression{},
				},
			},
		},
	}
	assertParse(t, input, want)
}
