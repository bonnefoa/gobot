PARSER  = "src/irc/message/parser.go"
SQLITE  = "src/github.com/mattn/go-sqlite3/sqlite3.go"

$(SQLITE): 
	go get github.com/mattn/go-sqlite3

$(PARSER):
	go tool yacc -o $(PARSER) src/irc/message/parser.y;

test: $(PARSER) $(SQLITE)
	go test -i irc/message
	go test irc/message

metapi.prof: 
	go build irc/metapi
	./metapi -cpuprofile=metapi.prof

gobot: $(PARSER) $(SQLITE)
	go build irc/gobot
