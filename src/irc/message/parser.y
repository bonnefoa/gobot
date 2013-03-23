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
%token USER
%token JOIN
%token PRIVMSG
%token PONG
%token EOF

%type <val> WORD

%%

goal:
    NICK WORD
    {
        yylex.(*lex).m = MsgNick{Name:$2}
    }

|   PONG WORD
    {
        yylex.(*lex).m = MsgPong{$2}
    }
%%

type lex struct {
    tokens []token
    m interface{}
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

func ParseMessage(msg string) interface{} {
  log.Printf("Parsing %q", msg)
  res := gen_lex("NICK", msg)
  go res.run()
  tokens := []token{}
  for el := range res.tokens {
    tokens = append(tokens, el)
  }
  l := &lex{tokens:tokens}
  yyParse(l)
  return l.m
}
