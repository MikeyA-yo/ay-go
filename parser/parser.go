package parser

import "log"

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

type left struct {
	Type string
	Name string
}

type right struct {
	Type     string
	Name     string
	Operator string
	Left     left
	Right    *right
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
	Left        left
	Right       right
	Body        *ASTNode
	Params      []left
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
