package parser

import (
	"fmt"
	"gloss/ast"
	"gloss/lexer"
	"gloss/token"
)

type Parser struct {
	lexer *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	Diagnostics DiagnosticList
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer: l,
	}
	p.nextToken()
	p.nextToken()
	return p
}

// Helpers
func (p *Parser) expectNext(t token.TokenType, msg string) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.Diagnostics.Error(p.curToken, msg)
	return false
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
	fmt.Println("next token is ", p.curToken.Literal)
}

func (p *Parser) parseDeclarations() ast.Node {
	switch p.curToken.Type {
	case token.FUNC:
		return p.parseFunc()
	default:
		return nil
	}
}

func (p *Parser) parseFunc() *ast.Func {
	fn := &ast.Func{}

	if !p.expectNext(token.IDENT, "Expected function name") {
		return nil
	}

	fn.Name = p.curToken.Literal

	if !p.expectNext(token.LPAREN, "Expected function parameters") {
		return nil
	}

	fn.Params = p.parseFuncParams()

	if p.peekToken.Type != token.LBRACE {
		// TODD: Should implement a parseType receiver for generics like Result<T>
		if p.expectNext(token.IDENT, "Expected return type") {
			fn.ReturnType = &ast.Type{
				Name: p.curToken.Literal,
			}
		} else {
			// TODO: Recover
			p.nextToken()
		}
	}

	if p.expectNext(token.LBRACE, "Expected '{'") {
		fn.Body = p.parseBlockStatement()
	} else {
		// TODO: Recover
	}

	return fn
}

func (p *Parser) parseFuncParams() []*ast.Parameter {
	// Handle functions with no arguments
	if p.peekToken.Type == token.RPAREN {
		return nil
	}

	var params []*ast.Parameter
	for {

		if !p.expectNext(token.IDENT, "Expected parameter name") {
			// TODO: recover
			continue
		}

		param := &ast.Parameter{
			Name: p.curToken.Literal,
		}

		if !p.expectNext(token.IDENT, "Expected parameter type") {
			// TODO: recover
			continue
		}

		// TODO: Assign default values
		param.Type = p.curToken.Literal
		params = append(params, param)

		if p.peekToken.Type != token.COMMA && p.peekToken.Type != token.RPAREN {
			p.Diagnostics.Error(p.peekToken, "Expected ',' or ')'")
			// TODO: recover
		}

		// Consume comma and keep processing parameters
		if p.peekToken.Type == token.COMMA {
			p.nextToken()
		}

		// We reached the end of the parameter list
		if p.peekToken.Type == token.RPAREN {
			p.nextToken()
			break
		}
	}

	return params
}

func (p *Parser) parseStatement() ast.Node {
	switch p.curToken.Type {
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		// TODO: Recover
		return nil
	}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	return &ast.ReturnStatement{
		// Value: p.parseExpression(),
	}
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{}

	for p.curToken.Type != token.EOF && p.curToken.Type != token.RBRACE {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	// TODO: should we always produce a block node
	if len(block.Statements) == 0 {
		return nil
	}

	return block
}

func (p *Parser) Parse() ast.SourceFile {
	file := ast.SourceFile{}

	for p.curToken.Type != token.EOF {
		fmt.Println(p.curToken.Type)
		decl := p.parseDeclarations()
		if decl != nil {
			file.Declarations = append(file.Declarations, decl)
		}
		p.nextToken()
	}

	return file
}
