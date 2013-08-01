package main

import (
        "github.com/bonnefoa/gobot/message"
        "log"
        "fmt"
        "github.com/bonnefoa/gobot/bet"
        "database/sql"
        "strings"
        "time"
        "errors"
        "math/rand"
        "github.com/bonnefoa/gobot/utils/utilstring"
        "github.com/bonnefoa/gobot/utils/html"
        "strconv"
        "github.com/bonnefoa/gobot/metapi"
        "github.com/bonnefoa/gobot/bsmeter"
        "github.com/bonnefoa/gobot/meteo"
)

func placeBet(state State, nick string, msg message.MsgPrivate, ts time.Time, isAdmin bool) {
        err := bet.AddUserBet(state.Db, nick, ts, isAdmin)
        if err == nil {
              ts_str := fmt.Sprintf("Za dude %s placed bet for %s", nick, ts.Format(time.RFC1123Z))
              state.ResponseChannel <- message.MsgSend{msg.Response(), ts_str}
        } else {
              ts_str := fmt.Sprintf("Hey %s, %s", nick, err)
              state.ResponseChannel <- message.MsgSend{msg.Response(), ts_str}
        }
}

func formatWinners(winners []string, conf *BotConf) string {
        if len(winners) == 0 {
                return fmt.Sprintf("No bet, no winners")
        }
        winStr := fmt.Sprintf("winners are %s", winners)
        if len(winners) == 1 {
                winStr = fmt.Sprintf("winner is %s", winners[0])
        }
        return fmt.Sprintf("Bet are closed, %s", winStr)
}

func closeBet(state State, msg message.MsgPrivate, ts time.Time) {
        winners := bet.CloseBet(state.Db, ts)
        resp := formatWinners(winners, state.Conf)
        state.ResponseChannel <- message.MsgSend{msg.Response(), resp}
}

func handleSpecificTop(state State, msg message.MsgPrivate) bool {
        if strings.ToLower(msg.User) != state.Conf.Topper &&
                strings.ToLower(msg.User) != state.Conf.Admin { return false }
        lower_msg := strings.ToLower(msg.Msg)
        if !strings.HasPrefix(lower_msg, "top") { return false }
        ts, err := time.Parse("top 15h04", lower_msg)
        if err != nil { return false }
        now := time.Now()
        ts = time.Date(now.Year(), now.Month(), now.Day(), ts.Hour(),
                       ts.Minute(), ts.Second(), 0, now.Location())
        closeBet(state, msg, ts)
        return true
}

func handleTop(state State, msg message.MsgPrivate) bool {
        if strings.ToLower(msg.Msg)!= "top" { return false }
        if strings.ToLower(msg.User) != state.Conf.Topper &&
                strings.ToLower(msg.User) != state.Conf.Admin { return false }
        closeBet(state, msg, time.Now())
        return true
}

func handleScores(state State, msg message.MsgPrivate) bool {
        if strings.ToLower(msg.Msg) != "scores" { return false }
        scores := bet.GetScores(state.Db)
        nicks := make([]string, len(scores))
        for _, score := range scores { nicks = append(nicks, score.String()) }
        maxSize := getMaxNickLength(nicks)

        for _, v := range scores {
          resp := fmt.Sprintf("%*s %d", maxSize, v.Nick, v.Score)
          state.ResponseChannel <- message.MsgSend{msg.Response(), resp}
        }
        return true
}

func handlePlaceBet(state State, msg message.MsgPrivate) bool {
        ts, err := bet.ParseDate(msg.Msg)
        if err != nil { return false }
        placeBet(state, msg.Nick(), msg, ts, false)
        return true
}

func handlePlaceSpecificBet(state State, msg message.MsgPrivate) bool {
        ts, err := bet.ParseDate(msg.Msg)
        if err != nil { return false }
        placeBet(state, msg.Nick(), msg, ts, false)
        return true
}

func handleAdminBet(state State, msg message.MsgPrivate) bool {
        if strings.ToLower(msg.User) != state.Conf.Admin { return false }
        splitted := strings.Split(msg.Msg, " ")
        if len(splitted) == 0 { return false }
        ts, err := bet.ParseDate(strings.Join(splitted[:len(splitted)-1], " ") )
        if err != nil { return false }
        nick := splitted[len(splitted) - 1]
        placeBet(state, nick, msg, ts, true)
        return true
}

func getMaxNickLength(nicks []string) int {
        maxSize := 0
        for _, user := range nicks {
                lenNick := len(user)
                if maxSize < lenNick { maxSize = lenNick }
        }
        return maxSize
}

func printSpecificBet(state State, betId int, location *time.Location, msg message.MsgPrivate) {
        bets := bet.GetUserBets(state.Db, betId)
        nicks := make([]string, len(bets))
        for _, bet := range bets { nicks = append(nicks, bet.String()) }
        maxSize := getMaxNickLength(nicks)
        for _, userBet := range bets {
          resp := fmt.Sprintf("%*s %s", maxSize, userBet.Nick, userBet.Time.In(location))
          state.ResponseChannel <- message.MsgSend{msg.Response(), resp}
        }
}

func printBet(state State, location *time.Location, msg message.MsgPrivate) {
        betId := bet.GetCurrentBet(state.Db)
        printSpecificBet(state, betId, location, msg)
}

func printLastBet(state State, location *time.Location, msg message.MsgPrivate) {
        betId, _ := bet.GetLastBet(state.Db)
        printSpecificBet(state, betId, location, msg)
}

func handleBet(state State, msg message.MsgPrivate) bool {
        location, _ := time.LoadLocation("Local")
        if strings.HasPrefix(strings.ToLower(msg.Msg), "bet") {
                printBet(state, location, msg)
                return true
        }
        if strings.HasPrefix(strings.ToLower(msg.Msg), "last bet") {
                printLastBet(state, location, msg)
                return true
        }
        return false
}

func handleBetSpecificTimeZone(state State, msg message.MsgPrivate) bool {
        if !strings.HasPrefix(strings.ToLower(msg.Msg),  "bet") { return false }
        location, err := time.LoadLocation(msg.Msg[4:])
        if err != nil {
                state.ResponseChannel <- message.MsgSend{msg.Response(), "Invalid timezone dude"}
        } else {
                printBet(state, location, msg)
        }
        return true
}

func parseTimezoneQuery(query string) (time.Time, error) {
        var err error
        var ts time.Time
        ts = time.Now()
        splitted := strings.Split(query, " ")
        if len(splitted) < 3 {
                return ts, errors.New("Too small")
        }
        possibleDate := strings.Join(splitted[:len(splitted)-2], " ")
        log.Printf("Possible date %s", possibleDate)
        ts, err = bet.ParseDate(possibleDate)
        if err != nil {
                return ts, err
        }
        locationStr := splitted[len(splitted) -1]
        fmt.Printf("Location is %s\n", locationStr)
        location, err := time.LoadLocation(locationStr)
        if err != nil {
                return ts, err
        }
        return ts.In(location), err
}

func handleTimezoneConversion(state State, msg message.MsgPrivate) bool {
        ts, err := parseTimezoneQuery(msg.Msg)
        if err != nil { return false }
        state.ResponseChannel <- message.MsgSend{msg.Response(), ts.String()}
        return true
}

func handleRollback(state State, msg message.MsgPrivate) bool {
        if strings.ToLower(msg.Msg) != "rollback" { return false }
        if strings.ToLower(msg.User) != state.Conf.Admin { return false }
        bet.RollbackLastBet(state.Db)
        state.ResponseChannel <- message.MsgSend{msg.Response(), "I have rollbacked the last bet, master."}
        return true
}

func handleReset(state State, msg message.MsgPrivate) bool {
        if strings.ToLower(msg.Msg) != "reset" { return false }
        if strings.ToLower(msg.User) != state.Conf.Admin { return false }
        bet.ResetBet(state.Db)
        state.ResponseChannel <- message.MsgSend{msg.Response(), "I have reset the bet, master."}
        return true
}

func pickMessage(possibleMessages []string) string {
        hearts := utilstring.GetHearts(4)
        exclamations := strings.Repeat("!", rand.Intn(4))
        msg := utilstring.RandomString(possibleMessages)
        return fmt.Sprintf("%s %s%s %s", hearts, msg, exclamations, hearts)
}

func getNickInMessage(db *sql.DB, msg string) string {
        users := bet.GetUsers(db)
        for _, n := range strings.Split(msg, " ") {
                if utilstring.StringContains(n, users) {
                        return fmt.Sprintf("%s: ", n)
                }
        }
        return ""
}

func getPukeMessage(msg string) string {
        hearts := utilstring.GetHearts(25)
        return fmt.Sprintf("%s", hearts)
}

func handleTrigger(state State, msg message.MsgPrivate, trigger Trigger) bool {
        lowerMsg := strings.ToLower(msg.Msg)
        lowerMsg = strings.Map(utilstring.KeepLettersAndSpace, lowerMsg)
        if !utilstring.TriggerIn(trigger.Words, lowerMsg) {
                return false
        }
        resp := ""
        nick := getNickInMessage(state.Db, msg.Msg)
        if trigger.IsPuke {
                resp = getPukeMessage(lowerMsg)
        } else {
                resp = pickMessage(trigger.Results)
        }
        state.ResponseChannel <- message.MsgSend{msg.Response(), fmt.Sprintf("%s%s", nick, resp)}
        return true
}

func handleTriggers(state State, msg message.MsgPrivate) bool {
        for _, trigger := range state.Conf.Triggers {
                if handleTrigger(state, msg, trigger) { return true }
        }
        return false
}

func handlePutf8(state State, msg message.MsgPrivate) bool {
        if !strings.Contains(strings.ToLower(msg.Msg), "putf8") { return false }
        numChars := rand.Intn(25)
        res := make([]string, numChars)
        for i := 0; i < numChars; i++ {
                res[i] = utilstring.ColorString(utilstring.GetRandUtf8())
        }
        nick := getNickInMessage(state.Db, msg.Msg)
        state.ResponseChannel <-
                message.MsgSend{msg.Response(), fmt.Sprintf("%s%s", nick, strings.Join(res, " "))}
        return true
}

func handleDodo(state State, msg message.MsgPrivate) bool {
        if !strings.Contains(strings.ToLower(msg.Msg), "dodo") { return false }
        zStr := strings.Repeat("Zz", rand.Intn(25))
        res := utilstring.ColorStringSlice(utilstring.ShuffleString(zStr))
        nick := getNickInMessage(state.Db, msg.Msg)
        state.ResponseChannel <- message.MsgSend{msg.Response(), fmt.Sprintf("%s%s", nick, res)}
        return true
}

func handleTroll(state State, msg message.MsgPrivate) bool {
        if !strings.HasPrefix(strings.ToLower(msg.Msg), "troll") { return false }
        winners := []string{msg.Nick()}
        resp := formatWinners(winners, state.Conf)
        state.ResponseChannel <- message.MsgSend{msg.Response(), resp}
        return true
}

func handleHelp(state State, msg message.MsgPrivate) bool {
        if strings.ToLower(msg.Msg) != "gobot: help" { return false }
        state.ResponseChannel <- message.MsgSend{msg.Response(), state.Conf.Help}
        return true
}

func handleMetapi(state State, msg message.MsgPrivate) bool {
        log.Printf("Check %s", msg.Msg)
        if !strings.HasPrefix(strings.ToLower(msg.Msg), "metapi") { return false }
        log.Printf("Passed for %s", msg.Msg)
        num, err := strconv.ParseInt(msg.Msg[7:], 10, 64)
        if err != nil {
                state.ResponseChannel <-
                        message.MsgSend{msg.Response(), fmt.Sprintf("Expect an int32 to search, error was %q", err)}
                return true
        } else {
                state.ResponseChannel <-
                        message.MsgSend{msg.Response(), "Launching big real data time calculation on the claoud"}
                state.PiQueryChannel <- metapi.PiQuery{num, msg.Response()}
        }
        return true
}

func handleRotate(state State, msg message.MsgPrivate) bool {
        if !strings.HasPrefix(strings.ToLower(msg.Msg), "rotate") { return false }
        state.ResponseChannel <- message.MsgSend{msg.Response(), utilstring.RotateString(msg.Msg[6:])}
        return true
}

func removeFirstWord(src string) string{
        return src[strings.Index(src," ") + 1:]
}

func handleBsRequest(state State, msg message.MsgPrivate) bool {
        lowerMsg := strings.ToLower(msg.Msg)
        bsQuery := bsmeter.BsQuery{Channel:msg.Response()}
        hasBsQuery := strings.HasPrefix(lowerMsg, "isbs")
        hasHttp := strings.Contains(lowerMsg, "http")
        if hasHttp {
                urls := html.ExtractUrls(msg.Msg)
                bsQuery.Urls = urls
        }
        if hasBsQuery {
                phrase := removeFirstWord(lowerMsg)
                for _, url := range bsQuery.Urls {
                        phrase = strings.Replace(phrase, url, "", -1)
                }
                bsQuery.Phrase = phrase
        }
        if hasHttp || hasBsQuery {
                state.BsQueryChannel <- bsQuery
                return true
        }
        return false
}

func handleBsTraining(state State, msg message.MsgPrivate) bool {
        lowerMsg := strings.ToLower(msg.Msg)
        if lowerMsg == "bsreload" {
                state.BsQueryChannel <- bsmeter.BsQuery{IsReload:true}
                return true
        }
        if !strings.HasPrefix(lowerMsg, "bs") && !strings.HasPrefix(lowerMsg, "nobs") {
                return false
        }
        bs := strings.HasPrefix(lowerMsg, "bs")
        urls := html.ExtractUrls(msg.Msg)
        if len(urls) == 0 {
                state.BsQueryChannel<- bsmeter.BsQuery{
                        Phrase:removeFirstWord(lowerMsg),
                        IsTraining:true, Bs:bs, Channel:msg.Response()}
        } else {
                state.BsQueryChannel<- bsmeter.BsQuery{Urls:urls,
                        IsTraining:true, Bs:bs, Channel:msg.Response()}
        }
        return true
}

func handleMeteo(state State, msg message.MsgPrivate) bool {
        lowerMsg := strings.ToLower(msg.Msg)
        if lowerMsg == "meteo" || lowerMsg == "météo" {
                weather := meteo.FetchWeatherFromUrl(state.Conf.Meteo.Url)
                log.Printf("Fetched %q", weather)
                state.ResponseChannel <- message.MsgSend{msg.Response(), strings.Join(weather, "|")}
                return true
        }
        return false
}

var handlers = []func(State, message.MsgPrivate) bool { handleBsTraining, handleBsRequest, handleHelp,
        handleTop, handleSpecificTop, handleScores, handlePlaceBet, handleAdminBet,
        handleBet, handleRollback, handleReset, handleBetSpecificTimeZone, handleTimezoneConversion,
        handlePutf8, handleDodo, handleTroll, handleTriggers, handleMetapi, handleRotate, handleMeteo}

func handleMessage(state State, msg message.MsgPrivate) {
        log.Printf("Received message %s", msg.Msg)
        for i, handler := range handlers {
                if handler(state, msg) {
                        log.Printf("Breaking on handler %v", i)
                        break
                }
        }
}
