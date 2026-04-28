package parser

import (
	"fmt"
	"gloss/ast"
	"testing"
)

func statements(content string) string {
	return fmt.Sprintf("fn main() { %s }", content)
}

func source(nodes ...ast.Node) ast.SourceFile {
	return ast.SourceFile{
		Declarations: []ast.Node{
			&ast.Func{
				Name: &ast.Identifier{Name: "main"},
				Body: &ast.BlockStatement{
					Statements: nodes,
				},
			},
		},
	}
}

func TestLoop(t *testing.T) {
	input := statements(`loop {}`)
	want := source(&ast.Loop{
		Body: &ast.BlockStatement{},
	})
	assertParse(t, input, want)
}

func TestForInLoop(t *testing.T) {
	input := statements(`
		for key, val in items {}
	`)
	want := source(
		&ast.ForEach{
			Key:      &ast.Identifier{Name: "key"},
			Value:    &ast.Identifier{Name: "val"},
			Iterable: &ast.Identifier{Name: "items"},
			Body:     &ast.BlockStatement{},
		},
		// &ast.ForEach{
		// 	Key:      &ast.Identifier{Name: "_"},
		// 	Value:    &ast.Identifier{Name: "val"},
		// 	Iterable: &ast.Identifier{Name: "items"},
		// 	Body:     &ast.BlockStatement{},
		// },
	)
	assertParse(t, input, want)
}
