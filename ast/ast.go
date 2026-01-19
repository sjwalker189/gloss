package ast

import "gloss/token"

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

type Identifier struct {
	BaseNode
	Token token.Token
	Name  string
}

type Expression interface {
	expressionNode()
}

type InfixExpression struct {
	Left     Expression
	Right    Expression
	Operator string
}

type PrefixExpression struct {
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

func (e InfixExpression) expressionNode()  {}
func (e PrefixExpression) expressionNode() {}
func (e ParenExpression) expressionNode()  {}
func (e CallExpression) expressionNode()   {}
func (e IntegerLiteral) expressionNode()   {}
func (e StringLiteral) expressionNode()    {}
func (e Boolean) expressionNode()          {}
func (e Identifier) expressionNode()       {}

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

type Parameter struct {
	BaseNode
	Name    string
	Type    string
	Default *Expression
}

type Type struct {
	Name string
}

type Func struct {
	BaseNode
	Name       string
	Params     []*Parameter
	Body       *BlockStatement
	ReturnType *Type
}

type ReturnStatement struct {
	BaseNode
}

type IntegerLiteral struct {
	BaseNode
	Value int64
}

type StringLiteral struct {
	BaseNode
	Value string
}

type Boolean struct {
	BaseNode
	Value bool
}
