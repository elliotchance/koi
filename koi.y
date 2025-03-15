%{
package main
%}

%token AND BREAK CONST CONTINUE ELSE IF IMPORT IN FOR FUNC MUT NOT OR;

%token OPEN_PAREN CLOSE_PAREN OPEN_CURLY CLOSE_CURLY;
%token EQUAL NOT_EQUAL LESS LESS_EQUAL GREATER GREATER_EQUAL;
%token PLUS MINUS TIMES DIVIDE MODULO;
%token ASSIGN COLON RANGE;

%token IDENTIFIER NUMBER STRING BOOLEAN;

%start program;

%left EQUAL NOT_EQUAL LESS LESS_EQUAL GREATER GREATER_EQUAL
%left AND OR NOT
%left PLUS MINUS TIMES DIVIDE MODULO

%%

program:
  | program import_stmt { file.Imports = append(file.Imports, $2.String) }
  | program func_stmt { file.Funcs = append(file.Funcs, $2.Func) }
  | program var_stmt { file.Vars = append(file.Vars, $2.Stmt.(VarStmt)) }

import_stmt:
    IMPORT IDENTIFIER { $$.String = $2.String }

func_stmt:
    FUNC OPEN_PAREN IDENTIFIER CLOSE_PAREN OPEN_CURLY stmts CLOSE_CURLY {
      $$.Func = Func{Name: $3.String, Stmts: $6.Stmts}
    }

stmts:
  | stmts expr { $$.Stmts = append($$.Stmts, ExprStmt{$2.Expr}) }
  | stmts var_stmt { $$.Stmts = append($$.Stmts, $2.Stmt) }
  | stmts for_stmt { $$.Stmts = append($$.Stmts, $2.Stmt) }
  | stmts assign_stmt { $$.Stmts = append($$.Stmts, $2.Stmt) }
  | stmts BREAK { $$.Stmts = append($$.Stmts, BreakStmt{}) }
  | stmts if_stmt { $$.Stmts = append($$.Stmts, $2.Stmt) }
  | stmts CONTINUE { $$.Stmts = append($$.Stmts, ContinueStmt{}) }

call_expr:
    IDENTIFIER OPEN_PAREN IDENTIFIER COLON expr CLOSE_PAREN {
      $$.Expr = CallExpr{
        Package: $1.String,
        Name: $3.String,
        Args: []Expr{$5.Expr},
      }
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

expr:
    STRING { $$.Expr = StringExpr($1.String) }
  | NUMBER { $$.Expr = NumberExpr($1.String) }
  | IDENTIFIER { $$.Expr = IdentifierExpr($1.String) }
  | BOOLEAN { $$.Expr = BoolExpr($1.Bool) }
  | call_expr { $$.Expr = $1.Expr }
  | NOT expr { $$.Expr = UnaryExpr{Expr: $2.Expr, Op: "!"} }
  | OPEN_PAREN expr CLOSE_PAREN { $$.Expr = $2.Expr }
  | expr PLUS expr { $$.Expr = BinaryExpr{Left: $1.Expr, Right: $3.Expr, Op: "+"} }
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

%%
