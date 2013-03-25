package message

import (
        "testing"
        "testing/assert"
        "fmt"
)

func testMessageParse(t *testing.T, msg fmt.Stringer) {
        res := ParseMessage(msg.String())
        assert.AssertEquals(t, res, msg)
}

func TestParseNick(t *testing.T) {
        testMessageParse(t, MsgNick{ "4" })
}

func TestParsePong(t *testing.T) {
        testMessageParse(t, MsgPong{"ping"})
}

func TestParseQuit(t *testing.T) {
        testMessageParse(t, MsgQuit{"quit bis repetita"})
}

func TestParseInvalid(t *testing.T) {
        assert.AssertEquals(t, ParseMessage("Invalid"), nil)
}

func TestParsePrivate(t *testing.T) {
        testMessageParse(t, MsgPrivate{"sora!~sora@mougnou.fr",
          "#geek2", "4"})
}
