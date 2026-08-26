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

	Context("Exponential", func() {
		DescribeTable("returns the expected duration when",
			func(value time.Duration, exponent int, expectedDuration time.Duration) {
				Expect(duration.Exponential(value, exponent)).To(Equal(expectedDuration))
			},
			Entry("the duration is zero", time.Duration(0), 5, time.Duration(0)),
			Entry("the duration and the exponent are zero", time.Duration(0), 0, time.Duration(0)),
			Entry("the exponent is zero", time.Minute, 0, time.Minute),
			Entry("the exponent is negative", time.Minute, -1, time.Minute),
			Entry("the exponent is one", time.Minute, 1, 2*time.Minute),
			Entry("the exponent is ten", time.Second, 10, 1024*time.Second),
			Entry("the duration is negative", -time.Minute, 3, -8*time.Minute),
			// The largest duration that still multiplies within range, and the first one that does not. Half the
			// maximum doubles to one nanosecond short of it, since the maximum is odd.
			Entry("the product is the largest that fits", time.Duration(math.MaxInt64/2), 1, time.Duration(math.MaxInt64-1)),
			Entry("the product is one past the largest that fits", time.Duration(math.MaxInt64/2+1), 1, time.Duration(math.MaxInt64)),
			Entry("the product is the largest that fits at a large exponent", time.Nanosecond, 62, time.Duration(1<<62)),
			Entry("the product is one past the largest that fits at a large exponent", 2*time.Nanosecond, 62, time.Duration(math.MaxInt64)),
			// The smallest duration is a power of two, so the bound is exact and equality multiplies to it precisely.
			Entry("the product is the smallest that fits", time.Duration(math.MinInt64/2), 1, time.Duration(math.MinInt64)),
			Entry("the product is one short of the smallest that fits", time.Duration(math.MinInt64/2+1), 1, time.Duration(math.MinInt64+2)),
			Entry("the product overflows the largest duration", time.Hour, 60, time.Duration(math.MaxInt64)),
			Entry("the product overflows the smallest duration", -time.Hour, 60, time.Duration(math.MinInt64)),
			// No exponent is special cased, so the bounds alone have to hold once the shift count reaches the width.
			Entry("the exponent is the width of a duration", time.Nanosecond, 63, time.Duration(math.MaxInt64)),
			Entry("the exponent is the width of a duration and the duration is negative", -time.Nanosecond, 63, time.Duration(math.MinInt64)),
			Entry("the exponent exceeds the width of a duration", time.Nanosecond, 64, time.Duration(math.MaxInt64)),
			Entry("the exponent exceeds the width of a duration and the duration is negative", -time.Nanosecond, 64, time.Duration(math.MinInt64)),
			Entry("the exponent is far beyond the width of a duration", time.Nanosecond, 1000, time.Duration(math.MaxInt64)),
			Entry("the exponent is the largest int", time.Nanosecond, math.MaxInt, time.Duration(math.MaxInt64)),
			Entry("the exponent is the largest int and the duration is negative", -time.Nanosecond, math.MaxInt, time.Duration(math.MinInt64)),
		)
	})

	Context("WithJitter", func() {
		DescribeTable("returns the duration unchanged when",
			func(value time.Duration, jitter float64) {
				Expect(duration.WithJitter(value, jitter)).To(Equal(value))
			},
			Entry("the duration is zero", time.Duration(0), 0.2),
			Entry("the jitter is zero", time.Minute, 0.0),
			Entry("the duration and the jitter are zero", time.Duration(0), 0.0),
			// The jitter magnitude truncates to zero, leaving a single possible value.
			Entry("the jitter is too small to reach a nanosecond", time.Nanosecond, 0.2),
			Entry("the jitter is too small to reach a nanosecond and the duration is negative", -time.Nanosecond, 0.2),
			// A jitter below zero is pinned to zero, leaving nothing to apply.
			Entry("the jitter is negative", time.Minute, -0.2),
			Entry("the jitter is negative and the duration is negative", -time.Minute, -0.2),
			Entry("the jitter is negative infinity", time.Minute, math.Inf(-1)),
			// A NaN magnitude is rejected before conversion, since converting it is implementation dependent.
			Entry("the jitter is not a number", time.Minute, math.NaN()),
			Entry("the jitter is not a number and the duration is zero", time.Duration(0), math.NaN()),
			Entry("the jitter magnitude reaches the largest that can be applied", time.Duration(1<<62), 1.0),
			Entry("the jitter magnitude exceeds the largest that can be applied", time.Duration(math.MaxInt64), 0.5),
		)

		DescribeTable("returns a duration within the jitter of the duration when",
			func(value time.Duration, jitter float64, expectedJitter time.Duration) {
				for range 100 {
					Expect(duration.WithJitter(value, jitter)).To(SatisfyAll(
						BeNumerically(">=", value-expectedJitter),
						BeNumerically("<=", value+expectedJitter),
					))
				}
			},
			Entry("the duration and the jitter are positive", time.Minute, 0.2, 12*time.Second),
			Entry("the duration is negative", -time.Minute, 0.2, 12*time.Second),
			Entry("the jitter is one, so the duration may be zero", time.Minute, 1.0, time.Minute),
			// A jitter above one is pinned to one, so it applies at most the whole duration.
			Entry("the jitter is greater than one", time.Minute, 2.5, time.Minute),
			Entry("the jitter is positive infinity", time.Minute, math.Inf(1), time.Minute),
			Entry("the jitter is the smallest that reaches a nanosecond", 4*time.Nanosecond, 0.25, time.Nanosecond),
			Entry("the duration is coarse", time.Hour, 0.25, 15*time.Minute),
			Entry("the jitter magnitude is just below the largest that can be applied", time.Duration(1<<62), 0.5, time.Duration(1<<61)),
		)

		DescribeTable("returns a duration clamped to the representable range when",
			func(value time.Duration, jitter float64, expectedMinimum time.Duration, expectedMaximum time.Duration) {
				for range 100 {
					Expect(duration.WithJitter(value, jitter)).To(SatisfyAll(
						BeNumerically(">=", expectedMinimum),
						BeNumerically("<=", expectedMaximum),
					))
				}
			},
			Entry("the duration is the largest", time.Duration(math.MaxInt64), 0.25, time.Duration(math.MaxInt64-(1<<61)), time.Duration(math.MaxInt64)),
			Entry("the duration is the largest and the jitter is nearly half", time.Duration(math.MaxInt64), 0.49, time.Duration(math.MaxInt64-(1<<62)), time.Duration(math.MaxInt64)),
			Entry("the duration is the smallest", time.Duration(math.MinInt64), 0.25, time.Duration(math.MinInt64), time.Duration(math.MinInt64+(1<<61))),
			Entry("the duration is the smallest and the jitter is nearly half", time.Duration(math.MinInt64), 0.49, time.Duration(math.MinInt64), time.Duration(math.MinInt64+(1<<62))),
		)

		DescribeTable("returns every duration across the jitter magnitude, inclusive of both extremes, when",
			func(value time.Duration, jitter float64, expectedDurations []time.Duration) {
				jittered := map[time.Duration]bool{}
				for range 1000 {
					jittered[duration.WithJitter(value, jitter)] = true
				}
				Expect(jittered).To(HaveLen(len(expectedDurations)))
				for _, expectedDuration := range expectedDurations {
					Expect(jittered).To(HaveKey(expectedDuration))
				}
			},
			Entry("the jitter magnitude is one nanosecond", 5*time.Nanosecond, 0.2,
				[]time.Duration{4 * time.Nanosecond, 5 * time.Nanosecond, 6 * time.Nanosecond}),
			Entry("the duration is negative", -5*time.Nanosecond, 0.2,
				[]time.Duration{-6 * time.Nanosecond, -5 * time.Nanosecond, -4 * time.Nanosecond}),
			Entry("the jitter magnitude reaches the duration", 3*time.Nanosecond, 1.0,
				[]time.Duration{0, time.Nanosecond, 2 * time.Nanosecond, 3 * time.Nanosecond, 4 * time.Nanosecond, 5 * time.Nanosecond, 6 * time.Nanosecond}),
			// Pinning has to cap the span, not merely keep it within range, so these belong here rather than above.
			Entry("the jitter is greater than one", 3*time.Nanosecond, 5.0,
				[]time.Duration{0, time.Nanosecond, 2 * time.Nanosecond, 3 * time.Nanosecond, 4 * time.Nanosecond, 5 * time.Nanosecond, 6 * time.Nanosecond}),
			Entry("the jitter is positive infinity", 3*time.Nanosecond, math.Inf(1),
				[]time.Duration{0, time.Nanosecond, 2 * time.Nanosecond, 3 * time.Nanosecond, 4 * time.Nanosecond, 5 * time.Nanosecond, 6 * time.Nanosecond}),
		)

		DescribeTable("returns a duration on the same side of zero as the duration when",
			func(value time.Duration, jitter float64) {
				for range 1000 {
					jittered := duration.WithJitter(value, jitter)
					if value >= 0 {
						Expect(jittered).To(SatisfyAll(BeNumerically(">=", 0), BeNumerically("<=", 2*value)))
					} else {
						Expect(jittered).To(SatisfyAll(BeNumerically("<=", 0), BeNumerically(">=", 2*value)))
					}
				}
			},
			Entry("the jitter is one", time.Minute, 1.0),
			Entry("the jitter is one and the duration is negative", -time.Minute, 1.0),
			Entry("the jitter is greater than one", time.Minute, 100.0),
			Entry("the jitter is greater than one and the duration is negative", -time.Minute, 100.0),
		)

		It("returns durations both above and below the duration", func() {
			var above bool
			var below bool
			for range 1000 {
				switch jittered := duration.WithJitter(time.Minute, 0.2); {
				case jittered > time.Minute:
					above = true
				case jittered < time.Minute:
					below = true
				}
			}
			Expect(above).To(BeTrue())
			Expect(below).To(BeTrue())
		})
	})
})
