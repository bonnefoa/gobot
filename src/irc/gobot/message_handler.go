package main

import (
        "irc/message"
        "log"
        "fmt"
        "irc/bet"
        "database/sql"
        "strings"
        "time"
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

func closeBet(db *sql.DB, msg message.MsgPrivate, ts time.Time, responseChannel chan fmt.Stringer) {
        winners := bet.CloseBet(db, ts)
        var resp string
        if len(winners) == 0 {
          resp = fmt.Sprintf("No bet, no winners")
        } else {
          resp = fmt.Sprintf("Bet are closed, winnners are %s", winners)
        }
        responseChannel <- message.MsgSend{msg.Response(), resp}
}

func handleSpecificTop(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.User) != conf.Topper && strings.ToLower(msg.User) != conf.Admin { return false }
        lower_msg := strings.ToLower(msg.Msg)
        if !strings.HasPrefix(lower_msg, "top") { return false }
        ts, err := time.Parse("top 15h04", lower_msg)
        if err != nil { return false }
        now := time.Now()
        ts = time.Date(now.Year(), now.Month(), now.Day(), ts.Hour(),
                       ts.Minute(), ts.Second(), 0, now.Location())
        closeBet(db, msg, ts, responseChannel)
        return true
}

func handleTop(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.Msg)!= "top" { return false }
        if strings.ToLower(msg.User) != conf.Topper && strings.ToLower(msg.User) != conf.Admin { return false }
        closeBet(db, msg, time.Now(), responseChannel)
        return true
}

func handleScores(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.Msg) != "scores" { return false }
        scores := bet.GetScores(db)
        for k, v := range scores {
          resp := fmt.Sprintf("%s %d", k, v)
          responseChannel <- message.MsgSend{msg.Response(), resp}
        }
        return true
}

func handlePlaceBet(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        ts, err := bet.ParseDate(msg.Msg)
        if err != nil { return false }
        placeBet(db, msg.Nick(), msg, ts, responseChannel, false)
        return true
}

func handlePlaceSpecificBet(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        ts, err := bet.ParseDate(msg.Msg)
        if err != nil { return false }
        placeBet(db, msg.Nick(), msg, ts, responseChannel, false)
        return true
}

func handleAdminBet(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        splitted := strings.Split(msg.Msg, " ")
        if len(splitted) == 0 { return false }
        ts, err := bet.ParseDate(strings.Join(splitted[:len(splitted)-1], " ") )
        if err != nil { return false }
        nick := splitted[len(splitted) - 1]
        placeBet(db, nick, msg, ts, responseChannel, true)
        return true
}

func handleBet(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.Msg) != "bet" { return false }
        betId := bet.GetCurrentBet(db)
        bets := bet.GetUserBets(db, betId)
        for k, ts := range bets {
          resp := fmt.Sprintf("%s %s", k, bet.ConvertTimeToLocal(ts))
          responseChannel <- message.MsgSend{msg.Response(), resp}
        }
        return true
}

func handleRollback(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.Msg) != "rollback" { return false }
        if strings.ToLower(msg.User) != conf.Admin { return false }
        bet.RollbackLastBet(db)
        responseChannel <- message.MsgSend{msg.Response(), "I have rollbacked the last bet, master."}
        return true
}

func handleReset(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
        if strings.ToLower(msg.Msg) != "reset" { return false }
        if strings.ToLower(msg.User) != conf.Admin { return false }
        bet.ResetBet(db)
        responseChannel <- message.MsgSend{msg.Response(), "I have reset the bet, master."}
        return true
}

func handleHelp(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) bool {
  if strings.ToLower(msg.Msg) != "gobot: help" { return false }
  responseChannel <- message.MsgSend{msg.Response(), conf.Help}
  return true
}

func handleMessage(db *sql.DB, conf BotConf, msg message.MsgPrivate,
                        responseChannel chan fmt.Stringer) {
        log.Printf("Received message %s", msg.Msg)
        handlers := []func(*sql.DB, BotConf, message.MsgPrivate, chan fmt.Stringer) bool { handleHelp,
          handleTop, handleSpecificTop, handleScores, handlePlaceBet, handleAdminBet,
          handleBet, handleRollback, handleReset}
        for _, handler := range handlers {
          if handler(db, conf, msg, responseChannel) { break }
        }
}
