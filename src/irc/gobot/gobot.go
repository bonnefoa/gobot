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
)

type BotConf struct {
        Server string
        Password string
        Channel string
        Db string
        Admin string
        Topper string
        Help string
        Name string
        RealName string
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

func dispatchMessage(db *sql.DB, conf BotConf, msg string, responseChannel chan fmt.Stringer) {
        for _, single_msg := range message.ParseMessage(msg) {
            switch parsed := single_msg.(type) {
            case message.MsgPing:
                    responseChannel <- message.MsgPong{parsed.Ping}
            case message.MsgPrivate:
                    log.Printf("Received message %s from %s\n", parsed.Msg, parsed.User)
                    handleMessage(db, conf, parsed, responseChannel)
            default:
                    log.Printf("No switch matched for %s!\n", parsed)
            }
      }
}

func join(conf BotConf, responseChannel chan fmt.Stringer) {
        responseChannel <- message.MsgPassword{ conf.Password }
        responseChannel <- message.MsgNick{ conf.Name }
        responseChannel <- message.MsgUser{ conf.Name , conf.RealName }
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
                        go dispatchMessage(db, conf, string(data), responseChannel)
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
