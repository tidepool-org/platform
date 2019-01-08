package test

import "time"

var now = time.Now().Truncate(time.Millisecond)

func Now() time.Time {
	return now
}

func PastFarTime() time.Time {
	return now.AddDate(-30, 0, 0)
}

func PastNearTime() time.Time {
	return now.AddDate(0, -1, 0)
}

func FutureNearTime() time.Time {
	return now.AddDate(0, 1, 0)
}

func FutureFarTime() time.Time {
	return now.AddDate(30, 0, 0)
}
