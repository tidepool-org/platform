package duration

import (
	"math"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/errors"
)

// Parse returns the duration represented by specified value. Value may be a standard Go duration string
// or a number string without units, which is then assumed to be in the specified units.
func Parse(value string, units time.Duration) (time.Duration, error) {
	if units <= 0 {
		return 0, errors.New("units is invalid")
	}

	// Attempt to parse as standard Go duration string first.
	if valueParsed, err := time.ParseDuration(value); err == nil {
		return valueParsed, nil
	}

	// Determine minimum and maximum possible values in the specified units.
	valueMinimum := math.MinInt64 / int64(units)
	valueMaximum := math.MaxInt64 / int64(units)

	// Attempt to parse as an integer then float without units, which is then assumed to be in the specified
	// units. Integer values are inclusive of the minimum and maximum, since the multiply below is exact. Float
	// values are exclusive, because float64 rounds the minimum and maximum themselves past the true bounds
	// whenever either exceeds 2^53, which happens for units finer than 1024ns.
	if valueInt, err := strconv.ParseInt(value, 10, 64); err == nil {
		if valueInt >= valueMinimum && valueInt <= valueMaximum {
			return time.Duration(valueInt * int64(units)), nil
		}
	} else if valueFloat, err := strconv.ParseFloat(value, 64); err == nil {
		if valueFloat > float64(valueMinimum) && valueFloat < float64(valueMaximum) {
			return time.Duration(valueFloat * float64(units)), nil
		}
	}

	return 0, errors.New("unable to parse duration")
}
