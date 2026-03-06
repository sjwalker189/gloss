package ast

// ----------------------------------------------------------------------------
// Interfaces
// ----------------------------------------------------------------------------

type Node interface {
	GetRange() Range
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Declaration interface {
	Node
	declarationNode()
}

type Type interface {
	Node
	typeNode()
}

type Pattern interface {
	Node
	patternNode()
}

// ----------------------------------------------------------------------------
// Base Node & Range
// ----------------------------------------------------------------------------

type Range struct {
	StartByte uint
	EndByte   uint
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

type BaseNode struct {
	Range
}

func (b BaseNode) GetRange() Range { return b.Range }

// ----------------------------------------------------------------------------
// Top Level
// ----------------------------------------------------------------------------

type SourceFile struct {
	BaseNode
	Declarations []Declaration
}

// ----------------------------------------------------------------------------
// Declarations
// ----------------------------------------------------------------------------

type UseDeclaration struct {
	BaseNode
	Module string
}

type TypeDeclaration struct {
	BaseNode
	IsPublic       bool
	IsExtern       bool
	Name           string
	TypeParameters []*TypeParameter
	Type           Type
}

type EnumDeclaration struct {
	BaseNode
	IsPublic bool
	IsExtern bool
	Name     string
	Variants []*EnumVariant
}

type EnumVariant struct {
	BaseNode
	Name  string
	Value Expression // Optional
}

type UnionDeclaration struct {
	BaseNode
	IsPublic       bool
	IsExtern       bool
	Name           string
	TypeParameters []*TypeParameter
	Variants       []*UnionVariant
}

type UnionVariant struct {
	BaseNode
	Name    string
	Payload Type // Optional (Type or StructBody)
}

type StructDeclaration struct {
	BaseNode
	IsPublic       bool
	IsExtern       bool
	Name           string
	TypeParameters []*TypeParameter
	Fields         []*StructField
	Methods        []*FunctionDeclaration
}

type StructField struct {
	BaseNode
	Name string
	Type Type
}

type FunctionDeclaration struct {
	BaseNode
	IsPublic       bool
	Name           string
	TypeParameters []*TypeParameter
	Parameters     []*Parameter
	ReturnType     Type
	Body           *BlockStatement // Body is expression or block in grammar? Grammar says `block`.
}

type VariableDeclaration struct {
	BaseNode
	IsPublic bool
	IsConst  bool // true for `const`, false for `let`
	Name     string
	Type     Type // Optional
	Value    Expression
}

type Parameter struct {
	BaseNode
	Name string
	Type Type // Optional
}

type TypeParameter struct {
	BaseNode
	Name string
}

func (d UseDeclaration) declarationNode()      {}
func (d TypeDeclaration) declarationNode()     {}
func (d EnumDeclaration) declarationNode()     {}
func (d UnionDeclaration) declarationNode()    {}
func (d StructDeclaration) declarationNode()   {}
func (d FunctionDeclaration) declarationNode() {}
func (d VariableDeclaration) declarationNode() {}

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

type TypeIdentifier struct {
	BaseNode
	Name string
}

type PrimitiveType struct {
	BaseNode
	Name string // int, string, bool, void, nil
}

type GenericType struct {
	BaseNode
	Name          string
	TypeArguments []Type
}

type SliceType struct {
	BaseNode
	Size    *IntegerLiteral // Optional, for [N]T
	Element Type
}

type TupleType struct {
	BaseNode
	Elements []Type
}

type StructTypeLiteral struct {
	BaseNode
	Fields []*StructField
}

func (t TypeIdentifier) typeNode()    {}
func (t PrimitiveType) typeNode()     {}
func (t GenericType) typeNode()       {}
func (t SliceType) typeNode()         {}
func (t TupleType) typeNode()         {}
func (t StructTypeLiteral) typeNode() {}

// ----------------------------------------------------------------------------
// Statements
// ----------------------------------------------------------------------------

type BlockStatement struct {
	BaseNode
	Statements []Statement
	Expression Expression // Optional trailing expression
}

type ReturnStatement struct {
	BaseNode
	Value Expression // Optional
}

type AssignmentStatement struct {
	BaseNode
	Left     Expression // Identifier, MemberExpression, IndexExpression
	Operator string     // =, +=, etc.
	Right    Expression
}

type IfStatement struct {
	BaseNode
	Condition   Expression
	Consequence *BlockStatement
	Alternative Statement // BlockStatement or IfStatement (else if)
}

type LoopStatement struct {
	BaseNode
	Body *BlockStatement
}

type WhileStatement struct {
	BaseNode
	Condition Expression
	Body      *BlockStatement
}

type ForStatement struct {
	BaseNode
	Initializer Statement // VariableDeclaration or Expression
	Condition   Expression
	Update      Expression
	Body        *BlockStatement
}

type ForInStatement struct {
	BaseNode
	Index string
	Value string
	Right Expression
	Body  *BlockStatement
}

type BreakStatement struct{ BaseNode }
type ContinueStatement struct{ BaseNode }

type ExpressionStatement struct {
	BaseNode
	Expression Expression
}

func (s BlockStatement) statementNode()      {}
func (s ReturnStatement) statementNode()     {}
func (s AssignmentStatement) statementNode() {}
func (s IfStatement) statementNode()         {}
func (s LoopStatement) statementNode()       {}
func (s WhileStatement) statementNode()      {}
func (s ForStatement) statementNode()        {}
func (s ForInStatement) statementNode()      {}
func (s BreakStatement) statementNode()      {}
func (s ContinueStatement) statementNode()   {}
func (s ExpressionStatement) statementNode() {}

// Also allow Declarations to be Statements in block contexts?
// Go doesn't allow declarations as statements in the interface usually, but the parser handles it.
// The grammar says `_statement` includes `variable_declaration`.
// So VariableDeclaration should implement Statement too?
func (d VariableDeclaration) statementNode() {}

// ----------------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------------

type Identifier struct {
	BaseNode
	Name string
}

type IntegerLiteral struct {
	BaseNode
	Value  int64 // Or string if handling big ints
	Source string
}

type FloatLiteral struct {
	BaseNode
	Value  float64
	Source string
}

type StringLiteral struct {
	BaseNode
	Value string
}

type BooleanLiteral struct {
	BaseNode
	Value bool
}

type UnaryExpression struct {
	BaseNode
	Operator string
	Operand  Expression
}

type BinaryExpression struct {
	BaseNode
	Left     Expression
	Operator string
	Right    Expression
}

type UpdateExpression struct {
	BaseNode
	Operator string // ++, --
	Operand  Expression
	Post     bool // true if x++, false if ++x
}

type ParenExpression struct {
	BaseNode
	Expression Expression
}

type TupleExpression struct {
	BaseNode
	Elements []Expression
}

type CompositeLiteral struct {
	BaseNode
	Type   Type // Optional or inferred? Grammar says `type` field is present.
	Fields []*FieldValue
	Values []Expression // For slices/arrays
}

type FieldValue struct {
	BaseNode
	Name  string
	Value Expression
}

type CallExpression struct {
	BaseNode
	Function      Expression
	TypeArguments []Type
	Arguments     []Expression
}

type MemberExpression struct {
	BaseNode
	Object   Expression
	Property string
	Optional bool // ?.
}

type IndexExpression struct {
	BaseNode
	Operand  Expression
	Index    Expression
	Optional bool // ?.[
}

type MatchExpression struct {
	BaseNode
	Value Expression
	Arms  []*MatchArm
}

type MatchArm struct {
	BaseNode
	Pattern Pattern
	Value   Expression // Expression or Block
}

type AnonymousFunction struct {
	BaseNode
	TypeParameters []*TypeParameter
	Parameters     []*Parameter
	ReturnType     Type
	Body           Node // BlockStatement or Expression
}

type JSXElement struct {
	BaseNode
	OpeningElement *JSXOpeningElement
	Children       []Node // JSXText, JSXElement, JSXExpression
	ClosingElement *JSXClosingElement
}

type JSXSelfClosingElement struct {
	BaseNode
	Name       Expression // Identifier, TypeIdentifier, MemberExpression
	Attributes []*JSXAttribute
}

type JSXFragment struct {
	BaseNode
	Children []Node
}

type JSXOpeningElement struct {
	BaseNode
	Name       Expression
	Attributes []*JSXAttribute
}

type JSXClosingElement struct {
	BaseNode
	Name Expression
}

type JSXAttribute struct {
	BaseNode
	Name  string
	Value Expression // StringLiteral or JSXExpression
}

type JSXExpression struct {
	BaseNode
	Expression Expression
}

type JSXText struct {
	BaseNode
	Value string
}

func (e Identifier) expressionNode()            {}
func (e IntegerLiteral) expressionNode()        {}
func (e FloatLiteral) expressionNode()          {}
func (e StringLiteral) expressionNode()         {}
func (e BooleanLiteral) expressionNode()        {}
func (e UnaryExpression) expressionNode()       {}
func (e BinaryExpression) expressionNode()      {}
func (e UpdateExpression) expressionNode()      {}
func (e ParenExpression) expressionNode()       {}
func (e TupleExpression) expressionNode()       {}
func (e CompositeLiteral) expressionNode()      {}
func (e CallExpression) expressionNode()        {}
func (e MemberExpression) expressionNode()      {}
func (e IndexExpression) expressionNode()       {}
func (e MatchExpression) expressionNode()       {}
func (e AnonymousFunction) expressionNode()     {}
func (e JSXElement) expressionNode()            {}
func (e JSXSelfClosingElement) expressionNode() {}
func (e JSXFragment) expressionNode()           {}
func (e JSXExpression) expressionNode()         {}

// ----------------------------------------------------------------------------
// Patterns
// ----------------------------------------------------------------------------

type WildcardPattern struct{ BaseNode }

type LiteralPattern struct {
	BaseNode
	Value Expression // Integer, String, Boolean
}

type IdentifierPattern struct {
	BaseNode
	Name string
}

type EnumPattern struct {
	BaseNode
	Name     string
	Elements []Pattern
}

func (p WildcardPattern) patternNode()   {}
func (p LiteralPattern) patternNode()    {}
func (p IdentifierPattern) patternNode() {}
func (p EnumPattern) patternNode()       {}
