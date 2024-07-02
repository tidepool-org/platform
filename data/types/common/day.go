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

func DaysOfWeek() []string {
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

type DaysOfWeekByDayIndex []string

func (d DaysOfWeekByDayIndex) Len() int {
	return len(d)
}
func (d DaysOfWeekByDayIndex) Swap(i int, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d DaysOfWeekByDayIndex) Less(i int, j int) bool {
	return DayIndex(d[i]) < DayIndex(d[j])
}

func DayIndex(day string) int {
	switch day {
	case DaySunday:
		return 1
	case DayMonday:
		return 2
	case DayTuesday:
		return 3
	case DayWednesday:
		return 4
	case DayThursday:
		return 5
	case DayFriday:
		return 6
	case DaySaturday:
		return 7
	default:
		return 0
	}
}
