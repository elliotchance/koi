package main

import (
	"fmt"
	"strings"
)

var yySym yySymType

type File struct {
	Imports []string
	Funcs   []*Func
	Vars    []VarStmt
	Types   []TypeStmt
}

func (f *File) GetType(name string) TypeStmt {
	for _, t := range f.Types {
		if t.Name == name {
			return t
		}
	}

	return TypeStmt{}
}

func (f *File) GetFunc(name string) *Func {
	for _, f := range f.Funcs {
		parts := strings.Split(f.FuncType.String(), "]")
		if parts[0]+"]" == name {
			return f
		}
	}

	panic(name)

	return nil
}

var file File

type CallExprArg struct {
	Name string
	Expr Expr
	Type string
}

type CallExpr struct {
	On   Expr
	Args []*CallExprArg
}

func (c *CallExpr) GoName(any bool) string {
	var a []string
	for _, arg := range c.Args {
		if arg.Type == "" {
			a = append(a, arg.Name)
		} else {
			if any {
				a = append(a, fmt.Sprintf("%s_any", arg.Name))
			} else {
				a = append(a, fmt.Sprintf("%s_%s", arg.Name, arg.Type))
			}
		}
	}

	return strings.Join(a, "__")
}

type KeyValueExpr struct {
	Key   string
	Value Expr
}

type AssignStmt struct {
	Name string
	Expr Expr
}

type VarStmt struct {
	Mut        bool
	AssignStmt AssignStmt
}

type Expr any
type Stmt any
type ExprStmt struct {
	Expr Expr
}

type BreakStmt struct{}

type ContinueStmt struct{}

type ForStmt struct {
	Expr  Expr
	Stmts []Stmt
}

type Else struct {
	Expr  Expr
	Stmts []Stmt
}

type IfStmt struct {
	Expr  Expr
	Stmts []Stmt
	Elses []Else
}

type ReturnStmt struct {
	Expr Expr
}

type ForRangeStmt struct {
	Name  string
	From  Expr
	To    Expr
	Stmts []Stmt
}

type StringExpr string
type NumberExpr string
type BoolExpr bool
type IdentifierExpr string

type BinaryExpr struct {
	Left, Right Expr
	Op          string
}

type IsExpr struct {
	Name string
	Type string
}

type TypeStmt struct {
	Name   string
	Fields []FuncType
}

type UnaryExpr struct {
	Expr Expr
	Op   string
}

type NewExpr struct {
	Type   string
	Fields []KeyValueExpr
}

type FuncArg struct {
	Prefix string
	Name   string
	Type   string
}

type Type []string

type FuncType struct {
	Type   string
	Args   []FuncArg
	Return Type
}

func (f FuncType) String() string {
	var a []string
	for _, arg := range f.Args {
		if arg.Type == "" {
			a = append(a, arg.Prefix)
		} else {
			a = append(a, fmt.Sprintf("%s:%s", arg.Prefix, arg.Type))
		}
	}
	if len(f.Return) == 0 {
		return f.Type + "[" + strings.Join(a, " ") + "]"
	}
	return f.Type + "[" + strings.Join(a, " ") + "] " + strings.Join(f.Return, " | ")
}

func (f FuncType) Prototype(includeTypes bool) string {
	var a []string
	for _, arg := range f.Args {
		if arg.Type == "" {
			a = append(a, arg.Prefix)
		} else {
			if includeTypes {
				a = append(a, fmt.Sprintf("%s:%s", arg.Prefix, arg.Type))
			} else {
				a = append(a, fmt.Sprintf("%s:", arg.Prefix))
			}
		}
	}

	return f.Type + "(" + strings.Join(a, " ") + ")"
}

func (f FuncType) GoName(includeType bool) string {
	var a []string
	for _, arg := range f.Args {
		if arg.Type == "" {
			a = append(a, arg.Prefix)
		} else {
			a = append(a, fmt.Sprintf("%s_%s", arg.Prefix, arg.Type))
		}
	}

	if includeType {
		return f.Type + "__" + strings.Join(a, "__")
	}
	return strings.Join(a, "__")
}

type Func struct {
	FuncType FuncType
	Stmts    []Stmt
	VarTypes map[string]string
}

type yySymType struct {
	// NUMBER STRING IDENTIFIER
	String string
	// BOOLEAN
	Bool bool

	Expr         Expr
	Stmt         Stmt
	Func         *Func
	Stmts        []Stmt
	Elses        []Else
	FuncArgs     []FuncArg
	CallExprArgs []*CallExprArg
	Type         Type
	FuncTypes    []FuncType
	Vars         []VarStmt

	r   any
	yys int
}

type lexer struct {
	s   string
	pos int
}

func isWordChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\n'
}

func isNumberChar(c byte) bool {
	return (c >= '0' && c <= '9') || c == '.' || c == 'e'
}

func (l *lexer) Lex(lval *yySymType) (result int) {
	for ; l.pos < len(l.s) && isWhitespace(l.s[l.pos]); l.pos++ {
	}

	if l.pos >= len(l.s) {
		return 0
	}

	l.pos++
	lval.yys = l.pos
	switch l.s[l.pos-1] {
	case ':':
		return COLON
	case '%':
		return MODULO
	case '[':
		return OPEN_SQUARE
	case ']':
		return CLOSE_SQUARE
	case '(':
		return OPEN_PAREN
	case ')':
		return CLOSE_PAREN
	case '{':
		return OPEN_CURLY
	case '}':
		return CLOSE_CURLY
	case '+':
		return PLUS
	case '-':
		return MINUS
	case '*':
		return TIMES
	case '/':
		return DIVIDE
	case ',':
		return COMMA
	case '|':
		return PIPE
	case '=':
		if l.pos < len(l.s) && l.s[l.pos] == '=' {
			l.pos++
			return EQUAL
		}
		return ASSIGN
	case '!':
		if l.pos < len(l.s) && l.s[l.pos] == '=' {
			l.pos++
			return NOT_EQUAL
		}
	case '<':
		if l.pos < len(l.s) && l.s[l.pos] == '=' {
			l.pos++
			return LESS_EQUAL
		}
		return LESS
	case '>':
		if l.pos < len(l.s) && l.s[l.pos] == '=' {
			l.pos++
			return GREATER_EQUAL
		}
		return GREATER
	case '.':
		if l.pos < len(l.s) && l.s[l.pos] == '.' {
			l.pos++
			return RANGE
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		s := string(l.s[l.pos-1])
		for ; l.pos <= len(l.s)-1 && isNumberChar(l.s[l.pos]); l.pos++ {
			s += string(l.s[l.pos])
		}
		lval.String = s
		return NUMBER
	case '"':
		{
			s := ""
			for ; l.pos <= len(l.s) && l.s[l.pos] != '"'; l.pos++ {
				s += string(l.s[l.pos])
			}
			l.pos++
			lval.String = s
			return STRING
		}
	case '@':
		{
			s := ""
			for ; l.pos <= len(l.s) && isWordChar(l.s[l.pos]); l.pos++ {
				s += string(l.s[l.pos])
			}
			l.pos++
			lval.String = s
			return TAG
		}
	}

	word := ""
	for ; l.pos <= len(l.s) && isWordChar(l.s[l.pos-1]); l.pos++ {
		word += string(l.s[l.pos-1])
	}
	l.pos--

	switch word {
	case "true":
		lval.Bool = true
		return BOOLEAN
	case "false":
		lval.Bool = false
		return BOOLEAN

	case "and":
		return AND
	case "break":
		return BREAK
	case "const":
		return CONST
	case "continue":
		return CONTINUE
	case "else":
		return ELSE
	case "for":
		return FOR
	case "func":
		return FUNC
	case "if":
		return IF
	case "import":
		return IMPORT
	case "in":
		return IN
	case "is":
		return IS
	case "mut":
		return MUT
	case "match":
		return MATCH
	case "map":
		return MAP
	case "new":
		return NEW
	case "not":
		return NOT
	case "or":
		return OR
	case "return":
		return RETURN
	case "type":
		return TYPE
	}

	lval.String = word
	return IDENTIFIER
}

func (l *lexer) Error(s string) {
	panic(fmt.Sprintf("%s at ...%s", s, l.s[l.pos:]))
}

func lexAppend[T any](a, b any) []T {
	if a == nil {
		return []T{b.(T)}
	}
	return append(a.([]T), b.(T))
}
