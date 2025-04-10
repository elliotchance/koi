%{
package main
%}

// keywords
%token AND BREAK CONST CONTINUE ELSE IF FALSE MATCH MAP;
%token IMPORT IN IS FOR FUNC MUT NEW NOT OR RETURN TYPE TRUE;

// operators
%token EQUAL NOT_EQUAL LESS_EQUAL GREATER_EQUAL RANGE;

// dynamic
%token IDENTIFIER NUMBER STRING TAG;

%start program;

%left EQUAL NOT_EQUAL LESS LESS_EQUAL GREATER GREATER_EQUAL
%left AND OR NOT
%left PLUS MINUS TIMES DIVIDE MODULO

%%

program:
  | program import_stmt { file.Imports = append(file.Imports, $2.r.(string)) }
  | program func_stmt { file.Funcs = append(file.Funcs, $2.r.(*FuncStmt)) }
  | program var_stmt { file.Vars = append(file.Vars, $2.r.(VarStmt)) }
  | program type_stmt { file.Types = append(file.Types, $2.r.(TypeStmt)) }

type_stmt: // *TypeStmt
    TYPE IDENTIFIER '{' type_stmt_fields '}' {
      $$.r = &TypeStmt{Name: $2.r.(string), Fields: $4.r.([]*FuncType)}
    }
  | TYPE IDENTIFIER '{' '}' {
      $$.r = &TypeStmt{Name: $2.r.(string)}
    }
  | TYPE IDENTIFIER '[' IDENTIFIER ']' '=' type { /* TODO */ }
  | TYPE IDENTIFIER '=' type { /* TODO */ }

type_stmt_fields: // []*FuncType
    func_type { $$.r = []*FuncType{$1.r.(*FuncType)} }
  | type_stmt_fields func_type { $$.r = append($1.r.([]*FuncType), $2.r.(*FuncType)) }

import_stmt: // string
    IMPORT IDENTIFIER { $$.r = $2.r.(string) }

func_stmt: // *FuncStmt
    func_type block {
      $$.r = &FuncStmt{ FuncType: $1.r.(*FuncType), Stmts: $2.r.([]Stmt) }
    }

func_args: // []*FuncArg
    IDENTIFIER { $$.r = []*FuncArg{{Prefix: $1.r.(string)}} }
  | func_args_multi { $$.r = $1.r }

func_args_multi: // []*FuncArg
    func_args_term { $$.r = []*FuncArg{$1.r.(*FuncArg)} }
  | func_args_multi func_args_term {
      $$.r = append($1.r.([]*FuncArg), $2.r.(*FuncArg))
    }

func_args_term: // *FuncArg
    IDENTIFIER ':' type {
      $$.r = &FuncArg{Prefix: $1.r.(string), Name: $1.r.(string), Type: $3.r.(Type)}
    }
  | IDENTIFIER IDENTIFIER ':' type {
      $$.r = &FuncArg{Prefix: $1.r.(string), Name: $2.r.(string), Type: $4.r.(Type)}
    }

stmts: // []Stmt
    { $$.r = []Stmt(nil) }
  | stmts single_expr { $$.r = append($$.r.([]Stmt), ExprStmt{$2.r}) }
  | stmts var_stmt { $$.r = append($$.r.([]Stmt), $2.r) }
  | stmts for_stmt { $$.r = append($$.r.([]Stmt), $2.r) }
  | stmts for_range_stmt { $$.r = append($$.r.([]Stmt), $2.r) }
  | stmts assign_stmt { $$.r = append($$.r.([]Stmt), $2.r) }
  | stmts BREAK { $$.r = append($$.r.([]Stmt), BreakStmt{}) }
  | stmts if_stmt { $$.r = append($$.r.([]Stmt), $2.r) }
  | stmts CONTINUE { $$.r = append($$.r.([]Stmt), ContinueStmt{}) }
  | stmts RETURN expr { $$.r = append($$.r.([]Stmt), ReturnStmt{Expr: $3.r}) }
  // TODO add match

match_expr:
    MATCH expr '{' match_cases '}'

match_cases:
    match_case
  | match_cases match_case

match_case:
    expr block
  | TAG IDENTIFIER block

block:
    '{' stmts '}' { $$.r = $2.r.([]Stmt) }

call_expr: // *CallExpr
    '(' IDENTIFIER ')' {
      $$.r = &CallExpr{Args: []*CallExprArg{{Name: $2.r.(string)}}}
    }
  | '(' call_args ')' {
      $$.r = &CallExpr{Args: $2.r.([]*CallExprArg)}
    }

call_args: // []*CallExprArg
    IDENTIFIER ':' expr {
      $$.r = []*CallExprArg{{Name: $1.r.(string), Expr: $3.r}}
    }
  | call_args IDENTIFIER ':' expr {
      $$.r = append($1.r.([]*CallExprArg),
        &CallExprArg{Name: $2.r.(string), Expr: $4.r})
    }

var_stmt: // *VarStmt
    CONST assign_stmt { $$.r = &VarStmt{Mut: false, AssignStmt: $2.r.(*AssignStmt)} }
  | MUT assign_stmt { $$.r = &VarStmt{Mut: true, AssignStmt: $2.r.(*AssignStmt)} }

assign_stmt: // *AssignStmt
    IDENTIFIER '=' expr {
      $$.r = AssignStmt{Name: $1.r.(string), Expr: $3.r}
    }

for_stmt: // *ForStmt
    FOR expr block { $$.r = &ForStmt{Expr: $2.r, Stmts: $3.r.([]Stmt)} }

for_range_stmt: // *ForRangeStmt
    FOR IDENTIFIER IN expr RANGE expr block {
      $$.r = &ForRangeStmt{
        Name: $2.r.(string),
        From: $4.r,
        To: $6.r,
        Stmts: $7.r.([]Stmt),
      }
    }

if_stmt: // *IfStmt
    IF expr block { $$.r = &IfStmt{Expr: $2.r, Stmts: $3.r.([]Stmt)} }
  | IF IDENTIFIER IS IDENTIFIER block {
      $$.r = &IfStmt{
        Expr: IsExpr{$2.r.(string), $4.r.(string)},
        Stmts: $5.r.([]Stmt),
      }
    }
  | IF expr block ELSE block {
      $$.r = &IfStmt{Expr: $2.r, Stmts: $3.r.([]Stmt), Else: $5.r.([]Stmt) }
    }

key_value_exprs: // []*KeyValueExpr
    key_value_expr { $$.r = []*KeyValueExpr{$1.r.(*KeyValueExpr)} }
  | key_value_exprs ',' key_value_expr {
      $$.r = append($1.r.([]*KeyValueExpr), $3.r.(*KeyValueExpr))
    }

key_value_expr: // *KeyValueExpr
    IDENTIFIER ':' expr { $$.r = &KeyValueExpr{$1.r.(string), $3.r} }

// Expressions

expr: // any
    single_expr { $$.r = $1.r }
  | binary_expr { $$.r = $1.r }
  | NOT expr { $$.r = UnaryExpr{Expr: $2.r, Op: "!"} }
  | MINUS expr { $$.r = UnaryExpr{Expr: $2.r, Op: "-"} }

single_expr:
    value { $$.r = $1.r }
  // | '(' expr ')' { $$.r = $2.Expr }
  | single_expr call_expr {
      call := $2.r.(*CallExpr)
      call.On = $1.r
      $$.r = call
    }
  | match_expr { /* TODO */ }

binary_expr:
    expr PLUS expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "+"} }
  | expr MINUS expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "-"} }
  | expr TIMES expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "*"} }
  | expr DIVIDE expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "/"} }
  | expr MODULO expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "%"} }
  | expr AND expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "&&"} }
  | expr OR expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "||"} }
  | expr EQUAL expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "=="} }
  | expr NOT_EQUAL expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "!="} }
  | expr LESS expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "<"} }
  | expr LESS_EQUAL expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: "<="} }
  | expr GREATER expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: ">"} }
  | expr GREATER_EQUAL expr { $$.r = BinaryExpr{Left: $1.r, Right: $3.r, Op: ">="} }

expr_list:
    expr
  | expr_list ',' expr

// Values

value:
    IDENTIFIER { $$.r = IdentifierExpr($1.r.(string)) }
  | STRING { $$.r = StringExpr($1.r.(string)) }
  | NUMBER { $$.r = NumberExpr($1.r.(string)) }
  | boolean_value
  | array_value
  | map_value
  | object_value

boolean_value:
    TRUE { $$.r = BoolExpr(true) }
  | FALSE { $$.r = BoolExpr(false) }

// new Person{name: "Bob"}
// new thing{}
object_value:
    NEW IDENTIFIER '{' key_value_exprs '}' {
      $$.r = NewExpr{$2.r.(string), $4.r.([]KeyValueExpr)}
    }
  | NEW IDENTIFIER '{' '}'

// [1, 2, 3]
// [1, 2,]
// new []int
array_value:
    '[' expr_list optional_comma ']'
  | NEW array_type

// {a: "foo", b: "bar"}
// {1: 5, 7: 8,}
// new map[string]string
map_value:
    '{' key_value_exprs optional_comma '}'
  | NEW map_type

optional_comma:
  | ','

// Types

type: // Type
    single_type { $$.r = Type{$1.r.(*SingleType)} }
  | '(' sum_type ')' { $$.r = $2.r.([]*SingleType) }

array_type: // *ArrayType
    '[' ']' single_type {
      $$.r = &ArrayType{Element: $3.r.(*SingleType)}
    }

map_type: // *MapType
    MAP '[' IDENTIFIER ']' single_type {
      $$.r = &MapType{
        Key: &SingleType{Type: $3.r.(string)},
        Value: $5.r.(*SingleType),
      }
    }

single_type: // *SingleType
    IDENTIFIER { $$.r = &SingleType{Type: $1.r.(string)} }
  | array_type { $$.r = &SingleType{Array: $1.r.(*ArrayType)} }
  | map_type { $$.r = &SingleType{Map: $1.r.(*MapType)} }
  | func_type { $$.r = &SingleType{Func: $1.r.(*FuncType)} }

sum_type: // []*SingleType
    single_type { $$.r = []*SingleType{$1.r.(*SingleType)} }
  | sum_type '|' single_type {
      $$.r = append($1.r.([]*SingleType), $3.r.(*SingleType))
    }

func_type: // *FuncType
    FUNC '(' func_args ')' type {
      $$.r = &FuncType{Args: $3.r.([]*FuncArg), Return: $5.r.(Type)}
    }
  | FUNC IDENTIFIER '(' func_args ')' type {
      $$.r = &FuncType{Type: $2.r.(string), Args: $2.r.([]*FuncArg), Return: $4.r.(Type)}
    }

%%
