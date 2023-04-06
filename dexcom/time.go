package dexcom

import (
	"regexp"
	"strings"
	"time"
)

type Time struct {
	time.Time
}

func NewTime() *Time {
	return &Time{}
}

func (t *Time) Raw() *time.Time {
	if t == nil {
		return nil
	}
	return &t.Time
}

func (t Time) MarshalText() ([]byte, error) {
	return []byte(t.format()), nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.format() + `"`), nil
}

func (t Time) MarshalBSON() (interface{}, error) {
	return t.format(), nil
}

func (t Time) format() string {
	return t.Time.Format(TimeFormatMilli)
}

func TimeFromRaw(raw *time.Time) *Time {
	if raw == nil {
		return nil
	}
	return &Time{
		Time: *raw,
	}
}

func TimeFromString(raw *string) *Time {
	if raw == nil {
		return nil
	}

	tzRegex := "(?:Z|[+-](?:2[0-3]|[01][0-9]):[0-5][0-9])$"

	stringValue := *raw
	var err error
	var timeValue time.Time

	if strings.HasSuffix(stringValue, "Z") {
		timeValue, err = time.Parse(TimeFormatMilliUTC, stringValue)
		if err != nil {
			return nil
		}
	} else {
		hasZone, _ := regexp.MatchString(tzRegex, stringValue)
		if hasZone {
			timeValue, err = time.Parse(TimeFormatMilliZ, stringValue)
			if err != nil {
				return nil
			}
		} else {
			timeValue, err = time.Parse(TimeFormatMilli, stringValue)
			if err != nil {
				return nil
			}
		}
	}

	return &Time{
		Time: timeValue,
	}
}
