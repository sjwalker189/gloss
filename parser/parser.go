package parser

import (
	"fmt"
	"gloss/token"
	"slices"
)

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

type SourceFile struct {
	BaseNode
	Declarations []Node
}

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

type Diagnostic struct {
	Line    int
	Column  int
	Message string
}

type Parser struct {
	current     int
	tokens      []token.Token
	Diagnostics []Diagnostic
}

func NewParser(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, current: 0}
}

// Parse is the entry point
func (p *Parser) Parse() SourceFile {
	file := SourceFile{}

	for {
		tok := p.tokens[p.current]
		if tok.Type == token.EOF {
			return file
		}

		var decl Node

		switch tok.Type {
		case token.ENUM:
			decl = p.parseEnum()
		default:
			fmt.Println("type is ", tok.Type, tok.Literal)
			panic("unhandled type")
		}

		if decl != nil {
			file.Declarations = append(file.Declarations, decl)
		}
	}
}

func (p *Parser) parseEnum() *EnumDeclaration {
	// 1. Consume 'enum' keyword
	if _, ok := p.consume(token.ENUM, "Expected 'enum' keyword"); !ok {
		return nil
	}

	// 2. Consume Identifier (Enum Name)
	nameToken, ok := p.consume(token.IDENT, "Expected identifier for enum name")
	if !ok {
		// Stop parsing this node if we lack a name, but keep going structurally if possible
		return nil
	}

	// 3. Consume '{'
	if _, ok := p.consume(token.LBRACE, "Expected '{' after enum name"); !ok {
		return nil
	}

	// 4. Parse Members
	members := []*EnumMember{}

	// Loop until '}' or EOF
	for !p.check(token.RBRACE) && !p.isAtEnd() {
		member := p.parseMember()
		if member != nil {
			members = append(members, member)
		}

		// If we are not at '}' and didn't just consume a comma, we enforce comma usage
		if !p.check(token.RBRACE) {
			if _, ok := p.consume(token.COMMA, "Expected ',' after enum member"); !ok {
				// ERROR RECOVERY: If comma is missing, we record error but likely
				// want to continue if the next token looks like an identifier.
				p.synchronize()
			}
		} else {
			// Optional trailing comma handling
			if p.check(token.COMMA) {
				p.advance()
			}
		}
	}

	// 5. Consume '}'
	p.consume(token.RBRACE, "Expected '}' to close enum body")

	return &EnumDeclaration{
		Name:    nameToken.Literal,
		Members: members,
	}
}

// parseMember handles: MemberName [= Value]
func (p *Parser) parseMember() *EnumMember {
	// We expect an identifier
	nameToken, ok := p.consume(token.IDENT, "Expected enum member name")
	if !ok {
		// ERROR RECOVERY: If we don't find a name, the stream is likely garbage.
		// We sync to the next comma or brace.
		p.synchronize()
		return nil
	}

	member := &EnumMember{Name: nameToken.Literal}

	// Check for optional initializer '= Value'
	if p.match(token.ASSIGN) {
		if p.check(token.INT) || p.check(token.STRING) {
			val := p.advance()
			member.Value = val.Literal
		} else {
			// Report error but don't crash
			p.error(p.peek(), "Expected number or string initializer")
			// We don't necessarily need to sync hard here, just skip the bad token
			p.advance()
		}
	}

	return member
}

// --- Error Recovery & Helpers ---

// synchronize skips tokens until it finds a boundary (Comma or RBrace)
// This is "Panic Mode" recovery.
func (p *Parser) synchronize() {
	p.advance() // Skip the token that caused the error

	for !p.isAtEnd() {
		// If we find a comma or closing brace, we are back in a known state
		if p.peek().Type == token.COMMA || p.peek().Type == token.RBRACE {
			return
		}
		p.advance()
	}
}

// consume checks current token type. If match, advances. If not, records error.
func (p *Parser) consume(tt token.TokenType, errorMsg string) (token.Token, bool) {
	if p.check(tt) {
		return p.advance(), true
	}
	p.error(p.previous(), errorMsg)
	return token.Token{}, false
}

// error adds a diagnostic
func (p *Parser) error(token token.Token, message string) {
	p.Diagnostics = append(p.Diagnostics, Diagnostic{
		Line:    token.Line,
		Column:  token.Column,
		Message: message,
	})
}

// match consumes the token if it matches the type
func (p *Parser) match(types ...token.TokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) check(t token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}

func (p *Parser) peek() token.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() token.Token {
	return p.tokens[p.current-1]
}
