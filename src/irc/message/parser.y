%{
package message

import (
    "log"
    "fmt"
)
%}

%union{
    tok int
    val string
}

%token WORD
%token NICK
%token EOF

%type <val> WORD

%%

goal:
    NICK WORD
    {
        yylex.(*lex).m = $2
    }

%%

type lex struct {
    tokens []token
    m string
}

func (l *lex) Lex(lval *yySymType) int {
    if len(l.tokens) == 0 {
        return 0
    }
    v := l.tokens[0]
    l.tokens = l.tokens[1:]
    lval.val = v.val
    if ( v.tok == EOF ) {
        return 0
    }
    return v.tok
}

func (l *lex) Error(e string) {
    log.Fatal(e)
}
