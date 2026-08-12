package duration_test

import (
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/duration"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Duration", func() {
	Context("Parse", func() {
		DescribeTable("returns the expected duration when the value",
			func(value string, units time.Duration, expectedDuration time.Duration) {
				Expect(test.MustDuration(duration.Parse(value, units))).To(Equal(expectedDuration))
			},
			Entry("has a single unit", "45s", time.Second, 45*time.Second),
			Entry("has multiple units", "1m30s", time.Second, 90*time.Second),
			Entry("has a fractional unit", "1.5h", time.Second, 90*time.Minute),
			Entry("has units, which override the units argument", "45s", time.Hour, 45*time.Second),
			Entry("has units and is negative", "-1m", time.Second, -time.Minute),
			Entry("has no units, using the units argument", "45", time.Second, 45*time.Second),
			Entry("has no units, using non-second units", "45", time.Millisecond, 45*time.Millisecond),
			Entry("has no units and is fractional", "1.5", time.Second, 1500*time.Millisecond),
			Entry("has no units and is fractional, using non-second units", "1.5", time.Hour, 90*time.Minute),
			Entry("has no units and is in exponent notation", "1e3", time.Second, 1000*time.Second),
			Entry("has no units and is negative", "-3", time.Second, -3*time.Second),
			Entry("has no units and is zero", "0", time.Second, time.Duration(0)),
			Entry("has no units and is finer than a nanosecond", "0.0000000001", time.Second, time.Duration(0)),
			// The integer bounds are inclusive and the multiply is exact, so the extremes convert without loss.
			Entry("has no units and is the largest duration", "9223372036854775807", time.Nanosecond, time.Duration(math.MaxInt64)),
			Entry("has no units and is the smallest duration", "-9223372036854775808", time.Nanosecond, time.Duration(math.MinInt64)),
			Entry("has no units and is the largest whole number of units", "9223372036", time.Second, 9223372036*time.Second),
			Entry("has no units and is the smallest whole number of units", "-9223372036", time.Second, -9223372036*time.Second),
			Entry("has no units and is an integer beyond float64 precision", "9007199254740993", time.Nanosecond, time.Duration(9007199254740993)),
			// The float bounds are exclusive, and float64 carries only 53 bits, so these two do lose the low bits.
			Entry("has no units, is fractional, and is the largest float64 below the largest duration", "9223372036854774784.0", time.Nanosecond, time.Duration(9223372036854774784)),
			Entry("has no units, is fractional, and is beyond float64 precision", "9223372035.5", time.Second, time.Duration(9223372035500000256)),
		)

		DescribeTable("returns an error when the value",
			func(value string, units time.Duration) {
				parsedDuration, err := duration.Parse(value, units)
				Expect(err).To(MatchError("unable to parse duration"))
				Expect(parsedDuration).To(BeZero())
			},
			Entry("is empty", "", time.Second),
			Entry("is not a number or a duration", "invalid", time.Second),
			Entry("has a leading space", " 45", time.Second),
			Entry("has a trailing space", "45 ", time.Second),
			Entry("has an unrecognized unit", "45y", time.Second),
			Entry("has an upper case unit", "45S", time.Second),
			Entry("has trailing characters after the units", "45sx", time.Second),
			Entry("is hexadecimal", "0x10", time.Second),
			Entry("is not a number", "NaN", time.Second),
			Entry("is not a number in lower case", "nan", time.Second),
			Entry("is positive infinity", "Inf", time.Second),
			Entry("is negative infinity", "-Inf", time.Second),
			Entry("is infinity spelled out", "infinity", time.Second),
			Entry("overflows the units in exponent notation", "1e30", time.Second),
			Entry("overflows the units", "9223372036854775807", time.Second),
			Entry("is one past the largest whole number of units", "9223372037", time.Second),
			Entry("is one past the smallest whole number of units", "-9223372037", time.Second),
			Entry("is one past the largest duration", "9223372036854775808", time.Nanosecond),
			// The float bounds are exclusive, so a fraction above the largest whole number of units is given up.
			Entry("is fractional and above the largest whole number of units", "9223372036.5", time.Second),
		)

		DescribeTable("returns an error when the units",
			func(value string, units time.Duration) {
				parsedDuration, err := duration.Parse(value, units)
				Expect(err).To(MatchError("units is invalid"))
				Expect(parsedDuration).To(BeZero())
			},
			Entry("is zero", "45", time.Duration(0)),
			Entry("is negative", "45", -time.Second),
			Entry("is the smallest duration", "45", time.Duration(math.MinInt64)),
			Entry("is zero and the value has units", "45s", time.Duration(0)),
			Entry("is zero and the value is not parsable", "invalid", time.Duration(0)),
		)
	})
})
