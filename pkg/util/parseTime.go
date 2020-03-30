package util

import (
	"github.com/pkg/errors"
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

type Time struct {
	time.Time `json:",inline"`
}

func (t *Time) UnmarshalJSON(b []byte) error {

	// strings come with quotes on them
	ts := strings.Trim(string(b), "\"")

	parsed, err := ParseTime(ts)
	if err != nil {
		return errors.Wrap(err, "parse time")
	}
	*t = Time{parsed}
	return nil
}
