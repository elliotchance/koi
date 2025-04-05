package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const header = `
type M func(...V) V

type V struct {
	N string
	V any
	F map[string]M
}

func (v V) String() string {
	return fmt.Sprintf("%v", v.V)
}

func (v V) C(method string) any {
	return v.F[method](v).V
}

func _static[T any](x T) V {
	return V{fmt.Sprintf("%T", x), x, nil}
}

func __static[T any](x T) M {
	return func(...V) V { return _static(x) }
}
`

type Compiler struct {
	file File
}

func (c *Compiler) CompileFile(file File) error {
	c.file = file
	p := "package main\n\n"

	for _, code := range file.Code {
		p += code
	}

	for _, imp := range file.Imports {
		if imp == "io" {
			p += "import \"fmt\"\n"
		} else {
			p += fmt.Sprintf("import \"%s\"\n", imp)
		}
	}

	p += header + "\n"

	for _, v := range file.Vars {
		p += c.compileStmt(v) + "\n\n"
	}

	for _, funcStmt := range file.Funcs {
		if funcStmt.FuncType.GoName(false) == "main" {
			p += "func main() {\n"
		} else {
			p += fmt.Sprintf("func %s(args ...V) V {\n", funcStmt.FuncType.GoName(true))
		}
		if funcStmt.FuncType.Type != "static" {
			p += fmt.Sprintf("\t%s := args[0]\n", funcStmt.FuncType.Type)
		}

		if funcStmt.FuncType.Args[0].Name != "" {
			for i, arg := range funcStmt.FuncType.Args {
				p += fmt.Sprintf("\t%s := args[%d]\n", arg.Name, i+1)
			}
		}

		p += c.compileStmts(funcStmt.Stmts)
		if funcStmt.FuncType.GoName(false) != "main" {
			p += "\treturn V{}\n"
		}
		p += "}\n\n"
	}

	err := os.WriteFile("out/main.go", []byte(p), 0755)
	if err != nil {
		return err
	}

	return nil
}

func (c *Compiler) compileStmts(stmts []Stmt) string {
	var lines []string
	for _, stmt := range stmts {
		lines = append(lines, c.compileStmt(stmt))
	}

	return strings.Join(lines, "\n") + "\n"
}

func (c *Compiler) compileStmt(stmt Stmt) string {
	switch s := stmt.(type) {
	case AssignStmt:
		return fmt.Sprintf("\t%s = %s", s.Name, c.compileExpr(s.Expr))
	case BreakStmt:
		return fmt.Sprintf("\tbreak")
	case ContinueStmt:
		return fmt.Sprintf("\tcontinue")
	case ForStmt:
		expr := ""
		if s.Expr != nil {
			expr = c.compileExpr(s.Expr)
		}
		return fmt.Sprintf("\tfor %s { %s }", expr, c.compileStmts(s.Stmts))
	case IfStmt:
		code := fmt.Sprintf("\tif %s {\n%s\n}", c.compileExpr(s.Expr), c.compileStmts(s.Stmts))
		for _, el := range s.Elses {
			if el.Expr == nil {
				code += fmt.Sprintf("else { %s }", c.compileStmts(el.Stmts))
			} else {
				code += fmt.Sprintf("else if %s { %s }", c.compileExpr(el.Expr),
					c.compileStmts(el.Stmts))
			}
		}

		return code
	case ForRangeStmt:
		return fmt.Sprintf("\tfor %s := %s; %s < %s; %s++ { %s }",
			s.Name, c.compileExpr(s.From), s.Name, c.compileExpr(s.To), s.Name,
			c.compileStmts(s.Stmts))
	case VarStmt:
		return fmt.Sprintf("\tvar %s = %s", s.AssignStmt.Name, c.compileExpr(s.AssignStmt.Expr))
	case ExprStmt:
		return c.compileExpr(s.Expr)
	case ReturnStmt:
		return "\treturn _static[float64](" + c.compileExpr(s.Expr) + ")\n"
	case CodeStmt:
		return s.Code
	}

	return fmt.Sprintf("ERROR: %T", stmt)
}

func (c *Compiler) compileExpr(expr Expr) string {
	switch e := expr.(type) {
	case StringExpr:
		if strings.Contains(string(e), "${") {
			re := regexp.MustCompile(`\$\{(.*?)\}`)
			var vals []string
			for _, val := range re.FindAllStringSubmatch(string(e), -1) {
				vals = append(vals, val[1])
			}

			return fmt.Sprintf("fmt.Sprintf(\"%s\", %s)",
				re.ReplaceAllString(string(e), "%v"),
				strings.Join(vals, ", "))
		}
		return fmt.Sprintf("V{V:\"%s\"}", string(e))
	case BoolExpr, NumberExpr, IdentifierExpr:
		return fmt.Sprintf("%v", e)
	case UnaryExpr:
		return fmt.Sprintf("(%s %s)", e.Op, c.compileExpr(e.Expr))
	case BinaryExpr:
		return fmt.Sprintf("(%s %s %s)", c.compileExpr(e.Left), e.Op, c.compileExpr(e.Right))
	case IsExpr:
		return fmt.Sprintf("%s.N == \"%s\"", e.Name, e.Type)
	case NewExpr:
		var fields []string
		for _, f := range e.Fields {
			fields = append(fields, fmt.Sprintf("\t\t\"%s\": __static[float64](%s)",
				f.Key, f.Value))
		}
		for _, f := range c.file.Funcs {
			if f.FuncType.Type == e.Type {
				fields = append(fields, fmt.Sprintf("\t\t\"%s\": %s", f.FuncType.GoName(false), f.FuncType.GoName(true)))
			}
		}
		return fmt.Sprintf("V{\"%s\", nil, map[string]M{\n%s,\n\t}}\n", e.Type, strings.Join(fields, ",\n"))
	case CallExpr:
		funcName := fmt.Sprintf("%s.C", e.On)
		if e.On == IdentifierExpr("static") {
			lookingFor := fmt.Sprintf("static(%s:)", e.Args[0].Name)
			for _, f := range c.file.Funcs {
				if lookingFor == f.FuncType.Prototype(false) {
					return fmt.Sprintf("\t%s(V{}, %s)", f.FuncType.GoName(true), c.compileExpr(e.Args[0].Expr))
				}
			}

			return "FIXME"
		}
		if e.On == IdentifierExpr("io") {
			return fmt.Sprintf("\tfmt.Println(%s)", c.compileExpr(e.Args[0].Expr))
		}
		return fmt.Sprintf("\t%s(\"%s\").(float64)", funcName, e.GoName())
	}

	return fmt.Sprintf("ERROR: %T", expr)
}
