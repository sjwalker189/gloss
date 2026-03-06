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
		_, _ = fmt.Fprintf(c.writer, "%s", strings.Repeat(" ", c.indentLevel*c.indentSize))
	}
}

func (c *Go) emit(format string, args ...any) {
	_, err := fmt.Fprintf(c.writer, format, args...)
	if err != nil {
		panic(err)
	}
}

func (c *Go) Compile(file *ast.SourceFile) {
	c.emit("package %s\n\n", c.packageName)

	for _, decl := range file.Declarations {
		c.compileDeclaration(decl)
		c.emit("\n\n")
	}
}

func (c *Go) compileDeclaration(decl ast.Declaration) {
	switch d := decl.(type) {
	case *ast.FunctionDeclaration:
		c.compileFunctionDeclaration(d)
	// TODO: Handle other declarations
	default:
		c.emit("// Unknown declaration type: %T", d)
	}
}

func (c *Go) compileNode(node ast.Node) {
	switch n := node.(type) {
	case ast.Statement:
		c.compileStatement(n)
	case ast.Expression:
		c.compileExpression(n)
	case ast.Declaration:
		c.compileDeclaration(n)
	default:
		c.emit("// Unknown node type: %T", n)
	}
}

func (c *Go) compileStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		c.compileReturnStatement(s)
	case *ast.BlockStatement:
		c.compileBlockStatement(s)
	// TODO: Other statements
	default:
		c.emit("// Unknown statement type: %T", s)
	}
}

func (c *Go) compileType(node ast.Type) {
	switch t := node.(type) {
	case *ast.TypeIdentifier:
		c.emit("%s", t.Name)
	case *ast.PrimitiveType:
		c.emit("%s", t.Name)
	default:
		c.emit("any") // Fallback
	}
}

// Declarations

func (c *Go) compileFunctionDeclaration(node *ast.FunctionDeclaration) {
	c.emit("func %s", node.Name)

	c.emit("(")
	for i, param := range node.Parameters {
		c.emit("%s ", param.Name)
		if param.Type != nil {
			c.compileType(param.Type)
		} else {
			c.emit("any")
		}
		if i < len(node.Parameters)-1 {
			c.emit(", ")
		}
	}
	c.emit(")")

	if node.ReturnType != nil {
		c.emit(" ")
		c.compileType(node.ReturnType)
	}

	c.emit(" ")
	if node.Body != nil {
		c.compileBlockStatement(node.Body)
	}
}

// Statements

func (c *Go) compileBlockStatement(node *ast.BlockStatement) {
	c.emit("{")
	c.indent()
	for _, stmt := range node.Statements {
		c.emit("\n")
		c.emitIndent()
		c.compileStatement(stmt)
	}
	c.outdent()
	if len(node.Statements) > 0 {
		c.emit("\n")
		c.emitIndent()
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
	default:
		c.emit("/* expr %T */", t)
	}
}

func (c *Go) compileUnaryExpression(exp *ast.UnaryExpression) {
	c.emit("%s", exp.Operator)
	c.compileExpression(exp.Operand)
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
