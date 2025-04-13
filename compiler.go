package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

type Compiler struct {
	Imports []string
	Vars    []*VarStmt
	Funcs   []*FuncStmt
}

const ImportPath = "lib"

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) GetFunc(name string) *FuncStmt {
	fmt.Printf("looking for %s in:\n", name)
	for _, f := range c.Funcs {
		fmt.Printf("  '%s'\n", f.FuncType.Prototype())
		if f.FuncType.Prototype() == name {
			return f
		}
	}

	panic(name)
}

func (c *Compiler) Finish() error {
	err := c.resolveTypes()
	if err != nil {
		return err
	}

	p := "package main\n\n"

	for _, imp := range c.Imports {
		p += fmt.Sprintf("import \"github.com/elliotchance/koi/lib/%s\"\n", imp)
	}

	for _, v := range c.Vars {
		p += c.compileStmt(v, nil) + "\n\n"
	}

	for _, funcStmt := range c.Funcs {
		if funcStmt.Extern {
			continue
		}

		if funcStmt.FuncType.String() == "(main)" {
			p += "func main() {\n"
		} else {
			p += fmt.Sprintf("func %s(args ...V) V {\n", funcStmt.FuncType.GoName())
		}

		if funcStmt.FuncType.Args[0].Name != "" {
			for i, arg := range funcStmt.FuncType.Args {
				p += fmt.Sprintf("\t%s := args[%d]\n", arg.Name, i+1)
			}
		}

		p += c.compileStmts(funcStmt.Stmts, funcStmt)
		if funcStmt.FuncType.String() != "(main)" {
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

func (c *Compiler) compileImport(imp string) error {
	c.Imports = append(c.Imports, imp)
	f, err := Parse(path.Join(ImportPath, imp, imp+".koi"))
	if err != nil {
		return err
	}
	err = c.compileFile(imp, f)
	if err != nil {
		return err
	}

	return nil
}

func (c *Compiler) compileFile(packageName string, file *File) error {
	c.Vars = append(c.Vars, file.Vars...)

	for _, fn := range file.Funcs {
		fn := *fn
		if fn.FuncType.Type == "" {
			fn.FuncType.Type = packageName
		}
		c.Funcs = append(c.Funcs, &fn)
	}

	for _, imp := range file.Imports {
		err := c.compileImport(imp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) compileStmts(stmts []Stmt, fn *FuncStmt) string {
	var lines []string
	for _, stmt := range stmts {
		lines = append(lines, c.compileStmt(stmt, fn))
	}

	return strings.Join(lines, "\n") + "\n"
}

func (c *Compiler) compileStmt(stmt Stmt, fn *FuncStmt) string {
	switch s := stmt.(type) {
	case *AssignStmt:
		code, _, _ := c.compileExpr(s.Expr, fn)
		return fmt.Sprintf("\t%s = %s", s.Name, code)
	case *BreakStmt:
		return fmt.Sprintf("\tbreak")
	case *ContinueStmt:
		return fmt.Sprintf("\tcontinue")
	case *ForStmt:
		expr := ""
		if s.Expr != nil {
			expr, _, _ = c.compileExpr(s.Expr, fn)
		}
		return fmt.Sprintf("\tfor %s { %s }", expr, c.compileStmts(s.Stmts, fn))
	case *IfStmt:
		expr, _, _ := c.compileExpr(s.Expr, fn)
		code := fmt.Sprintf("\tif %s {\n%s\n}", expr, c.compileStmts(s.Stmts, fn))
		if len(s.Else) > 0 {
			code += fmt.Sprintf("else { %s }", c.compileStmts(s.Else, fn))
		}

		return code
	case *ForRangeStmt:
		from, _, _ := c.compileExpr(s.From, fn)
		to, _, _ := c.compileExpr(s.To, fn)
		return fmt.Sprintf("\tfor %s := %s; %s < %s; %s++ { %s }",
			s.Name, from, s.Name, to, s.Name,
			c.compileStmts(s.Stmts, fn))
	case *VarStmt:
		expr, _, _ := c.compileExpr(s.AssignStmt.Expr, fn)
		return fmt.Sprintf("\tvar %s = %s", s.AssignStmt.Name, expr)
	case *ExprStmt:
		expr, _, _ := c.compileExpr(s.Expr, fn)
		return expr
	case *ReturnStmt:
		expr, _, _ := c.compileExpr(s.Expr, fn)
		return "\treturn " + expr + "\n"
	}

	return fmt.Sprintf("ERROR1: %T", stmt)
}

func (c *Compiler) compileExpr(expr Expr, fn *FuncStmt) (string, Type, error) {
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
				strings.Join(vals, ", ")), Type{{Type: "string"}}, nil
		}
		return fmt.Sprintf("koi.NewString(\"%s\")", string(e)), Type{{Type: "string"}}, nil
	case BoolExpr:
		return fmt.Sprintf("koi.Bool(%v)", e), Type{{Type: "bool"}}, nil
	case NumberExpr:
		if strings.Contains(string(e), ".") || strings.Contains(string(e), "e") {
			return fmt.Sprintf("koi.Number{V: %v}", e), Type{{Type: "float64"}}, nil
		}
		return fmt.Sprintf("koi.Number{V: %v}", e), Type{{Type: "int"}}, nil
	case IdentifierExpr:
		return string(e), fn.VarTypes[string(e)], nil
	case *UnaryExpr:
		code, typ, err := c.compileExpr(e.Expr, fn)
		if err != nil {
			return "", Type{}, err
		}
		return fmt.Sprintf("koi.Bool(%s %s.V.(bool))", e.Op, code), typ, nil
	case *BinaryExpr:
		left, leftType, err := c.compileExpr(e.Left, fn)
		if err != nil {
			return "", Type{}, err
		}
		right, rightType, err := c.compileExpr(e.Right, fn)
		if err != nil {
			return "", Type{}, err
		}
		if len(leftType) == 0 {
			panic("here")
		}
		if len(rightType) == 0 {
			panic("here")
		}
		return fmt.Sprintf("koi.Static_(%s.V.(%s) %s %s.V.(%s))",
			left, leftType[0].Type, e.Op, right, rightType[0].Type), leftType, nil
	case *IsExpr:
		return fmt.Sprintf("%s.N == \"%s\"", e.Name, e.Type), Type{{Type: "bool"}}, nil
	case *NewExpr:
		var fields []string
		for _, f := range e.Fields {
			fields = append(fields, fmt.Sprintf("\t\t\"%s\": koi.Static__[float64](%s)",
				f.Key, f.Value))
		}
		for _, f := range c.Funcs {
			if f.FuncType.Type == e.Type {
				fields = append(fields, fmt.Sprintf("\t\t\"%s\": %s", f.FuncType.GoName(), f.FuncType.GoName()))
			}
		}
		return fmt.Sprintf("V{\"%s\", nil, map[string]M{\n%s,\n\t}}\n", e.Type, strings.Join(fields, ",\n")),
			Type{{Type: "unknown5"}}, nil
	case *CallExpr:
		on, onType, err := c.compileExpr(e.On, fn)
		if err != nil {
			return "", Type{}, err
		}

		var args []string
		if len(e.Args) != 1 || e.Args[0].Expr != nil {
			for _, arg := range e.Args {
				a, _, err := c.compileExpr(arg.Expr, fn)
				if err != nil {
					return "", Type{}, err
				}

				args = append(args, a)
			}
		}

		fmt.Println("#", e.Prototype(onType.String()))
		return on + "." + e.GoName() + "(" + strings.Join(args, ", ") + ")",
			c.GetFunc(e.Prototype(onType.String())).FuncType.Return, nil
	}

	return fmt.Sprintf("ERROR2: %T", expr), Type{{Type: "unknown6"}}, nil
}

func (c *Compiler) typeOfExpr(expr Expr, fn *FuncStmt) (Type, error) {
	switch e := expr.(type) {
	case StringExpr:
		return Type{{Type: "string"}}, nil
	case BoolExpr:
		return Type{{Type: "bool"}}, nil
	case NumberExpr:
		if strings.Contains(string(e), ".") || strings.Contains(string(e), "e") {
			return Type{{Type: "float64"}}, nil
		}
		return Type{{Type: "int"}}, nil
	case IdentifierExpr:
		return fn.VarTypes[string(e)], nil
	case *UnaryExpr:
		return Type{{Type: "bool"}}, nil
	case *BinaryExpr:
		return c.typeOfExpr(e.Left, fn)
	case *IsExpr:
	case *NewExpr:
	case *CallExpr:
		return Type{{Type: "float64"}}, nil
	}

	return Type{{Type: fmt.Sprintf("unknown7 %T", expr)}}, nil
}

func (c *Compiler) fixType(expr Expr, fn *FuncStmt) error {
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
	for _, funcStmt := range c.Funcs {
		if debug {
			fmt.Println(funcStmt.FuncType)
		}

		funcStmt.VarTypes = map[string]Type{}

		for _, imp := range c.Imports {
			funcStmt.VarTypes[imp] = Type{&SingleType{Type: imp}}
		}

		for _, arg := range funcStmt.FuncType.Args {
			if len(arg.Type) != 0 {
				funcStmt.VarTypes[arg.Name] = arg.Type
			}
		}

		for _, stmt := range funcStmt.Stmts {
			// TODO(elliotchance): This is stupid to load in all globals into all
			// functions, but it's a hack for now.
			for _, v := range c.Vars {
				typ, err := c.typeOfExpr(v.AssignStmt.Expr, funcStmt)
				if err != nil {
					return err
				}
				funcStmt.VarTypes[v.AssignStmt.Name] = typ
			}

			if stmt, ok := stmt.(*VarStmt); ok {
				typ, err := c.typeOfExpr(stmt.AssignStmt.Expr, funcStmt)
				if err != nil {
					return err
				}
				funcStmt.VarTypes[stmt.AssignStmt.Name] = typ
			}
			if stmt, ok := stmt.(*ExprStmt); ok {
				c.fixType(stmt.Expr, funcStmt)
			}
		}

		if debug {
			for k, v := range funcStmt.VarTypes {
				fmt.Printf("  %s %s\n", k, v[0].Type)
			}
		}
	}

	return nil
}
