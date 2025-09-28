package parser

import (
	"strings"
)

// AST to JavaScript compiler for AY language
// This function takes an AST (array of nodes) and returns JavaScript code as a string
func CompileAST(ast []ASTNode) string {
	var compiled []string
	for _, node := range ast {
		compiled = append(compiled, compileNode(node))
	}
	return strings.Join(compiled, "\n")
}

func compileNode(node ASTNode) string {
	if node.Type == 0 && node.Value == "" {
		return ""
	}

	switch node.Type {
	case VariableDeclaration:
		if node.Initializer != nil {
			return "let " + node.Identifier + " = " + compileNode(*node.Initializer) + ";"
		} else {
			return "let " + node.Identifier + ";"
		}
	case LiteralD:
		return node.Value
	case FunctionDeclaration:
		var paramStrs []string
		for _, param := range node.Params {
			paramStrs = append(paramStrs, compileNode(param))
		}
		var bodyStrs []string
		for _, stmt := range node.Body {
			bodyStrs = append(bodyStrs, compileNode(stmt))
		}
		identifier := node.Identifier
		if identifier == "" {
			identifier = ""
		}
		return "function " + identifier + "(" + strings.Join(paramStrs, ", ") + ") {\n" + strings.Join(bodyStrs, "\n") + "\n}"
	case Return:
		if node.Initializer != nil {
			return "return " + compileNode(*node.Initializer) + ";"
		} else if node.Value != "" {
			return "return " + node.Value + ";"
		} else {
			return "return;"
		}
	case Break:
		return "break;"
	case Continue:
		return "continue;"
	case IfElse:
		return compileIfElse(node)
	case Loop:
		return compileLoop(node)
	case CallExpression:
		var argStrs []string
		for _, arg := range node.Args {
			argStrs = append(argStrs, compileNode(arg))
		}
		return node.Identifier + "(" + strings.Join(argStrs, ", ") + ")"
	default:
		// Handle binary, unary, array, etc.
		if node.Operator != "" && node.Left != nil && node.Right != nil {
			// Handle string concatenation and other binary operations
			left := compileNode(*node.Left)
			var right string
			if node.Right != nil && node.Right.Paren != nil {
				right = compileTest(*node.Right)
			} else {
				right = compileNode(*node.Right)
			}
			return left + " " + node.Operator + " " + right
		}
		if node.PostOp != "" && node.Identifier != "" {
			return node.Identifier + node.PostOp + ";"
		}
		if node.InfixOp != "" && node.Identifier != "" {
			return node.InfixOp + " " + node.Identifier + ";"
		}
		if len(node.Elements) > 0 {
			var elemStrs []string
			for _, elem := range node.Elements {
				elemStrs = append(elemStrs, compileNode(elem))
			}
			return "[" + strings.Join(elemStrs, ", ") + "]"
		}
		if node.Identifier != "" && len(node.Index) > 0 {
			var indexStrs []string
			for _, idx := range node.Index {
				indexStrs = append(indexStrs, compileNode(idx))
			}
			return node.Identifier + "[" + strings.Join(indexStrs, "][") + "]"
		}
		if node.Value != "" {
			return node.Value
		}
		return ""
	}
}

func compileIfElse(node ASTNode) string {
	var test string
	if node.Test != nil {
		test = compileTest(*node.Test)
	} else {
		test = ""
	}

	var cons []ASTNode
	if node.Consequent != nil {
		cons = node.Consequent.Body
	}

	var consStrs []string
	for _, stmt := range cons {
		consStrs = append(consStrs, compileNode(stmt))
	}

	code := "if (" + test + ") {\n" + strings.Join(consStrs, "\n") + "\n}"

	if node.Alternate != nil {
		if len(node.Alternate.Body) > 0 {
			var altStrs []string
			for _, stmt := range node.Alternate.Body {
				altStrs = append(altStrs, compileNode(stmt))
			}
			code += " else {\n" + strings.Join(altStrs, "\n") + "\n}"
		} else if node.Alternate.Type == IfElse {
			code += " else " + compileIfElse(*node.Alternate)
		}
	}

	return code
}

func compileTest(test ASTNode) string {
	if test.Paren != nil && test.Paren.Left != nil && test.Paren.Operator != "" && test.Paren.Right != nil {
		return "(" + compileNode(*test.Paren.Left) + " " + test.Paren.Operator + " " + compileNode(*test.Paren.Right) + ")"
	}
	if test.Operator != "" && test.Left != nil && test.Right != nil {
		return "(" + compileNode(*test.Left) + " " + test.Operator + " " + compileNode(*test.Right) + ")"
	}
	// Handle boolean values that come like { paren: "true" } or { paren: "false" }
	if test.Paren != nil && test.Paren.Value != "" {
		return "(" + test.Paren.Value + ")"
	}
	return compileNode(test)
}

func compileLoop(node ASTNode) string {
	// For loop
	if node.Initializer != nil && node.Test != nil && node.Upgrade != nil {
		init := compileNode(*node.Initializer)
		test := compileTest(*node.Test)
		upgrade := compileNode(*node.Upgrade)
		// Always remove trailing semicolon from upgrade
		upgrade = strings.TrimSuffix(upgrade, ";")

		var bodyStrs []string
		for _, stmt := range node.Body {
			bodyStrs = append(bodyStrs, compileNode(stmt))
		}

		return "for (" + init + " " + test + "; " + upgrade + ") {\n" + strings.Join(bodyStrs, "\n") + "\n}"
	}

	// While loop
	if node.Test != nil && len(node.Body) > 0 {
		test := compileTest(*node.Test)
		var bodyStrs []string
		for _, stmt := range node.Body {
			bodyStrs = append(bodyStrs, compileNode(stmt))
		}
		return "while " + test + " {\n" + strings.Join(bodyStrs, "\n") + "\n}"
	}

	return ""
}
