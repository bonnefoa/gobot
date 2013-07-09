package main

import (
        "irc/message"
        "log"
        "fmt"
        "irc/bet"
        "database/sql"
        "strings"
        "time"
        "errors"
        "math/rand"
        "utils/utilstring"
)

func placeBet(db *sql.DB, nick string, msg message.MsgPrivate, ts time.Time, responseChannel chan fmt.Stringer, isAdmin bool) {
        err := bet.AddUserBet(db, nick, ts, isAdmin)
        if err == nil {
              ts_str := fmt.Sprintf("Za dude %s placed bet for %s", nick, ts.Format(time.RFC1123Z))
              responseChannel <- message.MsgSend{msg.Response(), ts_str}
        } else {
              ts_str := fmt.Sprintf("Hey %s, %s", nick, err)
              responseChannel <- message.MsgSend{msg.Response(), ts_str}
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

func closeBet(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                ts time.Time, responseChannel chan fmt.Stringer) {
        winners := bet.CloseBet(db, ts)
        resp := formatWinners(winners, conf)
        responseChannel <- message.MsgSend{msg.Response(), resp}
}

func handleSpecificTop(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.User) != conf.Topper && strings.ToLower(msg.User) != conf.Admin { return false }
        lower_msg := strings.ToLower(msg.Msg)
        if !strings.HasPrefix(lower_msg, "top") { return false }
        ts, err := time.Parse("top 15h04", lower_msg)
        if err != nil { return false }
        now := time.Now()
        ts = time.Date(now.Year(), now.Month(), now.Day(), ts.Hour(),
                       ts.Minute(), ts.Second(), 0, now.Location())
        closeBet(db, conf, msg, ts, responseChannel)
        return true
}

func handleTop(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.Msg)!= "top" { return false }
        if strings.ToLower(msg.User) != conf.Topper && strings.ToLower(msg.User) != conf.Admin { return false }
        closeBet(db, conf, msg, time.Now(), responseChannel)
        return true
}

func handleScores(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.Msg) != "scores" { return false }
        scores := bet.GetScores(db)
        nicks := make([]string, len(scores))
        for _, score := range scores { nicks = append(nicks, score.String()) }
        maxSize := getMaxNickLength(nicks)

        for _, v := range scores {
          resp := fmt.Sprintf("%*s %d", maxSize, v.Nick, v.Score)
          responseChannel <- message.MsgSend{msg.Response(), resp}
        }
        return true
}

func handlePlaceBet(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        ts, err := bet.ParseDate(msg.Msg)
        if err != nil { return false }
        placeBet(db, msg.Nick(), msg, ts, responseChannel, false)
        return true
}

func handlePlaceSpecificBet(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        ts, err := bet.ParseDate(msg.Msg)
        if err != nil { return false }
        placeBet(db, msg.Nick(), msg, ts, responseChannel, false)
        return true
}

func handleAdminBet(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.User) != conf.Admin { return false }
        splitted := strings.Split(msg.Msg, " ")
        if len(splitted) == 0 { return false }
        ts, err := bet.ParseDate(strings.Join(splitted[:len(splitted)-1], " ") )
        if err != nil { return false }
        nick := splitted[len(splitted) - 1]
        placeBet(db, nick, msg, ts, responseChannel, true)
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

func printSpecificBet(betId int, db *sql.DB, location *time.Location, msg message.MsgPrivate, responseChannel chan fmt.Stringer) {
        bets := bet.GetUserBets(db, betId)
        nicks := make([]string, len(bets))
        for _, bet := range bets { nicks = append(nicks, bet.String()) }
        maxSize := getMaxNickLength(nicks)
        for _, userBet := range bets {
          resp := fmt.Sprintf("%*s %s", maxSize, userBet.Nick, userBet.Time.In(location))
          responseChannel <- message.MsgSend{msg.Response(), resp}
        }
}

func printBet(db *sql.DB, location *time.Location, msg message.MsgPrivate, responseChannel chan fmt.Stringer) {
        betId := bet.GetCurrentBet(db)
        printSpecificBet(betId, db, location, msg, responseChannel)
}

func printLastBet(db *sql.DB, location *time.Location, msg message.MsgPrivate, responseChannel chan fmt.Stringer) {
        betId, _ := bet.GetLastBet(db)
        printSpecificBet(betId, db, location, msg, responseChannel)
}

func handleBet(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        location, _ := time.LoadLocation("Local")
        if strings.HasPrefix(strings.ToLower(msg.Msg), "bet") {
                printBet(db, location, msg, responseChannel)
                return true
        }
        if strings.HasPrefix(strings.ToLower(msg.Msg), "last bet") {
                printLastBet(db, location, msg, responseChannel)
                return true
        }
        return false
}

func handleBetSpecificTimeZone(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if !strings.HasPrefix(strings.ToLower(msg.Msg),  "bet") { return false }
        location, err := time.LoadLocation(msg.Msg[4:])
        if err != nil {
          responseChannel <- message.MsgSend{msg.Response(), "Invalid timezone dude"}
        } else {
                printBet(db, location, msg, responseChannel)
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

func handleTimezoneConversion(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        ts, err := parseTimezoneQuery(msg.Msg)
        if err != nil { return false }
        responseChannel <- message.MsgSend{msg.Response(), ts.String()}
        return true
}

func handleRollback(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.Msg) != "rollback" { return false }
        if strings.ToLower(msg.User) != conf.Admin { return false }
        bet.RollbackLastBet(db)
        responseChannel <- message.MsgSend{msg.Response(), "I have rollbacked the last bet, master."}
        return true
}

func handleReset(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.Msg) != "reset" { return false }
        if strings.ToLower(msg.User) != conf.Admin { return false }
        bet.ResetBet(db)
        responseChannel <- message.MsgSend{msg.Response(), "I have reset the bet, master."}
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

func handleTrigger(db *sql.DB, msg message.MsgPrivate, trigger Trigger, responseChannel chan fmt.Stringer) bool {
        lowerMsg := strings.ToLower(msg.Msg)
        lowerMsg = strings.Map(utilstring.KeepLettersAndSpace, lowerMsg)
        if !utilstring.TriggerIn(trigger.Words, lowerMsg) {
                return false
        }
        resp := ""
        nick := getNickInMessage(db, msg.Msg)
        if trigger.IsPuke {
                resp = getPukeMessage(lowerMsg)
        } else {
                resp = pickMessage(trigger.Results)
        }
        responseChannel <- message.MsgSend{msg.Response(), fmt.Sprintf("%s%s", nick, resp)}
        return true
}

func handleTriggers(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        for _, trigger := range conf.Triggers {
                if handleTrigger(db, msg, trigger, responseChannel) { return true }
        }
        return false
}

func handlePutf8(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if !strings.Contains(strings.ToLower(msg.Msg), "putf8") { return false }
        numChars := rand.Intn(25)
        res := make([]string, numChars)
        for i := 0; i < numChars; i++ {
                res[i] = utilstring.ColorString(utilstring.GetRandUtf8())
        }
        nick := getNickInMessage(db, msg.Msg)
        responseChannel <- message.MsgSend{msg.Response(), fmt.Sprintf("%s%s", nick, strings.Join(res, " "))}
        return true
}

func handleDodo(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if !strings.Contains(strings.ToLower(msg.Msg), "dodo") { return false }
        zStr := strings.Repeat("Zz", rand.Intn(25))
        res := utilstring.ColorStringSlice(utilstring.ShuffleString(zStr))
        nick := getNickInMessage(db, msg.Msg)
        responseChannel <- message.MsgSend{msg.Response(), fmt.Sprintf("%s%s", nick, res)}
        return true
}

func handleTroll(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if !strings.HasPrefix(strings.ToLower(msg.Msg), "troll") { return false }
        winners := []string{msg.Nick()}
        resp := formatWinners(winners, conf)
        responseChannel <- message.MsgSend{msg.Response(), resp}
        return true
}

func handleHelp(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.Msg) != "gobot: help" { return false }
        responseChannel <- message.MsgSend{msg.Response(), conf.Help}
        return true
}

func handleRotate(db *sql.DB, conf *BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if !strings.HasPrefix(strings.ToLower(msg.Msg), "rotate") { return false }
        responseChannel <- message.MsgSend{msg.Response(), utilstring.RotateString(msg.Msg[6:])}
        return true
}

func handleMessage(db *sql.DB, conf *BotConf, msg message.MsgPrivate, responseChannel chan fmt.Stringer) {
        log.Printf("Received message %s", msg.Msg)
        handlers := []func(*sql.DB, *BotConf, message.MsgPrivate, chan fmt.Stringer) bool { handleHelp,
        handleTop, handleSpecificTop, handleScores, handlePlaceBet, handleAdminBet,
        handleBet, handleRollback, handleReset, handleBetSpecificTimeZone, handleTimezoneConversion,
        handlePutf8, handleDodo, handleTroll, handleTriggers, handleRotate}
        for _, handler := range handlers {
                if handler(db, conf, msg, responseChannel) {
                        log.Printf("Breaking on handler %s", handler)
                        break
                }
        }
}
