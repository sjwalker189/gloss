package ast

import (
	"gloss/token"
)

type SourceFile struct {
	BaseNode
	Declarations []Node
}

// Range tracks where this node is in the source code.
type Range struct {
	StartByte uint
	EndByte   uint
}

// BaseNode implements the common interface for all nodes.
type BaseNode struct {
	Range
}

func (b BaseNode) GetRange() Range { return b.Range }

type Node interface {
	GetRange() Range
}

type Type interface {
	typeNode()
}

type Expression interface {
	expressionNode()
}

type Statement interface {
	statementNode()
}

type Alternative interface {
	alternativeNode()
}

type Identifier struct {
	BaseNode
	Token token.Token
	Name  string
}

type BinaryExpression struct {
	Left     Expression
	Right    Expression
	Operator string
}

type UnaryExpression struct {
	Right    Expression
	Operator string
}

type ParenExpression struct {
	Expression Expression
}

type CallExpression struct {
	Function  Expression
	Arguments []Expression
}

// TODO: explore let assert ... syntax similar to gleam
type LetStatement struct {
	BaseNode
	Token token.Token
	Name  *Identifier
	Value Expression
}

type BlockStatement struct {
	BaseNode
	Statements []Node
}

// TODO: statement or expression?
type If struct {
	BaseNode
	Condition Expression
	Then      *BlockStatement
	Else      Alternative
	// TODO: Else can be a block or another if condition
}

type Loop struct {
	BaseNode
	Body *BlockStatement
}

type For struct {
	BaseNode
	Condition Expression
	Body      *BlockStatement
}

type Parameter struct {
	BaseNode
	Name    string
	Type    Type
	Default *Expression
}

type TypeIdentifier struct {
	BaseNode
	Name       string
	Parameters []*TypeParameter
}

type TypeParameter struct {
	BaseNode
	Name string
}

type TypeLiteral struct {
	BaseNode
	Type string
}

type Enum struct {
	BaseNode
	Name    string
	Members []*EnumMember
}

type EnumMember struct {
	BaseNode
	Name     string
	IntValue int64
	Value    Expression
}

type Union struct {
	BaseNode
	Name       string
	Fields     []*UnionField
	Parameters []*TypeParameter
}

type UnionField struct {
	BaseNode
	Name string
	Type Type // literal type | union type | struct body | type ref with or without parameters
}

type Func struct {
	BaseNode
	Name       string
	Params     []*Parameter
	TypeParams []*TypeParameter
	Body       *BlockStatement
	ReturnType Type
}

type ReturnStatement struct {
	BaseNode
	Value Expression
}

type BreakStatement struct{ BaseNode }
type ContinueStatement struct{ BaseNode }

type IntegerLiteral struct {
	BaseNode
	Value  int64
	Signed bool
}

type StringLiteral struct {
	BaseNode
	Value string
}

type Boolean struct {
	BaseNode
	Value bool
}

type Tuple struct {
	BaseNode
	Items []Type
}

type Struct struct {
	BaseNode
	Name   string
	Params []*TypeParameter
	Fields []*StructField
}

type StructBody struct {
	BaseNode
	Fields []*StructField
}

type StructField struct {
	BaseNode
	Name string
	Type Type
}

type TupleType struct {
	BaseNode
	Fields []Type
}

// Denote nodes which can be used as types
func (n TypeIdentifier) typeNode() {}
func (n TypeLiteral) typeNode()    {}
func (n StructBody) typeNode()     {}

// Denote expression nodes
func (e BinaryExpression) expressionNode() {}
func (e UnaryExpression) expressionNode()  {}
func (e ParenExpression) expressionNode()  {}
func (e CallExpression) expressionNode()   {}
func (e IntegerLiteral) expressionNode()   {}
func (e StringLiteral) expressionNode()    {}
func (e Boolean) expressionNode()          {}
func (e Identifier) expressionNode()       {}

// Denote alternative nodes (those that can be chained with if)
func (n If) alternativeNode()             {}
func (n BlockStatement) alternativeNode() {}
