package main

import (
	"os"
)

func main() {
	file, err := Parse(os.Args[1])
	if err != nil {
		panic(err)
	}

	c := &Compiler{}
	c.compileImport("koi")
	c.compileFile("", file)
	c.Finish()
}
