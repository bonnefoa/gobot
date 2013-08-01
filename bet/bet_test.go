package bet

import (
	"os"
	"testing"
	"github.com/bonnefoa/gobot/testing/assert"
	"time"
)

func TestParseDate(t *testing.T) {
	var err error
	now := time.Now()
	deltaDay := 0
	if now.Hour() > 11 || (now.Hour() == 11 && now.Minute() > 40) {
		deltaDay = 1
	}
	expected := time.Date(now.Year(), now.Month(),
		now.Day()+deltaDay, 11, 40, 0,
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

func checkUserScores(t *testing.T, first []UserScores, snd []UserScores) {
	if len(first) != len(snd) {
		t.Logf("Invalid size : First %v, snd %v\n", first, snd)
		t.FailNow()
	}

	for k := range first {
		assert.AssertEquals(t, first[k], snd[k])
	}
}

func TestSimpleBet(t *testing.T) {
	os.Remove("test.db")
	db := InitBase("test.db")
	GetOrCreateBet(db)
	AddUserBet(db, "after", time.Now().Add(time.Hour), false)
	AddUserBet(db, "other", time.Now().Add(time.Hour), false)
	good := time.Now().Add(time.Minute)
	AddUserBet(db, "near", good, false)
	AddUserBet(db, "near2", good, false)

	closestBet := CloserBet(db, time.Now())
	expecetedWinners := []string{"near", "near2"}
	t.Logf("Checking first bet")
	assert.AssertEquals(t, expecetedWinners[0], closestBet[0])
	assert.AssertEquals(t, expecetedWinners[1], closestBet[1])
	t.Logf("Closing first bet")
	CloseBet(db, time.Now())

	scores := GetScores(db)
	expectedScores := []UserScores{{"near", 1},
		{"near2", 1}, {"after", 0}, {"other", 0}}
	checkUserScores(t, expectedScores, scores)

	good2 := time.Now()
	good2 = good2.Add(time.Minute)
	AddUserBet(db, "near", good2, false)
	AddUserBet(db, "other", good2, false)
	AddUserBet(db, "after", good2.Add(time.Hour), false)

	CloseBet(db, good2)

	scores = GetScores(db)
	expectedScores2 := []UserScores{{"near", 2},
		{"near2", 1}, {"other", 1}, {"after", 0}}
	checkUserScores(t, expectedScores2, scores)

	RollbackLastBet(db)
	scores = GetScores(db)
	checkUserScores(t, expectedScores, scores)
}
