package message

import (
  "testing"
  "testing/assert"
)

func TestParse(t *testing.T) {

  msg := "NICK 4"
  res := gen_lex("NICK", msg)
  go res.run()
  items := []item{}
  for el := range res.items {
    items = append(items, el)
  }
  t.Logf("%s", items)

  l := &lex{tokens:items}
  yyParse(l)
  assert.AssertEquals(t, l.r, "4")

  //fmt.Printf("Res : \n")
  //fmt.Printf("Res : %r\n", l.m)
  //fmt.Printf("Expected : %r\n", expected)
  //assert.AssertEquals(t, l.m, expected)

  t.Fail()
}
