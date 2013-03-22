package message

import "fmt"

type MsgNick struct { name string }
type MsgUser struct { name, real_name string }
type MsgPong struct { ping string }
type MsgJoin struct { channel string }
type MsgPrivate struct { user, msg string }
type MsgQuit struct { reason string }

func (msg MsgNick) String() string {
  return fmt.Sprintf("NICK %s\r\n", msg.name)
}

func (msg MsgUser) String() string {
  return fmt.Sprintf("USER %s 0 * :%s\r\n", msg.name, msg.real_name)
}

func (msg MsgPong) String() string {
  return fmt.Sprintf("PONG %s\r\n", msg.ping)
}

func (msg MsgJoin) String() string {
  return fmt.Sprintf("JOIN %s\r\n", msg.channel)
}

func (msg MsgPrivate) String() string {
  return fmt.Sprintf("PRIVMSG %s %s\r\n", msg.user, msg.msg)
}

func (msg MsgQuit) String() string {
  return fmt.Sprintf("QUIT %s\r\n", msg.reason)
}
