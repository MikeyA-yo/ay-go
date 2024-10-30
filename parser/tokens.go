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
		}
	}
	if currentToken != "" {
		tokens = append(tokens, Token{currentType, currentToken})
	}
	return tokens
}
