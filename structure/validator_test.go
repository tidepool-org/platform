package structure_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/structure"
)

var _ = Describe("Validator", func() {
	Context("InRange", func() {
		DescribeTable("return true for values within the range; int",
			func(value int, minimum int, maximum int, expected bool) {
				Expect(structure.InRange(value, minimum, maximum)).To(Equal(expected))
			},
			Entry("succeeds", -1, -1, 1, true),
			Entry("succeeds", 0, 0, 1, true),
			Entry("succeeds", 1, 1, 1, true),
			Entry("succeeds", -1, 0, 1, false),
			Entry("succeeds", 2, 1, 1, false),
		)

		DescribeTable("return true for values within the range; float64",
			func(value float64, minimum float64, maximum float64, expected bool) {
				Expect(structure.InRange(value, minimum, maximum)).To(Equal(expected))
			},
			Entry("succeeds", -1.0, -1.0, 1.0, true),
			Entry("succeeds", 0.0, 0.0, 1.0, true),
			Entry("succeeds", 1.0, 1.0, 1.0, true),
			Entry("succeeds", -1.0, 0.0, 1.0, false),
			Entry("succeeds", 2.0, 1.0, 1.0, false),
		)

		DescribeTable("return true for values within the range; time.Duration",
			func(value time.Duration, minimum time.Duration, maximum time.Duration, expected bool) {
				Expect(structure.InRange(value, minimum, maximum)).To(Equal(expected))
			},
			Entry("succeeds", -1*time.Second, -1*time.Second, 1*time.Second, true),
			Entry("succeeds", 0*time.Second, 0*time.Second, 1*time.Second, true),
			Entry("succeeds", 1*time.Second, 1*time.Second, 1*time.Second, true),
			Entry("succeeds", -1*time.Second, 0*time.Second, 1*time.Second, false),
			Entry("succeeds", 2*time.Second, 1*time.Second, 1*time.Second, false),
		)
	})
})
