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
	Type  string
}

type Expression struct {
	BaseNode
	// TODO
}

type LetStatement struct {
	BaseNode
	Token token.Token
	Name  Identifier
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
