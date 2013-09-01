package bet

import (
	"time"
    "strings"
    "strconv"
)

func ConvertTimeToLocal(ts time.Time) time.Time {
	local, _ := time.LoadLocation("Local")
	ts = ts.In(local)
	return ts
}

func ConvertTimeToUTC(ts time.Time) time.Time {
	utc, _ := time.LoadLocation("UTC")
	ts = ts.In(utc)
	return ts
}

func FormatTimeInUtc(ts time.Time) string {
	ts = ConvertTimeToUTC(ts)
	return ts.Format("2006-01-02 15:04:05")
}

func ParseCron(cr string, ref time.Time) int64 {
    refWeekday := int64(ref.Weekday())

    splitted := strings.Split(cr, " ")
    min, _ := strconv.ParseInt(splitted[0], 0, 32)
    hour, _ := strconv.ParseInt(splitted[1], 0, 32)
    weekday, _ := strconv.ParseInt(splitted[2], 0, 32)

    deltaDay := refWeekday - weekday
    if refWeekday > weekday {
        deltaDay = refWeekday + 7 - weekday
    }
    ref.AddDate(0, 0, int(deltaDay))
    nextEvent := time.Date(ref.Year(), ref.Month(), ref.Day(),
        int(hour), int(min), 0, 0, nil)
    duration := ref.Sub(nextEvent)
    return int64(duration.Seconds())
}
