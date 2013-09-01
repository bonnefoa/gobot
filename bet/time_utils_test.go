package bet

import (
    "testing"
    "time"
)

func TestParseFunc(t *testing.T) {
	local, _ := time.LoadLocation("Local")
    ref := time.Date(2013, 9, 1, 16, 0, 0, 0, local)
    next := ParseCron("0 17 0", ref)
    if next != 3600 {
        t.Fatalf("Expected %q, got %q", 3600, next)
    }
}
