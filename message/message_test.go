package message

import (
        "testing"
        "testing/assert"
)

func TestMessage(t *testing.T) {
        a := MsgNick{ Name:"to"}
        assert.AssertEquals(t, a.String(), "NICK to\r\n")
}
