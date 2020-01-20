package util

import (
	"strings"
	"time"
)

func ParseTime(ts string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, ts)
	if err == nil {
		return parsed, nil
	}

	ts = strings.TrimSuffix(ts, "+0000 (UTC)")
	return time.Parse("Mon Jan 02 2006 15:04:05 MST", ts)
}
