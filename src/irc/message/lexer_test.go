package message

import "testing"
import "testing/assert"

func TestSimpleLexer(t *testing.T) {
  res := gen_lex("toto", "NICK lol\r\n")
  go res.run()

  first := <-res.tokens
  assert.AssertEquals(t, first.tok, NICK)
  assert.AssertEquals(t, first.val, "NICK")
  second := <-res.tokens
  assert.AssertEquals(t, second.tok, WORD)
  assert.AssertEquals(t, second.val, "lol")
  third := <-res.tokens
  assert.AssertEquals(t, third.tok, EOF)
}

func TestInvalidLexer(t *testing.T) {
  res := gen_lex("4", "4")
  go res.run()

  first := <-res.tokens
  assert.AssertEquals(t, first.tok, INVALID)
  assert.AssertEquals(t, first.val, "4")
}
