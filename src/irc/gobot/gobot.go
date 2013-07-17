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
        "flag"
        "math/rand"
        "math/big"
        cryptrand "crypto/rand"
        "irc/metapi"
        "runtime"
        "runtime/pprof"
)


type Trigger struct {
        Words []string
        Results []string
        IsPuke bool
}

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
        Triggers []Trigger
}

type State struct {
        Db *sql.DB
        Conf *BotConf
        ResponseChannel chan fmt.Stringer
        PiQueryChannel chan metapi.PiQuery
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

func dispatchMessage(state State, msg string) {
        for _, single_msg := range message.ParseMessage(msg) {
            switch parsed := single_msg.(type) {
            case message.MsgPing:
                    state.ResponseChannel <- message.MsgPong{parsed.Ping}
            case message.MsgPrivate:
                    log.Printf("Received message %s from %s\n", parsed.Msg, parsed.User)
                    handleMessage(state, parsed)
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

func initializeRandom(conf *BotConf) {
        max := big.NewInt(2^60)
        seed, _ := cryptrand.Int(cryptrand.Reader, max)
        rand.Seed(seed.Int64())
}

func connect() {
        confFile := flag.String("conffile", "gobot.json", "Conf path")
        flag.Parse()
        conf := readConfigurationFile(*confFile)
        initializeRandom(&conf)

        db := bet.InitBase(conf.Db)
        responseChannel := make(chan fmt.Stringer)
        piQueryChannel := make(chan metapi.PiQuery)
        state := State{Db: db, Conf: &conf, ResponseChannel: responseChannel,
                PiQueryChannel: piQueryChannel}
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
        go metapi.SearchWorker(piQueryChannel, responseChannel)

        for {
                select {
                case data := <-readChannel:
                        log.Printf("Got data %q\n", data)
                        go dispatchMessage(state, string(data))
                case err := <-errorChannel:
                        log.Fatal("Got error %q\n", err)
                case response := <-responseChannel:
                        log.Printf("Sending response %q\n", response.String())
                        fmt.Fprintf(conn, response.String())
                }
        }
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
        flag.Parse()
        if *cpuprofile != "" {
               f, err := os.Create(*cpuprofile)
               if err != nil {
                       log.Fatal(err)
               }
               pprof.StartCPUProfile(f)
               defer pprof.StopCPUProfile()
        }
        runtime.GOMAXPROCS(2)
        connect()
}
