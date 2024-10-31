package parser

import (
	"log"
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
)

type pNode struct {
	Type  int
	Value string
}

//	type right struct {
//		Type     int
//		Value     string
//		Operator string
//		Left     left
//		Right    *right
//	}
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
	Body        *ASTNode
	Params      []pNode
	Test        *ASTNode
	Consequent  *ASTNode
	Alternate   *ASTNode
}

type Parser struct {
	File      string
	Tokenizer *TokenGen
	Nodes     []ASTNode
	Parens    []string
	Braces    []string
	Bracs     []string
}

func NewParser(file string) *Parser {
	tokenizer := NewTokenGen(file)
	return &Parser{File: file, Tokenizer: tokenizer}
}

func (p *Parser) Consume() Token {
	token := p.Tokenizer.GetCurrentToken()
	p.Tokenizer.Next()
	return token
}
func (p *Parser) GroupBy(group string) string {
	mGroup := group
	closingChar := ")"
	switch group {
	case "{":
		closingChar = "}"
	case "[":
		closingChar = "]"
	case "(":
		closingChar = ")"
	default:
		closingChar = group
	}
	p.Consume()
	for p.Tokenizer.GetCurrentToken().Value != closingChar {
		if p.Tokenizer.GetCurrentToken().Value == "" {
			log.Fatal("Error: unmatched grouping,  expected: ", closingChar)
		}
		cGroup := ""
		curValue := p.Tokenizer.GetCurrentToken().Value
		if curValue == "(" || curValue == "[" || curValue == "{" {
			cGroup += p.GroupBy(curValue)
		} else {
			cGroup += p.Consume().Value
		}
		mGroup += cGroup
	}
	p.Consume()
	return mGroup + closingChar
}
func (p *Parser) ParseLiteral() ASTNode {
	return ASTNode{Type: LiteralD, Value: p.Consume().Value}
}

func (p *Parser) ParseBinaryExpr() ASTNode {
	operator := ""
	left := &ASTNode{Type: p.Tokenizer.GetCurrentToken().Type, Value: p.Tokenizer.GetCurrentToken().Value}
	rightS := &ASTNode{}

	p.Consume()
	if len(p.Tokenizer.GetTokenLeftLine()) != 0 {
		switch p.Tokenizer.GetCurrentToken().Type {
		case Operator:
			switch p.Tokenizer.GetCurrentToken().Value {
			case "+", "-", "*", "/":
				operator = p.Consume().Value
				token := p.Tokenizer.GetTokenLeftLine()
				if len(token) > 1 {
					val := p.ParseBinaryExpr()
					rightS = &val
				} else {
					val := p.Consume()
					rightS = &ASTNode{Type: val.Type, Value: val.Value}
				}
			default:
				//an error
			}
		default:
			//another error
		}
	}

	return ASTNode{Type: BinaryExpression, Left: left, Operator: operator, Right: rightS}
}

func (p *Parser) ParseNotnMinusExpr() ASTNode {
	initializer := p.Consume().Value
	if len(p.Tokenizer.GetTokenLeftLine()) > 1 {
		// todo
	} else {
		initializer += p.Consume().Value
	}
	return ASTNode{Type: NotExpression, Value: initializer}
}
func (p *Parser) ParseVarDecl() ASTNode {
	p.Tokenizer.Next()
	identifier := ""
	initializer := ASTNode{}
	if p.Tokenizer.GetCurrentToken().Type == Identifier {
		identifier = p.Consume().Value
		if len(p.Tokenizer.GetTokenLeftLine()) == 0 {
			return ASTNode{Type: VariableDeclaration, Identifier: identifier}
		}
		if p.Tokenizer.GetCurrentToken().Value == "=" {
			p.Consume()
			leftTokens := p.Tokenizer.GetTokenLeftLine()
			switch p.Tokenizer.GetCurrentToken().Type {
			case Identifier:
				if len(leftTokens) == 0 {
					initializer = p.ParseLiteral()
				} else {
					initializer = p.ParseBinaryExpr()
				}
			case Literal:
				if len(leftTokens) == 0 {
					initializer = p.ParseLiteral()
				} else {
					initializer = p.ParseBinaryExpr()
				}
			case StringLiteral:
				if len(leftTokens) == 0 {
					initializer = p.ParseLiteral()
				} else {
					initializer = p.ParseBinaryExpr()
				}
			case Punctuation:
				//todo
				initializer = ASTNode{Type: Punctuation, Value: p.GroupBy(p.Tokenizer.GetCurrentToken().Value)}
			case Operator:
				if p.Tokenizer.GetCurrentToken().Value == "!" || p.Tokenizer.GetCurrentToken().Value == "-" {
					initializer = p.ParseNotnMinusExpr()
				} else {
					// likely an error but i'm getting ideas to do something with my syntax
				}
			default:
				// errorrrr damn i need an error variable, like a slice or smth
			}
		} else {
			// error
		}
	} else {
		// error
	}
	return ASTNode{
		Type:        VariableDeclaration,
		Identifier:  identifier,
		Initializer: &initializer,
	}
}
func (p *Parser) CheckAndParse() {
	baseToken := p.Tokenizer.GetCurrentToken()
	switch baseToken.Type {
	case Keyword:
		switch baseToken.Value {
		case "l":
			node := p.ParseVarDecl()
			p.Nodes = append(p.Nodes, node)
		default:
			//todo
		}
	default:
		//todo
	}
}
func (p *Parser) Start() {
	for i := 0; i < len(p.Tokenizer.lines); i++ {
		p.CheckAndParse()
	}
}
