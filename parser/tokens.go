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

func IsAllowedKeyAsVal(key string) bool {
	var allowedKeysAsVal = []string{"true", "false", "null", "f", "new"}
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
		if char == "*" && nextChar == "/" && currentType == MultiLineComment && !sOpen {
			currentToken += "*/"
			tokens = append(tokens, Token{currentType, currentToken, 0, 0})
			i++
			currentToken = ""
			currentType = Identifier
			continue
		}

		// New lines
		if (char == "\r" && nextChar == "\n") && currentType != MultiLineComment {
			if currentType == SingleLineComment {
				tokens = append(tokens, Token{currentType, currentToken, 0, 0})
			}
			tokens = append(tokens, Token{NewLine, "\r\n", 0, 0})
			i++
			currentToken = ""
			currentType = Identifier
			continue
		} else if char == "\n" && currentType != MultiLineComment {
			if currentType == SingleLineComment {
				tokens = append(tokens, Token{currentType, currentToken, 0, 0})
			}
			tokens = append(tokens, Token{NewLine, "\n", 0, 0})
			currentToken = ""
			currentType = Identifier
			continue
		}

		// this checks if it's a string quote character, controls the value of sOpen
		// notice how we also make sure we are not in a comment by checking the type
		if (char == string('"') || char == "'") && currentType != SingleLineComment && currentType != MultiLineComment {
			qChar = char
			if sOpen {
				if string(currentToken[0]) == qChar {
					currentToken += qChar
					tokens = append(tokens, Token{currentType, currentToken, 0, 0})
					// cleanup
					currentToken = ""
					sOpen = false
					currentType = Identifier
				}
			} else {
				sOpen = true
				currentToken = qChar
			}
		}

		// keep adding every character as a comment, but since we only expect a line this is fine as it continues to the end of the line
		//keep adding string characters until sOpen is false, i.e it's closed with the ending quotechar
		if sOpen || currentType == SingleLineComment || currentType == MultiLineComment {
			currentToken += char
			continue
		}

		identTest := testRegex(`[a-zA-Z_@]`, char)
		opTest := testRegex(`[+*/%=<>&|!?^-]`, char)
		litTest := testRegex(`\d`, char)
		punctTest := testRegex(`[(){}[\]:;,.]`, char)

		if identTest && !sOpen && currentType != SingleLineComment && currentType != MultiLineComment {
			if currentType == Identifier {
				currentToken += char
			} else {
				currentType = Identifier
				currentToken = char
			}
			//checks if it's the last character or not, passes if not last char
			if len(line)-1 >= i+1 {
				if !testRegex(`[a-zA-Z_@0-9]`, string(line[i+1])) {
					if isKeyword(currentToken) {
						tokens = append(tokens, Token{Keyword, currentToken, 0, 0})
						currentToken = ""
					} else {
						tokens = append(tokens, Token{currentType, currentToken, 0, 0})
						currentToken = ""
					}
				}
			}
		} else if testRegex(`\s`, char) && !sOpen && currentType != SingleLineComment && currentType != MultiLineComment {
			currentType = Whitespace
			if len(currentToken) > 0 && testRegex(`\s`, currentToken) {
				currentToken += char
			} else {
				currentToken = char
			}

			if len(line)-1 >= i+1 {
				if !testRegex(`\s`, string(line[i+1])) {
					tokens = append(tokens, Token{currentType, currentToken, 0, 0})
					currentToken = ""
				}
			}
		} else if opTest && !sOpen && currentType != SingleLineComment && currentType != MultiLineComment {
			currentType = Operator
			if len(currentToken) > 0 && testRegex(`[+*/%=<>&|!?^-]`, currentToken) {
				switch len(currentToken) {
				case 1:
					if currentToken == "/" && char == "/" {
						// This is a single line comment //
						currentType = SingleLineComment
						currentToken += char
					} else if (currentToken == "=" && char == "=") || (currentToken == "!" && char == "=") ||
						(currentToken == "<" && char == "=") || (currentToken == ">" && char == "=") ||
						(currentToken == "&" && char == "&") || (currentToken == "|" && char == "|") ||
						(currentToken == "+" && char == "+") || (currentToken == "-" && char == "-") {
						currentToken += char
						tokens = append(tokens, Token{currentType, currentToken, 0, 0})
						currentToken = ""
					} else {
						tokens = append(tokens, Token{currentType, currentToken, 0, 0})
						currentToken = char
					}
				case 2:
					if (currentToken == ">>" || currentToken == "<<") && (char == ">" || char == "<") {
						currentToken += char
						tokens = append(tokens, Token{currentType, currentToken, 0, 0})
					}
				default:
					tokens = append(tokens, Token{Unknown, currentToken, 0, 0})
					currentToken = char
				}
			} else {
				currentToken = char
			}

			if len(line)-1 >= i+1 && currentType != SingleLineComment && currentType != MultiLineComment {
				if !testRegex(`[+*/%=<>&|!?^-]`, string(line[i+1])) {
					tokens = append(tokens, Token{currentType, currentToken, 0, 0})
					currentToken = ""
				}
			}
		} else if litTest && !sOpen && currentType != SingleLineComment && currentType != MultiLineComment {
			// If we're already building an identifier, add the digit to it
			if currentType == Identifier {
				currentToken += char
				// check if next char would end the identifier
				if len(line)-1 >= i+1 {
					if !testRegex(`[a-zA-Z_@0-9]`, nextChar) {
						tokens = append(tokens, Token{Identifier, currentToken, 0, 0})
						currentToken = ""
					}
				}
			} else {
				currentType = Literal
				if len(currentToken) > 0 && (testRegex(`\d`, currentToken) || strings.HasSuffix(currentToken, ".")) {
					if testRegex(`\d`, currentToken) || strings.HasSuffix(currentToken, ".") {
						currentToken += char
					}
				} else {
					currentToken = char
				}
				if len(line)-1 >= i+1 {
					if !testRegex(`\d`, nextChar) && nextChar != "." {
						tokens = append(tokens, Token{currentType, currentToken, 0, 0})
						currentToken = ""
					}
				}
			}
		} else if punctTest && !sOpen && currentType != SingleLineComment && currentType != MultiLineComment {
			if currentType == Literal && char == "." && !strings.Contains(currentToken, ".") && len(currentToken) > 0 {
				currentToken += char
			} else {
				currentType = Punctuation
				currentToken = char
				tokens = append(tokens, Token{currentType, currentToken, 0, 0})
				currentToken = ""
			}
		}
	}

	// Handle any remaining token at the end
	if currentToken != "" {
		if currentType == Identifier && isKeyword(currentToken) {
			currentType = Keyword
		}
		tokens = append(tokens, Token{currentType, currentToken, 0, 0})
	}

	// Add EOF token
	tokens = append(tokens, Token{EOF, "", 0, 0})

	// Filter out whitespace and comment tokens
	var filteredTokens []Token
	for _, token := range tokens {
		if token.Type != Whitespace && token.Type != SingleLineComment && token.Type != MultiLineComment {
			filteredTokens = append(filteredTokens, token)
		}
	}

	return filteredTokens
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
