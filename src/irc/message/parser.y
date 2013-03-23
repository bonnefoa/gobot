%{
package message

import (
    "log"
    "fmt"
    "strings"
)
%}

%union{
    tok int
    val string
    text string
}

%token WORD
%token NICK
%token USER
%token JOIN
%token PRIVMSG
%token PONG
%token PING
%token EOF
%token INVALID
%token QUIT

%type <val> WORD
%type <text> text

%%

goal:
    NICK text
    {
        yylex.(*lex).m = MsgNick{Name:$2}
    }
|   PING text
    {
        yylex.(*lex).m = MsgPing{$2}
    }
|   PONG text
    {
        yylex.(*lex).m = MsgPong{$2}
    }
|   QUIT text
    {
        yylex.(*lex).m = MsgQuit{$2}
    }
|   PRIVMSG WORD text
    {
        yylex.(*lex).m = MsgPrivate{$2, $3}
    }
|   JOIN text
    {
        yylex.(*lex).m = MsgJoin{$2}
    }
|   INVALID
    {
        yylex.(*lex).m = nil
    }

text:
    text WORD
    {
        s := []string{$1, $2}
        $$ = (strings.Join(s, " "))
    }
|   WORD
    {
        $$ = $1
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
