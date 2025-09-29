package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MikeyA-yo/ay-go/parser"
)

//go:embed functions/arr.js
var arrF string

//go:embed functions/mth.js
var mathF string

//go:embed functions/string.js
var stringF string

//go:embed functions/print.js
var printF string

//go:embed functions/fs.js
var fsF string

//go:embed functions/date.js
var dateF string

//go:embed functions/timer.js
var timeF string

//go:embed functions/http.js
var httpF string

const VERSION = "1.0.1"

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

Usage: ay-go <filename>
Example: ay-go myprogram.ay

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

	p := parser.NewParser(string(fileText))
	p.Start()

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

	// Generate output with embedded function libraries
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
