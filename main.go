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
	// pp := parser.NewParser("{buu, bb}")
	// fmt.Println(pp.GroupBy("{"))
	fmt.Println(tg.Peek(4), tg.GetRemainingToken())
}
