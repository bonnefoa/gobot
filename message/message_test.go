package message

import (
	"testing"
	"github.com/bonnefoa/gobot/testing/assert"
)

func TestMessage(t *testing.T) {
	a := MsgNick{Name: "to"}
	assert.AssertEquals(t, a.String(), "NICK to\r\n")
}
