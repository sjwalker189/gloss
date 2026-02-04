package compiler

import (
	"fmt"
	"gloss/lexer"
	"gloss/parser"

	"bytes"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func assertCompileResult(t *testing.T, input string, want string) {
	t.Helper()

	var w bytes.Buffer

	l := lexer.New([]byte(input))
	p := parser.NewParser(l)
	c := NewGoCompiler(&w)

	source := p.Parse()
	c.Compile(&source)

	gotBytes, err := io.ReadAll(&w)
	if err != nil {
		t.Error(err)
	}

	got := string(gotBytes)

	fmt.Println()
	fmt.Println(got)
	fmt.Println()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("go.Compile() mismatch (-want +got):\nInput:%s\n\nOutput:\n%s", input, diff)
	}
}

func TestCompilerMainFunc(t *testing.T) {
	input := `fn main() {}`
	want := "package main\n\nfunc main() {}"
	assertCompileResult(t, input, want)
}

// TODO: TABS VS SPACES
func TestCompiler(t *testing.T) {
	input := `fn sum(a int, b int) int {
	return a + b
	}`
	want := `package main

func sum(a int, b int) int {
    return a + b
}`
	assertCompileResult(t, input, want)
}
