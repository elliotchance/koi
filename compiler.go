package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const header = `
import . "github.com/elliotchance/koi/lib/koi"
`

type Compiler struct {
	file File
}

func (c *Compiler) CompileFile(file File) error {
	c.file = file

	c.file.Funcs = append(c.file.Funcs, &Func{
		FuncType: FuncType{
			Type:   "int",
			Args:   []FuncArg{{"float64", "", ""}},
			Return: []string{"float64"},
		},
	})

	err := c.resolveTypes()
	if err != nil {
		return err
	}

	p := "package main\n\n"

	for _, imp := range file.Imports {
		p += fmt.Sprintf("import \"github.com/elliotchance/koi/lib/%s\"\n", imp)
	}

	p += header + "\n"

	for _, v := range file.Vars {
		p += c.compileStmt(v, nil) + "\n\n"
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

		p += c.compileStmts(funcStmt.Stmts, funcStmt)
		if funcStmt.FuncType.GoName(false) != "main" {
			p += "\treturn V{}\n"
		}
		p += "}\n\n"
	}

	err = os.WriteFile("out/main.go", []byte(p), 0755)
	if err != nil {
		return err
	}

	return nil
}

func (c *Compiler) compileStmts(stmts []Stmt, fn *Func) string {
	var lines []string
	for _, stmt := range stmts {
		lines = append(lines, c.compileStmt(stmt, fn))
	}

	return strings.Join(lines, "\n") + "\n"
}

func (c *Compiler) compileStmt(stmt Stmt, fn *Func) string {
	switch s := stmt.(type) {
	case AssignStmt:
		code, _, _ := c.compileExpr(s.Expr, fn)
		return fmt.Sprintf("\t%s = %s", s.Name, code)
	case BreakStmt:
		return fmt.Sprintf("\tbreak")
	case ContinueStmt:
		return fmt.Sprintf("\tcontinue")
	case ForStmt:
		expr := ""
		if s.Expr != nil {
			expr, _, _ = c.compileExpr(s.Expr, fn)
		}
		return fmt.Sprintf("\tfor %s { %s }", expr, c.compileStmts(s.Stmts, fn))
	case IfStmt:
		expr, _, _ := c.compileExpr(s.Expr, fn)
		code := fmt.Sprintf("\tif %s {\n%s\n}", expr, c.compileStmts(s.Stmts, fn))
		for _, el := range s.Elses {
			if el.Expr == nil {
				code += fmt.Sprintf("else { %s }", c.compileStmts(el.Stmts, fn))
			} else {
				expr, _, _ := c.compileExpr(el.Expr, fn)
				code += fmt.Sprintf("else if %s { %s }", expr, c.compileStmts(el.Stmts, fn))
			}
		}

		return code
	case ForRangeStmt:
		from, _, _ := c.compileExpr(s.From, fn)
		to, _, _ := c.compileExpr(s.To, fn)
		return fmt.Sprintf("\tfor %s := %s; %s < %s; %s++ { %s }",
			s.Name, from, s.Name, to, s.Name,
			c.compileStmts(s.Stmts, fn))
	case VarStmt:
		expr, _, _ := c.compileExpr(s.AssignStmt.Expr, fn)
		return fmt.Sprintf("\tvar %s = %s", s.AssignStmt.Name, expr)
	case ExprStmt:
		expr, _, _ := c.compileExpr(s.Expr, fn)
		return expr
	case ReturnStmt:
		expr, _, _ := c.compileExpr(s.Expr, fn)
		return "\treturn " + expr + "\n"
	}

	return fmt.Sprintf("ERROR: %T", stmt)
}

func (c *Compiler) compileExpr(expr Expr, fn *Func) (string, string, error) {
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
				strings.Join(vals, ", ")), "string", nil
		}
		return fmt.Sprintf("Static_(\"%s\")", string(e)), "string", nil
	case BoolExpr:
		return fmt.Sprintf("Static_(%v)", e), "bool", nil
	case NumberExpr:
		if strings.Contains(string(e), ".") || strings.Contains(string(e), "e") {
			return fmt.Sprintf("Static_(%v)", e), "float64", nil
		}
		return fmt.Sprintf("Static_(%v)", e), "int", nil
	case IdentifierExpr:
		return string(e), fn.VarTypes[string(e)], nil
	case UnaryExpr:
		code, typ, err := c.compileExpr(e.Expr, fn)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("Static_(%s %s.V.(bool))", e.Op, code), typ, nil
	case BinaryExpr:
		left, leftType, err := c.compileExpr(e.Left, fn)
		if err != nil {
			return "", "", err
		}
		right, rightType, err := c.compileExpr(e.Right, fn)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("Static_(%s.V.(%s) %s %s.V.(%s))",
			left, leftType, e.Op, right, rightType), leftType, nil
	case IsExpr:
		return fmt.Sprintf("%s.N == \"%s\"", e.Name, e.Type), "bool", nil
	case NewExpr:
		var fields []string
		for _, f := range e.Fields {
			fields = append(fields, fmt.Sprintf("\t\t\"%s\": Static__[float64](%s)",
				f.Key, f.Value))
		}
		for _, f := range c.file.Funcs {
			if f.FuncType.Type == e.Type {
				fields = append(fields, fmt.Sprintf("\t\t\"%s\": %s", f.FuncType.GoName(false), f.FuncType.GoName(true)))
			}
		}
		return fmt.Sprintf("V{\"%s\", nil, map[string]M{\n%s,\n\t}}\n", e.Type, strings.Join(fields, ",\n")), "unknown5", nil
	case *CallExpr:
		if ex, ok := e.On.(IdentifierExpr); ok && fn.VarTypes[string(ex)] != "" {
			funcName := fmt.Sprintf("Koi_%s_%s", fn.VarTypes[string(ex)], e.GoName(false))
			a := []string{string(ex)}
			for _, arg := range e.Args {
				if arg.Expr == nil {
					continue
				}
				arg, _, err := c.compileExpr(arg.Expr, fn)
				if err != nil {
					return "", "", err
				}
				a = append(a, arg)
			}
			return fmt.Sprintf("%s(%s)", funcName, strings.Join(a, ", ")),
				c.file.GetFunc("int[float64]").FuncType.Return[0], nil
		}

		funcName := fmt.Sprintf("%s.C", e.On)
		if e.On == IdentifierExpr("static") {
			lookingFor := fmt.Sprintf("static(%s:)", e.Args[0].Name)
			for _, f := range c.file.Funcs {
				if lookingFor == f.FuncType.Prototype(false) {
					code, _, err := c.compileExpr(e.Args[0].Expr, fn)
					if err != nil {
						return "", "", err
					}
					return fmt.Sprintf("\t%s(V{}, %s)", f.FuncType.GoName(true), code), "unknown1", nil
				}
			}

			return "FIXME", "unknown2", nil
		}
		if e.On == IdentifierExpr("io") || e.On == IdentifierExpr("math") {
			code, _, err := c.compileExpr(e.Args[0].Expr, fn)
			if err != nil {
				return "", "", err
			}
			return fmt.Sprintf("\t%s.%s(%s)", e.On, e.GoName(true), code), "unknown3", nil
		}
		return fmt.Sprintf("\t%s(\"%s\")", funcName, e.GoName(false)), "float64", nil
	}

	return fmt.Sprintf("ERROR: %T", expr), "unknown6", nil
}

func (c *Compiler) typeOfExpr(expr Expr, fn *Func) (string, error) {
	switch e := expr.(type) {
	case StringExpr:
		return "string", nil
	case BoolExpr:
		return "bool", nil
	case NumberExpr:
		if strings.Contains(string(e), ".") || strings.Contains(string(e), "e") {
			return "float64", nil
		}
		return "int", nil
	case IdentifierExpr:
		return fn.VarTypes[string(e)], nil
	case UnaryExpr:
		return "bool", nil
	case BinaryExpr:
		return c.typeOfExpr(e.Left, fn)
	case IsExpr:
	case NewExpr:
	case *CallExpr:
		return "float64", nil
	}

	return fmt.Sprintf("unknown7 %T", expr), nil
}

func (c *Compiler) fixType(expr Expr, fn *Func) error {
	if expr, ok := expr.(*CallExpr); ok {
		for _, arg := range expr.Args {
			typ, err := c.typeOfExpr(arg.Expr, fn)
			if err != nil {
				return err
			}
			arg.Type = typ
			c.fixType(arg.Expr, fn)
		}
	}
	return nil
}

func (c *Compiler) resolveTypes() error {
	debug := true
	for _, funcStmt := range c.file.Funcs {
		if debug {
			fmt.Println(funcStmt.FuncType)
		}

		funcStmt.VarTypes = map[string]string{}

		for _, arg := range funcStmt.FuncType.Args {
			if arg.Type != "" {
				funcStmt.VarTypes[arg.Name] = arg.Type
			}
		}

		for _, stmt := range funcStmt.Stmts {
			// TODO(elliotchance): This is stupid to load in all globals into all
			// functions, but it's a hack for now.
			for _, v := range c.file.Vars {
				typ, err := c.typeOfExpr(v.AssignStmt.Expr, funcStmt)
				if err != nil {
					return err
				}
				funcStmt.VarTypes[v.AssignStmt.Name] = typ
			}

			if stmt, ok := stmt.(VarStmt); ok {
				typ, err := c.typeOfExpr(stmt.AssignStmt.Expr, funcStmt)
				if err != nil {
					return err
				}
				funcStmt.VarTypes[stmt.AssignStmt.Name] = typ
			}
			if stmt, ok := stmt.(ExprStmt); ok {
				c.fixType(stmt.Expr, funcStmt)
			}
		}

		if debug {
			for k, v := range funcStmt.VarTypes {
				fmt.Printf("  %s %s\n", k, v)
			}
		}
	}

	return nil
}
