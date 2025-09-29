package parser

// ASTNode represents a node in the Abstract Syntax Tree
type ASTNode struct {
	// Base properties
	Type       int    `json:"type"`
	Name       string `json:"name,omitempty"`
	Value      string `json:"value,omitempty"`
	Raw        string `json:"raw,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	DataType   string `json:"dataType,omitempty"`

	// Expression properties
	Operator string   `json:"operator,omitempty"`
	Left     *ASTNode `json:"left,omitempty"`
	Right    *ASTNode `json:"right,omitempty"`

	// Variable and initialization
	Initializer *ASTNode `json:"initializer,omitempty"`

	// Block statements and functions
	Body   []ASTNode `json:"body,omitempty"`
	Params []ASTNode `json:"params,omitempty"`

	// If-else statements
	Test       *ASTNode `json:"test,omitempty"`
	Consequent *ASTNode `json:"consequent,omitempty"`
	Alternate  *ASTNode `json:"alternate,omitempty"`
	Paren      *ASTNode `json:"paren,omitempty"`

	// Arrays and indexing
	Elements []ASTNode `json:"elements,omitempty"`
	Index    []ASTNode `json:"index,omitempty"`

	// Function calls
	Args []ASTNode `json:"args,omitempty"`

	// Increment/Decrement
	PostOp  string `json:"postOp,omitempty"`
	InfixOp string `json:"infixOp,omitempty"`

	// Loop specific
	Upgrade *ASTNode `json:"upgrade,omitempty"`
}

// Variable represents a variable in the parser's context
type Variable struct {
	DataType string `json:"dataType"`
	Val      string `json:"val"`
	NodePos  int    `json:"nodePos"`
}
