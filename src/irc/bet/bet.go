package bet

import (
        "database/sql"
        _ "github.com/mattn/go-sqlite3"
        "log"
        "time"
)

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
        res := GetBet(db)
        if res != 0 { return res }
        execInsert(db, "INSERT INTO bet values (NULL, NULL);")
        return GetBet(db)
}

func GetBet(db *sql.DB) int {
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

func AddUserBet(db *sql.DB, nick string, ts time.Time) {
        timeStr := ts.Format("2006-01-02 15:04:05")
        cur := GetOrCreateBet(db)
        if cur == 0 { log.Fatal("Could not get bet\n") }
        log.Printf("User %s is setting a bet (id %d) at %s\n", nick, cur, timeStr)
        execInsert(db, "REPLACE INTO userBet values (?, ?, ?);", nick , cur, timeStr)
        execInsert(db, "INSERT OR IGNORE INTO betScore(nick) values (?);", nick)
}

func CloserBet(db *sql.DB, ts time.Time) []string {
        var err error
        curBet := GetBet(db)
        if curBet == 0 { return []string {} }
        timeStr := ts.Format("2006-01-02 15:04:05")
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

func IncrementNicks(db *sql.DB, nicks []string) {
        for _, nick := range nicks {
                log.Printf("Incrementing nick %s\n", nick)
                execInsert(db,
                "UPDATE betScore SET score=score+1 WHERE nick = ?;", nick)
        }
}

func CloseBet(db *sql.DB, ts time.Time) []string {
        timeStr := ts.Format("2006-01-02 15:04:05")
        log.Printf("Closing bet with time %s\n", timeStr)
        nicks := CloserBet(db, ts)
        IncrementNicks(db, nicks)
        execInsert(db, "UPDATE bet SET time = ? WHERE time IS NULL;", timeStr)
        return nicks
}

func GetScores(db *sql.DB) map[string] int {
        rows := prepareSelect(db, `
             SELECT nick, score
             FROM betScore
             ORDER BY score DESC;
             `)
        var err error
        res := make(map[string]int)
        for rows.Next() {
                var nick string
                var score int
                err = rows.Scan(&nick, &score)
                res[nick] = score
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
                FOREIGN KEY(betId) REFERENCES bet(id)
        );`)
        return db
}

func CheckBetMessage(msg string) (time.Time, error) {
        now := time.Now()
        ts, err := time.Parse("15h04", msg)
        if err != nil {
                return now, err
        }
        res := time.Date(now.Year(), now.Month(), now.Day(),
                ts.Hour(), ts.Minute(),0 ,0, now.Location())
        return res, err
}
