%{
package main
%}

// keywords
%token AND BREAK CONST CONTINUE ELSE IF MATCH MAP;
%token IMPORT IN IS FOR FUNC MUT NEW NOT OR RETURN TYPE;

// operators
%token OPEN_PAREN CLOSE_PAREN OPEN_CURLY CLOSE_CURLY OPEN_SQUARE CLOSE_SQUARE;
%token EQUAL NOT_EQUAL LESS LESS_EQUAL GREATER GREATER_EQUAL;
%token PLUS MINUS TIMES DIVIDE MODULO;
%token ASSIGN COLON RANGE COMMA PIPE;

// dynamic
%token IDENTIFIER NUMBER STRING BOOLEAN TAG;

%start program;

%left EQUAL NOT_EQUAL LESS LESS_EQUAL GREATER GREATER_EQUAL
%left AND OR NOT
%left PLUS MINUS TIMES DIVIDE MODULO

%%

program:
  | program import_stmt { file.Imports = append(file.Imports, $2.String) }
  | program func_stmt { file.Funcs = append(file.Funcs, $2.Func) }
  | program var_stmt { file.Vars = append(file.Vars, $2.Stmt.(VarStmt)) }
  | program type_stmt { file.Types = append(file.Types, $2.Stmt.(TypeStmt)) }

type_stmt:
    TYPE IDENTIFIER OPEN_CURLY type_stmt_fields CLOSE_CURLY {
      $$.Stmt = TypeStmt{Name: $2.String, Fields: $4.FuncTypes}
    }
  | TYPE IDENTIFIER OPEN_CURLY CLOSE_CURLY { $$.Stmt = TypeStmt{Name: $2.String} }
  | TYPE IDENTIFIER ASSIGN sum_type

type_stmt_fields:
    OPEN_SQUARE func_args CLOSE_SQUARE type { $$.FuncTypes = $1.FuncTypes }
  | type_stmt_fields OPEN_SQUARE func_args CLOSE_SQUARE type { $$.FuncTypes = append($1.FuncTypes, $2.FuncTypes...) }

import_stmt:
    IMPORT IDENTIFIER { $$.String = $2.String }

func_stmt:
    FUNC IDENTIFIER OPEN_SQUARE func_args CLOSE_SQUARE type OPEN_CURLY stmts CLOSE_CURLY {
      // $3.FuncTypes[0].Type = $2.String
      // $$.Func = &Func{FuncType: $3.FuncTypes[0], Stmts: $5.Stmts}
    }

func_args:
    IDENTIFIER { $$.FuncArgs = []FuncArg{{Prefix: $1.String}} }
  | func_args_2 { $$.FuncArgs = $1.FuncArgs }

func_args_2:
    IDENTIFIER COLON IDENTIFIER {
      $$.FuncArgs = []FuncArg{{Prefix: $1.String, Name: $1.String, Type: $3.String}}
    }
  | IDENTIFIER IDENTIFIER COLON IDENTIFIER {
      $$.FuncArgs = []FuncArg{{Prefix: $1.String, Name: $2.String, Type: $4.String}}
    }
  | func_args_2 IDENTIFIER COLON IDENTIFIER {
      $$.FuncArgs = append($1.FuncArgs,
        FuncArg{Prefix: $2.String, Name: $2.String, Type: $4.String})
    }
  | func_args_2 IDENTIFIER IDENTIFIER COLON IDENTIFIER {
      $$.FuncArgs = append($1.FuncArgs,
        FuncArg{Prefix: $2.String, Name: $3.String, Type: $5.String})
    }

stmts:
    { $$.Stmts = nil }
  | stmts expr_base { $$.Stmts = append($$.Stmts, ExprStmt{$2.Expr}) }
  | stmts var_stmt { $$.Stmts = append($$.Stmts, $2.Stmt) }
  | stmts for_stmt { $$.Stmts = append($$.Stmts, $2.Stmt) }
  | stmts assign_stmt { $$.Stmts = append($$.Stmts, $2.Stmt) }
  | stmts BREAK { $$.Stmts = append($$.Stmts, BreakStmt{}) }
  | stmts if_stmt { $$.Stmts = append($$.Stmts, $2.Stmt) }
  | stmts CONTINUE { $$.Stmts = append($$.Stmts, ContinueStmt{}) }
  | stmts RETURN expr { $$.Stmts = append($$.Stmts, ReturnStmt{Expr: $3.Expr}) }

match_expr:
    MATCH expr OPEN_CURLY match_cases CLOSE_CURLY

match_cases:
    match_case
  | match_cases match_case

match_case:
    expr block
  | TAG IDENTIFIER block

block:
    OPEN_CURLY stmts CLOSE_CURLY

call_expr:
    OPEN_SQUARE IDENTIFIER CLOSE_SQUARE {
      $$.Expr = &CallExpr{
        Args: []*CallExprArg{{Name: $2.String}},
      }
    }
  | OPEN_SQUARE call_args CLOSE_SQUARE {
      $$.Expr = &CallExpr{
        Args: $2.CallExprArgs,
      }
    }

call_args:
    IDENTIFIER COLON expr {
      $$.CallExprArgs = []*CallExprArg{{Name: $1.String, Expr: $3.Expr}}
    }
  | call_args IDENTIFIER COLON expr {
      $$.CallExprArgs = append($1.CallExprArgs, &CallExprArg{Name: $2.String, Expr: $4.Expr})
    }

var_stmt:
    CONST assign_stmt { $$.Stmt = VarStmt{Mut: false, AssignStmt: $2.Stmt.(AssignStmt)} }
  | MUT assign_stmt { $$.Stmt = VarStmt{Mut: true, AssignStmt: $2.Stmt.(AssignStmt)} }

assign_stmt:
    IDENTIFIER ASSIGN expr { $$.Stmt = AssignStmt{Name: $1.String, Expr: $3.Expr} }

for_stmt:
    FOR expr OPEN_CURLY stmts CLOSE_CURLY {
      $$.Stmt = ForStmt{Expr: $2.Expr, Stmts: $4.Stmts}
    }
  | FOR IDENTIFIER IN expr RANGE expr OPEN_CURLY stmts CLOSE_CURLY {
      $$.Stmt = ForRangeStmt{Name: $2.String, From: $4.Expr, To: $6.Expr, Stmts: $8.Stmts}
    }
  | FOR OPEN_CURLY stmts CLOSE_CURLY {
      $$.Stmt = ForStmt{Stmts: $3.Stmts}
    }

if_stmt:
    IF expr OPEN_CURLY stmts CLOSE_CURLY {
      $$.Stmt = IfStmt{Expr: $2.Expr, Stmts: $4.Stmts}
    }
  | IF IDENTIFIER IS IDENTIFIER OPEN_CURLY stmts CLOSE_CURLY {
      $$.Stmt = IfStmt{
        Expr: IsExpr{$2.String, $4.String},
        Stmts: $6.Stmts,
      }
    }
  | IF expr OPEN_CURLY stmts CLOSE_CURLY ELSE OPEN_CURLY stmts CLOSE_CURLY {
      $$.Stmt = IfStmt{Expr: $2.Expr, Stmts: $4.Stmts, Elses: $8.Elses}
    }
  | IF expr OPEN_CURLY stmts CLOSE_CURLY else_ifs ELSE OPEN_CURLY stmts CLOSE_CURLY {
      $$.Stmt = IfStmt{Expr: $2.Expr, Stmts: $4.Stmts, Elses: $8.Elses}
    }

else_ifs:
    else_if { $$.Elses = $1.Elses }
  | else_ifs else_if { $$.Elses = append($1.Elses, $2.Elses...) }

else_if:
    ELSE IF expr OPEN_CURLY stmts CLOSE_CURLY {
      $$.Elses = append($$.Elses, Else{Expr: $3.Expr, Stmts: $5.Stmts})
    }

key_value_exprs:
    key_value_expr { $$.r = []KeyValueExpr{$1.r.(KeyValueExpr)} }
  | key_value_exprs COMMA key_value_expr { $$.r = lexAppend[KeyValueExpr]($1.r, $3.r) }

key_value_expr:
    IDENTIFIER COLON expr { $$.r = KeyValueExpr{$1.String, $3.Expr} }

value:
    IDENTIFIER { $$.Expr = IdentifierExpr($1.String) }
  | NEW IDENTIFIER OPEN_CURLY key_value_exprs CLOSE_CURLY {
      $$.Expr = NewExpr{$2.String, $4.r.([]KeyValueExpr)}
    }
  | STRING { $$.Expr = StringExpr($1.String) }
  | NUMBER { $$.Expr = NumberExpr($1.String) }
  | BOOLEAN { $$.Expr = BoolExpr($1.Bool) }

binary_expr:
    expr PLUS expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "+"} }
  | expr MINUS expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "-"} }
  | expr TIMES expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "*"} }
  | expr DIVIDE expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "/"} }
  | expr MODULO expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "%"} }
  | expr AND expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "&&"} }
  | expr OR expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "||"} }
  | expr EQUAL expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "=="} }
  | expr NOT_EQUAL expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "!="} }
  | expr LESS expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "<"} }
  | expr LESS_EQUAL expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "<="} }
  | expr GREATER expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: ">"} }
  | expr GREATER_EQUAL expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: ">="} }

expr_base:
    value { $$.Expr = $1.Expr }
  | OPEN_PAREN expr CLOSE_PAREN { $$.Expr = $2.Expr }
  | expr_base call_expr {
      call := $2.Expr.(*CallExpr)
      call.On = $1.Expr
      $$.Expr = call
    }
  | match_expr

expr:
    expr_base { $$.Expr = $1.Expr }
  | binary_expr { $$.Expr = $1.Expr }
  | NOT expr { $$.Expr = UnaryExpr{Expr: $2.Expr, Op: "!"} }
  | MINUS expr { $$.Expr = UnaryExpr{Expr: $2.Expr, Op: "+"} }
  | array
  | map

map:
    map_type OPEN_CURLY key_value_exprs CLOSE_CURLY
  | map_type OPEN_CURLY CLOSE_CURLY

array:
    array_type OPEN_CURLY expr_list CLOSE_CURLY
  | array_type OPEN_CURLY CLOSE_CURLY

expr_list:
    expr
  | expr_list COMMA expr

// Types

array_type:
    OPEN_SQUARE CLOSE_SQUARE IDENTIFIER

map_type:
    MAP OPEN_SQUARE IDENTIFIER CLOSE_SQUARE IDENTIFIER

single_type:
    IDENTIFIER
  | array_type
  | map_type

multi_type:
    single_type
  | OPEN_PAREN type_list CLOSE_PAREN

sum_type:
    multi_type
  | sum_type PIPE multi_type

type:
    sum_type
  | func_type

func_type:
    FUNC OPEN_SQUARE func_args CLOSE_SQUARE type {
      $$.FuncTypes = []FuncType{{Args: $2.FuncArgs, Return: $4.Type}}
    }

type_list:
    type
  | type_list COMMA type

%%
