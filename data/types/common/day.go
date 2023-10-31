package common

const (
	DaySunday    = "sunday"
	DayMonday    = "monday"
	DayTuesday   = "tuesday"
	DayWednesday = "wednesday"
	DayThursday  = "thursday"
	DayFriday    = "friday"
	DaySaturday  = "saturday"
)

func Days() []string {
	return []string{
		DaySunday,
		DayMonday,
		DayTuesday,
		DayWednesday,
		DayThursday,
		DayFriday,
		DaySaturday,
	}
}
