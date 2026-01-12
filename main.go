package main

import (
	"fmt"
	"gloss/lexer"
	"gloss/parser"
)

var source = []byte(`
enum State {
	On,
	Off,
end
`)

func main() {
	l := lexer.New(source)
	tokens := l.Tokenize()
	fmt.Printf("\n--- Tokens ---\n")
	for _, tok := range tokens {
		fmt.Printf("%+v\n", tok)
	}

	p := parser.NewParser(tokens)
	file := p.Parse()

	fmt.Println("\ndeclarations: ", len(file.Declarations))

	for _, node := range file.Declarations {
		fmt.Printf("\n--- AST ---\n")

		switch decl := node.(type) {
		case *parser.EnumDeclaration:
			fmt.Printf("Enum: %s\n", decl.Name)
			for _, m := range decl.Members {
				val := m.Value
				if val == "" {
					val = "(auto)"
				}
				fmt.Printf("  Member: %-10s Value: %s\n", m.Name, val)
			}
		default:
			fmt.Println("unhandled type", decl)
		}
	}

	if len(p.Diagnostics) > 0 {
		fmt.Printf("\n--- Diagnostics (Errors Found) ---\n")
		for _, diag := range p.Diagnostics {
			fmt.Printf("%s at line %d col %d \n", diag.Message, diag.Line, diag.Column)
		}
	} else {
		fmt.Println("\nSuccess! No errors.")
	}
}
