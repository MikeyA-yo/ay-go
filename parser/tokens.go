package parser

import (
	"log"
	"regexp"
	"slices"
	"strings"
)

var Keywords = []string{
	"l",
	"define",
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

const (
	Identifier int = 0 + iota
	Operator
	Keyword
	Literal
	StringLiteral
	WhiteSpace
	Punctuation
	SingleLineComment
	Unkown
)

type Token struct {
	Type  int
	Value string
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
	//keeps track of current token and type
	currentToken := ""
	currentType := Identifier
	//keeps track of whether a string is open or not
	qChar := ""
	sOpen := false
	lineArr := strings.Split(line, "")
	for i := 0; i < len(lineArr); i++ {
		char := string(line[i])
		if (char == string('"') || char == "'") && currentType != SingleLineComment {
			qChar = char
			if sOpen {
				if string(currentToken[0]) == qChar {
					currentToken += qChar
					tokens = append(tokens, Token{currentType, currentToken})
					// cleanup
					currentToken = ""
					sOpen = false
				}
				// we know it's the start of a new string not end
			} else {
				currentType = StringLiteral
				sOpen = true
			}
		}
		if sOpen || currentType == SingleLineComment {
			currentToken += char
		}
		identTest := testRegex(`[a-zA-Z_@]`, char)
		opTest := testRegex(`[+*/%=<>&|!?^-]`, char)
		litTest := testRegex(`\d`, char)
		punctTest := testRegex(`[(){}[\]:;,.]`, char)
		if identTest && !sOpen && currentType != SingleLineComment {
			if currentType == Identifier {
				currentToken += char
			} else {
				currentType = Identifier
				currentToken = char
			}
			if len(lineArr)-1 >= i+1 {
				if !testRegex(`[a-zA-Z_@]`, string(lineArr[i+1])) {
					if isKeyword(currentToken) {
						tokens = append(tokens, Token{Keyword, currentToken})
						currentToken = ""
					} else {
						tokens = append(tokens, Token{currentType, currentToken})
						currentToken = ""
					}
				}
			}
		} else if testRegex(`\s`, char) && !sOpen && currentType != SingleLineComment {
			currentType = WhiteSpace
			if len(currentToken) > 0 && testRegex(`\s`, currentToken) {
				currentToken += char
			} else {
				currentToken = char
			}
			if len(lineArr)-1 >= i+1 {
				if !testRegex(`\s`, string(lineArr[i+1])) {
					tokens = append(tokens, Token{currentType, currentToken})
					currentToken = ""
				}
			}
		} else if opTest && !sOpen && currentType != SingleLineComment {
			currentType = Operator
			if len(currentToken) > 0 && testRegex(`[+*/%=<>&|!?-]`, currentToken) {
				switch len(currentToken) {
				case 1:
					if currentToken != "/" && currentToken != "^" {
						if currentToken == char {
							currentToken += char
						} else if char == "=" {
							currentToken += char
						} else {
							tokens = append(tokens, Token{currentType, currentToken})
							currentToken = char
						}
					} else {
						currentType = SingleLineComment
						currentToken += char
					}
				case 2:
					if (currentToken == ">>" || currentToken == "<<") && (char == ">" || char == "<") {
						currentToken += char
					}
				default:
					currentType = Unkown
					currentToken = char
				}
			} else {
				currentToken = char
			}
			if len(lineArr)-1 >= i+1 && currentType != SingleLineComment {
				if !testRegex(`[+*/%=<>&|!?-]`, string(line[i+1])) {
					tokens = append(tokens, Token{currentType, currentToken})
					currentToken = ""
				}
			}
		} else if litTest && !sOpen && currentType != SingleLineComment {
			currentType = Literal
			if len(currentToken) > 0 && (testRegex(`\d`, currentToken) || string(currentToken[len(currentToken)-1]) == ".") {
				currentToken += char
			} else {
				currentToken = char
			}
			if len(lineArr)-1 >= i+1 && currentType != SingleLineComment {
				if !testRegex(`\d`, string(line[i+1])) && string(line[i+1]) != "." {
					tokens = append(tokens, Token{currentType, currentToken})
					currentToken = ""
				}
			}
		} else if punctTest && !sOpen && currentType != SingleLineComment {
			if currentType == Literal && char == "." && !strings.Contains(currentToken, ".") && len(currentToken) > 0 {
				currentToken += char
			} else {
				currentType = Punctuation
				currentToken = char
				tokens = append(tokens, Token{currentType, currentToken})
				currentToken = ""
			}
		}
	}
	if currentToken != "" {
		tokens = append(tokens, Token{currentType, currentToken})
	}
	var returnTokens []Token
	for _, v := range tokens {
		if v.Type != WhiteSpace {
			returnTokens = append(returnTokens, v)
		}
	}
	return returnTokens
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
