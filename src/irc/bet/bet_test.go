package bet

import "testing"
import "time"
import "os"
import "testing/assert"

func TestParseDate(t *testing.T) {
        var err error
        now := time.Now()
        deltaDay := 0
        if now.Hour() > 11 || (now.Hour() == 11 && now.Minute() > 40) {
          deltaDay = 1
        }
        expected := time.Date(now.Year(), now.Month(),
                now.Day() + deltaDay,11, 40, 0,
                0, now.Location())
        var res time.Time

        res, _ = ParseDate("11h40")
        assert.AssertEquals(t, expected, res)

        res, _ = ParseDate("11:40")
        assert.AssertEquals(t, expected, res)

        res, _ = ParseDate("11:40:00")
        assert.AssertEquals(t, expected, res)

        utc, _ := time.LoadLocation("UTC")
        expected = time.Date(2006, 1, 2, 15, 4, 5, 0, utc)

        res, err = ParseDate("2006-01-02T15:04:05Z")
        assert.AssertNotNil(t, err)
        assert.AssertEquals(t, expected, res)
}

func TestSimpleBet(t *testing.T) {
        os.Remove("test.db")
        db := InitBase("test.db")
        GetOrCreateBet(db)
        AddUserBet(db, "after", time.Now().Add(time.Hour), false)
        good := time.Now().Add(time.Minute)
        AddUserBet(db, "near", good, false)
        AddUserBet(db, "near2", good, false)

        closestBet := CloserBet(db, time.Now())
        expecetedWinners := []string { "near", "near2" }
        t.Logf("Checking first bet")
        assert.AssertEquals(t, expecetedWinners[0], closestBet[0])
        assert.AssertEquals(t, expecetedWinners[1], closestBet[1])
        t.Logf("Closing first bet")
        CloseBet(db, time.Now())

        var scores map[string]int
        scores = GetScores(db)
        expectedScores := map[string] int {"near":1, "near2":1,
                "before":0, "after":0}
        assert.AssertMapEquals(t, expectedScores, scores)

        good2 := time.Now()
        good2 = good2.Add(time.Minute)
        AddUserBet(db, "near", good2, false)
        AddUserBet(db, "before", good2, false)
        AddUserBet(db, "after", good2.Add(time.Hour), false)

        CloseBet(db, good2)

        scores = GetScores(db)
        expectedScores2 := map[string] int {"near":2, "near2":1,
                "before":1, "after":0}
        assert.AssertMapEquals(t, expectedScores2, scores)

        RollbackLastBet(db)
        scores = GetScores(db)
        assert.AssertMapEquals(t, expectedScores, scores)
}
