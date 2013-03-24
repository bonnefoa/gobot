package bet

import "testing"
import "time"
import "os"
import "testing/assert"

func testSimpleBet(t *testing.T) {
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

        scores := GetScores(db)
        expectedScores := map[string] int {"near":1, "near2":1}
        assert.AssertEquals(t, expectedScores, scores)
}
