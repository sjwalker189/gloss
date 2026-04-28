package parser

import (
	"gloss/ast"
	"testing"
)

func TestStruct(t *testing.T) {
	input := `
		struct Rect {
			width: int,
			height: int,
		}
	`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.StructDeclaration{
				Name: &ast.Identifier{Name: "Rect"},
				Fields: []*ast.FieldDeclaration{
					{Name: &ast.Identifier{Name: "width"}, Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "int"}}},
					{Name: &ast.Identifier{Name: "height"}, Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "int"}}},
				},
			},
		},
	}

	assertParse(t, input, want)
}

func TestGenericStruct(t *testing.T) {
	input := `
		struct Box<T> {
			value: T,
		}
	`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.StructDeclaration{
				Name: &ast.Identifier{Name: "Box"},
				TypeParameters: []ast.Type{
					&ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
				},
				Fields: []*ast.FieldDeclaration{
					{
						Name: &ast.Identifier{Name: "value"},
						Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
					},
				},
			},
		},
	}

	assertParse(t, input, want)
}

func TestGenericCompositeStruct(t *testing.T) {
	input := `
		struct Box<T> {
			value: []T,
		}
	`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.StructDeclaration{
				Name: &ast.Identifier{Name: "Box"},
				TypeParameters: []ast.Type{
					&ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
				},
				Fields: []*ast.FieldDeclaration{
					{
						Name: &ast.Identifier{Name: "value"},
						Type: &ast.ArrayType{
							ElementType: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
						},
					},
				},
			},
		},
	}

	assertParse(t, input, want)
}

//	func TestStructWithReceivers(t *testing.T) {
//		input := `
//			struct Rect {
//				width: int,
//				height: int,
//				fn area() int {}
//			}
//		`
//		want := ast.SourceFile{
//			Declarations: []ast.Node{
//				&ast.StructDeclaration{
//					Name: &ast.Identifier{Name: "Box"},
//					TypeParameters: []ast.Type{
//						&ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
//					},
//					Fields: []*ast.FieldDeclaration{
//						{
//							Name: &ast.Identifier{Name: "value"},
//							Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
//						},
//					},
//				},
//			},
//		}
//
//		assertParse(t, input, want)
//	}
func TestStructExpression(t *testing.T) {
	input := `
		let user = User{ id: 1 }
	`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.LetStatement{
				Name: &ast.Identifier{Name: "user"},
				Value: &ast.CompositeLiteral{
					Type: &ast.Identifier{Name: "User"},
					Elements: []ast.Expression{
						&ast.KeyValuePair{
							Key:   &ast.Identifier{Name: "id"},
							Value: &ast.IntegerLiteral{Value: 1},
						},
					},
				},
			},
		},
	}

	assertParse(t, input, want)
}

func TestStructSliceExpression(t *testing.T) {
	input := `
		let users = []User{
			{ id: 1 },
			{ id: 2 },
		}
	`
	want := ast.SourceFile{
		Declarations: []ast.Node{
			&ast.LetStatement{
				Name: &ast.Identifier{Name: "users"},
				Value: &ast.ArrayTypeExpression{
					BaseType: &ast.CompositeLiteral{
						Type: &ast.Identifier{Name: "User"},
						Elements: []ast.Expression{
							&ast.CompositeLiteral{
								Type: nil,
								Elements: []ast.Expression{
									&ast.KeyValuePair{
										Key:   &ast.Identifier{Name: "id"},
										Value: &ast.IntegerLiteral{Value: 1},
									},
								},
							},
							&ast.CompositeLiteral{
								Type: nil,
								Elements: []ast.Expression{
									&ast.KeyValuePair{
										Key:   &ast.Identifier{Name: "id"},
										Value: &ast.IntegerLiteral{Value: 2},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	assertParse(t, input, want)
}
