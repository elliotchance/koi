package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
)

type Compiler struct {
	types map[string]ast.Expr
}

func (c *Compiler) compileStmts(stmts []any) []ast.Stmt {
	var s []ast.Stmt
	for _, stmt := range stmts {
		s = append(s, c.compileStmt(stmt))
	}

	return s
}

func (c *Compiler) compileFuncStmt(funcStmt *FuncStmt) ast.Decl {
	stmts := c.compileThis(funcStmt.Type.Args)
	stmts = append(stmts, c.compileStmts(funcStmt.Block)...)

	return &ast.FuncDecl{
		Name: ast.NewIdent(funcStmt.Name),
		Type: &ast.FuncType{
			Params: c.compileFields(funcStmt.Type.Args),
		},
		Body: &ast.BlockStmt{List: stmts},
	}
}

func (c *Compiler) compileThis(args []*Field) []ast.Stmt {
	if len(args) == 0 {
		return nil
	}

	stmts := []ast.Stmt{
		// this := map[string]any{}
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("this"),
			},
			Tok: token.DEFINE, // :=
			Rhs: []ast.Expr{
				&ast.CompositeLit{
					Type: &ast.MapType{
						Key:   &ast.Ident{Name: "string"},
						Value: &ast.Ident{Name: "any"},
					},
					Elts: nil, // empty literal {}
				},
			},
		},
	}

	for _, arg := range args {
		stmts = append(stmts, &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.IndexExpr{
					X:     ast.NewIdent("this"),
					Index: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, arg.Name)},
				},
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				ast.NewIdent(arg.Name),
			},
		})
		c.types[arg.Name] = c.compileType(arg.Type)
	}

	return stmts
}

func (c *Compiler) compileTypeStmt(typeStmt *TypeStmt) ast.Decl {
	stmts := c.compileThis(typeStmt.Args)
	stmts = append(stmts, c.compileStmts(typeStmt.Block)...)
	stmts = append(stmts, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.Ident{Name: "this"},
		},
	})

	return &ast.FuncDecl{
		Name: ast.NewIdent(typeStmt.Name),
		Type: &ast.FuncType{
			Params: c.compileFields(typeStmt.Args),
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.MapType{
							Key:   ast.NewIdent("string"),
							Value: ast.NewIdent("any"),
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{List: stmts},
	}
}

func (c *Compiler) compileType(t *Type) ast.Expr {
	switch t.Name {
	case "float64":
		return ast.NewIdent(t.Name)
	}
	return &ast.MapType{
		Key:   ast.NewIdent("string"),
		Value: ast.NewIdent("any"),
	}
}

func (c *Compiler) compileField(f *Field) *ast.Field {
	names := []*ast.Ident{ast.NewIdent(f.Name)}
	typ := c.compileType(f.Type)

	if f.Type.IsFunc {
		var returns []*ast.Field
		if f.Type.Returns != nil {
			returns = []*ast.Field{{Type: ast.NewIdent(f.Type.Returns.Name)}}
		}

		typ = &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{List: returns},
		}
	}

	return &ast.Field{
		Names: names,
		Type:  typ,
	}
}

func (c *Compiler) compileFields(fields []*Field) *ast.FieldList {
	fieldList := &ast.FieldList{}
	for _, field := range fields {
		fieldList.List = append(fieldList.List, c.compileField(field))
	}
	return fieldList
}

func (c *Compiler) compileExprs(exprs []any) []ast.Expr {
	var result []ast.Expr
	for _, expr := range exprs {
		result = append(result, c.compileExpr(expr))
	}
	return result
}

func (c *Compiler) compileExpr(expr any) ast.Expr {
	switch e := expr.(type) {
	case *CallExpr:
		return &ast.CallExpr{
			Fun:  c.compileExpr(e.Expr),
			Args: c.compileExprs(e.Args),
		}

	case *DotExpr:
		if e.Expr.(*Identifier).Name == "io" {
			return ast.NewIdent(fmt.Sprintf("%s__%s", e.Expr.(*Identifier).Name, e.Identifier.(*Identifier).Name))
		}

		return &ast.TypeAssertExpr{
			X: &ast.IndexExpr{
				X:     c.compileExpr(e.Expr),
				Index: ast.NewIdent("\"" + e.Identifier.(*Identifier).Name + "\""),
			},
			Type: &ast.FuncType{
				Results: &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("float64")}}},
			},
		}

	case *BinaryExpr:
		var op token.Token
		switch e.Op {
		case "+":
			op = token.ADD
		case "-":
			op = token.SUB
		case "*":
			op = token.MUL
		case "/":
			op = token.QUO
		case "%":
			op = token.REM
		case "and":
			op = token.LAND
		case "or":
			op = token.LOR
		}
		return &ast.BinaryExpr{
			X:  c.compileExpr(e.Left),
			Op: op,
			Y:  c.compileExpr(e.Right),
		}

	case *UnaryExpr:
		return &ast.UnaryExpr{
			Op: token.NOT,
			X:  c.compileExpr(e.Expr),
		}

	case *FuncStmt:
		return &ast.FuncLit{
			Type: &ast.FuncType{
				Params: c.compileFields(e.Type.Args),
				Results: &ast.FieldList{
					List: []*ast.Field{
						{
							Type: ast.NewIdent("float64"),
						},
					},
				},
			},
			Body: &ast.BlockStmt{
				List: c.compileStmts(e.Block),
			},
		}

	case *Identifier:
		// if _, ok := c.types[e.Name]; !ok {
		// 	fmt.Printf("%s not in %+#v", e.Name, c.types)
		// }
		if e.Name == "rect" || e.Name == "measure" {
			return ast.NewIdent(e.Name)
		}
		return &ast.TypeAssertExpr{
			X: &ast.IndexExpr{
				X:     ast.NewIdent("this"),
				Index: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, e.Name)},
			},
			Type: c.types[e.Name],
		}

	case *StringExpr:
		call := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("fmt"),
				Sel: ast.NewIdent("Sprintf"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind: token.STRING,
				},
			},
		}

		fmtString := ""
		for _, p := range e.Parts {
			if s, ok := p.(string); ok {
				fmtString += s
			} else {
				fmtString += "%v"
				call.Args = append(call.Args, c.compileExpr(p))
			}
		}
		call.Args[0].(*ast.BasicLit).Value = "\"" + fmtString + "\""

		return call

	case bool:
		return ast.NewIdent(fmt.Sprintf("%v", e))

	case *ast.Ident:
		return e
	case *ast.BasicLit:
		return e
	}

	panic(expr)
}

func (c *Compiler) compileStmt(stmt any) ast.Stmt {
	switch s := stmt.(type) {
	case *AssignStmt:
		return &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.IndexExpr{
					X:     ast.NewIdent("this"),
					Index: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, s.Variable)},
				},
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				c.compileExpr(s.Expr),
			},
		}
	case *FuncStmt:
		return &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.IndexExpr{
					X:     ast.NewIdent("this"),
					Index: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, s.Name)},
				},
			},
			Tok: token.ASSIGN, // =
			Rhs: []ast.Expr{
				c.compileExpr(s),
			},
		}
	case *ReturnStmt:
		return &ast.ReturnStmt{
			Results: []ast.Expr{
				c.compileExpr(s.Expr),
			},
		}
	}

	return &ast.ExprStmt{
		X: c.compileExpr(stmt),
	}
}

func (c *Compiler) Compile(stmts []any) {
	fset := token.NewFileSet()

	// Build AST for: import "fmt"
	importSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"fmt"`,
		},
	}

	// Combine into *ast.File
	file := &ast.File{
		Name: ast.NewIdent("main"),
		Decls: []ast.Decl{
			&ast.GenDecl{
				Tok:   token.IMPORT,
				Specs: []ast.Spec{importSpec},
			},
			&ast.FuncDecl{
				Name: ast.NewIdent("io__printLine"),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("v")},
								Type:  ast.NewIdent("any"),
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent("fmt"),
									Sel: ast.NewIdent("Println"),
								},
								Args: []ast.Expr{
									ast.NewIdent("v"),
								},
							},
						},
					},
				},
			},
		},
	}

	for _, stmt := range stmts {
		c.types = map[string]ast.Expr{
			"r": &ast.MapType{
				Key:   ast.NewIdent("string"),
				Value: ast.NewIdent("any"),
			},
		}
		if stmt, ok := stmt.(*FuncStmt); ok {
			file.Decls = append(file.Decls, ast.Decl(c.compileFuncStmt(stmt)))
		}
		if stmt, ok := stmt.(*TypeStmt); ok {
			file.Decls = append(file.Decls, ast.Decl(c.compileTypeStmt(stmt)))
		}
	}

	// Create output file
	out, err := os.Create("tests/out.go")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Format and write source
	if err := format.Node(out, fset, file); err != nil {
		panic(err)
	}
}
