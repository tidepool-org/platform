package common_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/structure/validator"
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
			func(day string, expectedIndex int, expectedErr error) {
				actualIndex, actualError := common.DayIndex(day)
				Expect(actualIndex).To(Equal(expectedIndex))
				if expectedErr == nil {
					Expect(actualError).To(BeNil())
				} else {
					Expect(actualError).To(Equal(expectedErr))
				}
			},
			Entry("is an empty string", "", 0, errors.New("invalid day of the week")),
			Entry("is sunday", "sunday", 1, nil),
			Entry("is constant sunday", common.DaySunday, 1, nil),
			Entry("is monday", "monday", 2, nil),
			Entry("is constant monday", common.DayMonday, 2, nil),
			Entry("is tuesday", "tuesday", 3, nil),
			Entry("is constant tuesday", common.DayTuesday, 3, nil),
			Entry("is wednesday", "wednesday", 4, nil),
			Entry("is constant  wednesday", common.DayWednesday, 4, nil),
			Entry("is thursday", "thursday", 5, nil),
			Entry("is constant thursday", common.DayThursday, 5, nil),
			Entry("is friday", "friday", 6, nil),
			Entry("is constant friday", common.DayFriday, 6, nil),
			Entry("is saturday", "saturday", 7, nil),
			Entry("is constant saturday", common.DaySaturday, 7, nil),
			Entry("is an invalid string", "invalid", 0, errors.New("invalid day of the week")),
		)
	})

	Context("ValidateDayOfWeek", func() {
		DescribeTable("return error when invalid",
			func(day string, expectedErr error) {
				actualError := common.ValidateDayOfWeek(day)
				if expectedErr == nil {
					Expect(actualError).To(BeNil())
				} else {
					Expect(actualError.Error()).To(Equal(expectedErr.Error()))
				}
			},
			Entry("ok when same case", "tuesday", nil),
			Entry("ok when mixed case", "FriDAY", nil),
			Entry("ok when uppercase", "SUNDAY", nil),
			Entry("invalid when not a day of the week", "monday2", validator.ErrorValueStringNotOneOf("monday2", common.DaysOfWeek())),
		)
	})
})
