package common

import (
	"slices"
	"strings"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"
)

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
	iDay := d[i]
	jDay := d[j]
	iDayIndex, iErr := DayIndex(iDay)
	jDayIndex, jErr := DayIndex(jDay)
	if iErr != nil {
		if jErr != nil {
			return iDay < jDay
		} else {
			return false
		}
	} else if jErr != nil {
		return true
	} else {
		return iDayIndex < jDayIndex
	}
}

func DayIndex(day string) (int, error) {
	switch day {
	case DaySunday:
		return 1, nil
	case DayMonday:
		return 2, nil
	case DayTuesday:
		return 3, nil
	case DayWednesday:
		return 4, nil
	case DayThursday:
		return 5, nil
	case DayFriday:
		return 6, nil
	case DaySaturday:
		return 7, nil
	default:
		return 0, errors.New("invalid day of the week")
	}
}

func ValidateDayOfWeek(value string) error {
	if value == "" {
		return validator.ErrorValueEmpty()
	} else if !slices.Contains(DaysOfWeek(), strings.ToLower(value)) {
		return validator.ErrorValueStringNotOneOf(value, DaysOfWeek())
	}
	return nil
}

func DayOfWeekValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateDayOfWeek(value))
}
