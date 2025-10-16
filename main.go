package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := Parse(os.Args[1])
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+#v\n", file)

	// compile()

	c := &Compiler{}
	c.Compile(yyResult)
	// c.compileImport("koi")
	// c.compileFile("", file)
	// c.Finish()
}
