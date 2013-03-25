package main

import (
        "irc/message"
        "log"
        "crypto/tls"
        "fmt"
        "encoding/json"
        "os"
        "irc/bet"
        "database/sql"
        "strings"
        "time"
)

type BotConf struct {
        Server string
        Password string
        Channel string
        Db string
}

func readConfigurationFile(filename string) BotConf {
        file, err := os.Open(filename)
        if err != nil {
                log.Fatal("Could not open file %s, %s\n", filename, err)
        }
        dec := json.NewDecoder(file)
        var conf BotConf
        if err := dec.Decode(&conf); err != nil {
                log.Fatal(err)
        }
        return conf
}

func readConnection(conn *tls.Conn, readChannel chan []byte, errorChannel chan error) {
        for {
                data := make([]byte, 512)
                num, err := conn.Read(data)
                if err != nil {
                        errorChannel<- err
                        return
                }
                readChannel<- data[:num]
        }
}

func handleMessage(db *sql.DB, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) {
        ts, err := bet.CheckBetMessage(msg.Msg)
        if err == nil {
                nick := msg.Nick()
                bet.AddUserBet(db, nick, ts)
                ts_str := fmt.Sprintf("Za dude %s placed bet for %s", nick, ts.Format("02/01 15:04"))
                responseChannel <- message.MsgSend{msg.Dest, ts_str}
                return
        }
        if strings.ToLower(msg.Msg) == "top" {
                winners := bet.CloseBet(db, time.Now())
                var resp string
                if len(winners) == 0 {
                        resp = fmt.Sprintf("No bet, no winners")
                } else {
                        resp = fmt.Sprintf("Bet are closed, winnners are %s", winners)
                }
                responseChannel <- message.MsgSend{msg.Dest, resp}
        }

        if strings.ToLower(msg.Msg) == "scores" {
                scores := bet.GetScores(db)
                for k, v := range scores {
                        resp := fmt.Sprintf("%s %d", k, v)
                        responseChannel <- message.MsgSend{msg.Dest, resp}
                }
        }
}

func dispatchMessage(db *sql.DB, msg string, responseChannel chan fmt.Stringer) {
        switch parsed := message.ParseMessage(msg).(type) {
        case message.MsgPing:
                responseChannel <- message.MsgPong{parsed.Ping}
        case message.MsgPrivate:
                log.Printf("Received message %s from %s\n", parsed.Msg, parsed.User)
                handleMessage(db, parsed, responseChannel)
        default:
                log.Printf("No switch matched for %s!\n", parsed)
        }
}

func join(conf BotConf, responseChannel chan fmt.Stringer) {
        responseChannel <- message.MsgPassword{ conf.Password }
        responseChannel <- message.MsgNick{ "Gobot" }
        responseChannel <- message.MsgUser{ "Gobot", "GoGoBot" }
        responseChannel <- message.MsgJoin{ conf.Channel }
}

func connect() {
        conf := readConfigurationFile("gobot.json")

        db := bet.InitBase(conf.Db)
        responseChannel := make(chan fmt.Stringer)
        readChannel := make(chan []byte)
        errorChannel := make(chan error)

        config := &tls.Config{ InsecureSkipVerify : true }
        log.Printf("Connecting to %s\n", conf.Server)
        conn, err := tls.Dial("tcp", conf.Server, config)
        if err != nil {
                log.Fatal(err)
        }
        log.Printf("Joining %s\n", conf.Channel)

        go readConnection(conn, readChannel, errorChannel)
        go join(conf, responseChannel)

        for {
                select {
                case data := <-readChannel:
                        log.Printf("Got data %q\n", data)
                        go dispatchMessage(db, string(data), responseChannel)
                case err := <-errorChannel:
                        log.Fatal("Got error %q\n", err)
                case response := <-responseChannel:
                        log.Printf("Sending response %q\n", response.String())
                        fmt.Fprintf(conn, response.String())
                }
        }
}

func main() {
        connect()
}
