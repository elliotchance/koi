package main

import (
	"fmt"
	"os"
)

var yySym yySymType
var yyResult []any

func Parse(filePath string) ([]any, error) {
	dat, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// yyResult = []any{}
	yyParse(&lexer{s: string(dat)})

	return yyResult, nil
}

type yySymType struct {
	r   any
	yys int
}

type lexer struct {
	s        string
	pos      int
	inString int // 0=no; 1=yes; 2=in-expr
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
	// defer func() {
	// 	if result > 256 {
	// 		fmt.Printf("%s %+#v\n", yyToknames[result-57346+3], lval.r)
	// 	} else {
	// 		fmt.Printf("'%c' %+#v\n", result, lval.r)
	// 	}
	// }()

	lval.r = nil

	// If we're in a string, we need to capture all raw string parts
	if l.inString == 1 {
		s := ""
		for ; l.pos <= len(l.s) && l.s[l.pos] != '"' && l.s[l.pos] != '{'; l.pos++ {
			s += string(l.s[l.pos])
		}
		if len(s) > 0 {
			lval.r = s
			return STRING
		}
	}

	for ; l.pos < len(l.s); l.pos++ {
		if isWhitespace(l.s[l.pos]) {
			continue
		}

		if l.s[l.pos] == '#' {
			for l.s[l.pos] != '\n' {
				l.pos++
			}
			continue
		}

		break
	}

	if l.pos >= len(l.s) {
		return 0
	}

	l.pos++
	lval.yys = l.pos
	switch l.s[l.pos-1] {
	case '+', '-', '*', '/', ':', '%', ',', '|', '(', ')', '[', ']', '&', '~':
		return int(l.s[l.pos-1])
	case '{':
		if l.inString == 1 {
			l.inString = 2
		}
		return int(l.s[l.pos-1])
	case '}':
		if l.inString == 2 {
			l.inString = 1
		}
		return int(l.s[l.pos-1])
	case '=':
		if l.pos < len(l.s) && l.s[l.pos] == '=' {
			l.pos++
			return EQUAL
		}
		return '='
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
		return '<'
	case '>':
		if l.pos < len(l.s) && l.s[l.pos] == '=' {
			l.pos++
			return GREATER_EQUAL
		}
		return '>'
	case '.':
		if l.pos < len(l.s) && l.s[l.pos] == '.' {
			l.pos++
			return RANGE
		}
		return '.'
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		s := string(l.s[l.pos-1])
		for ; l.pos <= len(l.s)-1 && isNumberChar(l.s[l.pos]); l.pos++ {
			s += string(l.s[l.pos])
		}
		lval.r = s
		return NUMBER
	case '"':
		if l.inString == 1 {
			l.inString = 0
			return STRING_END
		}
		l.inString = 1
		return STRING_START
		// case '@':
		// 	{
		// 		s := ""
		// 		for ; l.pos <= len(l.s) && isWordChar(l.s[l.pos]); l.pos++ {
		// 			s += string(l.s[l.pos])
		// 		}
		// 		l.pos++
		// 		lval.r = s
		// 		return TAG
		// 	}
	}

	word := ""
	for ; l.pos <= len(l.s) && isWordChar(l.s[l.pos-1]); l.pos++ {
		word += string(l.s[l.pos-1])
	}
	l.pos--

	switch word {
	case "and":
		return AND
	case "break":
		return BREAK
	case "continue":
		return CONTINUE
	case "else":
		return ELSE
	case "false":
		return FALSE
	case "for":
		return FOR
	case "func":
		return FUNC
	case "if":
		return IF
	case "import":
		return IMPORT
	case "is":
		return IS
	case "match":
		return MATCH
	case "map":
		return MAP
	case "not":
		return NOT
	case "or":
		return OR
	case "return":
		return RETURN
	case "true":
		return TRUE
	case "type":
		return TYPE
	}

	lval.r = word
	return IDENTIFIER
}

func (l *lexer) Error(s string) {
	panic(fmt.Sprintf("%s at ...%s", s, l.s[l.pos:]))
}

type DotExpr struct {
	Expr       any
	Identifier any
}

type CallExpr struct {
	Expr any
	Args []any
}

type ImportStmt struct{}

type FuncStmt struct {
	Name  string
	Type  *Type
	Block []any
}

type Field struct {
	Name string
	Type *Type
}

type AssignStmt struct {
	Variable string
	Expr     any
}

type TypeStmt struct {
	Name  string
	Args  []*Field
	Block []any
}

type ReturnStmt struct {
	Expr any
}

type BinaryExpr struct {
	Left  any
	Op    string
	Right any
}

type Identifier struct {
	Name string
}

type Type struct {
	Name    string
	IsFunc  bool
	Args    []*Field
	Returns *Type
}

type StringExpr struct {
	Parts []any
}

type UnaryExpr struct {
	Op   string
	Expr any
}
