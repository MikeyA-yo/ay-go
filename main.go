package main

import (
	"ay/parser"
	"fmt"
)

func main() {
	tg := parser.NewTokenGen("l b = 12\nl a = 'Heyy'")
	fmt.Println(tg.GetCurrentToken())
	tg.Next()
	fmt.Println(tg.GetCurrentToken())
	pp := parser.NewParser("l b = 12\nl a = 'Heyy'")
	pp.Start()
	fmt.Println(pp.Nodes)
	// fmt.Println(tg.Peek(4), tg.GetRemainingToken())
}
