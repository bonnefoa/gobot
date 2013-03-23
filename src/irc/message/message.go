package message

import "fmt"

type MsgNick struct { Name string }
type MsgUser struct { Name, RealName string }
type MsgPong struct { Ping string }
type MsgJoin struct { Channel string }
type MsgPrivate struct { UserName, Msg string }
type MsgQuit struct { Reason string }

func (msg MsgNick) String() string {
  return fmt.Sprintf("NICK %s\r\n", msg.Name)
}

func (msg MsgUser) String() string {
  return fmt.Sprintf("USER %s 0 * :%s\r\n", msg.Name, msg.RealName)
}

func (msg MsgPong) String() string {
  return fmt.Sprintf("PONG %s\r\n", msg.Ping)
}

func (msg MsgJoin) String() string {
  return fmt.Sprintf("JOIN %s\r\n", msg.Channel)
}

func (msg MsgPrivate) String() string {
  return fmt.Sprintf("PRIVMSG %s %s\r\n", msg.UserName, msg.Msg)
}

func (msg MsgQuit) String() string {
  return fmt.Sprintf("QUIT %s\r\n", msg.Reason)
}
