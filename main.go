package main

import (
	"fmt"
	"reflect"

	"github.com/MikeyA-yo/ay-go/parser"
)

func main() {
	tg := parser.NewTokenGen("l b = 12\nl a = 'Heyy'")
	fmt.Println(tg.GetCurrentToken())
	tg.Next()
	fmt.Println(tg.GetCurrentToken())
	pp := parser.NewParser("l b = 12-2\nl a = 'Heyy', ' yo'")
	pp.Start()
	for _, v := range pp.Nodes {
		PrintNonNilFields(v.Initializer)
	}
	// fmt.Println(tg.Peek(4), tg.GetRemainingToken())
}
func PrintNonNilFields(data interface{}) {
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	fmt.Println("Non-nil fields:")
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name

		// Check if field is nil (for pointers, slices, maps, and interfaces)
		if field.Kind() == reflect.Ptr || field.Kind() == reflect.Slice || field.Kind() == reflect.Map || field.Kind() == reflect.Interface {
			if field.IsNil() {
				continue
			}
		}

		// Print the field name and value
		fmt.Printf("%s: %v\n", fieldName, field.Interface())
	}
}
