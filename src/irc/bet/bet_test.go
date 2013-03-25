package bet

import "testing"
import "time"
import "os"
import "testing/assert"

func TestCheckBetMessage(t *testing.T) {
        now := time.Now()
        expected := time.Date(now.Year(), now.Month(), now.Day(),
                11, 40, 0, 0, now.Location())
        res, _ := CheckBetMessage("11h40")
        assert.AssertEquals(t, expected, res)

        _, err := CheckBetMessage("11h40")
        assert.AssertNotNil(t, err)
}

func TestSimpleBet(t *testing.T) {
        os.Remove("test.db")
        db := InitBase("test.db")
        GetOrCreateBet(db)
        AddUserBet(db, "before", time.Now().Add(time.Hour))
        AddUserBet(db, "after", time.Now().Add(-time.Hour))
        good := time.Now()
        AddUserBet(db, "near", good)
        AddUserBet(db, "near2", good)

        closestBet := CloserBet(db, time.Now())
        expecetedWinners := []string { "near", "near2" }
        assert.AssertEquals(t, expecetedWinners[0], closestBet[0])
        assert.AssertEquals(t, expecetedWinners[1], closestBet[1])
        CloseBet(db, time.Now())

        var scores map[string]int
        scores = GetScores(db)
        expectedScores := map[string] int {"near":1, "near2":1,
                "before":0, "after":0}
        assert.AssertMapEquals(t, expectedScores, scores)

        good2 := time.Now()
        AddUserBet(db, "near", good2)
        AddUserBet(db, "before", good2)
        AddUserBet(db, "after", good2.Add(time.Hour))

        CloseBet(db, good2)

        scores = GetScores(db)
        expectedScores2 := map[string] int {"near":2, "near2":1,
                "before":1, "after":0}
        assert.AssertMapEquals(t, expectedScores2, scores)
}
