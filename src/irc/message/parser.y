%{
package message

import (
    "fmt"
    "log"
)
%}

%union{
    typ int
    val string
}

%token NAME
%token EOF

%type <val> NAME
%type <eof> EOF

%%

goal:
        'NICK' NAME EOF
    {
        yylex.(*lex).r = $2
    }

%%

type lex struct {
    tokens []item
    r string
}

func (l *lex) Lex(lval *yySymType) int {
    if len(l.tokens) == 0 {
        return 0
    }
    v := l.tokens[0]
    l.tokens = l.tokens[1:]
    lval.val = v.val
    return int(v.typ)
}

func (l *lex) Error(e string) {
    log.Fatal(e)
}
