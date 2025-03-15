package main

import (
	"fmt"
	"os"
	"strings"
)

func Compile(file File) error {
	p := "package main\n\n"

	for _, imp := range file.Imports {
		if imp == "io" {
			p += "import \"fmt\"\n"
		} else {
			p += fmt.Sprintf("import \"%s\"\n", imp)
		}
	}
	p += "\n"

	for _, v := range file.Vars {
		p += compileStmt(v) + "\n\n"
	}

	for _, funcStmt := range file.Funcs {
		p += fmt.Sprintf("func %s() {\n", funcStmt.Name)
		p += compileStmts(funcStmt.Stmts)
		p += "}\n"
	}

	err := os.WriteFile("out/main.go", []byte(p), 0755)
	if err != nil {
		return err
	}

	return nil
}

func compileStmts(stmts []Stmt) string {
	var lines []string
	for _, stmt := range stmts {
		lines = append(lines, compileStmt(stmt))
	}

	return strings.Join(lines, "\n")
}

func compileStmt(stmt Stmt) string {
	switch s := stmt.(type) {
	case AssignStmt:
		return fmt.Sprintf("\t%s = %s", s.Name, compileExpr(s.Expr))
	case BreakStmt:
		return fmt.Sprintf("\tbreak")
	case ContinueStmt:
		return fmt.Sprintf("\tcontinue")
	case ForStmt:
		expr := ""
		if s.Expr != nil {
			expr = compileExpr(s.Expr)
		}
		return fmt.Sprintf("\tfor %s { %s }", expr, compileStmts(s.Stmts))
	case IfStmt:
		code := fmt.Sprintf("\tif %s { %s }", compileExpr(s.Expr), compileStmts(s.Stmts))
		for _, el := range s.Elses {
			if el.Expr == nil {
				code += fmt.Sprintf("else { %s }", compileStmts(el.Stmts))
			} else {
				code += fmt.Sprintf("else if %s { %s }", compileExpr(el.Expr),
					compileStmts(el.Stmts))
			}
		}

		return code
	case ForRangeStmt:
		return fmt.Sprintf("\tfor %s := %s; %s < %s; %s++ { %s }",
			s.Name, compileExpr(s.From), s.Name, compileExpr(s.To), s.Name,
			compileStmts(s.Stmts))
	case VarStmt:
		prefix := "const"
		if s.Mut {
			prefix = "var"
		}
		return fmt.Sprintf("\t%s %s = %s", prefix, s.AssignStmt.Name, compileExpr(s.AssignStmt.Expr))
	case ExprStmt:
		return compileExpr(s.Expr)
	}

	return fmt.Sprintf("ERROR: %T", stmt)
}

func compileExpr(expr Expr) string {
	switch e := expr.(type) {
	case StringExpr:
		return fmt.Sprintf("\"%s\"", string(e))
	case BoolExpr, NumberExpr, IdentifierExpr:
		return fmt.Sprintf("%v", e)
	case UnaryExpr:
		return fmt.Sprintf("(%s %s)", e.Op, compileExpr(e.Expr))
	case BinaryExpr:
		return fmt.Sprintf("(%s %s %s)", compileExpr(e.Left), e.Op, compileExpr(e.Right))
	case CallExpr:
		funcName := fmt.Sprintf("%s.%s", e.Package, e.Name)
		if e.Package == "io" && e.Name == "printLine" {
			funcName = "fmt.Println"
		}
		if e.Package == "math" && e.Name == "sin" {
			funcName = "math.Sin"
		}
		return fmt.Sprintf("\t%s(%s)", funcName, compileExpr(e.Args[0]))
	}

	return fmt.Sprintf("ERROR: %T", expr)
}
