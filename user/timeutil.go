package user

import (
	"fmt"
	"strconv"
	"time"
)

func ParseTimestamp(timestamp string) (time.Time, error) {
	return time.Parse(TimestampFormat, timestamp)
}

func TimestampToUnixString(timestamp string) (unix string, err error) {
	parsed, err := ParseTimestamp(timestamp)
	if err != nil {
		return
	}
	unix = fmt.Sprintf("%v", parsed.Unix())
	return
}

func UnixStringToTimestamp(unixString string) (timestamp string, err error) {
	i, err := strconv.ParseInt(unixString, 10, 64)
	if err != nil {
		return
	}
	t := time.Unix(i, 0)
	timestamp = t.Format(TimestampFormat)
	return
}
