package times

import "time"

func Clamp(value time.Time, minimum time.Time, maximum time.Time) time.Time {
	if maximum.Before(minimum) {
		return value
	} else if value.Before(minimum) {
		return minimum
	} else if value.After(maximum) {
		return maximum
	} else {
		return value
	}
}
