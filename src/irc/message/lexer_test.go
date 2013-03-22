package message

import "testing"
import "testing/assert"

func TestLexer(t *testing.T) {
  res := gen_lex("toto", "toto lol")
  go res.run()

  first := <-res.items
  assert.AssertEquals(t, first.typ, itemWord)
  assert.AssertEquals(t, first.val, "toto")
  second := <-res.items
  assert.AssertEquals(t, second.typ, itemWord)
  assert.AssertEquals(t, second.val, "lol")
  third := <-res.items
  assert.AssertEquals(t, third.typ, itemEOF)
}
