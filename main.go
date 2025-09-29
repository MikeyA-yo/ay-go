package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MikeyA-yo/ay-go/parser"
)

const VERSION = "1.0.0"

const AY_FancyName = `
   █████╗ ██╗   ██╗
  ██╔══██╗╚██╗ ██╔╝
  ███████║ ╚████╔╝ 
  ██╔══██║  ╚██╔╝  
  ██║  ██║   ██║   
  ╚═╝  ╚═╝   ╚═╝   
`

var welcome = fmt.Sprintf(`%s
AY Programming Language Compiler v%s

A modern, expressive programming language that compiles to JavaScript.
Features: Variables (l), Functions (f), Comments, Control Flow, Async Operations, and more!

Usage: ayc <filename>
Example: ayc myprogram.ay

Visit: https://github.com/MikeyA-yo/ay-go
`, AY_FancyName, VERSION)

func main() {
	// Check if filename is provided
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, welcome)
		fmt.Fprintln(os.Stderr, "⚠️  No filename provided")
		os.Exit(1)
	}

	fileName := os.Args[1]

	// Get current working directory and construct file path
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	filePath := filepath.Join(cwd, fileName)

	// Read the source file
	fileText, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", fileName, err)
		os.Exit(1)
	}

	// Check file extension
	fileNameParts := strings.Split(fileName, ".")
	if len(fileNameParts) < 2 || fileNameParts[len(fileNameParts)-1] != "ay" {
		fmt.Fprintln(os.Stderr, welcome)
		fmt.Fprintln(os.Stderr, "⚠️  Invalid file extension. Please use .ay files only.")
		os.Exit(1)
	}

	// Get functions directory relative to current working directory
	functionsDir := filepath.Join(cwd, "functions")

	// Read function files (with error handling if files don't exist)
	arrF := readFunctionFile(filepath.Join(functionsDir, "arr.js"))
	mathF := readFunctionFile(filepath.Join(functionsDir, "mth.js"))
	stringF := readFunctionFile(filepath.Join(functionsDir, "string.js"))
	printF := readFunctionFile(filepath.Join(functionsDir, "print.js"))
	fsF := readFunctionFile(filepath.Join(functionsDir, "fs.js"))
	dateF := readFunctionFile(filepath.Join(functionsDir, "date.js"))
	timeF := readFunctionFile(filepath.Join(functionsDir, "timer.js"))
	httpF := readFunctionFile(filepath.Join(functionsDir, "http.js"))

	// Parse the source code
	fmt.Printf("=== Starting Parser ===\n")
	p := parser.NewParser(string(fileText))
	fmt.Printf("=== Calling Start() ===\n")
	p.Start()
	fmt.Printf("=== Parser Finished ===\n")

	// Debug: Print AST structure as JSON
	fmt.Printf("\n=== AST DEBUG (JSON) ===\n")
	astJson, err := json.MarshalIndent(p.Nodes, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling AST to JSON: %v\n", err)
		// Fallback to simple debug
		for i, node := range p.Nodes {
			fmt.Printf("Node %d: Type=%d, Value='%s', Identifier='%s'\n", i, node.Type, node.Value, node.Identifier)
		}
	} else {
		fmt.Printf("%s\n", string(astJson))
	}
	fmt.Printf("=== END AST DEBUG ===\n\n")

	// Check for parsing errors
	if len(p.Errors) > 0 {
		fmt.Printf("Found %d parsing errors:\n", len(p.Errors))
		for _, error := range p.Errors {
			fmt.Printf("Error: %s\n", error)
		}
		fmt.Fprintf(os.Stderr, "%s Error encountered\nError compiling %s\n\n", AY_FancyName, fileName)
		fmt.Fprintln(os.Stderr, "Errors:")
		for _, error := range p.Errors {
			fmt.Fprintln(os.Stderr, error)
		}
		os.Exit(1)
	}

	// Compile AST to JavaScript
	compiled := parser.CompileAST(p.Nodes)

	// Generate output with function libraries
	output := fmt.Sprintf(`
%s
%s
%s
%s
%s
%s
%s
%s
%s
`, arrF, mathF, stringF, printF, fsF, dateF, timeF, compiled, httpF)

	// Generate output filename
	baseName := strings.Join(fileNameParts[:len(fileNameParts)-1], ".")
	// Remove path from baseName if present
	if strings.Contains(baseName, string(filepath.Separator)) {
		baseName = filepath.Base(baseName)
	}
	outputFileName := baseName + ".js"

	// Write output file
	err = os.WriteFile(outputFileName, []byte(output), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file %s: %v\n", outputFileName, err)
		os.Exit(1)
	}

	fmt.Printf("✅ Compiled %s to %s\n", fileName, outputFileName)
	fmt.Printf("🚀 Run with: node %s\n", outputFileName)
}

// Helper function to read function files with error handling
func readFunctionFile(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Return empty string if function file doesn't exist
		// You could also return a default implementation or log a warning
		return fmt.Sprintf("// Function file %s not found\n", filepath.Base(filePath))
	}
	return string(content)
}
