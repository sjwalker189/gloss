package ast

type EnumDeclaration struct {
	BaseNode
	Name    string
	Members []*EnumMember
}

type EnumMember struct {
	BaseNode
	Name  string
	Value string
}
