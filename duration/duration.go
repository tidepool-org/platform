package duration

import (
	"math"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/crypto"
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

// Exponential returns the duration multiplied by two raised to the exponent, saturating rather than overflowing. An
// exponent of zero or less returns the duration unchanged.
//
// The two bounds are not symmetric. The positive one is a floor, so a duration equal to it still multiplies within
// range, while the negative one is exact, so a duration equal to it multiplies to precisely the smallest duration.
// Both also cover an exponent of 63 or more, where the shifts yield zero and negative one for every count.
func Exponential(duration time.Duration, exponent int) time.Duration {
	if exponent <= 0 {
		return duration
	} else if duration > 0 && duration > durationMaximum>>exponent {
		return durationMaximum
	} else if duration < 0 && duration <= durationMinimum>>exponent {
		return durationMinimum
	} else {
		return time.Duration(int64(duration) * (1 << exponent))
	}
}

// WithJitter returns the duration offset by a random amount, uniformly distributed over plus or minus the specified
// jitter fraction of the duration, inclusive. The jitter is pinned to the range zero through one, so a negative jitter
// applies none at all and a jitter above one applies at most the whole duration, keeping the result on the same side of
// zero as the duration, save for a duration beyond the exact range of a float64, where the magnitude can round up past
// the duration by a few nanoseconds. The duration is returned unchanged whenever the resulting jitter magnitude is not
// a number, is less than a nanosecond, or is too large to apply.
func WithJitter(duration time.Duration, jitter float64) time.Duration {
	// A NaN jitter survives the pinning, since min and max both yield NaN for it, and is rejected below.
	jitter = min(max(jitter, 0), 1)

	// Written as a negated acceptance test rather than a rejection test, since every comparison against a NaN
	// magnitude is false, and on the float rather than the converted integer, since converting a value the integer
	// cannot hold is implementation dependent.
	durationJitterMagnitude := math.Abs(float64(duration)) * jitter
	if !(durationJitterMagnitude >= 1 && durationJitterMagnitude < durationJitterMagnitudeMaximum) {
		return duration
	}

	// Clamp the duration before applying the jitter, since applying it first could overflow.
	durationJitterMaximum := int64(durationJitterMagnitude)
	durationJitter := time.Duration(crypto.RandomInt64N(2*durationJitterMaximum+1) - durationJitterMaximum)
	if durationJitter >= 0 {
		return min(duration, durationMaximum-durationJitter) + durationJitter
	} else {
		return max(duration, durationMinimum-durationJitter) + durationJitter
	}
}

const (
	durationMaximum = time.Duration(math.MaxInt64)
	durationMinimum = time.Duration(math.MinInt64)

	// The largest jitter magnitude that can be doubled to span both directions without overflowing a duration.
	durationJitterMagnitudeMaximum = float64(int64(1) << 62)
)
