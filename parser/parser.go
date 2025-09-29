package parser

import (
	"fmt"
	"strings"
)

// AST Node Types (constants matching TypeScript ASTNodeType enum)
const (
	Program int = 0 + iota
	VariableDeclaration
	Expression
	LiteralD
	IdentifierD
	NotExpression
	TernaryExpression
	BinaryExpression
	UnaryExpression
	FunctionDeclaration
	BlockStmt
	DefDecl
	Return
	Break
	Continue
	IfElse
	Loop
	CallExpression
	ArrayExpr
	ArrayIndex
	IncDec
	Error
)

// Parser represents the parser state
type Parser struct {
	tokenizer *TokenGen
	Nodes     []ASTNode
	Errors    []string
	vars      []Variable
	defines   map[string]string
}

// NewParser creates a new parser instance
func NewParser(file string) *Parser {
	return &Parser{
		tokenizer: NewTokenGen(file),
		Nodes:     []ASTNode{},
		Errors:    []string{},
		vars:      []Variable{},
		defines:   make(map[string]string),
	}
}

// addError adds an error message with current token context
func (p *Parser) addError(message string) {
	line := p.tokenizer.GetCurrentLineNumber()
	col := p.tokenizer.GetCurrentColNumber()
	currentToken := p.tokenizer.GetCurrentToken()

	// Get the actual source line
	var actualSourceLine string
	if line-1 < len(p.tokenizer.Lines) {
		actualSourceLine = p.tokenizer.Lines[line-1]
	} else {
		actualSourceLine = "(empty line)"
	}

	// Create a pointer to show where the error is
	pointer := strings.Repeat(" ", max(0, col-1)) + "^"

	errorMsg := fmt.Sprintf(`
Error at Line %d, Column %d: %s
%s
%s

Current token: "%s" (%d)`, line, col, message, actualSourceLine, pointer, currentToken.Value, currentToken.Type)

	p.Errors = append(p.Errors, errorMsg)
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// resolveDefine resolves defined aliases
func (p *Parser) resolveDefine(value string) string {
	if resolved, exists := p.defines[value]; exists {
		return resolved
	}
	return value
}

// consume returns current token and advances to next
func (p *Parser) consume() Token {
	token := p.tokenizer.GetCurrentToken()
	// Create a new token with resolved value if it's a define
	resolvedToken := Token{
		Type:  token.Type,
		Value: p.resolveDefine(token.Value),
		Line:  token.Line,
		Col:   token.Col,
	}
	p.tokenizer.Next()
	return resolvedToken
}

// expectPeek checks if next token matches expected type
func (p *Parser) expectPeek(t int) bool {
	pk := p.tokenizer.Peek(0)
	return pk.Type == t
}

// expectToken checks if current token matches expected type
func (p *Parser) expectToken(t int) bool {
	tk := p.tokenizer.GetCurrentToken()
	return tk.Type == t
}

// expectPeekVal checks if next token matches expected value
func (p *Parser) expectPeekVal(v string) bool {
	pk := p.tokenizer.Peek(0)
	resolvedValue := p.resolveDefine(pk.Value)
	return pk.Value == v || resolvedValue == v
}

// expectTokenVal checks if current token matches expected value
func (p *Parser) expectTokenVal(v string) bool {
	tk := p.tokenizer.GetCurrentToken()
	resolvedValue := p.resolveDefine(tk.Value)
	return tk.Value == v || resolvedValue == v
}

// consumeOptionalSemicolon consumes a semicolon if present (semicolons are optional in AY)
func (p *Parser) consumeOptionalSemicolon() {
	if p.expectTokenVal(";") {
		p.consume()
	}
}

// Start begins parsing the token stream
func (p *Parser) Start() {
	for p.tokenizer.GetCurrentToken().Type != EOF {
		// Skip newlines at the top level
		if p.expectToken(NewLine) {
			p.tokenizer.Next()
			continue
		}

		currentPos := p.tokenizer.CurrentTokenNo
		node := p.parseStatement()
		if node != nil {
			p.Nodes = append(p.Nodes, *node)
		}

		// Safety check: if we haven't advanced, force advance to prevent infinite loop
		if p.tokenizer.CurrentTokenNo == currentPos && p.tokenizer.GetCurrentToken().Type != EOF {
			fmt.Printf("[DEBUG] Force advancing from token: %+v\n", p.tokenizer.GetCurrentToken())
			p.tokenizer.Next()
		}
	}
}

// parseStatement parses a statement
func (p *Parser) parseStatement() *ASTNode {
	token := p.tokenizer.GetCurrentToken()

	// Handle define statements: def identifier -> value
	if p.expectTokenVal("def") {
		return p.parseDefine()
	}

	// Variable declaration: l identifier = expression
	if p.expectTokenVal("l") {
		return p.parseVariableDeclaration()
	}

	// Function declaration: f identifier(params) { body }
	if p.expectTokenVal("f") {
		return p.parseFunction()
	}

	// Return statement
	if p.expectTokenVal("return") {
		return p.parseReturn()
	}

	// Break statement
	if p.expectTokenVal("break") {
		p.consume()
		return &ASTNode{Type: Break}
	}

	// Continue statement
	if p.expectTokenVal("continue") {
		p.consume()
		return &ASTNode{Type: Continue}
	}

	// If statement
	if p.expectTokenVal("if") {
		return p.parseIfElse()
	}

	// For/while loop
	if p.expectTokenVal("for") || p.expectTokenVal("while") {
		return p.parseLoop()
	}

	// Function call or expression statement
	if p.expectToken(Identifier) {
		// Check if it's a function call
		if p.expectPeekVal("(") {
			node := p.parseCallExpr()
			if node != nil {
				// Consume optional semicolon after function call statement
				p.consumeOptionalSemicolon()
			}
			return node
		}
		// Check if it's an assignment or other expression
		node := p.parseExpression()
		if node != nil {
			// Consume optional semicolon after expression statement
			p.consumeOptionalSemicolon()
		}
		return node
	}
	// Skip semicolons at statement level (they're optional in AY)
	if p.expectTokenVal(";") {
		p.consume() // consume the semicolon
		return nil  // return nil since semicolon is not a statement
	}

	// Skip unknown tokens with error
	if token.Type != EOF {
		p.addError(fmt.Sprintf("Unexpected token: %s", token.Value))
		p.tokenizer.Next()
	}

	return nil
}

// parseDefine parses define statements (def identifier -> value)
func (p *Parser) parseDefine() *ASTNode {
	p.consume() // consume 'def'

	if !p.expectToken(Identifier) {
		p.addError("Expected identifier after 'def'")
		return nil
	}

	identifier := p.consume().Value

	if !p.expectTokenVal("-") {
		p.addError("Expected '-' after identifier in define statement")
		return nil
	}
	p.consume() // consume '-'

	if !p.expectTokenVal(">") {
		p.addError("Expected '>' after '-' in define statement")
		return nil
	}
	p.consume() // consume '>'

	// Get the value/replacement
	if p.tokenizer.GetCurrentToken().Type == EOF {
		p.addError("Expected value after '->' in define statement")
		return nil
	}

	value := p.consume().Value

	// Store the define mapping
	p.defines[identifier] = value

	return &ASTNode{
		Type:       DefDecl,
		Identifier: identifier,
		Value:      value,
	}
}

// parseVariableDeclaration parses variable declarations
func (p *Parser) parseVariableDeclaration() *ASTNode {
	node := p.parseVariableDeclarationNoSemicolon()
	if node != nil {
		// Consume optional semicolon
		p.consumeOptionalSemicolon()
	}
	return node
}

// parseVariableDeclarationNoSemicolon parses variable declarations without consuming semicolon
func (p *Parser) parseVariableDeclarationNoSemicolon() *ASTNode {
	p.consume() // consume 'l'

	if !p.expectToken(Identifier) {
		p.addError("Expected identifier after 'l'")
		return nil
	}

	identifier := p.consume().Value

	var initializer *ASTNode
	if p.expectTokenVal("=") {
		p.consume() // consume '='
		initializer = p.parseExpression()
	}

	// Add variable to scope
	p.vars = append(p.vars, Variable{
		DataType: "unknown",
		Val:      identifier,
		NodePos:  len(p.Nodes),
	})

	return &ASTNode{
		Type:        VariableDeclaration,
		Identifier:  identifier,
		Initializer: initializer,
	}
} // parseFunction parses function declarations
func (p *Parser) parseFunction() *ASTNode {
	p.consume() // consume 'f'

	// Function name (optional for anonymous functions)
	var identifier string
	if p.expectToken(Identifier) {
		identifier = p.consume().Value
	}

	// Parameters
	if !p.expectTokenVal("(") {
		p.addError("Expected '(' after function identifier")
		return nil
	}
	p.consume() // consume '('

	var params []ASTNode
	for !p.expectTokenVal(")") && p.tokenizer.GetCurrentToken().Type != EOF {
		if p.expectToken(Identifier) {
			param := p.consume()
			params = append(params, ASTNode{
				Type:  IdentifierD,
				Value: param.Value,
			})

			if p.expectTokenVal(",") {
				p.consume()
			} else if !p.expectTokenVal(")") {
				p.addError("Expected ',' or ')' in parameter list")
				break
			}
		} else {
			p.addError("Expected parameter name")
			break
		}
	}

	if !p.expectTokenVal(")") {
		p.addError("Expected ')' after parameters")
		return nil
	}
	p.consume() // consume ')'

	// Function body
	if !p.expectTokenVal("{") {
		p.addError("Expected '{' to start function body")
		return nil
	}

	body := p.parseBlockStatement()
	if body == nil {
		return nil
	}

	return &ASTNode{
		Type:       FunctionDeclaration,
		Identifier: identifier,
		Params:     params,
		Body:       body.Body,
	}
}

// parseReturn parses return statements
func (p *Parser) parseReturn() *ASTNode {
	p.consume() // consume 'return'

	var initializer *ASTNode
	if !p.expectToken(NewLine) && p.tokenizer.GetCurrentToken().Type != EOF {
		initializer = p.parseExpression()
	}

	node := &ASTNode{
		Type:        Return,
		Initializer: initializer,
	}

	// Consume optional semicolon
	p.consumeOptionalSemicolon()

	return node
}

// parseBlockStatement parses block statements
func (p *Parser) parseBlockStatement() *ASTNode {
	if !p.expectTokenVal("{") {
		p.addError("Expected '{'")
		return nil
	}
	p.consume() // consume '{'

	var body []ASTNode
	for !p.expectTokenVal("}") && p.tokenizer.GetCurrentToken().Type != EOF {
		// Skip newlines in block
		if p.expectToken(NewLine) {
			p.tokenizer.Next()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			body = append(body, *stmt)
		}
	}

	if !p.expectTokenVal("}") {
		p.addError("Expected '}' to close block")
		return nil
	}
	p.consume() // consume '}'

	return &ASTNode{
		Type: BlockStmt,
		Body: body,
	}
}

// parseExpression parses expressions
func (p *Parser) parseExpression() *ASTNode {
	var left *ASTNode

	if p.expectTokenVal("(") {
		// Handle parenthesized expressions
		left = p.parseParenExpr()
	} else if p.expectTokenVal("!") || p.expectTokenVal("-") {
		left = p.parseNotMinusExpression()
	} else if p.expectToken(Identifier) && p.expectPeekVal("(") {
		left = p.parseCallExpr()
	} else if p.expectTokenVal("[") {
		left = p.parseArray()
	} else if p.expectToken(Identifier) && p.expectPeekVal("[") {
		left = p.parseArrIndex()
	} else if p.expectTokenVal("f") {
		left = p.parseFunction()
	} else if p.expectTokenVal("--") || p.expectTokenVal("++") ||
		(p.expectToken(Identifier) && (p.expectPeekVal("--") || p.expectPeekVal("++"))) {
		left = p.parseIncDec()
	} else {
		left = p.parsePrimary()
	}

	if left == nil {
		return nil
	}

	// Handle binary operations
	for p.expectToken(Operator) && p.isBinaryOperator() {
		operator := p.consume().Value
		right := p.parseExpression()
		if right == nil {
			break
		}

		left = &ASTNode{
			Type:     BinaryExpression,
			Operator: operator,
			Left:     left,
			Right:    right,
		}
	}

	return left
}

// isBinaryOperator checks if current token is a binary operator
func (p *Parser) isBinaryOperator() bool {
	tk := p.tokenizer.GetCurrentToken()
	binaryOps := []string{"+", "-", "*", "/", "%", "==", "!=", "<", ">", "<=", ">=", "&&", "||", "=", "+=", "-=", "*=", "/=", "%="}
	for _, op := range binaryOps {
		if tk.Value == op {
			return true
		}
	}
	return false
}

// parsePrimary parses primary expressions (literals, identifiers)
func (p *Parser) parsePrimary() *ASTNode {
	token := p.tokenizer.GetCurrentToken()

	switch token.Type {
	case Literal:
		return &ASTNode{
			Type:  LiteralD,
			Value: p.consume().Value,
		}
	case StringLiteral:
		return &ASTNode{
			Type:  LiteralD,
			Value: p.consume().Value,
		}
	case Identifier:
		return &ASTNode{
			Type:  IdentifierD,
			Value: p.consume().Value,
		}
	case Keyword:
		if IsAllowedKeyAsVal(token.Value) {
			return &ASTNode{
				Type:  LiteralD,
				Value: p.consume().Value,
			}
		}
	}

	p.addError(fmt.Sprintf("Unexpected token: %s", token.Value))
	return nil
}

// parseParenExpr parses parenthesized expressions
func (p *Parser) parseParenExpr() *ASTNode {
	p.consume() // consume '('

	expr := p.parseExpression()
	if expr == nil {
		return nil
	}

	if !p.expectTokenVal(")") {
		p.addError("Expected ')' after expression")
		return nil
	}
	p.consume() // consume ')'

	return &ASTNode{
		Type:  Expression,
		Paren: expr,
	}
}

// parseNotMinusExpression parses unary expressions (! and -)
func (p *Parser) parseNotMinusExpression() *ASTNode {
	operator := p.consume().Value // Consume the unary operator

	var operand *ASTNode
	if p.expectTokenVal("(") {
		operand = p.parseParenExpr()
	} else {
		operand = p.parseExpression()
	}

	return &ASTNode{
		Type:     UnaryExpression,
		Operator: operator,
		Left:     operand,
	}
}

// parseCallExpr parses function call expressions
func (p *Parser) parseCallExpr() *ASTNode {
	identifier := p.consume().Value // Consume the function identifier

	if !p.expectTokenVal("(") {
		p.addError(fmt.Sprintf("Expected '(' after function identifier '%s'", identifier))
		return nil
	}
	p.consume() // consume '('

	var args []ASTNode

	// Check for empty argument list
	if p.expectTokenVal(")") {
		p.consume() // consume ')'
		return &ASTNode{
			Type:       CallExpression,
			Identifier: identifier,
			Args:       args,
		}
	}

	// Parse arguments
	for !p.expectTokenVal(")") && p.tokenizer.GetCurrentToken().Type != EOF {
		arg := p.parseExpression()
		if arg == nil {
			p.addError(fmt.Sprintf("Invalid argument in function call '%s'", identifier))
			break
		}
		args = append(args, *arg)

		if p.expectTokenVal(",") {
			p.consume() // consume ','
		} else if !p.expectTokenVal(")") {
			p.addError(fmt.Sprintf("Expected ',' or ')' in function call '%s'", identifier))
			break
		}
	}

	if !p.expectTokenVal(")") {
		p.addError(fmt.Sprintf("Unmatched parentheses in function call '%s'", identifier))
		return nil
	}
	p.consume() // consume ')'

	return &ASTNode{
		Type:       CallExpression,
		Identifier: identifier,
		Args:       args,
	}
}

// parseArray parses array expressions
func (p *Parser) parseArray() *ASTNode {
	p.consume() // consume '['

	var elements []ASTNode

	if p.expectTokenVal("]") {
		p.consume()
		return &ASTNode{
			Type:     ArrayExpr,
			Elements: elements,
		}
	}

	for !p.expectTokenVal("]") && p.tokenizer.GetCurrentToken().Type != EOF {
		// Skip newlines
		if p.expectToken(NewLine) {
			p.tokenizer.Next()
			continue
		}

		element := p.parseExpression()
		if element == nil {
			p.addError("Invalid array element")
			break
		}
		elements = append(elements, *element)

		if p.expectTokenVal(",") {
			p.consume() // consume ','
		} else if !p.expectTokenVal("]") {
			p.addError("Expected ',' or ']' in array")
			break
		}
	}

	if !p.expectTokenVal("]") {
		p.addError("Unmatched brackets in array")
		return nil
	}
	p.consume() // consume ']'

	return &ASTNode{
		Type:     ArrayExpr,
		Elements: elements,
	}
}

// parseArrIndex parses array index expressions
func (p *Parser) parseArrIndex() *ASTNode {
	identifier := p.consume().Value // Consume the array identifier

	if !p.expectTokenVal("[") {
		p.addError(fmt.Sprintf("Expected '[' after array identifier '%s'", identifier))
		return nil
	}

	var indexNodes []ASTNode

	// Handle multiple nested indices like ident[0][1][2]
	for p.expectTokenVal("[") {
		p.consume() // consume '['

		index := p.parseExpression()
		if index == nil {
			p.addError(fmt.Sprintf("Invalid array index for '%s'", identifier))
			return nil
		}
		indexNodes = append(indexNodes, *index)

		if !p.expectTokenVal("]") {
			p.addError(fmt.Sprintf("Expected ']' after array index for '%s'", identifier))
			return nil
		}
		p.consume() // consume ']'
	}

	return &ASTNode{
		Type:       ArrayIndex,
		Identifier: identifier,
		Index:      indexNodes,
	}
}

// parseIncDec parses increment/decrement expressions
func (p *Parser) parseIncDec() *ASTNode {
	if p.expectToken(Operator) && p.expectPeek(Identifier) {
		// Prefix: ++identifier or --identifier
		infixOp := p.consume().Value
		identifier := p.consume().Value
		return &ASTNode{
			Type:       IncDec,
			InfixOp:    infixOp,
			Identifier: identifier,
		}
	} else {
		// Postfix: identifier++ or identifier--
		identifier := p.consume().Value
		postOp := p.consume().Value
		return &ASTNode{
			Type:       IncDec,
			PostOp:     postOp,
			Identifier: identifier,
		}
	}
}

// parseIfElse parses if-else statements
func (p *Parser) parseIfElse() *ASTNode {
	p.consume() // consume 'if'

	if !p.expectTokenVal("(") {
		p.addError("Expected '(' after 'if'")
		return nil
	}
	p.consume() // consume '('

	test := p.parseExpression()
	if test == nil {
		return nil
	}

	if !p.expectTokenVal(")") {
		p.addError("Expected ')' after if condition")
		return nil
	}
	p.consume() // consume ')'

	consequent := p.parseBlockStatement()
	if consequent == nil {
		return nil
	}

	var alternate *ASTNode
	if p.expectTokenVal("else") {
		p.consume() // consume 'else'

		if p.expectTokenVal("if") {
			// else if
			alternate = p.parseIfElse()
		} else {
			// else block
			alternate = p.parseBlockStatement()
		}
	}

	return &ASTNode{
		Type:       IfElse,
		Test:       test,
		Consequent: consequent,
		Alternate:  alternate,
	}
}

// parseLoop parses for and while loops
func (p *Parser) parseLoop() *ASTNode {
	loopType := p.consume().Value // consume 'for' or 'while'

	if !p.expectTokenVal("(") {
		p.addError(fmt.Sprintf("Expected '(' after '%s'", loopType))
		return nil
	}
	p.consume() // consume '('

	if loopType == "for" {
		// For loop: for (init; test; upgrade)
		// Parse initializer - this is typically a variable declaration
		var initializer *ASTNode
		if p.expectTokenVal("l") {
			initializer = p.parseVariableDeclarationNoSemicolon()
		} else {
			// Could be an expression or assignment
			initializer = p.parseExpression()
		}

		if !p.expectTokenVal(";") {
			p.addError("Expected ';' after for loop initializer")
			return nil
		}
		p.consume() // consume ';'

		test := p.parseExpression()

		if !p.expectTokenVal(";") {
			p.addError("Expected ';' after for loop test")
			return nil
		}
		p.consume() // consume ';'

		upgrade := p.parseExpression()

		if !p.expectTokenVal(")") {
			p.addError("Expected ')' after for loop")
			return nil
		}
		p.consume() // consume ')'

		body := p.parseBlockStatement()
		if body == nil {
			return nil
		}

		return &ASTNode{
			Type:        Loop,
			Initializer: initializer,
			Test:        test,
			Upgrade:     upgrade,
			Body:        body.Body,
		}
	} else {
		// While loop: while (test)
		test := p.parseExpression()

		if !p.expectTokenVal(")") {
			p.addError("Expected ')' after while condition")
			return nil
		}
		p.consume() // consume ')'

		body := p.parseBlockStatement()
		if body == nil {
			return nil
		}

		return &ASTNode{
			Type: Loop,
			Test: test,
			Body: body.Body,
		}
	}
}
