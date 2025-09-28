package parser

import (
	"log"
	"regexp"
	"slices"
	"strings"
)

var Keywords = []string{
	"l",
	"def",
	"f",
	"defer",
	"for",
	"if",
	"else",
	"while",
	"continue",
	"return",
	"break",
	"do",
	"imp@",
	"exp@",
	"print",
	"from",
	"false",
	"true",
	"class",
	"const",
	"debugger",
	"delete",
	"extends",
	"finally",
	"in",
	"instanceof",
	"new",
	"null",
	"super",
	"switch",
	"this",
	"throw",
	"try",
	"typeof",
	"void",
	"with",
	"yield",
}

var allowedKeysAsVal = []string{"true", "false", "null", "f", "new"}

func isAllowedKeyAsVal(key string) bool {
	return slices.Contains(allowedKeysAsVal, key)
}

const (
	Identifier int = 0 + iota
	Operator
	Keyword
	Literal
	StringLiteral
	Whitespace
	Punctuation
	SingleLineComment
	MultiLineComment
	NewLine
	EOF
	Unknown
)

type Token struct {
	Type  int
	Value string
	Line  int
	Col   int
}

func isKeyword(key string) bool {
	return slices.Contains(Keywords, key)
}
func testRegex(p, s string) bool {
	re, er := regexp.Compile(p)
	if er != nil {
		log.Fatal("Unexpected regex error: ", er)
	}
	test := re.MatchString(s)

	return test
}
func Tokenize(line string) []Token {
	var tokens []Token
	var currentToken string
	var currentType int
	var sOpen bool
	var qChar string

	for i := 0; i < len(line); i++ {
		char := string(line[i])
		nextChar := string(line[i+1])

		// multi line comment start
		if char == "/" && nextChar == "*" && currentType != SingleLineComment && !sOpen {
			currentType = MultiLineComment
			currentToken = "/*"
			continue

		}
	}
}

type TokenGen struct {
	lines          []string
	currentLine    int
	currentTokenNo int
}

func NewTokenGen(line string) *TokenGen {
	var lines []string
	if strings.Contains(line, "\r\n") {
		lines = strings.Split(line, "\r\n")
	} else {
		lines = strings.Split(line, "\n")
	}
	return &TokenGen{lines: lines, currentLine: 0, currentTokenNo: 0}
}
func (t *TokenGen) Next() {
	var currentLineToken = Tokenize(t.lines[t.currentLine])
	if t.currentTokenNo < len(currentLineToken)-1 {
		t.currentTokenNo++
	} else {
		if t.currentLine < len(t.lines)-1 {
			t.currentTokenNo = 0
			t.currentLine++
		}
	}
}
func (t *TokenGen) Back() {
	if t.currentTokenNo != 0 {
		t.currentTokenNo--
	} else {
		if t.currentLine != 0 {
			t.currentLine--
			currentLineToken := Tokenize(t.lines[t.currentLine])
			t.currentTokenNo = len(currentLineToken)
		}
	}
}

func (t *TokenGen) Peek(steps int) Token {
	var token Token
	for i := 0; i < steps; i++ {
		t.Next()
	}
	token = t.GetCurrentToken() // todo replace with getCurrentToken method
	for i := 0; i < steps; i++ {
		t.Back()
	}
	return token
}

func (t *TokenGen) Skip(steps int) Token {
	var token Token
	for i := 0; i < steps; i++ {
		t.Next()
	}
	token = t.GetCurrentToken() // todo replace with getCurrentToken method
	return token
}
func (t *TokenGen) GetCurrentToken() Token {
	if t.currentLine >= len(t.lines) || t.currentLine < 0 {
		return Token{}
	}
	return Tokenize(t.lines[t.currentLine])[t.currentTokenNo]
}
func (t *TokenGen) GetRemainingToken() []Token {
	tokensLeft := Tokenize(t.lines[t.currentLine])[t.currentTokenNo+1:]
	linesLeft := t.lines[t.currentLine+1:]
	for i := 0; i < len(linesLeft); i++ {
		lineTokens := Tokenize(linesLeft[i])
		tokensLeft = slices.Concat(tokensLeft, lineTokens)
	}
	return tokensLeft
}
func (t *TokenGen) GetTokenLeftLine() []Token {
	if t.currentLine >= len(t.lines) {
		return []Token{}
	}
	return Tokenize(t.lines[t.currentLine])[t.currentTokenNo+1:]
}
func (t *TokenGen) GetFullLineToken() []Token {
	return Tokenize(t.lines[t.currentLine])
}
