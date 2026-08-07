package keycloak

import (
	"fmt"
	"strconv"
	"time"
)

// timestampFormat is the format used for the terms accepted timestamp of a
// user.
const timestampFormat = "2006-01-02T15:04:05-07:00"

func timestampToUnixString(timestamp string) (string, error) {
	parsed, err := time.Parse(timestampFormat, timestamp)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", parsed.Unix()), nil
}

func unixStringToTimestamp(unixString string) (string, error) {
	i, err := strconv.ParseInt(unixString, 10, 64)
	if err != nil {
		return "", err
	}
	return time.Unix(i, 0).Format(timestampFormat), nil
}
