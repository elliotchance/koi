package main

import (
	"os"
)

func main() {
	dat, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	yyParse(&lexer{s: string(dat)})

	c := Compiler{}
	c.CompileFile(file)
}
