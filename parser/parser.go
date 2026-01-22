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
	return p
}

// Helpers

func (p *Parser) expect(t token.TokenType, msg string) bool {
	if p.curToken.Type == t {
		return true
	}
	p.Diagnostics.Error(p.curToken, msg)
	return false
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
	case token.UNION:
		return p.parseUnion()
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

	// Return types are optional, but if present it must be a type node
	if p.curToken.Type != token.LBRACE {
		fn.ReturnType = p.parseType()
		if fn.ReturnType == nil {
			p.Diagnostics.Error(p.curToken, "Expected type")
			// TODO: recover
			p.nextToken()
		}
	}

	if p.expect(token.LBRACE, "Expected '{'") {
		fn.Body = p.parseBlockStatement()
	} else {
		// TODO: Recover
	}

	return fn
}

func (p *Parser) parseFuncParams() []*ast.Parameter {
	if p.curToken.Type != token.LPAREN {
		panic(fmt.Sprintf("cannot parse func params. invalid token type: %s", p.curToken.Type))
	}

	// Handle functions with no arguments
	if p.peekToken.Type == token.RPAREN {
		p.nextToken() // eat (
		p.nextToken() // eat )
		return nil
	}

	var params []*ast.Parameter
	for p.curToken.Type != token.EOF && p.curToken.Type != token.RPAREN {
		if p.curToken.Type != token.IDENT {
			p.Diagnostics.Error(p.curToken, "Expected parameter name")
			// TODO: recover
			p.nextToken()
			continue
		}

		param := &ast.Parameter{
			Name: p.curToken.Literal,
		}
		p.nextToken()

		param.Type = p.parseType()
		if param.Type == nil {
			p.Diagnostics.Error(p.curToken, "Expected parameter type")
			// TODO: recover
			p.nextToken()
		}

		params = append(params, param)

		if p.curToken.Type != token.COMMA && p.curToken.Type != token.RPAREN {
			p.Diagnostics.Error(p.peekToken, "Expected ',' or ')'")
			// TODO: recover
			p.nextToken()
		}

		// Consume comma and keep processing parameters
		if p.curToken.Type == token.COMMA {
			p.nextToken()
		}

		// We reached the end of the parameter list
		if p.curToken.Type == token.RPAREN {
			break
		}
	}

	p.nextToken() // eat )

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
	if p.curToken.Type != token.LBRACE {
		panic("cannot parse block statement. invalid token")
	}

	block := &ast.BlockStatement{}

	p.nextToken() // eat {

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
	p.nextToken()

	if p.curToken.Type == token.LPAREN {
		union.Parameters = p.parseTypeParameters()
	}

	if !p.expect(token.LBRACE, "Expected '{'") {
		// TODO: Recover
		p.nextToken()
	}

	for p.curToken.Type != token.EOF && p.curToken.Type != token.RBRACE {
		if m := p.parseUnionField(); m != nil {
			union.Fields = append(union.Fields, m)
		}

		if p.curToken.Type == token.RBRACE {
			p.nextToken()
			break
		}

		if p.curToken.Type != token.COMMA {
			p.Diagnostics.Error(p.curToken, "Expected ','")
			// TODO: Recover
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

	switch p.curToken.Type {
	case token.COMMA:
		return m
	case token.LPAREN:
		p.nextToken() // eat (

		m.Type = p.parseType()
		// Should end on ')'
		if p.curToken.Type != token.RPAREN {
			p.Diagnostics.Error(p.curToken,
				"Expected ')'")
		}

		p.nextToken() // eat )
	default:
		// TODO: Recover
		p.nextToken()
	}

	return m
}

func (p *Parser) parseType() ast.Type {
	switch p.curToken.Type {
	case token.LBRACE:
		return p.parseStructBody()
	case token.TYPE_INT, token.TYPE_BOOL, token.TYPE_STRING:
		t := &ast.TypeLiteral{
			Type: p.curToken.Literal,
		}
		p.nextToken()
		return t
	case token.IDENT:
		t := &ast.TypeIdentifier{
			Name: p.curToken.Literal,
		}

		p.nextToken() // eat ident

		if p.curToken.Type == token.LANGLE {
			t.Parameters = p.parseTypeParameters()
		}

		return t
	}

	return nil
}

func (p *Parser) parseTypeParameters() []*ast.TypeParameter {
	p.nextToken() // eat (

	var params []*ast.TypeParameter

	for p.curToken.Type != token.EOF && p.curToken.Type != token.RPAREN {
		if !p.expect(token.IDENT, "Expected type parameter") {
			// TODO: Recover
			p.nextToken()
			continue
		}

		// TODO: Default values and constraints
		param := &ast.TypeParameter{Name: p.curToken.Literal}
		params = append(params, param)
		p.nextToken()

		if p.curToken.Type == token.RPAREN {
			break
		}

		if p.curToken.Type != token.COMMA {
			p.Diagnostics.Error(p.curToken, "Expected ','")
			// TODO: Recover
		}

		p.nextToken()
	}

	p.nextToken() // eat )

	if len(params) > 0 {
		return params
	}

	return nil
}

func (p *Parser) parseTypeLiteral() ast.Type {
	switch p.curToken.Type {
	case token.TYPE_INT, token.TYPE_BOOL, token.TYPE_STRING:
		return &ast.TypeLiteral{
			Type: p.curToken.Literal,
		}
	}
	return nil
}

func (p *Parser) parseStructBody() *ast.StructBody {
	if p.curToken.Type != token.LBRACE {
		panic("cannot parse struct body. invalid token")
	}

	body := &ast.StructBody{}

	p.nextToken() // eat {

	for p.curToken.Type != token.EOF && p.curToken.Type != token.RBRACE {
		if field := p.parseStructField(); field != nil {
			body.Fields = append(body.Fields, field)
		}

		if p.curToken.Type == token.RBRACE {
			break
		}

		if p.curToken.Type == token.COMMA {
			p.nextToken()
			continue
		}

		p.Diagnostics.Error(p.curToken, "Expected ',' or '}'")
	}

	p.nextToken() // eat }

	return body
}

func (p *Parser) parseStructField() *ast.StructField {
	if p.curToken.Type != token.IDENT {
		return nil
	}

	field := &ast.StructField{
		Name: p.curToken.Literal,
	}

	fmt.Println(p.curToken.Literal)
	p.nextToken() // eat ident

	fmt.Println(p.curToken.Literal)
	if p.curToken.Type != token.COLON {
		panic(fmt.Sprintf("%s not a colon\n", p.curToken.Literal))
	}

	p.nextToken() // eat :
	fmt.Println(p.curToken.Literal)

	field.Type = p.parseType()
	if field.Type == nil {
		p.Diagnostics.Error(p.curToken, "Expected type")
		return nil
	}

	return field
}

func (p *Parser) parseTupleType() *ast.TupleType {
	return nil
	// tuple := &ast.Tuple{}
	// p.nextToken()
	//
	// for p.curToken.Type != token.EOF && p.curToken.Type != token.RPAREN {
	// 	if t := p.parseType(); t != nil {
	// 		tuple.Fields = append(tuple.Fields, t)
	// 	}
	//
	// 	p.nextToken()
	// 	if p.curToken.Type == token.RPAREN {
	// 		break
	// 	}
	//
	// 	if p.curToken.Type == token.COMMA {
	// 		p.nextToken()
	// 		if p.curToken.Type == token.RPAREN {
	// 			// TODO: Support trailing commas in tuples?
	// 		}
	// 	}
	// }
	//
	// p.nextToken() // eat )
	//
	// return tuple
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

func (p *Parser) parseBinaryExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpression{
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
