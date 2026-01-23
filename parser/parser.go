package parser

import (
	"gloss/ast"
	"gloss/lexer"
	"gloss/token"
	"strconv"
)

type (
	prefixExpParseFunc func() ast.Expression
	binaryExpParseFunc func(ast.Expression) ast.Expression
)

type Parser struct {
	lexer *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	Diagnostics *DiagnosticList

	prefixParseFunc map[token.TokenType]prefixExpParseFunc
	infixParseFunc  map[token.TokenType]binaryExpParseFunc
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:       l,
		Diagnostics: &DiagnosticList{},
	}
	p.init()
	return p
}

func (p *Parser) init() {
	p.prefixParseFunc = map[token.TokenType]prefixExpParseFunc{
		token.BOOL:   p.parseBoolean,
		token.INT:    p.parseIntegerLiteral,
		token.STRING: p.parseStringLiteral,
		token.IDENT:  p.parseIdent,
		token.MINUS:  p.parsePrefixExpression,
		token.BANG:   p.parsePrefixExpression,
		token.LPAREN: p.parseGroupedExpression,
	}

	p.infixParseFunc = map[token.TokenType]binaryExpParseFunc{
		token.PLUS:   p.parseBinaryExpression,
		token.MINUS:  p.parseBinaryExpression,
		token.MUL:    p.parseBinaryExpression,
		token.DIV:    p.parseBinaryExpression,
		token.MOD:    p.parseBinaryExpression,
		token.EQ:     p.parseBinaryExpression,
		token.NOT_EQ: p.parseBinaryExpression,
		token.LT:     p.parseBinaryExpression,
		token.GT:     p.parseBinaryExpression,
		token.AND:    p.parseBinaryExpression,
		token.LPAREN: p.parseCallExpression,
	}
	p.nextToken()
	p.nextToken()
}

// Helpers

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) expectNext(t token.TokenType, msg string) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.Diagnostics.Error(p.peekToken, msg)
	return false
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

// Top Level Parsing

func (p *Parser) Parse() ast.SourceFile {
	file := ast.SourceFile{}
	for p.curToken.Type != token.EOF {
		decl := p.parseDeclarations()
		if decl != nil {
			file.Declarations = append(file.Declarations, decl)
		}
		p.nextToken()
	}
	return file
}

func (p *Parser) parseDeclarations() ast.Node {
	switch p.curToken.Type {
	case token.ENUM:
		return p.parseEnum()
	case token.UNION:
		return p.parseUnion()
	case token.STRUCT:
		return p.parseStruct()
	case token.LET:
		return p.parseLetStatement()
	case token.FUNC:
		return p.parseFunc()
	default:
		return nil
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	let := &ast.LetStatement{}
	if !p.expectNext(token.IDENT, "Expected name") {
		return nil
	}
	let.Name = &ast.Identifier{Name: p.curToken.Literal}

	if !p.expectNext(token.ASSIGN, "Expected '='") {
		return nil
	}
	p.nextToken()
	let.Value = p.parseExpression(LOWEST)
	return let
}

func (p *Parser) parseFunc() *ast.Func {
	fn := &ast.Func{}
	if !p.expectNext(token.IDENT, "Expected name") {
		return nil
	}
	fn.Name = p.curToken.Literal

	if p.peekToken.Type == token.LANGLE {
		p.nextToken()
		fn.TypeParams = p.parseTypeParameters()
	}

	if !p.expectNext(token.LPAREN, "Expected '('") {
		return nil
	}
	fn.Params = p.parseFuncParams()

	if p.peekToken.Type != token.LBRACE {
		p.nextToken()
		fn.ReturnType = p.parseType()
	}

	if p.peekToken.Type == token.LBRACE {
		p.nextToken()
		fn.Body = p.parseBlockStatement()
	}
	return fn
}

func (p *Parser) parseFuncParams() []*ast.Parameter {
	var params []*ast.Parameter
	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return params
	}
	for {
		p.nextToken()
		param := &ast.Parameter{Name: p.curToken.Literal}
		p.nextToken()
		param.Type = p.parseType()
		params = append(params, param)
		if p.peekToken.Type != token.COMMA {
			break
		}
		p.nextToken()
	}
	p.expectNext(token.RPAREN, "Expected ')'")
	return params
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{}
	if p.peekToken.Type != token.RBRACE && p.peekToken.Type != token.SEMICOLON && p.peekToken.Type != token.EOF {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}
	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{}
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		switch p.curToken.Type {
		case token.RETURN:
			block.Statements = append(block.Statements, p.parseReturnStatement())
		case token.LET:
			block.Statements = append(block.Statements, p.parseLetStatement())
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseEnum() *ast.Enum {
	enum := &ast.Enum{}
	p.expectNext(token.IDENT, "Expected name")
	enum.Name = p.curToken.Literal
	p.expectNext(token.LBRACE, "Expected '{'")

	var curInt int64
	for p.peekToken.Type != token.RBRACE && p.peekToken.Type != token.EOF {
		p.nextToken()
		m := &ast.EnumMember{Name: p.curToken.Literal}
		if p.peekToken.Type == token.ASSIGN {
			p.nextToken()
			p.nextToken()
			m.Value = p.parseExpression(LOWEST)
			if itl, ok := m.Value.(*ast.IntegerLiteral); ok {
				curInt = itl.Value
			}
		}
		m.IntValue = curInt
		curInt++
		enum.Members = append(enum.Members, m)
		if p.peekToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	p.expectNext(token.RBRACE, "Expected '}'")
	return enum
}

func (p *Parser) parseUnion() *ast.Union {
	u := &ast.Union{}
	p.expectNext(token.IDENT, "Expected name")
	u.Name = p.curToken.Literal

	if p.peekToken.Type == token.LANGLE {
		p.nextToken()
		u.Parameters = p.parseTypeParameters()
	}

	p.expectNext(token.LBRACE, "Expected '{'")
	for p.peekToken.Type != token.RBRACE && p.peekToken.Type != token.EOF {
		p.nextToken()
		f := &ast.UnionField{Name: p.curToken.Literal}
		if p.peekToken.Type == token.LPAREN {
			p.nextToken()
			p.nextToken()
			f.Type = p.parseType()
			p.expectNext(token.RPAREN, "Expected ')'")
		}
		u.Fields = append(u.Fields, f)
		if p.peekToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	p.expectNext(token.RBRACE, "Expected '}'")
	return u
}

func (p *Parser) parseStruct() *ast.Struct {
	u := &ast.Struct{}
	p.expectNext(token.IDENT, "Expected name")
	u.Name = p.curToken.Literal

	if p.peekToken.Type == token.LANGLE {
		p.nextToken()
		u.Params = p.parseTypeParameters()
	}

	p.expectNext(token.LBRACE, "Expected '{'")

	for p.peekToken.Type != token.RBRACE && p.peekToken.Type != token.EOF {
		p.nextToken()

		f := &ast.StructField{Name: p.curToken.Literal}
		p.expectNext(token.COLON, "Expected ':'")
		p.nextToken()

		f.Type = p.parseType()
		u.Fields = append(u.Fields, f)

		if p.peekToken.Type == token.COMMA {
			p.nextToken()
		}
	}

	p.expectNext(token.RBRACE, "Expected '}'")
	return u
}

// Types

func (p *Parser) parseType() ast.Type {
	switch p.curToken.Type {
	case token.LBRACE:
		return p.parseStructBody()
	case token.TYPE_INT, token.TYPE_BOOL, token.TYPE_STRING:
		return &ast.TypeLiteral{Type: p.curToken.Literal}
	case token.IDENT:
		t := &ast.TypeIdentifier{Name: p.curToken.Literal}
		if p.peekToken.Type == token.LANGLE {
			p.nextToken()
			t.Parameters = p.parseTypeParameters()
		}
		return t
	default:
		return nil
	}
}

func (p *Parser) parseTypeParameters() []*ast.TypeParameter {
	var params []*ast.TypeParameter
	for p.peekToken.Type != token.RANGLE && p.peekToken.Type != token.EOF {
		p.nextToken()
		params = append(params, &ast.TypeParameter{Name: p.curToken.Literal})
		if p.peekToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	p.expectNext(token.RANGLE, "Expected '>'")
	return params
}

func (p *Parser) parseStructBody() *ast.StructBody {
	body := &ast.StructBody{}
	for p.peekToken.Type != token.RBRACE && p.peekToken.Type != token.EOF {
		p.nextToken()
		field := &ast.StructField{Name: p.curToken.Literal}
		p.expectNext(token.COLON, "Expected ':'")
		p.nextToken()
		field.Type = p.parseType()
		body.Fields = append(body.Fields, field)
		if p.peekToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	p.expectNext(token.RBRACE, "Expected '}'")
	return body
}

// Expressions (Pratt)

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFunc[p.curToken.Type]
	if prefix == nil {
		return nil
	}
	leftExp := prefix()

	for p.peekToken.Type != token.SEMICOLON && p.peekToken.Type != token.RBRACE && precedence < p.peekPrecedence() {
		infix := p.infixParseFunc[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseIdent() ast.Expression {
	return &ast.Identifier{Name: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	val, _ := strconv.ParseInt(p.curToken.Literal, 0, 64)
	return &ast.IntegerLiteral{Value: val}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Value: p.curToken.Literal[1 : len(p.curToken.Literal)-1]}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Value: p.curToken.Literal == "true"}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{Operator: p.curToken.Literal}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseBinaryExpression(left ast.Expression) ast.Expression {
	expr := &ast.BinaryExpression{Operator: p.curToken.Literal, Left: left}
	prec := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(prec)
	return expr
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	expr := p.parseExpression(LOWEST)
	p.expectNext(token.RPAREN, "Expected ')'")
	return &ast.ParenExpression{Expression: expr}
}

func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Function: fn, Arguments: []ast.Expression{}}
	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return exp
	}

	p.nextToken()

	exp.Arguments = append(exp.Arguments, p.parseExpression(LOWEST))
	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		exp.Arguments = append(exp.Arguments, p.parseExpression(LOWEST))
	}
	p.expectNext(token.RPAREN, "Expected ')'")
	return exp
}
