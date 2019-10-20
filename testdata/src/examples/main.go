package main

import (
	"fmt"
	"strings"
)

func main() {
	// A call on another package's function should be replaced by one on a proxy variable.
	fmt.Printf("Hello, World!\n") // want `call of untestable function/method: fmt\.Printf`
	printf("Hello, World!\n")

	// A call on a builtin function is okay.
	a := make([]string, 1)
	a = append(a, "foo")

	// A call on a function within the same package is okay.
	private()

	// A call on another package's method should be replaced by one on a proxy variable.
	{
		var b strings.Builder
		b.WriteString("Hello, World!\n") // want `call of untestable function/method: \(\*strings.Builder\)\.WriteString`
		printf(b.String())               // want `call of untestable function/method: \(\*strings.Builder\)\.String`
	}
	{
		var b strings.Builder
		stringsBuilderWriteString(&b, "Hello, World!\n")
		printf(stringsBuilderString(&b))
	}
}

var printf = fmt.Printf

func private() {

}

var stringsBuilderWriteString = (*strings.Builder).WriteString
var stringsBuilderString = (*strings.Builder).String
