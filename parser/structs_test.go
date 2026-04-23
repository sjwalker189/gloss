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

// func TestGenericCompositeStruct(t *testing.T) {
// 	input := `
// 		struct Box<T> {
// 			value: []T,
// 		}
// 	`
// 	want := ast.SourceFile{
// 		Declarations: []ast.Node{
// 			&ast.StructDeclaration{
// 				Name: &ast.Identifier{Name: "Box"},
// 				TypeParameters: []ast.Type{
// 					&ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
// 				},
// 				Fields: []*ast.FieldDeclaration{
// 					{
// 						Name: &ast.Identifier{Name: "value"},
// 						Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
// 					},
// 				},
// 			},
// 		},
// 	}
//
// 	assertParse(t, input, want)
// }

// func TestStructWithReceivers(t *testing.T) {
// 	input := `
// 		struct Rect {
// 			width: int,
// 			height: int,
// 			fn area() int {}
// 		}
// 	`
// 	want := ast.SourceFile{
// 		Declarations: []ast.Node{
// 			&ast.StructDeclaration{
// 				Name: &ast.Identifier{Name: "Box"},
// 				TypeParameters: []ast.Type{
// 					&ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
// 				},
// 				Fields: []*ast.FieldDeclaration{
// 					{
// 						Name: &ast.Identifier{Name: "value"},
// 						Type: &ast.TypeIdentifier{Name: &ast.Identifier{Name: "T"}},
// 					},
// 				},
// 			},
// 		},
// 	}
//
// 	assertParse(t, input, want)
// }
