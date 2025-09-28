package parser

import "strings"

func compileNode(Node ASTNode) string {

	switch Node.Type {
	case VariableDeclaration:
		if Node.Initializer != nil {
			return "let " + Node.Identifier + " = " + compileNode(*Node.Initializer)
		} else {
			return "let " + Node.Identifier
		}
	case Literal:
		return Node.Value
	case FunctionDeclaration:
		params := Node.Params
		paramStrs := []string{}
		for _, param := range params {
			paramStrs = append(paramStrs, compileNode(param))
		}
		body := Node.Body
		bodyStrs := []string{}
		for _, stmt := range body {
			bodyStrs = append(bodyStrs, compileNode(stmt))
		}
		return "function " + Node.Identifier + "(" + strings.Join(paramStrs, ", ") + ")" + " {\n" + strings.Join(bodyStrs, "\n") + "\n}"
	case Return:
		if Node.Initializer != nil {
			return "return " + compileNode(*Node.Initializer)
		} else if Node.Value != "" {
			return "return " + Node.Value
		} else {
			return "return"
		}
	case Break:
		return "break"
	case Continue:
		return "continue"

	}
	return ""
}

func compileIfElse(node ASTNode) string {}
func compileLoop(node ASTNode) string   {}
func compileTest(node ASTNode) string {
	if node.Paren != nil && node.Paren.Left != nil && node.Paren.Operator != "" && node.Paren.Right != nil {
		return "(" + compileNode(*node.Paren.Left) + " " + node.Paren.Operator + " " + compileNode(*node.Paren.Right) + ")"
	}
	return ""
}
func CompileAST(ast []ASTNode) {

}
