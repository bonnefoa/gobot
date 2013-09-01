package main

import (
	cryptrand "crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/bonnefoa/gobot/bet"
	"github.com/bonnefoa/gobot/bsmeter"
	"github.com/bonnefoa/gobot/message"
	"github.com/bonnefoa/gobot/metapi"
	"github.com/bonnefoa/gobot/meteo"
	"github.com/bonnefoa/gobot"
	"log"
	"math/big"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
    "time"
)

func readConnection(conn *tls.Conn, readChannel chan []byte, errorChannel chan error) {
	for {
		data := make([]byte, 512)
		num, err := conn.Read(data)
		if err != nil {
			errorChannel <- err
			return
		}
		readChannel <- data[:num]
	}
}

func dispatchMessage(state gobot.State, msg string) {
	for _, single_msg := range message.ParseMessage(msg) {
		switch parsed := single_msg.(type) {
		case message.MsgPing:
			state.ResponseChannel <- message.MsgPong{parsed.Ping}
		case message.MsgPrivate:
			log.Printf("Received message %s from %s\n", parsed.Msg, parsed.User)
			gobot.HandleMessage(state, parsed)
		default:
			log.Printf("No switch matched for %s!\n", parsed)
		}
	}
}

func cronMessage(conf gobot.BotConf, responseChannel chan fmt.Stringer) {
    for _, tr := range conf.Triggers {
        if tr.Cron == "" {
            continue
        }
        next := bet.ParseCron(tr.Cron, time.Now())
        <-time.After(time.Duration(next))
        for {
            responseChannel <- message.MsgSend{tr.Dest, tr.Results[0]}
            <-time.After(time.Hour * 24 * 7)
        }
    }
}

func join(conf gobot.BotConf, responseChannel chan fmt.Stringer) {
	responseChannel <- message.MsgPassword{conf.Password}
	responseChannel <- message.MsgNick{conf.Name}
	responseChannel <- message.MsgUser{conf.Name, conf.RealName}
	responseChannel <- message.MsgJoin{conf.Channel}
}

func initializeRandom(conf *gobot.BotConf) {
	max := big.NewInt(2 ^ 60)
	seed, _ := cryptrand.Int(cryptrand.Reader, max)
	rand.Seed(seed.Int64())
}

func connect() {
	conf := gobot.ReadConfigurationFile(*confFile)
	initializeRandom(&conf)

	db := bet.InitBase(conf.Db)
	responseChannel := make(chan fmt.Stringer)
	piQueryChannel := make(chan metapi.PiQuery)
	bsQueryChannel := make(chan bsmeter.BsQuery)
	state := gobot.State{Db: db, Conf: &conf, ResponseChannel: responseChannel,
		PiQueryChannel: piQueryChannel,
		BsQueryChannel: bsQueryChannel}
	readChannel := make(chan []byte)
	errorChannel := make(chan error)

	config := &tls.Config{InsecureSkipVerify: true}
	log.Printf("Connecting to %s\n", conf.Server)
	conn, err := tls.Dial("tcp", conf.Server, config)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Joining %s\n", conf.Channel)

	go readConnection(conn, readChannel, errorChannel)
	go join(conf, responseChannel)
	go metapi.SearchWorker(piQueryChannel, responseChannel)
	go bsmeter.BsWorker(conf.BsConf, bsQueryChannel, responseChannel)
	go meteo.RainWatcher(conf.Meteo, responseChannel)
	go cronMessage(conf, responseChannel)

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
var confFile = flag.String("conffile", "~/.gobot/gobot.json", "Conf path")

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
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	connect()
}
