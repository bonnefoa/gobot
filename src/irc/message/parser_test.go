package message

import (
  "testing"
  "testing/assert"
)

func TestParse(t *testing.T) {
  msg := "NICK 4"
  res := gen_lex("NICK", msg)
  go res.run()
  tokens := []token{}
  for el := range res.tokens {
    tokens = append(tokens, el)
  }
  t.Logf("%s", tokens)

  l := &lex{tokens:tokens}

  yyParse(l)
  t.Logf("%s\n", l.m)
  assert.AssertEquals(t, l.m, "4")
}
