package parser

import (
	"fmt"
	"gloss/ast"
	"gloss/lexer"
	"gloss/token"
	"strconv"
)

type (
	prefixParseFunc func() ast.Expression
	infixParseFunc  func(ast.Expression) ast.Expression
)

type Parser struct {
	lexer *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	Diagnostics DiagnosticList

	prefixParseFunc map[token.TokenType]prefixParseFunc
	infixParseFunc  map[token.TokenType]infixParseFunc
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer: l,
	}

	p.prefixParseFunc = map[token.TokenType]prefixParseFunc{
		token.BOOL:   p.parseBoolean,
		token.INT:    p.parseIntegerLiteral,
		token.STRING: p.parseStringLiteral,
		token.IDENT:  p.parseIdent,
		token.MINUS:  p.parsePrefixExpression,
		token.BANG:   p.parsePrefixExpression,
		token.LPAREN: p.parseGroupedExpression,
	}

	p.infixParseFunc = map[token.TokenType]infixParseFunc{
		token.PLUS:   p.parseInfixExpression,
		token.MINUS:  p.parseInfixExpression,
		token.MUL:    p.parseInfixExpression,
		token.DIV:    p.parseInfixExpression,
		token.MOD:    p.parseInfixExpression,
		token.EQ:     p.parseInfixExpression,
		token.NOT_EQ: p.parseInfixExpression,
		token.LT:     p.parseInfixExpression,
		token.GT:     p.parseInfixExpression,
		token.AND:    p.parseInfixExpression,
		token.LPAREN: p.parseCallExpression,
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

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) parseDeclarations() ast.Node {
	switch p.curToken.Type {
	case token.ENUM:
		return p.parseEnum()
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

	if !p.expectNext(token.IDENT, "Expected variable name") {
		return nil
	}

	let.Name = &ast.Identifier{
		Name: p.curToken.Literal,
	}

	if !p.expectNext(token.ASSIGN, "Expected '=") {
		return nil
	}

	p.nextToken() // Skip '='

	expr := p.parseExpression(LOWEST)
	if expr == nil {
		// TODO: Recover
		p.Diagnostics.Error(p.curToken, "Expected expression")
	}

	let.Value = expr
	return let
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
	stmt := &ast.ReturnStatement{}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	return stmt
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

func (p *Parser) parseEnum() *ast.Enum {
	enum := &ast.Enum{}

	if !p.expectNext(token.IDENT, "Expected name") {
		return nil
	}

	enum.Name = p.curToken.Literal

	if !p.expectNext(token.LBRACE, "Expected '{'") {
		// TODO: Recover
		p.nextToken()
	}

	var curIntValue int64

	for p.curToken.Type != token.EOF && p.curToken.Type != token.RBRACE {
		if m := p.parseEnumMember(); m != nil {
			integer, ok := m.Value.(*ast.IntegerLiteral)
			if ok {
				m.IntValue = integer.Value
				curIntValue = integer.Value
			} else {
				m.IntValue = curIntValue
			}
			enum.Members = append(enum.Members, m)
			curIntValue++
		}
		p.nextToken()
	}

	return enum
}

func (p *Parser) parseEnumMember() *ast.EnumMember {
	if p.curToken.Type != token.IDENT {
		return nil
	}

	m := &ast.EnumMember{
		Name: p.curToken.Literal,
	}

	if p.peekToken.Type == token.ASSIGN {
		p.nextToken()
		p.nextToken()
		switch p.curToken.Type {
		case token.STRING:
			m.Value = p.parseStringLiteral()
		case token.INT:
			m.Value = p.parseIntegerLiteral()
		default:
			// TODO: Recover
			p.Diagnostics.Error(p.curToken, "Expected an int or string value")
			p.nextToken()
		}
	}

	if !p.expectNext(token.COMMA, "Missing comma") {
		// TODO: Recover
		p.nextToken()
	}

	return m
}

func (p *Parser) parseUnion() *ast.Union {
	union := &ast.Union{}

	if !p.expectNext(token.IDENT, "Expected name") {
		return nil
	}

	union.Name = p.curToken.Literal

	if !p.expectNext(token.LBRACE, "Expected '{'") {
		// TODO: Recover
		p.nextToken()
	}

	for p.curToken.Type != token.EOF && p.curToken.Type != token.RBRACE {
		if m := p.parseUnionField(); m != nil {
			union.Fields = append(union.Fields, m)
		}
		p.nextToken()
	}

	return union
}

func (p *Parser) parseUnionField() *ast.UnionField {
	if p.curToken.Type != token.IDENT {
		return nil
	}

	m := &ast.UnionField{
		Name: p.curToken.Literal,
	}

	p.nextToken()

	switch p.peekToken.Type {
	case token.COMMA:
		return m
	case token.LBRACE:
		// m.Type = p.parseStructBody()
	case token.LPAREN:
		// m.Type = p.parseTuple()
	case token.IDENT, token.BOOL:
		// m.Type = p.parseType()
	default:
		// TODO: Recover
		p.nextToken()
	}

	if !p.expectNext(token.COMMA, "Missing comma") {
		// TODO: Recover
		p.nextToken()
	}

	return m
}

func (p *Parser) parseTupleType() *ast.TupleLiteral {
	tuple := &ast.Tuple{}
	p.nextToken()

	for p.curToken.Type != token.EOF && p.curToken.Type != token.RPAREN {
		if t := p.parseType(); t != nil {
			tuple.Fields = append(tuple.Fields, t)
		}

		p.nextToken()
		if p.curToken.Type == token.RPAREN {
			break
		}

		if p.curToken.Type == token.COMMA {
			p.nextToken()
			if p.curToken.Type == token.RPAREN {
				// TODO: Support trailing commas in tuples?
			}
		}
	}

	p.nextToken() // eat )

	return tuple
}

func (p *Parser) parseType() *ast.Type {
	// TODO: token.TYPE_BOOL meaning "bool" instead of BOOL which is literal "true" and "false"
	// TODO: need to lex more explict tokens for types
	// TODO: need to handle pointer types? or always wrap in Maybe<T>?
	if p.curToken.Type != token.IDENT {
		return nil
	}

	t := &ast.Type{
		Name: p.curToken.Literal,
	}

	switch p.curToken.Literal {
	case "bool", "int", "string":
		return t
	}

	// check for type paramters, for example: Either<A,B>
	p.nextToken()
	return t
}

func (p *Parser) parseIdent() ast.Expression {
	if p.curToken.Type != token.IDENT {
		panic(fmt.Sprintf("cannot parse token of type %s as identifier", p.curToken.Type))
	}
	return &ast.Identifier{
		Name: p.curToken.Literal,
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	if p.curToken.Type != token.INT {
		panic(fmt.Sprintf("cannot parse token of type %s as int", p.curToken.Type))
	}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		panic(err)
	}

	return &ast.IntegerLiteral{
		Value: value,
	}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	if p.curToken.Type != token.STRING {
		panic(fmt.Sprintf("cannot parse token of type %s as string", p.curToken.Type))
	}

	return &ast.StringLiteral{
		// Omit surrounding quotes
		Value: p.curToken.Literal[1 : len(p.curToken.Literal)-1],
	}
}

func (p *Parser) parseBoolean() ast.Expression {
	if p.curToken.Type != token.BOOL {
		panic(fmt.Sprintf("cannot parse token of type %s as boolean", p.curToken.Type))
	}
	switch p.curToken.Literal {
	case "true":
		return &ast.Boolean{Value: true}
	case "false":
		return &ast.Boolean{Value: false}
	default:
		panic(fmt.Sprintf("Boolean token does not hold a boolean value. Got '%s'", p.curToken.Literal))
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	if p.curToken.Type == token.EOF {
		p.Diagnostics.Error(p.curToken, "Unexpected EOF, expected expression")
		return nil
	}

	// 1. Parse the "Left" side (Prefix)
	// e.g., in "-1 + 2", this parses "-1"
	prefix := p.prefixParseFunc[p.curToken.Type]
	if prefix == nil {
		return nil
	}

	leftExp := prefix()

	// 2. Loop while the NEXT token has higher binding power (precedence)
	// e.g., if we are at "1" and see "+", we continue.
	// But if we are inside "1 * 2" and see "+", we stop (because * > +).
	for p.peekToken.Type != token.SEMICOLON && precedence < p.peekPrecedence() {
		infix := p.infixParseFunc[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken() // Move to the operator (+, *, etc.)

		// 3. Parse the "Right" side, passing the "Left" side in
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	// Recursively parse the right side
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	// Use PREFIX precedence so `-5 * 5` parses as `(-5) * 5`
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // eat "("

	exp := p.parseExpression(LOWEST)

	if !p.expectNext(token.RPAREN, "Expected ')'") {
		return nil
	}

	return &ast.ParenExpression{
		Expression: exp,
	}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{
		Function: function,
	}
	exp.Arguments = p.parseCallArguments()

	return exp
}

// Helper for comma-separated arguments
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	// Handle empty call "foo()"
	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return args
	}

	p.nextToken() // Eat '('

	// Parse first arg
	args = append(args, p.parseExpression(LOWEST))

	// Loop for remaining args
	for p.peekToken.Type == token.COMMA {
		p.nextToken() // Eat previous expr
		p.nextToken() // Eat comma
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectNext(token.RPAREN, "Expected ')'") {
		// TODO: Recover
	}
	if len(args) == 0 {
		return nil
	}

	return args
}

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
