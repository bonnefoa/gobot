package message

import "fmt"

type MsgOther struct { Text string }
type MsgNick struct { Name string }
type MsgUser struct { Name, RealName string }
type MsgPing struct { Ping string }
type MsgPong struct { Pong string }
type MsgJoin struct { Channel string }
type MsgPrivate struct { User, Dest, Msg string }
type MsgPassword struct { Password string }
type MsgQuit struct { Reason string }

func (msg MsgNick) String() string {
        return fmt.Sprintf("NICK %s\r\n", msg.Name)
}

func (msg MsgUser) String() string {
        return fmt.Sprintf("USER %s 0 * :%s\r\n", msg.Name, msg.RealName)
}

func (msg MsgPing) String() string {
        return fmt.Sprintf("PING %s\r\n", msg.Ping)
}

func (msg MsgPong) String() string {
        return fmt.Sprintf("PONG %s\r\n", msg.Pong)
}

func (msg MsgJoin) String() string {
        return fmt.Sprintf("JOIN %s\r\n", msg.Channel)
}

func (msg MsgPrivate) String() string {
        return fmt.Sprintf(":%s PRIVMSG %s :%s\r\n", msg.User,
                msg.Dest, msg.Msg)
}

func (msg MsgQuit) String() string {
        return fmt.Sprintf("QUIT %s\r\n", msg.Reason)
}

func (msg MsgPassword) String() string {
        return fmt.Sprintf("PASS %s\r\n", msg.Password)
}
