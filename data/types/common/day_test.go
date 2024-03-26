package common_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/common"
)

var _ = Describe("Day", func() {

	It("DaySunday is expected", func() {
		Expect(common.DaySunday).To(Equal("sunday"))
	})

	It("DayMonday is expected", func() {
		Expect(common.DayMonday).To(Equal("monday"))
	})

	It("DayTuesday is expected", func() {
		Expect(common.DayTuesday).To(Equal("tuesday"))
	})

	It("DayWednesday is expected", func() {
		Expect(common.DayWednesday).To(Equal("wednesday"))
	})

	It("DayThursday is expected", func() {
		Expect(common.DayThursday).To(Equal("thursday"))
	})

	It("DayFriday is expected", func() {
		Expect(common.DayFriday).To(Equal("friday"))
	})

	It("DaySaturday is expected", func() {
		Expect(common.DaySaturday).To(Equal("saturday"))
	})

	It("DaysOfWeek returns expected", func() {
		Expect(common.DaysOfWeek()).To(Equal([]string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}))
		Expect(common.DaysOfWeek()).To(Equal([]string{
			common.DaySunday,
			common.DayMonday,
			common.DayTuesday,
			common.DayWednesday,
			common.DayThursday,
			common.DayFriday,
			common.DaySaturday,
		}))
	})

	Context("DayIndex", func() {
		DescribeTable("return the expected index when the day",
			func(day string, expectedIndex int) {
				Expect(common.DayIndex(day)).To(Equal(expectedIndex))
			},
			Entry("is an empty string", "", 0),
			Entry("is sunday", "sunday", 1),
			Entry("is constant sunday", common.DaySunday, 1),
			Entry("is monday", "monday", 2),
			Entry("is constant monday", common.DayMonday, 2),
			Entry("is tuesday", "tuesday", 3),
			Entry("is constant tuesday", common.DayTuesday, 3),
			Entry("is wednesday", "wednesday", 4),
			Entry("isconstant  wednesday", common.DayWednesday, 4),
			Entry("is thursday", "thursday", 5),
			Entry("is constant thursday", common.DayThursday, 5),
			Entry("is friday", "friday", 6),
			Entry("is constant friday", common.DayFriday, 6),
			Entry("is saturday", "saturday", 7),
			Entry("is constant saturday", common.DaySaturday, 7),
			Entry("is an invalid string", "invalid", 0),
		)
	})
})
