package bet

import (
        "database/sql"
        _ "github.com/mattn/go-sqlite3"
        "log"
        "time"
        "errors"
)

type UserScores struct {
        Nick string
        Score int
}

type UserBet struct {
        Nick string
        Time time.Time
}

func (user *UserScores) String() string { return user.Nick }
func (user *UserBet) String() string { return user.Nick }

func createTable(db *sql.DB, query string) {
        tx, err := db.Begin()
        _, err = tx.Exec(query)
        err = tx.Commit()
        if err != nil { log.Fatal(err) }
}

func prepareSelect(db *sql.DB, query string, parameters... interface{}) *sql.Rows {
        var err error
        var stmt *sql.Stmt
        var rows *sql.Rows
        stmt, err = db.Prepare(query)
        rows, err = stmt.Query(parameters...)
        if err != nil { log.Fatalf("Select error %q\n", err) }
        return rows
}

func execInsert(db *sql.DB, query string, parameters... interface{}) {
        var err error
        var tx *sql.Tx
        var stmt  *sql.Stmt
        stmt, err = db.Prepare(query)
        tx, err = db.Begin()
        _, err = tx.Stmt(stmt).Exec(parameters...)
        err = tx.Commit()
        if err != nil { log.Fatalf("Insert error %q\n", err) }
}

func CreateUser(db *sql.DB, nick string) {
        execInsert(db, "REPLACE INTO user values (?);", nick)
}

func GetOrCreateBet(db *sql.DB) int {
        res := GetCurrentBet(db)
        if res != 0 { return res }
        execInsert(db, "INSERT INTO bet values (NULL, NULL);")
        return GetCurrentBet(db)
}

func GetLastBet(db *sql.DB) (int, time.Time) {
        var err error
        rows := prepareSelect(db, "SELECT id, time FROM bet ORDER BY time DESC LIMIT 1;")
        var id int
        var ts time.Time
        if rows.Next() {
                err = rows.Scan(&id, &ts)
                log.Printf("id %s, ts val %s\n", id, ts)
                err = rows.Close()
                if err != nil { log.Fatal(err) }
                return id, ts
        }
        return 0, time.Now()
}

func GetCurrentBet(db *sql.DB) int {
        var err error
        rows := prepareSelect(db, "SELECT id FROM bet WHERE time IS NULL;")
        var id int
        if rows.Next() {
                err = rows.Scan(&id)
                err = rows.Close()
                if err != nil { log.Fatal(err) }
                return id
        }
        return 0
}

func AddUserBet(db *sql.DB, nick string, ts time.Time, isAdmin bool) error {
        now := ConvertTimeToUTC(time.Now())
        ts = ConvertTimeToUTC(ts)
        timeStr := FormatTimeInUtc(ts)
        cur := GetOrCreateBet(db)
        if cur == 0 { log.Fatal("Could not get bet\n") }
        existing_ts := GetUserBet(db, cur, nick)
        if !isAdmin && ts.Before(now) {
            return errors.New("You are betting in the past, noob.")
        }
        if !isAdmin && existing_ts != nil {
          if existing_ts.After(now) && existing_ts.Before(now.Add(time.Minute * 4))  {
              return errors.New("Your bet time is in less than 4 minutes, you can't change it.")
          }
          if existing_ts.Before(now) && existing_ts.Add(time.Hour * 1).After(now) {
              return errors.New("Your bet is in the past, you can't change it.")
          }
      }
      log.Printf("User %s is setting a bet (id %d) at %s\n", nick, cur, timeStr)
      execInsert(db, "REPLACE INTO userBet values (?, ?, ?);", nick , cur, timeStr)
      execInsert(db, "INSERT OR IGNORE INTO betScore(nick) values (?);", nick)
      return nil
}

func CloserBet(db *sql.DB, ts time.Time) []string {
        var err error
        curBet := GetCurrentBet(db)
        if curBet == 0 { return []string {} }
        timeStr := FormatTimeInUtc(ts)
        log.Printf("Getting closer bet to %s\n", timeStr)
        rows := prepareSelect(db, `
             SELECT nick
             FROM userBet,
                  (SELECT min( ABS( julianday(userBet.time) - julianday(?) ) )
                   AS minTime FROM userBet WHERE betId = ? ) m
             WHERE m.minTime = ABS( julianday(userBet.time) - julianday(?) )
                   AND betId = ?;
            `, timeStr, curBet, timeStr, curBet)
        var nicks []string
        for rows.Next() {
                var nick string
                err = rows.Scan(&nick)
                nicks = append(nicks, nick)
        }
        err = rows.Close()
        if err != nil { log.Fatal(err) }
        return nicks
}

func IncrementNicks(db *sql.DB, nicks []string, val int) {
        for _, nick := range nicks {
                log.Printf("Incrementing nick %s by %d\n", nick, val)
                execInsert(db,
                "UPDATE betScore SET score=score+? WHERE nick = ?;", val, nick)
        }
}

func CloseBet(db *sql.DB, ts time.Time) []string {
        ts = ConvertTimeToUTC(ts)
        timeStr := FormatTimeInUtc(ts)
        log.Printf("Closing bet with time %s\n", timeStr)
        nicks := CloserBet(db, ts)
        IncrementNicks(db, nicks, 1)
        execInsert(db, "UPDATE bet SET time = ? WHERE time IS NULL;", timeStr)
        return nicks
}

func ResetBet(db *sql.DB) {
  cur := GetOrCreateBet(db)
  execInsert(db, "DELETE FROM userBet WHERE betId = ?;", cur)
}

func RollbackLastBet(db *sql.DB) {
         execInsert(db, "DELETE FROM bet WHERE time IS NULL;")
         id, ts := GetLastBet(db)
         execInsert(db, "UPDATE bet SET time = NULL WHERE id = ?;", id)
         nicks := CloserBet(db, ts)
         IncrementNicks(db, nicks, -1)
}

func GetUserBet(db *sql.DB, betId int, nick string) * time.Time {
        rows := prepareSelect(db, `
             SELECT time
             FROM userBet
             WHERE betId = ? AND nick = ?;
             `, betId, nick)
        var err error
        for rows.Next() {
                var ts time.Time
                err = rows.Scan(&ts)
                err = rows.Close()
                if err != nil { log.Fatal(err) }
                return &ts
        }
        return nil
}

func GetUserBets(db *sql.DB, betId int) []UserBet {
        rows := prepareSelect(db, `
             SELECT nick, time
             FROM userBet
             WHERE betId = ?
             ORDER BY time ASC;
             `, betId)
        var err error
        res := make([]UserBet, 0)
        for rows.Next() {
                var nick string
                var ts time.Time
                err = rows.Scan(&nick, &ts)
                res = append(res, UserBet{nick, ts})
        }
        err = rows.Close()
        if err != nil { log.Fatal(err) }
        return res
}

func GetUsers(db *sql.DB) []string {
        rows := prepareSelect(db, `
             SELECT nick
             FROM betScore;
             `)
        var err error
        res := make([]string, 0)
        for rows.Next() {
                var nick string
                err = rows.Scan(&nick)
                res = append(res, nick)
        }
        err = rows.Close()
        if err != nil { log.Fatal(err) }
        return res
}

func GetScores(db *sql.DB) []UserScores {
        rows := prepareSelect(db, `
             SELECT nick, score
             FROM betScore
             ORDER BY score DESC, nick ASC;
             `)
        var err error
        res := make([]UserScores, 0)
        for rows.Next() {
                var nick string
                var score int
                err = rows.Scan(&nick, &score)
                res = append(res, UserScores{nick, score} )
        }
        err = rows.Close()
        if err != nil { log.Fatal(err) }
        return res
}

func InitBase(dbPath string) *sql.DB {
        var db *sql.DB
        var err error
        db, err = sql.Open("sqlite3", dbPath)
        if err != nil {
                log.Fatal(err)
        }
        createTable(db, `create table if not exists betScore(
                nick text not null primary key,
                score integer DEFAULT 0
        );`)
        createTable(db, `create table if not exists bet(
                id integer not null primary key,
                time datetime
        );`)
        createTable(db, `create table if not exists userBet(
                nick text not null,
                betId not null,
                time datetime not null,
                PRIMARY KEY (nick, betId)
        );`)
        return db
}

func ParseHourLayout(msg string) (time.Time, error) {
        now := time.Now()
        var err error
        var ts time.Time
        hour_layouts := []string {  "15h04", "15:04", "15:04:05" }
        for _, layout := range hour_layouts {
          ts, err = time.Parse(layout, msg)
          if err == nil { break }
        }
        if err != nil {
                return now, err
        }
        res := time.Date(now.Year(), now.Month(), now.Day(),
                ts.Hour(), ts.Minute(), ts.Second(), 0, now.Location())
        if res.Before(now) {
          res = res.Add(time.Hour * 24)
        }
        return res, err
}

func ParseDate(msg string) (time.Time, error) {
        var err error
        var ts time.Time

        layouts := []string { time.RFC822, time.RFC850, time.RFC3339, time.RFC1123 }
        for _, layout := range layouts  {
          ts, err = time.Parse(layout, msg)
          if err == nil { return ts, err }
        }
        return ParseHourLayout(msg)
}
