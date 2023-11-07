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
})
