package times_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/times"
)

var _ = Describe("times", func() {
	Context("Clamp", func() {
		var first = test.RandomTime()
		var second = test.RandomTimeAfter(first)
		var third = test.RandomTimeAfter(second)

		DescribeTable("clamps the datum as expected",
			func(value time.Time, minimum time.Time, maximum time.Time, expected time.Time) {
				Expect(times.Clamp(value, minimum, maximum)).To(Equal(expected))
			},
			Entry("maximum is before minimum", first, third, second, first),
			Entry("value is before minimum", first, second, third, second),
			Entry("value is equal to minimum", first, first, third, first),
			Entry("value is between minimum and maximum", second, first, third, second),
			Entry("value is equal to minimum and maximum", second, second, second, second),
			Entry("value is equal to maximum", third, first, third, third),
			Entry("value is after maximum", third, first, second, second),
		)
	})
})
