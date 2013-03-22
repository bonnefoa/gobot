package message

import (
  "testing"
)

func AssertEquals(t *testing.T, a interface{}, b interface{}) {
  if a != b {
    t.Logf("expected %v to equal %v", a, b)
    t.Fail()
  }
}

func TestMessage(t *testing.T) {
  a := MsgNick{ name:"to"}
  AssertEquals(t, a.String(), "NICK to\r\n")
}
