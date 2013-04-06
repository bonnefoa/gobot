package bet

import (
        "time"
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

func FormatTimeInUTC(ts time.Time) string {
        ts = ConvertTimeToUTC(ts)
        return ts.Format("2006-01-02 15:04:05")
}
