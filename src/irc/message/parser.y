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
    msg interface{}
    messages []interface{}
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
%token EOL
%token INVALID
%token COLUMN
%token QUIT

%type <val> WORD COLUMN PRIVMSG text
%type <msg> msg
%type <messages> messages

%%

goal:
         messages
{
            yylex.(*lex).m = $1
}

messages:
        messages msg
{
        $$ = append($1, $2)
}
|       msg
{
        m := []interface{}{$1}
        $$ = m
}

msg:
        NICK text
        {
            $$ = MsgNick{Name:$2}
        }
    |   PING COLUMN text
        {
            $$ = MsgPing{$3}
        }
    |   PONG text
        {
            $$ = MsgPong{$2}
        }
    |   QUIT text
        {
            $$ = MsgQuit{$2}
        }
    |   COLUMN WORD PRIVMSG WORD COLUMN text
        {
            $$ = MsgPrivate{$2, $4, $6}
        }
    |   JOIN text
        {
            $$ = MsgJoin{$2}
        }
    |   INVALID
        {
            $$ = nil
        }

text:
        text WORD EOF
        {
            s := []string{$1, $2}
            $$ = (strings.Join(s, " "))
        }
    |   text WORD
        {
            s := []string{$1, $2}
            $$ = (strings.Join(s, " "))
        }
    |   WORD EOF
        {
            $$ = $1
        }
    |   WORD
        {
            $$ = $1
        }


%%

type lex struct {
    tokens []token
    m []interface{}
}

func (l *lex) Lex(lval *yySymType) int {
    if len(l.tokens) == 0 {
        return 0
    }
    v := l.tokens[0]
    l.tokens = l.tokens[1:]
    lval.val = v.val
    if ( v.tok == EOL ) {
        return 0
    }
    return v.tok
}

func (l *lex) Error(e string) {
    log.Printf(e)
}

func ParseMessage(msg string) []interface{} {
  log.Printf("Parsing %q", msg)
  res := gen_lex("Irc parser", msg)
  go res.run()
  tokens := []token{}
  for el := range res.tokens {
    tokens = append(tokens, el)
  }
  l := &lex{tokens:tokens}
  yyParse(l)
  return l.m
}
