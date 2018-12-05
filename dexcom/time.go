package dexcom

import (
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

func (t Time) GetBSON() (interface{}, error) {
	return t.format(), nil
}

func (t Time) format() string {
	return t.Time.Format(TimeFormat)
}

func TimeFromRaw(raw *time.Time) *Time {
	if raw == nil {
		return nil
	}
	return &Time{
		Time: *raw,
	}
}
