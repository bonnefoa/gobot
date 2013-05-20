package message

import (
        "testing"
        "testing/assert"
        "fmt"
)

func testMessageParse(t *testing.T, msg fmt.Stringer) {
        res := ParseMessage(msg.String())
        t.Logf("Parsed message is %q", res)
        assert.AssertEquals(t, res[0], msg)
}

func TestParseNick(t *testing.T) {
        testMessageParse(t, MsgNick{ "4" })
}

func TestParsePong(t *testing.T) {
        testMessageParse(t, MsgPong{"ping"})
}

func TestParsePing(t *testing.T) {
        testMessageParse(t, MsgPing{"adyxax.org"})
}

func TestParseQuit(t *testing.T) {
        testMessageParse(t, MsgQuit{"quit bis repetita"})
}

func TestParseInvalid(t *testing.T) {
        assert.AssertEquals(t, ParseMessage("Invalid")[0], nil)
}

func TestParsePrivate(t *testing.T) {
        testMessageParse(t, MsgPrivate{"sora!~sora@mougnou.fr", "#geek2", "4"})
}

func TestParseDoublePing(t *testing.T) {
        msg := ":Ga!~Ga@graou.Ga.org PRIVMSG #geek :il te reste sora :p\r\nPING :ga.ga.org\r\n"
        res := ParseMessage(msg)
        t.Logf("Parsed message is %#v\n", res)
        assert.AssertEquals(t, len(res), 2)
        assert.AssertEquals(t, res[1], MsgPing{"ga.ga.org"})
}

func TestParseDouble(t *testing.T) {
        msg1 := MsgPing{"ping"}
        msg2 := MsgPong{"pong"}
        conc := fmt.Sprintf("%s%s", msg1.String(), msg2.String());
        res := ParseMessage(conc)
        assert.AssertEquals(t, res[0], msg1)
        assert.AssertEquals(t, res[1], msg2)
}
