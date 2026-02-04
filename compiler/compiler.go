package compiler

import (
	"fmt"
	"gloss/ast"
	"io"
	"strings"
)

type Compiler interface {
	Compile(file *ast.SourceFile)
}

type Go struct {
	writer      io.Writer
	packageName string
	indentLevel int
	indentSize  int
}

func NewGoCompiler(writer io.Writer) Compiler {
	return &Go{
		writer:      writer,
		packageName: "main",
		indentLevel: 0,
		indentSize:  4,
	}
}

func (c *Go) indent() {
	c.indentLevel++
}

func (c *Go) outdent() {
	if c.indentLevel > 0 {
		c.indentLevel--
	}
}

func (c *Go) emitIndent() {
	if c.indentLevel > 0 {
		c.emit("%s", strings.Repeat(" ", c.indentLevel*c.indentSize))
	}
}

func (c *Go) Compile(file *ast.SourceFile) {
	c.emit("package %s\n\n", c.packageName)

	for _, node := range file.Declarations {
		c.compileNode(node)
	}
}

func (c *Go) emit(format string, args ...any) {
	_, err := fmt.Fprintf(c.writer, format, args...)
	if err != nil {
		panic(err)
	}
}

func (c *Go) compileNode(node ast.Node) {
	switch n := node.(type) {
	case *ast.IntegerLiteral:
		c.compileIntegerLiteral(n)
	case *ast.ReturnStatement:
		c.compileReturnStatement(n)
	case *ast.Func:
		c.compileFunc(n)
	default:
	}
}

func (c *Go) compileType(node ast.Type) {
	switch t := node.(type) {
	case *ast.TypeIdentifier:
		c.compileTypeIdentifier(t)
	case *ast.TypeLiteral:
		c.emit("%s", t.Type)
	}
}

func (c *Go) compileTypeIdentifier(node *ast.TypeIdentifier) {
	c.emit("%s", node.Name)
	// TODO: node.Parameters
}

// Declarations

func (c *Go) compileFunc(node *ast.Func) {
	// TODO: Format func name for go conventions, e.g. pub fn sum = func Sum
	c.emit("func %s", node.Name)

	count := len(node.Params)

	// TODO: generic arguments

	c.emit("(")
	for i := range count {
		param := node.Params[i]
		c.emit("%s ", param.Name)
		c.compileType(param.Type)
		if i < count-1 {
			c.emit(", ")
		}
	}
	c.emit(")")

	if node.ReturnType != nil {
		c.emit(" ")
		c.compileType(node.ReturnType)
	}

	c.emit(" ")
	c.compileBlockStatement(node.Body)
}

// Statements

func (c *Go) compileBlockStatement(node *ast.BlockStatement) {
	c.emit("{")
	c.indent()
	for _, stmt := range node.Statements {
		c.emit("\n")
		c.emitIndent()
		c.compileNode(stmt)
	}
	c.outdent()
	if len(node.Statements) > 0 {
		c.emit("\n")
	}

	c.emit("}")
}

func (c *Go) compileReturnStatement(node *ast.ReturnStatement) {
	c.emit("return")
	if node.Value != nil {
		c.emit(" ")
		c.compileExpression(node.Value)
	}
}

// Expressions

func (c *Go) compileExpression(exp ast.Expression) {
	switch t := exp.(type) {
	case *ast.BinaryExpression:
		c.compileBinaryExpression(t)
	case *ast.UnaryExpression:
		c.compileUnaryExpression(t)
	case *ast.IntegerLiteral:
		c.compileIntegerLiteral(t)
	case *ast.Identifier:
		c.compileIdentifier(t)
	}
}

func (c *Go) compileUnaryExpression(exp *ast.UnaryExpression) {
	c.emit("%s", exp.Operator)
	c.compileExpression(exp.Right)
}

func (c *Go) compileBinaryExpression(exp *ast.BinaryExpression) {
	c.compileExpression(exp.Left)
	c.emit(" %s ", exp.Operator)
	c.compileExpression(exp.Right)
}

func (c *Go) compileIdentifier(node *ast.Identifier) {
	c.emit("%s", node.Name)
}

func (c *Go) compileIntegerLiteral(node *ast.IntegerLiteral) {
	c.emit("%d", node.Value)
}
