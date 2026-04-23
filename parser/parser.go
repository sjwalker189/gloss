package parser

import (
	"fmt"
	"gloss/ast"
	"gloss/diagnostic"
	"gloss/lexer"
	"gloss/token"
	"strconv"
)

func dump(v any) {
	fmt.Printf("%+v\n", v)
}

type (
	unaryExprParseFunc  func() ast.Expression
	binaryExprParseFunc func(ast.Expression) ast.Expression
)

type Parser struct {
	lexer *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	Diagnostics *diagnostic.MessageList

	unaryExprParseFunc  map[token.TokenType]unaryExprParseFunc
	binaryExprParseFunc map[token.TokenType]binaryExprParseFunc

	typeDepth int // track generic type depth
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:       l,
		Diagnostics: &diagnostic.MessageList{},
	}
	p.init()
	return p
}

func (p *Parser) init() {
	p.unaryExprParseFunc = map[token.TokenType]unaryExprParseFunc{
		token.BOOL:   p.parseBoolean,
		token.INT:    p.parseIntegerLiteral,
		token.STRING: p.parseStringLiteral,
		token.IDENT:  p.parseIdent,
		token.MINUS:  p.parseUnaryExpression,
		token.BANG:   p.parseUnaryExpression,
		token.LPAREN: p.parseGroupedExpression,
	}

	p.binaryExprParseFunc = map[token.TokenType]binaryExprParseFunc{
		// Mathmatical
		token.PLUS:  p.parseBinaryExpression,
		token.MINUS: p.parseBinaryExpression,
		token.MUL:   p.parseBinaryExpression,
		token.DIV:   p.parseBinaryExpression,
		token.MOD:   p.parseBinaryExpression,

		// Equality
		token.EQ:     p.parseBinaryExpression,
		token.NOT_EQ: p.parseBinaryExpression,

		// Comparison
		token.LANGLE: p.parseBinaryExpression,
		token.RANGLE: p.parseBinaryExpression,
		token.LT:     p.parseBinaryExpression,
		token.GT:     p.parseBinaryExpression,
		token.LT_EQ:  p.parseBinaryExpression,
		token.GT_EQ:  p.parseBinaryExpression,

		// Boolean
		token.AND: p.parseBinaryExpression,
		token.OR:  p.parseBinaryExpression,

		// Bitwise
		token.BITWISE_AND: p.parseBinaryExpression,
		token.BITWISE_OR:  p.parseBinaryExpression,
		token.BITSHIFTL:   p.parseBinaryExpression,
		token.BITSHIFTR:   p.parseBinaryExpression,

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
	case token.CONST:
		return p.parseConstStatement()
	case token.FUNC:
		return p.parseFunc()
	// TODO: Following only present to support testing, should move to parseStatements only
	case token.IF:
		return p.parseIfStatement()
	case token.LOOP:
		return p.parseLoopStatement()
	case token.FOR:
		return p.parseForStatement()
	default:
		// TODO: Raise error
		return nil
	}
}

func (p *Parser) parseStatements() ast.Node {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.LOOP:
		return p.parseLoopStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.BREAK:
		return &ast.BreakStatement{}
	case token.CONTINUE:
		return &ast.ContinueStatement{}
	default:
		// TODO: Raise error
		return nil
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{}
	if !p.expectNext(token.IDENT, "Expected name") {
		return nil
	}
	stmt.Name = &ast.Identifier{Name: p.curToken.Literal}

	if p.peekToken.Type == token.COLON {
		p.nextToken()
		p.nextToken()
		stmt.Type = p.parseType()

		if p.peekToken.Type == token.RANGLE {
			p.nextToken()
		}
	}

	if p.peekToken.Type != token.ASSIGN {
		return stmt
	}

	p.nextToken()
	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{}
	if !p.expectNext(token.IDENT, "Expected name") {
		return nil
	}

	stmt.Name = &ast.Identifier{Name: p.curToken.Literal}

	if p.peekToken.Type == token.COLON {
		p.nextToken()
		p.nextToken()
		stmt.Type = p.parseType()

		if p.peekToken.Type == token.RANGLE {
			p.nextToken()
		}
	}

	p.expectNext(token.ASSIGN, "Expected =")
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
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
	p.nextToken()

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		if stmt := p.parseStatements(); stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseLoopStatement() *ast.Loop {
	loop := &ast.Loop{}
	p.expectNext(token.LBRACE, "Expected '{'")
	loop.Body = p.parseBlockStatement()
	return loop
}

func (p *Parser) parseForStatement() *ast.For {
	loop := &ast.For{}
	p.nextToken()
	loop.Condition = p.parseExpression(LOWEST)
	p.expectNext(token.LBRACE, "Expected '{'")
	loop.Body = p.parseBlockStatement()
	return loop
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
	default:
		// TODO: Should this be an exhaustive case?
		t := &ast.TypeIdentifier{
			Name: &ast.Identifier{Name: p.curToken.Literal},
		}

		if p.peekToken.Type == token.LANGLE {
			depth := p.typeDepth
			p.nextToken()

			t.Parameters = p.parseTypeParameters()

			for {
				if p.typeDepth == depth {
					break
				}
				if p.typeDepth > 2 {
					switch p.peekToken.Type {
					case token.BITSHIFTR:
						p.nextToken()
						p.typeDepth -= 2
					case token.RANGLE:
						p.nextToken()
						p.typeDepth--
					default:
						p.Diagnostics.Error(p.peekToken, "Expected >")
					}
				}
			}
		}

		return t
	}
}

func (p *Parser) parseTypeParameters() []ast.Type {
	var params []ast.Type

	for p.peekToken.Type != token.RANGLE &&
		p.peekToken.Type != token.BITSHIFTR && // Allow >> here
		p.peekToken.Type != token.EOF {

		p.nextToken()
		typ := p.parseType()
		if typ != nil {
			params = append(params, typ)
		}

		if p.peekToken.Type == token.COMMA {
			p.nextToken()
		}
	}

	// Logic to consume the closing bracket
	if p.peekToken.Type == token.BITSHIFTR {
		// "Split" the >> by transforming it into a >
		// and leaving the other > for the next caller
		p.peekToken.Type = token.RANGLE
		p.peekToken.Literal = ">"
		// We don't call nextToken() yet because the parent
		// parseType call will consume this "new" RANGLE
	}

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

// Control flow

func (p *Parser) parseIfStatement() *ast.If {
	stmt := &ast.If{}
	p.nextToken()

	stmt.Condition = p.parseExpression(LOWEST)

	p.expectNext(token.LBRACE, "Expected '{'")
	stmt.Then = p.parseBlockStatement()

	if p.peekToken.Type == token.ELSE {
		p.nextToken()

		if p.peekToken.Type == token.IF {
			p.nextToken()
			stmt.Else = p.parseIfStatement()
		} else {
			p.expectNext(token.LBRACE, "Expected '{'")
			stmt.Else = p.parseBlockStatement()
		}
	}

	return stmt
}

// Expressions (Pratt)

func (p *Parser) parseExpression(precedence int) ast.Expression {
	fmt.Printf("DEBUG: Starting parseExpression. curToken: %s (%s)\n", p.curToken.Literal, p.curToken.Type)
	prefix := p.unaryExprParseFunc[p.curToken.Type]
	if prefix == nil {
		fmt.Printf("DEBUG: No prefix function found for %s\n", p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for p.peekToken.Type != token.SEMICOLON && p.peekToken.Type != token.RBRACE && precedence < p.peekPrecedence() {
		infix := p.binaryExprParseFunc[p.peekToken.Type]
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

func (p *Parser) parseUnaryExpression() ast.Expression {
	expr := &ast.UnaryExpression{Operator: p.curToken.Literal}
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
