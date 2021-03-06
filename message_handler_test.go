package gobot

import (
	"testing"
	"github.com/bonnefoa/gobot/testing/assert"
	"time"
)

func TestTimezoneTranslate(t *testing.T) {
	utc, _ := time.LoadLocation("UTC")
	expected := time.Date(2006, 1, 2, 15, 4, 5, 0, utc)
	ts, _ := parseTimezoneQuery("Mon, 02 Jan 2006 15:04:05 MST in UTC")
	assert.AssertEquals(t, expected, ts)
}
