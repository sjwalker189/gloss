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

	allowCompositeLiterals bool
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:                  l,
		Diagnostics:            &diagnostic.MessageList{},
		allowCompositeLiterals: true,
	}
	p.init()
	return p
}

func (p *Parser) init() {
	p.unaryExprParseFunc = map[token.TokenType]unaryExprParseFunc{
		token.BOOL:        p.parseBoolean,
		token.INT:         p.parseIntegerLiteral,
		token.STRING:      p.parseStringLiteral,
		token.IDENT:       p.parseIdent,
		token.TYPE_INT:    p.parseIdent,
		token.TYPE_STRING: p.parseIdent,
		token.TYPE_BOOL:   p.parseIdent,
		token.MINUS:       p.parseUnaryExpression,
		token.BANG:        p.parseUnaryExpression,
		token.LPAREN:      p.parseGroupedExpression,
		token.LBRACE:      p.parseAnonymousCompositeLiteral,
		token.LBRACKET:    p.parseArrayTypeExpression,
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
		token.LBRACE: p.parseCompositeLiteral,
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
		return p.parseStructDeclaration()
	case token.LET:
		return p.parseLetStatement()
	case token.CONST:
		return p.parseConstStatement()
	case token.FUNC:
		return p.parseFunc()
	default:
		return nil
	}
}

func (p *Parser) parseStatements() ast.Node {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.CONST:
		return p.parseConstStatement()
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

	fn.Name = &ast.Identifier{Name: p.curToken.Literal}

	if p.peekToken.Type == token.LANGLE {
		p.nextToken()
		fn.TypeParameters = p.parseTypeParameters()
		p.expectNext(token.RANGLE, "Expected >")
	}

	if !p.expectNext(token.LPAREN, "Expected '('") {
		return nil
	}
	fn.Parameters = p.parseFuncParams()

	if p.peekToken.Type != token.LBRACE {
		p.nextToken()
		fn.ReturnType = p.parseType()

		if p.peekToken.Type == token.RANGLE {
			p.nextToken()
		}
	}

	if !p.expectNext(token.LBRACE, "expected {") {
		return nil
	}

	fn.Body = p.parseBlockStatement()
	return fn
}

func (p *Parser) parseFuncParams() []*ast.Parameter {
	var params []*ast.Parameter

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return params
	}

	for {
		p.nextToken() // consume opening ( or previous ,

		param := &ast.Parameter{Name: &ast.Identifier{Name: p.curToken.Literal}}
		p.expectNext(token.COLON, "expected :")
		p.nextToken()

		param.Type = p.parseType()
		if param.Type == nil {
			p.Diagnostics.Error(p.curToken, "expected type")
		} else if p.peekToken.Type == token.RANGLE {
			p.nextToken()
		}

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

func (p *Parser) parseLoopStatement() ast.Iter {
	loop := &ast.Loop{}
	p.expectNext(token.LBRACE, "Expected '{'")
	loop.Body = p.parseBlockStatement()
	return loop
}

func (p *Parser) parseForStatement() ast.Iter {
	// 1. Infinite loop (same as loop {})
	if p.peekToken.Type == token.LBRACE {
		p.nextToken() // move to {
		return &ast.For{
			Body: p.parseBlockStatement(),
		}
	}

	if !p.expectNext(token.IDENT, "expected identifier") {
		return nil
	}

	firstIdent := &ast.Identifier{Name: p.curToken.Literal}

	if p.peekToken.Type == token.COMMA {
		p.nextToken() // consume comma
		if !p.expectNext(token.IDENT, "Expected second identifier") {
			return nil
		}
		secondIdent := &ast.Identifier{Name: p.curToken.Literal}

		loop := &ast.ForEach{
			Key:   firstIdent,
			Value: secondIdent,
		}

		if !p.expectNext(token.IN, "expected 'in' keyword") {
			return nil
		}

		p.nextToken()
		p.allowCompositeLiterals = false
		exp := p.parseExpression(LOWEST)
		p.allowCompositeLiterals = true
		if exp == nil {
			p.Diagnostics.Error(p.curToken, "Expected expression to follow")
			return nil
		}

		loop.Iterable = exp
		p.expectNext(token.LBRACE, "expected {")
		loop.Body = p.parseBlockStatement()
		return loop
	}

	loop := &ast.For{
		Init: p.parseLetStatement(),
	}

	if loop.Init == nil {
		p.Diagnostics.Error(p.curToken, "expected initializer")
		return nil
	}

	if !p.expectNext(token.SEMICOLON, "expected ;") {
		return nil
	}

	p.allowCompositeLiterals = false
	loop.Condition = p.parseExpression(LOWEST)
	p.allowCompositeLiterals = true
	if loop.Condition == nil {
		p.Diagnostics.Error(p.curToken, "expected initializer")
		return nil
	}

	if !p.expectNext(token.SEMICOLON, "expected ;") {
		return nil
	}

	loop.Post = p.parseExpression(LOWEST)
	if loop.Post == nil {
		p.Diagnostics.Error(p.curToken, "expected post expression")
		return nil
	}

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

func (p *Parser) parseStructDeclaration() *ast.StructDeclaration {
	u := &ast.StructDeclaration{}

	p.expectNext(token.IDENT, "Expected name")
	u.Name = &ast.Identifier{Name: p.curToken.Literal}

	if p.peekToken.Type == token.LANGLE {
		p.nextToken()
		u.TypeParameters = p.parseTypeParameters()
		if p.peekToken.Type == token.RANGLE {
			p.nextToken()
		}
	}

	p.expectNext(token.LBRACE, "Expected '{'")

	for p.peekToken.Type != token.RBRACE && p.peekToken.Type != token.EOF {
		p.nextToken()

		f := &ast.FieldDeclaration{Name: &ast.Identifier{Name: p.curToken.Literal}}
		p.expectNext(token.COLON, "Expected ':'")
		p.nextToken()

		f.Type = p.parseType()
		if p.peekToken.Type == token.RANGLE {
			p.nextToken()
		}

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
	case token.LBRACKET:
		if !p.expectNext(token.RBRACKET, "Expected ]") {
			// TODO: consider tuple syntax like: [string, int]
			return nil
		}
		array := &ast.ArrayType{}
		p.nextToken()
		array.ElementType = p.parseType()
		if array.ElementType == nil {
			// TODO: raise error
		}
		return array

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
		field := &ast.FieldDeclaration{Name: &ast.Identifier{Name: p.curToken.Literal}}
		p.expectNext(token.COLON, "Expected ':'")
		p.nextToken()
		field.Type = p.parseType()
		if p.peekToken.Type == token.RANGLE {
			p.nextToken()
		}
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

	p.allowCompositeLiterals = false
	stmt.Condition = p.parseExpression(LOWEST)
	p.allowCompositeLiterals = true

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
	prefix := p.unaryExprParseFunc[p.curToken.Type]
	if prefix == nil {
		return nil
	}
	leftExp := prefix()

	for p.peekToken.Type != token.SEMICOLON &&
		p.peekToken.Type != token.RBRACE &&
		!(p.peekToken.Type == token.LBRACE && !p.allowCompositeLiterals) &&
		precedence < p.peekPrecedence() {

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

func (p *Parser) parseCompositeLiteral(left ast.Expression) ast.Expression {
	lit := &ast.CompositeLiteral{
		Type: left,
	}

	p.nextToken() // move past {

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		element := p.parseCompositeElement()
		if element != nil {
			lit.Elements = append(lit.Elements, element)
		}

		if p.peekToken.Type == token.COMMA {
			p.nextToken()
		}
		p.nextToken() // This moves to the start of the next element or the '}'
	}

	return lit
}

func (p *Parser) parseCompositeElement() ast.Expression {
	val := p.parseExpression(LOWEST)

	if p.peekToken.Type == token.COLON {
		p.nextToken() // Move to ':'
		p.nextToken() // Move to start of value

		return &ast.KeyValuePair{
			Key:   val,
			Value: p.parseExpression(LOWEST),
		}
	}

	return val
}

func (p *Parser) parseAnonymousCompositeLiteral() ast.Expression {
	// This represents a literal where the type is inferred, e.g., { id: 1 }
	return p.parseCompositeLiteral(nil)
}

func (p *Parser) parseArrayTypeExpression() ast.Expression {
	// curToken is '['. Move to ']'
	// TODO: can pre-allocate size with int next token?
	if !p.expectNext(token.RBRACKET, "Expected ']'") {
		return nil
	}

	p.nextToken()

	arrayType := &ast.ArrayTypeExpression{
		// Use a high precedence so we don't accidentally
		// consume the '{' inside this prefix function
		BaseType: p.parseExpression(PREFIX),
	}

	// When composite literals are disallowed in the outer context (e.g. for-each
	// iterable), the infix loop won't consume '{'. Eagerly parse it here so that
	// []Type{...} literals still work in those positions.
	if !p.allowCompositeLiterals && p.peekToken.Type == token.LBRACE {
		p.nextToken() // move to '{'
		return p.parseCompositeLiteral(arrayType)
	}

	return arrayType
}
