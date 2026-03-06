package parser

import (
	"gloss/ast"
	"gloss/diagnostic"
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

	Diagnostics *diagnostic.MessageList

	prefixParseFns map[token.TokenType]prefixParseFunc
	infixParseFns  map[token.TokenType]infixParseFunc
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:       l,
		Diagnostics: &diagnostic.MessageList{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFunc)
	p.infixParseFns = make(map[token.TokenType]infixParseFunc)

	// Register Prefix Parsers
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.TYPE_IDENT, p.parseTypeIdentifierExpression) // Handles Struct Literals
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BOOL, p.parseBooleanLiteral)
	p.registerPrefix(token.BANG, p.parseUnaryExpression)
	p.registerPrefix(token.MINUS, p.parseUnaryExpression)
	p.registerPrefix(token.PLUS, p.parseUnaryExpression)
	p.registerPrefix(token.BITWISE_NOT, p.parseUnaryExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	
	p.registerPrefix(token.LBRACKET, p.parseSliceCompositeLiteral) // [ starts a slice literal
	p.registerPrefix(token.MATCH, p.parseMatchExpression)
	p.registerPrefix(token.FUNC, p.parseAnonymousFunction)
	p.registerPrefix(token.ELEMENT_OPEN_START, p.parseJSXElement) // <div
	p.registerPrefix(token.LT, p.parseJSXFragment)                // <>

	// Register Infix Parsers
	p.registerInfix(token.PLUS, p.parseBinaryExpression)
	p.registerInfix(token.MINUS, p.parseBinaryExpression)
	p.registerInfix(token.MUL, p.parseBinaryExpression)
	p.registerInfix(token.DIV, p.parseBinaryExpression)
	p.registerInfix(token.MOD, p.parseBinaryExpression)
	p.registerInfix(token.EQ, p.parseBinaryExpression)
	p.registerInfix(token.NOT_EQ, p.parseBinaryExpression)
	p.registerInfix(token.LT, p.parseBinaryExpression) // Also handles generic calls?
	p.registerInfix(token.LT_EQ, p.parseBinaryExpression)
	p.registerInfix(token.GT, p.parseBinaryExpression)
	p.registerInfix(token.GT_EQ, p.parseBinaryExpression)
	p.registerInfix(token.AND, p.parseBinaryExpression)
	p.registerInfix(token.OR, p.parseBinaryExpression)
	p.registerInfix(token.BITWISE_AND, p.parseBinaryExpression)
	p.registerInfix(token.BITWISE_OR, p.parseBinaryExpression)
	p.registerInfix(token.BITWISE_XOR, p.parseBinaryExpression)
	p.registerInfix(token.BITSHIFTL, p.parseBinaryExpression)
	p.registerInfix(token.BITSHIFTR, p.parseBinaryExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.PERIOD, p.parseMemberExpression)
	p.registerInfix(token.QUESTION_DOT, p.parseMemberExpression) // Optional chaining
	p.registerInfix(token.INC, p.parsePostfixExpression)
	p.registerInfix(token.DEC, p.parsePostfixExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFunc) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFunc) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	p.Diagnostics.Error(p.peekToken, "Expected "+string(t)+", got "+string(p.peekToken.Type))
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

// ----------------------------------------------------------------------------
// Parsing Entry Point
// ----------------------------------------------------------------------------

func (p *Parser) Parse() ast.SourceFile {
	file := ast.SourceFile{}
	file.StartLine = p.curToken.Line
	file.StartCol = p.curToken.Column

	for p.curToken.Type != token.EOF {
		decl := p.parseDeclaration()
		if decl != nil {
			file.Declarations = append(file.Declarations, decl)
		}
		p.nextToken()
	}

	return file
}

// ----------------------------------------------------------------------------
// Declarations
// ----------------------------------------------------------------------------

func (p *Parser) parseDeclaration() ast.Declaration {
	// Handle Visibility
	isPublic := false
	if p.curTokenIs(token.PUB) {
		isPublic = true
		p.nextToken()
	}

	// Handle Extern
	isExtern := false
	if p.curTokenIs(token.EXTERN) {
		isExtern = true
		p.nextToken()
	}

	switch p.curToken.Type {
	case token.USE:
		return p.parseUseDeclaration()
	case token.TYPE:
		return p.parseTypeDeclaration(isPublic, isExtern)
	case token.ENUM:
		return p.parseEnumDeclaration(isPublic, isExtern)
	case token.UNION:
		return p.parseUnionDeclaration(isPublic, isExtern)
	case token.STRUCT:
		return p.parseStructDeclaration(isPublic, isExtern)
	case token.FUNC:
		return p.parseFunctionDeclaration(isPublic)
	case token.CONST:
		return p.parseVariableDeclaration(isPublic, true)
	case token.LET:
		return p.parseVariableDeclaration(isPublic, false)
	default:
		p.Diagnostics.Error(p.curToken, "Unexpected token at top level: "+p.curToken.Literal)
		return nil
	}
}

func (p *Parser) parseUseDeclaration() *ast.UseDeclaration {
	stmt := &ast.UseDeclaration{}
	p.nextToken() // Skip USE
	if p.expectPeek(token.STRING) {
		stmt.Module = p.curToken.Literal
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseTypeDeclaration(isPub, isExtern bool) *ast.TypeDeclaration {
	decl := &ast.TypeDeclaration{IsPublic: isPub, IsExtern: isExtern}
	p.nextToken() // Skip TYPE

	if !p.curTokenIs(token.TYPE_IDENT) && !p.curTokenIs(token.IDENT) {
		return nil
	}
	decl.Name = p.curToken.Literal

	// Generics
	if p.peekTokenIs(token.LT) {
		decl.TypeParameters = p.parseTypeParameters()
	}

	p.nextToken()
	decl.Type = p.parseType()

	return decl
}

func (p *Parser) parseEnumDeclaration(isPub, isExtern bool) *ast.EnumDeclaration {
	decl := &ast.EnumDeclaration{IsPublic: isPub, IsExtern: isExtern}
	p.nextToken() // Skip ENUM

	if !p.curTokenIs(token.TYPE_IDENT) && !p.curTokenIs(token.IDENT) {
		return nil
	}
	decl.Name = p.curToken.Literal

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		variant := &ast.EnumVariant{Name: p.curToken.Literal}
		if p.peekTokenIs(token.ASSIGN) {
			p.nextToken() // =
			p.nextToken() // expr
			variant.Value = p.parseExpression(LOWEST)
		}
		decl.Variants = append(decl.Variants, variant)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	p.expectPeek(token.RBRACE)
	return decl
}

func (p *Parser) parseUnionDeclaration(isPub, isExtern bool) *ast.UnionDeclaration {
	decl := &ast.UnionDeclaration{IsPublic: isPub, IsExtern: isExtern}
	p.nextToken() // Skip UNION

	if !p.curTokenIs(token.TYPE_IDENT) && !p.curTokenIs(token.IDENT) {
		return nil
	}
	decl.Name = p.curToken.Literal
	if p.peekTokenIs(token.LT) {
		decl.TypeParameters = p.parseTypeParameters()
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		variant := &ast.UnionVariant{Name: p.curToken.Literal}
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken() // (
			p.nextToken() // Type
			variant.Payload = p.parseType()
			p.expectPeek(token.RPAREN)
		}
		decl.Variants = append(decl.Variants, variant)
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	p.expectPeek(token.RBRACE)
	return decl
}

func (p *Parser) parseStructDeclaration(isPub, isExtern bool) *ast.StructDeclaration {
	decl := &ast.StructDeclaration{IsPublic: isPub, IsExtern: isExtern}
	p.nextToken() // Skip STRUCT

	if !p.curTokenIs(token.TYPE_IDENT) && !p.curTokenIs(token.IDENT) {
		return nil
	}
	decl.Name = p.curToken.Literal
	if p.peekTokenIs(token.LT) {
		decl.TypeParameters = p.parseTypeParameters()
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		// Field or Method?
		if p.curTokenIs(token.FUNC) {
			// Method
			method := p.parseFunctionDeclaration(false)
			decl.Methods = append(decl.Methods, method)
		} else {
			// Field
			field := &ast.StructField{Name: p.curToken.Literal}
			if !p.expectPeek(token.COLON) {
				return nil
			}
			p.nextToken()
			field.Type = p.parseType()
			decl.Fields = append(decl.Fields, field)
		}

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	p.expectPeek(token.RBRACE)
	return decl
}

func (p *Parser) parseFunctionDeclaration(isPub bool) *ast.FunctionDeclaration {
	fn := &ast.FunctionDeclaration{IsPublic: isPub}
	// FUNC token already consumed if from declaration loop?
	// parseDeclaration calls p.parseFunctionDeclaration(isPublic) when p.curToken is FUNC.
	// So we are AT FUNC.
	
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	fn.Name = p.curToken.Literal

	if p.peekTokenIs(token.LT) {
		fn.TypeParameters = p.parseTypeParameters()
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	fn.Parameters = p.parseParameters()

	// Return Type
	if !p.peekTokenIs(token.LBRACE) && !p.peekTokenIs(token.ASSIGN) && !p.peekTokenIs(token.ARROW) {
		p.nextToken()
		fn.ReturnType = p.parseType()
	}

	// Body
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		fn.Body = p.parseBlockStatement()
	}

	return fn
}

func (p *Parser) parseVariableDeclaration(isPub, isConst bool) *ast.VariableDeclaration {
	decl := &ast.VariableDeclaration{IsPublic: isPub, IsConst: isConst}
	p.nextToken() // Skip LET/CONST

	if !p.curTokenIs(token.IDENT) {
		return nil
	}
	decl.Name = p.curToken.Literal

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		p.nextToken()
		decl.Type = p.parseType()
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	decl.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return decl
}

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

func (p *Parser) parseType() ast.Type {
	switch p.curToken.Type {
	case token.IDENT:
		if isPrimitive(p.curToken.Literal) {
			return &ast.PrimitiveType{Name: p.curToken.Literal}
		}
		name := p.curToken.Literal
		if p.peekTokenIs(token.LT) {
			return p.parseGenericType(name)
		}
		return &ast.TypeIdentifier{Name: name}

	case token.TYPE_IDENT:
		name := p.curToken.Literal
		if p.peekTokenIs(token.LT) {
			return p.parseGenericType(name)
		}
		return &ast.TypeIdentifier{Name: name}

	case token.TYPE_INT, token.TYPE_BOOL, token.TYPE_STRING, token.TYPE_VOID:
		return &ast.PrimitiveType{Name: p.curToken.Literal}

	case token.LBRACKET: // Slice [N]T or []T
		slice := &ast.SliceType{}
		p.nextToken()
		// Optional Size
		if !p.curTokenIs(token.RBRACKET) {
			if p.curTokenIs(token.INT) {
				val, _ := strconv.ParseInt(p.curToken.Literal, 0, 64)
				slice.Size = &ast.IntegerLiteral{Value: val}
				p.nextToken()
			}
		}
		if p.curToken.Type != token.RBRACKET {
			// Error
		}
		p.nextToken() // Read Element Type
		slice.Element = p.parseType()
		return slice

	case token.LPAREN: // Tuple (T, T)
		tuple := &ast.TupleType{}
		p.nextToken()
		for !p.curTokenIs(token.RPAREN) {
			tuple.Elements = append(tuple.Elements, p.parseType())
			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
				p.nextToken()
			} else {
				p.nextToken()
			}
		}
		return tuple
	}
	return nil
}

func (p *Parser) parseGenericType(name string) *ast.GenericType {
	gt := &ast.GenericType{Name: name}
	p.nextToken() // <
	p.nextToken() // First type
	for {
		gt.TypeArguments = append(gt.TypeArguments, p.parseType())
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			break
		}
	}
	p.expectPeek(token.GT)
	return gt
}

func (p *Parser) parseTypeParameters() []*ast.TypeParameter {
	var params []*ast.TypeParameter
	p.nextToken() // Eat <
	for !p.curTokenIs(token.GT) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.IDENT) || p.curTokenIs(token.TYPE_IDENT) {
			params = append(params, &ast.TypeParameter{Name: p.curToken.Literal})
		}
		p.nextToken()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	return params
}

func (p *Parser) parseParameters() []*ast.Parameter {
	var params []*ast.Parameter
	p.nextToken() // Eat (
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.IDENT) {
			param := &ast.Parameter{Name: p.curToken.Literal}
			if p.peekTokenIs(token.COLON) {
				p.nextToken()
				p.nextToken()
				param.Type = p.parseType()
			}
			params = append(params, param)
		}
		p.nextToken()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	return params
}

func isPrimitive(s string) bool {
	return s == "int" || s == "bool" || s == "string" || s == "void" || s == "nil"
}

// ----------------------------------------------------------------------------
// Statements
// ----------------------------------------------------------------------------

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET, token.CONST:
		return p.parseVariableDeclaration(false, p.curToken.Type == token.CONST)
	case token.RETURN:
		return p.parseReturnStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.LOOP:
		return p.parseLoopStatement()
	case token.BREAK:
		return &ast.BreakStatement{}
	case token.CONTINUE:
		return &ast.ContinueStatement{}
	case token.LBRACE:
		return p.parseBlockStatement()
	default:
		expr := p.parseExpression(LOWEST)
		
		if p.peekTokenIs(token.ASSIGN) || p.peekTokenIs(token.PLUS_ASSIGN) || p.peekTokenIs(token.MINUS_ASSIGN) {
			return p.parseAssignmentStatement(expr)
		}
		
		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		return &ast.ExpressionStatement{Expression: expr}
	}
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{}
	p.nextToken() // {
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{}
	if !p.peekTokenIs(token.SEMICOLON) && !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseAssignmentStatement(left ast.Expression) *ast.AssignmentStatement {
	stmt := &ast.AssignmentStatement{Left: left}
	p.nextToken() // Operator
	stmt.Operator = p.curToken.Literal
	p.nextToken() // Right
	stmt.Right = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if p.peekTokenIs(token.IF) {
			p.nextToken()
			stmt.Alternative = p.parseIfStatement()
		} else {
			if !p.expectPeek(token.LBRACE) {
				return nil
			}
			stmt.Alternative = p.parseBlockStatement()
		}
	}
	return stmt
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseLoopStatement() *ast.LoopStatement {
	stmt := &ast.LoopStatement{}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseForStatement() ast.Statement {
	p.nextToken() // Skip FOR
	
	// Basic implementation assuming C-style `for init; cond; update {` for now
	// Or `for x in y {`
	
	stmt := &ast.ForStatement{}
	
	// Check for Variable Declaration
	if p.curTokenIs(token.LET) || p.curTokenIs(token.CONST) {
		stmt.Initializer = p.parseVariableDeclaration(false, p.curTokenIs(token.CONST))
		p.nextToken()
		stmt.Condition = p.parseExpression(LOWEST)
		p.expectPeek(token.SEMICOLON)
		p.nextToken()
		stmt.Update = p.parseExpression(LOWEST)
		p.expectPeek(token.LBRACE)
		stmt.Body = p.parseBlockStatement()
		return stmt
	}
	
	return stmt
}

// ----------------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------------

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.Diagnostics.Error(p.curToken, "No prefix parse function for "+p.curToken.Literal)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Name: p.curToken.Literal}
}

func (p *Parser) parseTypeIdentifierExpression() ast.Expression {
	name := p.curToken.Literal
	if p.peekTokenIs(token.LBRACE) {
		return p.parseCompositeLiteral(name)
	}
	return &ast.Identifier{Name: name}
}

func (p *Parser) parseCompositeLiteral(typeName string) ast.Expression {
	lit := &ast.CompositeLiteral{Type: &ast.TypeIdentifier{Name: typeName}}
	p.nextToken() // Move to {
	p.parseCompositeLiteralBody(lit)
	return lit
}

func (p *Parser) parseSliceCompositeLiteral() ast.Expression {
	// [ ... 
	t := p.parseType() // Consumes [ and ] and type
	if t == nil { return nil }
	
	lit := &ast.CompositeLiteral{Type: t}
	
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		p.parseCompositeLiteralBody(lit)
		return lit
	}
	// Fallback if just type? (e.g. `[10]int` in expression?)
	// Not typical expression. Return nil?
	return nil
}

func (p *Parser) parseCompositeLiteralBody(lit *ast.CompositeLiteral) {
	// Expect {
	if !p.curTokenIs(token.LBRACE) { return }

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		if p.peekTokenIs(token.RBRACE) { break }
		
		p.nextToken()
		
		// Lookahead for Key: Value
		if p.peekTokenIs(token.COLON) {
			name := p.curToken.Literal
			p.nextToken() // :
			p.nextToken() // Val
			val := p.parseExpression(LOWEST)
			lit.Fields = append(lit.Fields, &ast.FieldValue{Name: name, Value: val})
		} else {
			val := p.parseExpression(LOWEST)
			lit.Values = append(lit.Values, val)
		}
		
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	p.expectPeek(token.RBRACE)
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	val, _ := strconv.ParseInt(p.curToken.Literal, 0, 64)
	return &ast.IntegerLiteral{Value: val}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Value: p.curToken.Literal == "true"}
}

func (p *Parser) parseUnaryExpression() ast.Expression {
	exp := &ast.UnaryExpression{Operator: p.curToken.Literal}
	p.nextToken()
	exp.Operand = p.parseExpression(UNARY)
	return exp
}

func (p *Parser) parseBinaryExpression(left ast.Expression) ast.Expression {
	exp := &ast.BinaryExpression{Left: left, Operator: p.curToken.Literal}
	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)
	return exp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	p.expectPeek(token.RPAREN)
	return &ast.ParenExpression{Expression: exp}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	var list []ast.Expression
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	p.expectPeek(end)
	return list
}

func (p *Parser) parseMemberExpression(object ast.Expression) ast.Expression {
	exp := &ast.MemberExpression{Object: object}
	exp.Optional = p.curToken.Type == token.QUESTION_DOT
	p.nextToken()
	exp.Property = p.curToken.Literal
	return exp
}

func (p *Parser) parseIndexExpression(operand ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Operand: operand}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	p.expectPeek(token.RBRACKET)
	return exp
}

func (p *Parser) parseMatchExpression() ast.Expression {
	exp := &ast.MatchExpression{}
	p.nextToken() // Match
	exp.Value = p.parseExpression(LOWEST)
	p.expectPeek(token.LBRACE)
	
	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		arm := &ast.MatchArm{}
		arm.Pattern = p.parsePattern()
		if !p.expectPeek(token.ARROW) { return nil }
		p.nextToken()
		arm.Value = p.parseExpression(LOWEST)
		exp.Arms = append(exp.Arms, arm)
		
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	p.expectPeek(token.RBRACE)
	return exp
}

func (p *Parser) parsePattern() ast.Pattern {
	if p.curTokenIs(token.IDENT) {
		return &ast.IdentifierPattern{Name: p.curToken.Literal}
	}
	return &ast.WildcardPattern{}
}

func (p *Parser) parseAnonymousFunction() ast.Expression {
	fn := &ast.AnonymousFunction{}
	if p.peekTokenIs(token.LT) {
		fn.TypeParameters = p.parseTypeParameters()
	}
	p.expectPeek(token.LPAREN)
	fn.Parameters = p.parseParameters()
	
	if !p.peekTokenIs(token.LBRACE) && !p.peekTokenIs(token.ARROW) {
		p.nextToken()
		fn.ReturnType = p.parseType()
	}
	
	if p.peekTokenIs(token.ARROW) {
		p.nextToken()
		p.nextToken()
		fn.Body = p.parseExpression(LOWEST)
	} else {
		p.nextToken()
		fn.Body = p.parseBlockStatement()
	}
	return fn
}

func (p *Parser) parseJSXElement() ast.Expression {
	el := &ast.JSXElement{}
	open := &ast.JSXOpeningElement{}
	
	// Name
	p.nextToken() // assume name (or check)
	open.Name = &ast.Identifier{Name: p.curToken.Literal}
	
	// Attributes
	for !p.curTokenIs(token.ELEMENT_OPEN_END) && !p.curTokenIs(token.ELEMENT_VOID_END) && !p.curTokenIs(token.EOF) {
		p.nextToken()
		if p.curTokenIs(token.ELEMENT_ATTR) {
			attr := &ast.JSXAttribute{Name: p.curToken.Literal}
			if p.peekTokenIs(token.ASSIGN) {
				p.nextToken()
				p.nextToken()
				if p.curTokenIs(token.LBRACE) {
					attr.Value = p.parseJSXExpression()
				} else {
					attr.Value = p.parseStringLiteral()
				}
			}
			open.Attributes = append(open.Attributes, attr)
		}
	}
	
	el.OpeningElement = open
	
	if p.curTokenIs(token.ELEMENT_VOID_END) {
		return &ast.JSXSelfClosingElement{Name: open.Name, Attributes: open.Attributes}
	}
	
	// Children
	for !p.peekTokenIs(token.ELEMENT_CLOSE_START) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		if p.curTokenIs(token.ELEMENT_TEXT) {
			el.Children = append(el.Children, &ast.JSXText{Value: p.curToken.Literal})
		} else if p.curTokenIs(token.ELEMENT_OPEN_START) {
			el.Children = append(el.Children, p.parseJSXElement())
		} else if p.curTokenIs(token.LBRACE) {
			el.Children = append(el.Children, p.parseJSXExpression())
		}
	}
	
	p.expectPeek(token.ELEMENT_CLOSE_START)
	p.nextToken() // Name
	p.expectPeek(token.ELEMENT_CLOSE_END)
	
	return el
}

func (p *Parser) parseJSXFragment() ast.Expression {
	if !p.peekTokenIs(token.GT) { return nil }
	p.nextToken() // >
	frag := &ast.JSXFragment{}
	// Children...
	return frag
}

func (p *Parser) parseJSXExpression() ast.Expression {
	expr := p.parseExpression(LOWEST)
	p.expectPeek(token.RBRACE)
	return &ast.JSXExpression{Expression: expr}
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	return &ast.UpdateExpression{Operand: left, Operator: p.curToken.Literal, Post: true}
}
