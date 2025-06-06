%{
package main
%}

// keywords
%token AND BREAK CONST CONTINUE ELSE EXTERN IF FALSE MATCH MAP;
%token IMPORT IN IS FOR FUNC MUT NEW NOT OR RETURN TYPE TRUE;

// operators
%token EQUAL NOT_EQUAL LESS_EQUAL GREATER_EQUAL RANGE;

// dynamic
%token IDENTIFIER NUMBER STRING TAG;

%start program;

%left EQUAL NOT_EQUAL '<' LESS_EQUAL '>' GREATER_EQUAL
%left AND OR NOT
%left '+' '-'
%left '*' '/' '%'

%%

program:
  | program import_stmt { yyFile.Imports = append(yyFile.Imports, $2.r.(string)) }
  | program func_stmt { yyFile.Funcs = append(yyFile.Funcs, $2.r.(*FuncStmt)) }
  | program var_stmt { yyFile.Vars = append(yyFile.Vars, $2.r.(*VarStmt)) }
  | program type_stmt { yyFile.Types = append(yyFile.Types, $2.r.(*TypeStmt)) }
  | program eos

eos: '\n' | ';'

type_stmt: // *TypeStmt
    TYPE generic_type '{' type_stmt_fields '}' {
      $$.r = &TypeStmt{Type: $2.r.(*SingleType), Fields: $4.r.([]*FuncType)}
    }
  | TYPE generic_type '=' type { /* TODO */ }

type_stmt_fields: // []*FuncType
    { $$.r = []*FuncType{} }
  | type_stmt_fields eos { $$.r = $1.r }
  | type_stmt_fields extern_func_type eos { $$.r = append($1.r.([]*FuncType), $2.r.(*FuncType)) }

import_stmt: // string
    IMPORT IDENTIFIER eos { $$.r = $2.r.(string) }

func_stmt: // *FuncStmt
    func_type block {
      $$.r = &FuncStmt{ FuncType: $1.r.(*FuncType), Stmts: $2.r.([]Stmt) }
    }
  | EXTERN func_type {
      funcType := $2.r.(*FuncType)
      funcType.Extern = true
      $$.r = &FuncStmt{ FuncType: funcType }
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

stmt: // Stmt
    single_expr { $$.r = &ExprStmt{$1.r} }
  | var_stmt { $$.r = $1.r }
  | for_stmt { $$.r = $1.r }
  | for_range_stmt { $$.r = $1.r }
  | assign_stmt { $$.r = $1.r }
  | if_stmt { $$.r = $1.r }
  | BREAK { $$.r = &BreakStmt{} }
  | CONTINUE { $$.r = &ContinueStmt{} }
  | RETURN expr { $$.r = &ReturnStmt{Expr: $2.r} }

stmts: // []Stmt
    { $$.r = []Stmt(nil) }
  | stmts eos { $$.r = $1.r }
  | stmts stmt eos { $$.r = append($$.r.([]Stmt), $2.r) }

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
      $$.r = &CallExpr{HasArgs: true, Args: $2.r.([]*CallExprArg)}
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
      $$.r = &AssignStmt{Name: $1.r.(string), Expr: $3.r}
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
        Expr: &IsExpr{$2.r.(string), $4.r.(string)},
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
  | '-' expr { $$.r = UnaryExpr{Expr: $2.r, Op: "-"} }

index_expr:
  '[' expr ']' { $$.r = $2.r }

single_expr:
    value { $$.r = $1.r }
  // | '(' expr ')' { $$.r = $2.Expr }
  | single_expr call_expr {
      call := $2.r.(*CallExpr)
      call.On = $1.r
      $$.r = call
    }
  | single_expr index_expr {
      $$.r = &IndexExpr{On: $1.r, Index: $2.r}
      // call := $2.r.(*CallExpr)
      // call.On = $1.r
      // $$.r = call
    }
  | match_expr { /* TODO */ }

binary_expr:
    expr '+' expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "+"} }
  | expr '-' expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "-"} }
  | expr '*' expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "*"} }
  | expr '/' expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "/"} }
  | expr '%' expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "%"} }
  | expr AND expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "&&"} }
  | expr OR expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "||"} }
  | expr EQUAL expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "=="} }
  | expr NOT_EQUAL expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "!="} }
  | expr '<' expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "<"} }
  | expr LESS_EQUAL expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: "<="} }
  | expr '>' expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: ">"} }
  | expr GREATER_EQUAL expr { $$.r = &BinaryExpr{Left: $1.r, Right: $3.r, Op: ">="} }

expr_list: // []Expr
    expr { $$.r = []Expr{$1.r} }
  | expr_list ',' expr { $$.r = append($1.r.([]Expr), $3.r) }

// Values

value:
    IDENTIFIER { $$.r = IdentifierExpr($1.r.(string)) }
  | STRING { $$.r = StringExpr($1.r.(string)) }
  | NUMBER { $$.r = NumberExpr($1.r.(string)) }
  | boolean_value { $$.r = $1.r }
  | array_value { $$.r = $1.r }
  | map_value { $$.r = $1.r }
  | object_value { $$.r = $1.r }

boolean_value:
    TRUE { $$.r = BoolExpr(true) }
  | FALSE { $$.r = BoolExpr(false) }

// new Person{name: "Bob"}
// new thing{}
object_value: // *NewExpr
    NEW generic_type '{' key_value_exprs '}' {
      $$.r = &NewExpr{$2.r.(*SingleType), $4.r.([]KeyValueExpr)}
    }
  | NEW generic_type '{' '}' { $$.r = &NewExpr{Type: $2.r.(*SingleType)} }

// [1, 2, 3]
// [1, 2,]
// new []int
array_value: // *ArrayValue
    '[' expr_list optional_comma ']' {
      $$.r = &ArrayValue{Elements: $2.r.([]Expr)}
    }
  // | NEW array_type { $$.r = &ArrayValue{Type: Type{&SingleType{Array: $2.r.(*ArrayType)}}} }

// {a: "foo", b: "bar"}
// {1: 5, 7: 8,}
// new map[string]string
map_value:
    '{' key_value_exprs optional_comma '}'
  // | NEW map_type

optional_comma:
  | ','

// Types

type: // Type
    single_type { $$.r = Type{$1.r.(*SingleType)} }
  | '(' sum_type ')' { $$.r = $2.r.([]*SingleType) }

// array_type: // *ArrayType
//     '[' ']' single_type {
//       $$.r = &ArrayType{Element: $3.r.(*SingleType)}
//     }

// map_type: // *MapType
//     MAP '[' IDENTIFIER ']' single_type {
//       $$.r = &MapType{
//         Key: &SingleType{Type: $3.r.(string)},
//         Value: $5.r.(*SingleType),
//       }
//     }

single_type: // *SingleType
    generic_type { $$.r = $1.r }
  // | array_type { $$.r = &SingleType{Array: $1.r.(*ArrayType)} }
  // | map_type { $$.r = &SingleType{Map: $1.r.(*MapType)} }
  | func_type { $$.r = &SingleType{Func: $1.r.(*FuncType)} }

generic_type: // *SingleType
    IDENTIFIER { $$.r = &SingleType{Type: $1.r.(string)} }
  | IDENTIFIER '[' IDENTIFIER ']' {
      $$.r = &SingleType{Type: $1.r.(string), Generics: []string{$3.r.(string)}}
    }

sum_type: // []*SingleType
    single_type { $$.r = []*SingleType{$1.r.(*SingleType)} }
  | sum_type '|' single_type {
      $$.r = append($1.r.([]*SingleType), $3.r.(*SingleType))
    }

extern_func_type:
    func_type { $$.r = $1.r }
  | EXTERN func_type {
      $2.r.(*FuncType).Extern = true
      $$.r = $2.r
    }

func_type: // *FuncType
    FUNC '(' func_args ')' type {
      $$.r = &FuncType{Args: $3.r.([]*FuncArg), Return: $5.r.(Type)}
    }
  | FUNC IDENTIFIER '(' func_args ')' type {
      $$.r = &FuncType{Type: $2.r.(string), Args: $4.r.([]*FuncArg), Return: $6.r.(Type)}
    }

%%
