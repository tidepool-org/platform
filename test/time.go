package test

import (
	"math/rand"
	"time"
)

const TimeLimit = 365 * 24 * time.Hour

func TimeZones() []string {
	return []string{
		"America/Anchorage",
		"America/Boise",
		"America/Chicago",
		"America/Denver",
		"America/Detroit",
		"America/Edmonton",
		"America/Indiana/Indianapolis",
		"America/Los_Angeles",
		"America/Montreal",
		"America/New_York",
		"America/Phoenix",
		"America/Toronto",
		"America/Vancouver",
		"America/Winnipeg",
		"Asia/Riyadh",
		"Asia/Tokyo",
		"Australia/Brisbane",
		"Australia/Melbourne",
		"Australia/Sydney",
		"Europe/Berlin",
		"Europe/Dublin",
		"Europe/Lisbon",
		"Europe/London",
		"Europe/Madrid",
		"Europe/Prague",
		"Europe/Rome",
		"Europe/Stockholm",
		"Europe/Vienna",
		"Europe/Zurich",
		"Pacific/Auckland",
		"Pacific/Honolulu",
		"US/Central",
		"US/Eastern",
		"US/Mountain",
		"US/Pacific",
	}
}

func TimeMaximum() time.Time {
	return time.Now().Add(TimeLimit)
}

func TimeMinimum() time.Time {
	return time.Now().Add(-TimeLimit)
}

func NewTime() time.Time {
	return NewTimeInRange(TimeMinimum(), TimeMaximum())
}

func NewTimeInRange(earliest time.Time, latest time.Time) time.Time {
	return earliest.Add(time.Duration(rand.Int63n(int64(latest.Sub(earliest)))))
}

func NewTimeZone() string {
	timeZones := TimeZones()
	return timeZones[rand.Intn(len(timeZones))]

}
