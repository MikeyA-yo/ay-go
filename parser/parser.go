package parser

import (
	"strings"
)

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
)

type pNode struct {
	Type  int
	Value string
}

type ASTNode struct {
	Type        int
	Name        string
	Value       string
	Raw         string
	Identifier  string
	Initializer *ASTNode
	DataType    string
	Operator    string
	Left        *ASTNode
	Right       *ASTNode
	Body        []ASTNode
	Params      []ASTNode
	Test        *ASTNode
	Consequent  *ASTNode
	Alternate   *ASTNode
	Args        []ASTNode
	Elements    []ASTNode
	Index       []ASTNode
	Operand     *ASTNode
	Upgrade     *ASTNode
	Paren       *ASTNode
	InfixOp     string
	PostOp      string
}

type Variable struct {
	dataType string
	val      string
	nodePos  int
}

type Parser struct {
	Tokenizer *TokenGen
	Nodes     []ASTNode
	Parens    []string
	Braces    []string
	Bracs     []string
	Defines   map[string]string
	Vars      []Variable
	Errors    []string
}

func NewParser(file string) *Parser {
	tokenizer := NewTokenGen(file)
	return &Parser{Tokenizer: tokenizer, Defines: make(map[string]string)}
}
func (p *Parser) AddError(err string) {
	line := p.Tokenizer.GetCurrentLineNumber()
	col := p.Tokenizer.GetCurrentColNumber()
	currentToken := p.Tokenizer.GetCurrentToken()

	actualLineSource := p.Tokenizer.Lines[line-1]
	if actualLineSource == "" {
		actualLineSource = "(empty line)"
	}

	lnPointer := strings.Repeat(" ", col-1) + "^"

	errorMessage := "Error at line " + string(rune(line)) + ", column " + string(rune(col)) + ": " + err + "\n" +
		actualLineSource + "\n" +
		lnPointer + " (near '" + currentToken.Value + "')"
	p.Errors = append(p.Errors, errorMessage)
}

// Resolve defined aliases - replace any defined keyword with its actual value
func (p *Parser) ResolveDefine(value string) string {
	if defined, exists := p.Defines[value]; exists {
		return defined
	}
	return value
}

func (p *Parser) Consume() Token {
	token := p.Tokenizer.GetCurrentToken()
	// Create a new token with resolved value if it's a define
	resolvedToken := Token{
		Type:  token.Type,
		Value: p.ResolveDefine(token.Value),
		Line:  token.Line,
		Col:   token.Col,
	}
	p.Tokenizer.Next()
	return resolvedToken
}

func (p *Parser) ExpectPeek(t int) bool {
	pk := p.Tokenizer.Peek(0)
	return pk.Type == t
}

func (p *Parser) ExpectToken(t int) bool {
	tk := p.Tokenizer.GetCurrentToken()
	return tk.Type == t
}

func (p *Parser) ExpectPeekVal(v string) bool {
	pk := p.Tokenizer.Peek(0)
	resolvedValue := p.ResolveDefine(pk.Value)
	return pk.Value == v || resolvedValue == v
}

func (p *Parser) ExpectTokenVal(v string) bool {
	tk := p.Tokenizer.GetCurrentToken()
	resolvedValue := p.ResolveDefine(tk.Value)
	return tk.Value == v || resolvedValue == v
}

func (p *Parser) ParseLiteral() ASTNode {
	token := p.Consume()
	if p.ExpectToken(NewLine) || p.ExpectToken(EOF) {
		p.Consume()
	}
	return ASTNode{
		Type:  LiteralD,
		Value: token.Value,
	}
}

func (p *Parser) IsDefinedVar(v string) bool {
	for _, val := range p.Vars {
		if val.val == v {
			return true
		}
	}
	return false
}

func (p *Parser) ParseNotnMinusExpression() ASTNode {
	operatorToken := p.Consume() // Consume the unary operator (! or -)
	var operand ASTNode
	if p.ExpectTokenVal("(") {
		operand = p.ParseParenExpr()
	} else {
		operand = p.ParseExpression()
	}
	return ASTNode{
		Type:     UnaryExpression,
		Operator: operatorToken.Value,
		Operand:  &operand,
	}
}

func (p *Parser) ParseCallExpr() ASTNode {
	identifier := p.Consume().Value // Consume the function identifier
	var args []ASTNode

	if !p.ExpectTokenVal("(") {
		p.AddError("SyntaxError: Expected '(' after function identifier '" + identifier + "'")
		return ASTNode{} // Return empty node if parentheses are not found
	}

	p.Consume() // Consume the opening '('

	// Check for an empty argument list
	if p.ExpectTokenVal(")") {
		p.Consume() // Consume the closing ')'
		return ASTNode{
			Type:       CallExpression,
			Identifier: identifier,
			Args:       args,
		}
	}

	// Parse the arguments
	for !p.ExpectTokenVal(")") {
		arg := p.ParseExpression()
		args = append(args, arg) // Collect the parsed argument
		if p.ExpectTokenVal(",") {
			p.Consume() // Consume the comma separator
		} else if !p.ExpectTokenVal(")") {
			p.AddError("SyntaxError: Expected ',' or ')' in function call '" + identifier + "'")
			break
		}
	}

	if p.ExpectTokenVal(")") {
		p.Consume() // Consume the closing ')'
	} else {
		p.AddError("SyntaxError: Unmatched parentheses in function call '" + identifier + "'")
	}

	return ASTNode{
		Type:       CallExpression,
		Identifier: identifier,
		Args:       args,
	}
}

func (p *Parser) ParseArray() ASTNode {
	var elements []ASTNode
	p.Consume() // Consume the opening '['

	if p.ExpectTokenVal("]") {
		p.Consume()
		return ASTNode{
			Type:     ArrayExpr,
			Elements: elements,
		}
	}

	for !p.ExpectTokenVal("]") {
		// Skip newlines
		if p.ExpectToken(NewLine) {
			p.Consume()
			continue
		}

		element := p.ParseExpression()
		elements = append(elements, element)

		if p.ExpectTokenVal(",") {
			p.Consume() // Consume the comma separator
		} else if !p.ExpectTokenVal("]") {
			p.AddError("SyntaxError: Expected ',' or ']' in array")
			break
		}
	}

	if p.ExpectTokenVal("]") {
		p.Consume() // Consume the closing ']'
	} else {
		p.AddError("SyntaxError: Unmatched brackets in array")
	}

	return ASTNode{
		Type:     ArrayExpr,
		Elements: elements,
	}
}

func (p *Parser) ParseIncDec() ASTNode {
	if p.ExpectToken(Operator) && p.ExpectPeek(Identifier) {
		infixop := p.Consume().Value
		identifier := p.Consume().Value
		return ASTNode{
			Type:       IncDec,
			InfixOp:    infixop,
			Identifier: identifier,
		}
	} else {
		identifier := p.Consume().Value
		postop := p.Consume().Value
		return ASTNode{
			Type:       IncDec,
			PostOp:     postop,
			Identifier: identifier,
		}
	}
}

func (p *Parser) ParseArrIndex() ASTNode {
	identifier := p.Consume().Value // Consume the array identifier
	if !p.ExpectTokenVal("[") {
		p.AddError("SyntaxError: Expected '[' after array identifier '" + identifier + "'")
		return ASTNode{}
	}

	var indexNodes []ASTNode

	// Handle multiple nested indices like ident[0][1][2]
	for p.ExpectTokenVal("[") {
		p.Consume() // Consume the opening '['

		index := p.ParseExpression()
		indexNodes = append(indexNodes, index)

		if !p.ExpectTokenVal("]") {
			p.AddError("SyntaxError: Expected ']' after array index for '" + identifier + "'")
			return ASTNode{}
		}
		p.Consume() // Consume the closing ']'
	}

	return ASTNode{
		Type:       ArrayIndex,
		Identifier: identifier,
		Index:      indexNodes,
	}
}

func (p *Parser) ParseExpression() ASTNode {
	var left interface{}

	if p.ExpectTokenVal("(") {
		// Handle parenthesized expressions
		left = p.ParseParenExpr()
	} else if p.ExpectTokenVal("!") || p.ExpectTokenVal("-") {
		left = p.ParseNotnMinusExpression()
	} else if p.ExpectToken(Identifier) && p.ExpectPeekVal("(") {
		left = p.ParseCallExpr()
	} else if p.ExpectTokenVal("[") {
		left = p.ParseArray()
	} else if p.ExpectToken(Identifier) && p.ExpectPeekVal("[") {
		left = p.ParseArrIndex()
	} else if p.ExpectTokenVal("f") {
		left = p.ParseFunc()
	} else if p.ExpectTokenVal("--") || p.ExpectTokenVal("++") ||
		(p.ExpectToken(Identifier) && (p.ExpectPeekVal("--") || p.ExpectPeekVal("++"))) {
		left = p.ParseIncDec()
	} else {
		// Consume basic literals/identifiers
		left = p.Consume().Value
	}

	// Check if there's an operator next
	if p.ExpectToken(Operator) || p.ExpectTokenVal("~") {
		if p.ExpectTokenVal("~") {
			// Modify the current token to be an operator
			currentToken := p.Tokenizer.GetCurrentToken()
			currentToken.Value = "+"
			currentToken.Type = Operator
		}
		op := p.Consume().Value
		// Validate the next token for the right-hand side
		if !p.ExpectToken(Identifier) && !p.ExpectToken(Literal) &&
			!p.ExpectToken(StringLiteral) && !p.ExpectTokenVal("(") {
			p.AddError("Invalid expression after operator '" + op + "' - Expected identifier, number, string, or parenthesized expression")
			return ASTNode{Value: left.(string)} // Return what we have so far
		}

		// Handle the right-hand side of the expression
		var right interface{}
		if p.ExpectPeek(EOF) || p.ExpectPeek(NewLine) || p.ExpectPeekVal(";") ||
			p.ExpectPeekVal(")") || p.ExpectPeekVal(",") {
			right = p.Consume().Value // Simple right-hand expression
			if !p.ExpectTokenVal(")") && !p.ExpectTokenVal(",") {
				p.Consume()
			}
		} else {
			right = p.ParseExpression() // Recursively parse complex expressions
		}

		return ASTNode{
			Type:     BinaryExpression,
			Operator: op,
			Left:     &ASTNode{Value: left.(string)},
			Right:    &ASTNode{Value: right.(string)},
		}
	}

	switch v := left.(type) {
	case ASTNode:
		return v
	case string:
		return ASTNode{Value: v}
	default:
		return ASTNode{}
	}
}

func (p *Parser) ParseParenExpr() ASTNode {
	p.Consume() // Consume the opening '('
	// Important Error checks:
	if p.ExpectTokenVal(")") {
		p.AddError("Empty parentheses - Expected an expression inside parentheses")
	}
	expression := p.ParseExpression() // Parse the inner expression
	if p.ExpectTokenVal(")") {
		p.Consume() // Consume the closing ')'
	} else {
		p.AddError("Unmatched parentheses - Missing closing ')' for opening '('")
	}

	return ASTNode{
		Type:  Expression,
		Paren: &expression,
	} // Return the parsed inner expression
}

func (p *Parser) ParseVariable() ASTNode {
	p.Tokenizer.Next()
	var identifier string
	var initializer ASTNode
	dT := "unknown"

	if p.ExpectToken(Identifier) {
		identifier = p.Consume().Value
		// Check if it's just a plain declaration
		if p.ExpectToken(EOF) || p.ExpectToken(NewLine) || p.ExpectTokenVal(";") {
			p.Consume()
			return ASTNode{
				Type:       VariableDeclaration,
				Identifier: identifier,
			}
		}
		// Declaration and initialization
		if p.ExpectTokenVal("=") {
			p.Consume()
			leftTokenValues := p.Tokenizer.GetTokenLeftLine()
			switch p.Tokenizer.GetCurrentToken().Type {
			case Identifier:
				if len(leftTokenValues) == 1 {
					idtV := p.Tokenizer.GetCurrentToken().Value
					initializer = p.ParseLiteral()
					if p.IsDefinedVar(idtV) {
						for _, v := range p.Vars {
							if v.val == idtV {
								dT = v.dataType
								break
							}
						}
					}
					bL := len(p.Nodes)
					p.Vars = append(p.Vars, Variable{dataType: dT, val: identifier, nodePos: bL})
				} else {
					initializer = p.ParseExpression()
				}
			case Literal:
				if len(leftTokenValues) == 1 {
					initializer = p.ParseLiteral()
					dT = "number"
					bL := len(p.Nodes)
					p.Vars = append(p.Vars, Variable{dataType: dT, val: identifier, nodePos: bL})
				} else {
					initializer = p.ParseExpression()
				}
			case StringLiteral:
				if len(leftTokenValues) == 1 || p.ExpectPeek(NewLine) {
					initializer = p.ParseLiteral()
					dT = "string"
					bL := len(p.Nodes)
					p.Vars = append(p.Vars, Variable{dataType: dT, val: identifier, nodePos: bL})
				} else {
					initializer = p.ParseExpression()
				}
			case Punctuation:
				initializer = p.ParseExpression()
			case Operator:
				if p.ExpectTokenVal("!") || p.ExpectTokenVal("-") ||
					p.ExpectTokenVal("--") || p.ExpectTokenVal("++") {
					initializer = p.ParseExpression()
				} else {
					p.AddError("Unexpected operator in variable initialization")
				}
			case Keyword:
				if p.ExpectTokenVal("true") || p.ExpectTokenVal("false") {
					initializer = p.ParseExpression()
					p.Vars = append(p.Vars, Variable{
						dataType: "boolean",
						val:      identifier,
						nodePos:  len(p.Nodes),
					})
				} else {
					if IsAllowedKeyAsVal(p.Tokenizer.GetCurrentToken().Value) {
						switch p.Tokenizer.GetCurrentToken().Value {
						case "f":
							initializer = p.ParseFunc()
						default:
							initializer = p.ParseLiteral()
						}
					}
				}
			default:
				p.AddError("Unexpected token '" + p.Tokenizer.GetCurrentToken().Value + "' of type " + string(rune(p.Tokenizer.GetCurrentToken().Type)) + " at variable initialization for '" + identifier + "' - Expected a value")
				p.Tokenizer.ToNewLine()
			}
			return ASTNode{
				Type:        VariableDeclaration,
				Identifier:  identifier,
				Initializer: &initializer,
			}
		} else {
			p.AddError("Unexpected token '" + p.Tokenizer.GetCurrentToken().Value + "' after variable identifier '" + identifier + "' - Expected '=' for variable assignment")
			p.Consume()
			p.Tokenizer.ToNewLine()
		}
	} else {
		p.AddError("Unexpected token '" + p.Tokenizer.GetCurrentToken().Value + "' of type " + string(rune(p.Tokenizer.GetCurrentToken().Type)) + " in variable declaration - Expected identifier after 'l' keyword")
		p.Consume()
		p.Tokenizer.ToNewLine()
	}
	return ASTNode{}
}

func (p *Parser) ParseDefine() ASTNode {
	if p.ExpectPeek(Identifier) {
		p.Consume()
		identifier := p.Consume().Value
		var initializer ASTNode
		if p.ExpectTokenVal("-") {
			p.Consume()
			if p.ExpectTokenVal(">") {
				p.Consume()
				initializer = p.ParseLiteral()
				p.Defines[identifier] = initializer.Value
			} else {
				p.AddError("Unexpected token: " + p.Consume().Value + ", expected >")
			}
		} else {
			p.AddError("Unexpected token: " + p.Consume().Value + ", expected def chain ->")
		}
		return ASTNode{
			Type:        DefDecl,
			Identifier:  identifier,
			Initializer: &initializer,
		}
	} else {
		p.AddError("Unexpected token type: '" + p.Tokenizer.Peek(0).Value + "' (" + string(rune(p.Tokenizer.Peek(0).Type)) + ") after 'def' keyword - Expected identifier to define")
		p.Tokenizer.ToNewLine()
	}
	return ASTNode{}
}

func (p *Parser) ParseReturn() ASTNode {
	if p.ExpectPeek(NewLine) || p.ExpectPeekVal(";") || p.ExpectPeekVal("}") {
		rtToken := p.Consume()
		return ASTNode{
			Type:  Return,
			Value: rtToken.Value,
		}
	} else {
		if p.ExpectPeek(Identifier) || p.ExpectPeek(Literal) ||
			p.ExpectPeek(StringLiteral) || IsAllowedKeyAsVal(p.Tokenizer.Peek(0).Value) {
			p.Consume()
			tk := p.ParseExpression()
			return ASTNode{
				Type:        Return,
				Initializer: &tk,
			}
		}
		p.AddError("Unexpected token '" + p.Tokenizer.GetCurrentToken().Value + "' after return statement - Expected newline, semicolon, or end of block")
		p.Consume()
	}
	return ASTNode{}
}

func (p *Parser) ParseBreakNCont() ASTNode {
	keyword := p.Consume() // Get the break or continue keyword

	if p.ExpectToken(NewLine) || p.ExpectToken(EOF) ||
		p.ExpectTokenVal(";") || p.ExpectTokenVal("}") {
		// Valid termination for break/continue
		if p.ExpectToken(NewLine) || p.ExpectTokenVal(";") {
			p.Consume()
		}

		nodeType := Break
		if keyword.Value == "continue" {
			nodeType = Continue
		}
		return ASTNode{
			Type:  nodeType,
			Value: keyword.Value,
		}
	} else {
		p.AddError("Unexpected token '" + p.Tokenizer.GetCurrentToken().Value + "' after " + keyword.Value + " keyword - Expected newline, semicolon, or end of block")
		p.Tokenizer.ToNewLine()
		nodeType := Break
		if keyword.Value == "continue" {
			nodeType = Continue
		}
		return ASTNode{
			Type:  nodeType,
			Value: keyword.Value,
		}
	}
}

func (p *Parser) ParseFunc() ASTNode {
	p.Consume()
	var identifier string
	var params []ASTNode
	var body []ASTNode

	if p.ExpectToken(Identifier) {
		// Named function
		identifier = p.Consume().Value
		if p.ExpectTokenVal("(") {
			params = p.ParseFuncParams()
			if p.ExpectTokenVal("{") {
				body = p.ParseBlockStmt()
			} else {
				p.AddError("Expected '{' for function body")
			}
			return ASTNode{
				Type:       FunctionDeclaration,
				Identifier: identifier,
				Params:     params,
				Body:       body,
			}
		} else {
			p.AddError("Expected '(' at function declaration")
			return ASTNode{} // Stop parsing this function
		}
	} else {
		// Anonymous function
		if p.ExpectTokenVal("(") {
			params = p.ParseFuncParams()
			if p.ExpectTokenVal("{") {
				body = p.ParseBlockStmt()
			} else {
				p.AddError("Expected '{' for function body")
			}
			return ASTNode{
				Type:   FunctionDeclaration,
				Params: params,
				Body:   body,
			}
		} else {
			p.AddError("Expected '(' at anonymous function declaration")
		}
	}
	return ASTNode{}
}

func (p *Parser) ParseFuncParams() []ASTNode {
	p.Consume()
	var params []ASTNode
	for !p.ExpectTokenVal(")") {
		arg := p.ParseLiteral()
		params = append(params, arg) // Collect the parsed argument
		if p.ExpectTokenVal(",") {
			p.Consume() // Consume the comma separator
		} else if !p.ExpectTokenVal(")") {
			p.AddError("SyntaxError: Expected ',' or ')' in function declaration parameters - Found '" + p.Tokenizer.GetCurrentToken().Value + "' instead")
			break
		}
	}
	if p.ExpectTokenVal(")") {
		p.Consume()
	} else {
		p.AddError("SyntaxError: Unmatched parentheses in function declaration - Expected closing ')'")
	}
	return params
}

func (p *Parser) ParseBlockStmt() []ASTNode {
	p.Consume()
	var body []ASTNode
	for !p.ExpectTokenVal("}") {
		// Skip newlines and statement terminator
		if p.ExpectToken(NewLine) || p.ExpectTokenVal(";") {
			p.Consume()
			continue
		}
		node := p.CheckParseReturn()
		body = append(body, node)
	}
	if p.ExpectTokenVal("}") {
		p.Consume()
	} else {
		p.AddError("Block not closed properly - Expected '}' to close block opened with '{'")
	}
	return body
}

func (p *Parser) ParseIfElse() ASTNode {
	p.Consume()
	var test ASTNode
	var consequence []ASTNode
	var alternate *ASTNode

	if p.ExpectTokenVal("(") {
		test = p.ParseExpression()
	} else {
		p.AddError("SyntaxError: Expected '(' for if condition, got '" + p.Tokenizer.GetCurrentToken().Value + "' - If statements require parentheses around the condition")
		return ASTNode{}
	}
	if !p.ExpectTokenVal("{") {
		p.AddError("SyntaxError: Expected '{' to start if statement body, got '" + p.Tokenizer.GetCurrentToken().Value + "' - Code blocks must be wrapped in curly braces")
		return ASTNode{}
	}
	consequence = p.ParseBlockStmt()
	if !p.ExpectTokenVal("else") {
		return ASTNode{
			Type:       IfElse,
			Test:       &test,
			Consequent: &ASTNode{Body: consequence},
		}
	}
	p.Consume()
	if p.ExpectTokenVal("if") || p.ExpectTokenVal("{") {
		if p.ExpectTokenVal("{") {
			alternateBody := p.ParseBlockStmt()
			alternate = &ASTNode{Body: alternateBody}
		} else {
			alt := p.ParseIfElse()
			alternate = &alt
		}
		return ASTNode{
			Type:       IfElse,
			Test:       &test,
			Consequent: &ASTNode{Body: consequence},
			Alternate:  alternate,
		}
	} else {
		p.AddError("SyntaxError: Unexpected token '" + p.Tokenizer.GetCurrentToken().Value + "' after else keyword - Expected 'if' for else-if or '{' for else block")
		return ASTNode{}
	}
}

func (p *Parser) ParseWhileLoop() ASTNode {
	p.Consume()
	var test ASTNode
	var body []ASTNode

	if !p.ExpectTokenVal("(") {
		p.AddError("SyntaxError: Expected '(' for while condition, got '" + p.Tokenizer.GetCurrentToken().Value + "' - While loops require parentheses around the condition")
		return ASTNode{}
	}
	test = p.ParseExpression()
	if !p.ExpectTokenVal("{") {
		p.AddError("SyntaxError: Expected '{' to start while loop body, got '" + p.Tokenizer.GetCurrentToken().Value + "' - Code blocks must be wrapped in curly braces")
		return ASTNode{}
	}
	body = p.ParseBlockStmt()
	return ASTNode{
		Type: Loop,
		Test: &test,
		Body: body,
	}
}

func (p *Parser) ParseForLoop() ASTNode {
	p.Consume()
	var initializer ASTNode
	var test ASTNode
	var upgrade ASTNode
	var body []ASTNode

	if !p.ExpectTokenVal("(") {
		p.AddError("SyntaxError: Expected '(' for for loop, got '" + p.Tokenizer.GetCurrentToken().Value + "' - For loops require parentheses around the initialization, condition, and update")
		return ASTNode{}
	}
	p.Consume()
	initializer = p.ParseVariable()
	if p.ExpectToken(NewLine) || p.ExpectTokenVal(";") {
		p.Consume()
	}
	test = p.ParseExpression()
	if p.ExpectToken(NewLine) || p.ExpectTokenVal(";") {
		p.Consume()
	}
	upgrade = p.ParseExpression()
	p.Consume()
	if p.ExpectToken(NewLine) || p.ExpectTokenVal(";") {
		p.Consume()
	}
	if !p.ExpectTokenVal("{") {
		p.AddError("SyntaxError: Expected '{' to start for loop body, got '" + p.Tokenizer.GetCurrentToken().Value + "' - Code blocks must be wrapped in curly braces")
		return ASTNode{}
	}
	body = p.ParseBlockStmt()
	return ASTNode{
		Type:        Loop,
		Initializer: &initializer,
		Test:        &test,
		Upgrade:     &upgrade,
		Body:        body,
	}
}

func (p *Parser) CheckParseReturn() ASTNode {
	baseToken := p.Tokenizer.GetCurrentToken()
	var node ASTNode

	// First check if this is a keyword, then resolve any defines
	resolvedValue := p.ResolveDefine(baseToken.Value)

	// Check if the resolved value is a keyword, even if the original wasn't
	isResolvedKeyword := baseToken.Type == Keyword ||
		(baseToken.Type == Identifier && p.Defines[baseToken.Value] != "")

	if isResolvedKeyword {
		switch resolvedValue {
		case "l":
			node = p.ParseVariable()
		case "def":
			node = p.ParseDefine()
		case "return":
			node = p.ParseReturn()
		case "break", "continue":
			node = p.ParseBreakNCont()
		case "f":
			node = p.ParseFunc()
		case "if":
			node = p.ParseIfElse()
		case "while":
			node = p.ParseWhileLoop()
		case "for":
			node = p.ParseForLoop()
		default:
			// Handle other cases
		}
	} else {
		switch baseToken.Type {
		case Punctuation:
			if baseToken.Value == ";" {
				p.Tokenizer.Next()
			} else {
				p.AddError("Unexpected punctuation: '" + baseToken.Value + "' - Cannot start a statement with this punctuation")
				p.Tokenizer.Next()
			}
		case NewLine:
			p.Tokenizer.Next()
		case Identifier, Literal, StringLiteral:
			node = p.ParseExpression()
		case Operator:
			if p.ExpectTokenVal("!") || p.ExpectTokenVal("-") ||
				p.ExpectTokenVal("--") || p.ExpectTokenVal("++") {
				node = p.ParseExpression()
			} else {
				p.AddError("Unexpected operator: '" + baseToken.Value + "' - Cannot start a statement with this operator (expected prefix operators like !, -, ++, --)")
				p.Tokenizer.Next()
			}
		default:
			p.AddError("Unexpected statement start: '" + baseToken.Value + "' - Expected variable declaration (l), function (f), if statement, loop, or expression")
			p.Tokenizer.Next()
		}
	}
	return node
}

func (p *Parser) CheckAndParse() {
	baseToken := p.Tokenizer.GetCurrentToken()

	// First check if this is a keyword, then resolve any defines
	resolvedValue := p.ResolveDefine(baseToken.Value)

	// Check if the resolved value is a keyword, even if the original wasn't
	isResolvedKeyword := baseToken.Type == Keyword ||
		(baseToken.Type == Identifier && p.Defines[baseToken.Value] != "")

	if isResolvedKeyword {
		switch resolvedValue {
		case "l":
			nodeV := p.ParseVariable()
			p.Nodes = append(p.Nodes, nodeV)
		case "def":
			nodeD := p.ParseDefine()
			p.Nodes = append(p.Nodes, nodeD)
		case "return":
			nodeR := p.ParseReturn()
			p.Nodes = append(p.Nodes, nodeR)
		case "break", "continue":
			nodeBC := p.ParseBreakNCont()
			p.Nodes = append(p.Nodes, nodeBC)
		case "f":
			nodeF := p.ParseFunc()
			p.Nodes = append(p.Nodes, nodeF)
		case "if":
			nodeIf := p.ParseIfElse()
			p.Nodes = append(p.Nodes, nodeIf)
		case "while":
			nodeW := p.ParseWhileLoop()
			p.Nodes = append(p.Nodes, nodeW)
		case "for":
			nodeFo := p.ParseForLoop()
			p.Nodes = append(p.Nodes, nodeFo)
		default:
			p.Consume()
		}
	} else {
		switch baseToken.Type {
		case Punctuation:
			if baseToken.Value == ";" {
				p.Tokenizer.Next()
			} else {
				p.AddError("Unexpected punctuation: '" + baseToken.Value + "' - Cannot start a statement with this punctuation")
				p.Tokenizer.Next()
			}
		case NewLine:
			p.Tokenizer.Next()
		case Identifier, Literal, StringLiteral:
			nodeE := p.ParseExpression()
			p.Nodes = append(p.Nodes, nodeE)
		case Operator:
			if p.ExpectTokenVal("!") || p.ExpectTokenVal("-") ||
				p.ExpectTokenVal("--") || p.ExpectTokenVal("++") {
				nodeO := p.ParseExpression()
				p.Nodes = append(p.Nodes, nodeO)
			} else {
				p.AddError("Unexpected operator: '" + baseToken.Value + "' - Cannot start a statement with this operator (expected prefix operators like !, -, ++, --)")
				p.Tokenizer.Next()
			}
		default:
			p.AddError("Unexpected statement start: '" + baseToken.Value + "' - Expected variable declaration (l), function (f), if statement, loop, or expression")
			p.Tokenizer.Next()
		}
	}
}

func (p *Parser) Start() {
	for p.Tokenizer.GetCurrentToken().Type != EOF {
		p.CheckAndParse()
	}
}
