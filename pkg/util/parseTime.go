package util

import (
	"strings"
	"time"
)

func ParseTime(ts string) (time.Time, error) {
	ts = strings.TrimSuffix(ts, "+0000 (UTC)")

	createdAt, err := time.Parse("Mon Jan 02 2006 15:04:05 MST", ts)

	return createdAt, err
}
