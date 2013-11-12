package gobot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/bonnefoa/gobot/bsmeter"
	"github.com/bonnefoa/gobot/metapi"
	"github.com/bonnefoa/gobot/meteo"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

type Trigger struct {
	Words       []string
	Results     []string
	Repeated    bool
	ProcCommand []string
	Cron        string
	Dest        string
	PutHearts   bool
}

type BotConf struct {
	Server   string
	Password string
	Channel  string
	Db       string
	Admin    string
	Topper   string
	Help     string
	Name     string
	RealName string
	Triggers []Trigger
	Meteo    meteo.Meteo
	BsConf   bsmeter.BsConf
}

type State struct {
	Db              *sql.DB
	Conf            *BotConf
	ResponseChannel chan fmt.Stringer
	PiQueryChannel  chan metapi.PiQuery
	BsQueryChannel  chan bsmeter.BsQuery
}

func expandTilde(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path[:2] == "~/" {
		path = filepath.Join(dir, path[2:])
	}
	return path
}

func ReadConfigurationFile(filename string) BotConf {
	expandedFilename := expandTilde(filename)
	file, err := os.Open(expandedFilename)
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
