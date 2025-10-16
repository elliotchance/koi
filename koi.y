%{
package main

import (
	"go/ast"
	"go/token"
)

%}

// keywords
%token AND BREAK CONST CONTINUE ELSE EXTERN IF FALSE MATCH MAP;
%token IMPORT IN IS FOR FUNC MUT NEW NOT OR RETURN TYPE TRUE;

// operators
%token EQUAL NOT_EQUAL LESS_EQUAL GREATER_EQUAL RANGE;

// dynamic
%token TYPE_NAME IDENTIFIER NUMBER STRING TAG STRING_START STRING_END;

%start program;

%left EQUAL NOT_EQUAL '<' LESS_EQUAL '>' GREATER_EQUAL
%left AND OR NOT
%left '+' '-'
%left '*' '/' '%'

%%

program:
    program_stmts { yyResult = $1.r.([]any) }

program_stmts:
    program_stmt { $$.r = []any{$1.r} }
  | program program_stmt { $$.r = append($1.r.([]any), $2.r) }

program_stmt:
    func_stmt { $$.r = $1.r }
  | import_stmt { $$.r = $1.r }
  | type_stmt

type_stmt:
    TYPE IDENTIFIER '(' opt_func_args ')' block {
      $$.r = &TypeStmt{$2.r.(string), $4.r.([]*Field), $6.r.([]any)}
    }

func_stmt:
    FUNC IDENTIFIER func_type block {
      $$.r = &FuncStmt{$2.r.(string), $3.r.(*Type), $4.r.([]any)}
    }

func_type:
    '(' opt_func_args ')' opt_type {
      $$.r = &Type{IsFunc: true, Args: $2.r.([]*Field), Returns: $4.r.(*Type)}
    }

opt_type:
    /* empty */ { $$.r = (*Type)(nil) }
  | type { $$.r = $1.r }

type:
    IDENTIFIER { $$.r = &Type{Name: $1.r.(string)} }
  | '[' ']' type
  | MAP '[' type ']' type
  | func_type { $$.r = $1.r }

opt_func_args:
    /* empty */ { $$.r = []*Field{} }
  | func_args { $$.r = $1.r }
  | func_args ',' { $$.r = $1.r }

func_args:
    func_arg { $$.r = []*Field{$1.r.(*Field)} }
  | func_args ',' func_arg { $$.r = append($1.r.([]*Field), $3.r.(*Field)) }

func_arg:
    IDENTIFIER type { $$.r = &Field{$1.r.(string), $2.r.(*Type)} }

import_stmt:
    IMPORT IDENTIFIER { $$.r = &ImportStmt{} }

block:
    '{' '}' { $$.r = []any(nil) }
  | '{' stmts '}' { $$.r = $2.r }

stmts:
    stmt { $$.r = []any{$1.r} }
  | stmts stmt { $$.r = append($1.r.([]any), $2.r) }

stmt:
    expr { $$.r = $1.r }
  | IDENTIFIER '=' expr { $$.r = &AssignStmt{$1.r.(string), $3.r} }
  | func_stmt { $$.r = $1.r }
  | return_stmt { $$.r = $1.r }

return_stmt:
    RETURN expr { $$.r = &ReturnStmt{$2.r} }

expr:
    postfix_expr { $$.r = $1.r }
  | expr '+' expr { $$.r = &BinaryExpr{$1.r, "+", $3.r} }
  | expr '-' expr { $$.r = &BinaryExpr{$1.r, "-", $3.r} }
  | expr '*' expr { $$.r = &BinaryExpr{$1.r, "*", $3.r} }
  | expr '/' expr { $$.r = &BinaryExpr{$1.r, "/", $3.r} }
  | expr '%' expr { $$.r = &BinaryExpr{$1.r, "%", $3.r} }
  | expr AND expr { $$.r = &BinaryExpr{$1.r, "and", $3.r} }
  | expr OR expr { $$.r = &BinaryExpr{$1.r, "or", $3.r} }
  | NOT expr { $$.r = &UnaryExpr{"not", $2.r} }

postfix_expr:
    primary_expr { $$.r = $1.r }
  | postfix_expr '.' identifier { $$.r = &DotExpr{$1.r, $3.r} }
  | postfix_expr '(' opt_expr_list ')' { $$.r = &CallExpr{$1.r, $3.r.([]any)} }

primary_expr:
    identifier { $$.r = $1.r }
  | STRING { $$.r = &StringExpr{Parts: []any{$1.r}} }
  | NUMBER {
      $$.r = &ast.BasicLit{
        Kind:  token.FLOAT,
        Value: $1.r.(string),
      }
    }
  | boolean_expr
  | string_expr
  | array_expr
  | object_expr

boolean_expr:
    TRUE { $$.r = true }
  | FALSE { $$.r = false }

string_expr:
    STRING_START string_parts STRING_END {
      $$.r = &StringExpr{Parts: $2.r.([]any)}
    }

string_parts:
    /* empty */ { $$.r = []any{} }
  | string_parts string_part { $$.r = append($1.r.([]any), $2.r) }

string_part:
    STRING { $$.r = $1.r }
  | '{' expr '}' { $$.r = $2.r }

identifier:
    IDENTIFIER { $$.r = &Identifier{$1.r.(string)} }

array_expr:
    '[' opt_expr_list ']'

object_expr:
    '{' '}'
  | '{' object_keys '}'

object_keys:
    object_key
  | object_keys ',' object_key

object_key:
    IDENTIFIER ':' expr
  | '[' expr ']' ':' expr

opt_expr_list:
    /* empty */ { $$.r = []any{} }
  | expr_list { $$.r = $1.r }

expr_list:
    expr { $$.r = []any{$1.r} }
  | expr_list ',' expr { $$.r = append($1.r.([]any), $3.r) }

%%
