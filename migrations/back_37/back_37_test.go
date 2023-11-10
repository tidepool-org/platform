package main

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
)

var _ = Describe("back-37", func() {

	Context("getBGValuePlatformPrecision", func() {
		DescribeTable("return the expected mmol/L value",
			func(jellyfishVal float64, expectedVal float64) {
				actual := getBGValuePlatformPrecision(jellyfishVal)
				Expect(actual).To(Equal(expectedVal))
			},
			Entry("original mmol/L value", 10.1, 10.1),
			Entry("converted mgd/L of 100", 5.550747991045533, 5.55075),
			Entry("converted mgd/L of 150", 8.3261219865683, 8.32612),
			Entry("converted mgd/L of 65", 3.6079861941795968, 3.60799),
		)
	})

	Context("updateIfExistsPumpSettingsBolus", func() {
		var bolusData map[string]interface{}

		BeforeEach(func() {
			bolusData = map[string]interface{}{
				"bolous-1": pumpTest.NewRandomBolus(),
				"bolous-2": pumpTest.NewRandomBolus(),
			}
			Expect(bolusData).ToNot(BeNil())
		})

		DescribeTable("should",
			func(input bson.M, expected interface{}, expectError bool) {
				actual, err := updateIfExistsPumpSettingsBolus(input)
				if expectError {
					Expect(err).ToNot(BeNil())
					Expect(actual).To(BeNil())
					return
				}
				Expect(err).To(BeNil())
				if expected != nil {
					Expect(actual).To(Equal(expected))
				} else {
					Expect(actual).To(BeNil())
				}
			},
			Entry("do nothing when wrong type", bson.M{"type": "other"}, nil, false),
			Entry("do nothing when has no bolus", bson.M{"type": "pumpSettings"}, nil, false),
			Entry("error when bolus is invalid", bson.M{"type": "pumpSettings", "bolus": "wrong"}, nil, true),
			Entry("return bolus when valid", bson.M{"type": "pumpSettings", "bolus": bolusData}, bolusData, false),
		)
	})

	Context("updateIfExistsPumpSettingsSleepSchedules", func() {
		var sleepSchedulesExpected *pump.SleepScheduleMap
		var sleepSchedulesStored *pump.SleepScheduleMap
		var sleepSchedulesInvalidDays *pump.SleepScheduleMap

		BeforeEach(func() {
			sleepSchedulesExpected = &pump.SleepScheduleMap{
				"schedule-1": pumpTest.RandomSleepSchedule(),
				"schedule-2": pumpTest.RandomSleepSchedule(),
			}
			sleepSchedulesInvalidDays = pumpTest.CloneSleepSchedules(sleepSchedulesExpected)
			(*sleepSchedulesInvalidDays)["schedule-2"].Days = &[]string{"not-a-day", common.DayFriday}

			sleepSchedulesStored = pumpTest.CloneSleepSchedules(sleepSchedulesExpected)

			s1Days := (*sleepSchedulesStored)["schedule-1"].Days
			for key, day := range *s1Days {
				(*s1Days)[key] = strings.ToUpper(day)
			}
			(*sleepSchedulesStored)["schedule-1"].Days = s1Days

			s2Days := (*sleepSchedulesStored)["schedule-2"].Days
			for key, day := range *s2Days {
				(*s2Days)[key] = strings.ToUpper(day)
			}
			(*sleepSchedulesStored)["schedule-2"].Days = s2Days

			//ensure sorting
			sleepSchedulesExpected.Normalize(normalizer.New())

			Expect(sleepSchedulesExpected).ToNot(BeNil())
			Expect(sleepSchedulesStored).ToNot(BeNil())
			Expect(sleepSchedulesInvalidDays).ToNot(BeNil())
		})

		It("does nothing when wrong type", func() {
			actual, err := updateIfExistsPumpSettingsSleepSchedules(bson.M{"type": "other"})
			Expect(err).To(BeNil())
			Expect(actual).To(BeNil())
		})
		It("does nothing when no sleepSchedules", func() {
			actual, err := updateIfExistsPumpSettingsSleepSchedules(bson.M{"type": "pumpSettings"})
			Expect(err).To(BeNil())
			Expect(actual).To(BeNil())
		})
		It("returns error when sleepSchedules is invalid", func() {
			actual, err := updateIfExistsPumpSettingsSleepSchedules(bson.M{"type": "pumpSettings", "sleepSchedules": "wrong"})
			Expect(err).ToNot(BeNil())
			Expect(actual).To(BeNil())
		})
		It("returns updated sleepSchedules when valid", func() {
			actual, err := updateIfExistsPumpSettingsSleepSchedules(bson.M{"type": "pumpSettings", "sleepSchedules": sleepSchedulesStored})
			Expect(err).To(BeNil())
			Expect(actual).ToNot(BeNil())
			actualSchedules, ok := actual.(*pump.SleepScheduleMap)
			Expect(ok).To(BeTrue())
			Expect(actualSchedules).To(Equal(sleepSchedulesExpected))
		})
		It("returns updated sleepSchedules when valid", func() {
			actual, err := updateIfExistsPumpSettingsSleepSchedules(bson.M{"type": "pumpSettings", "sleepSchedules": sleepSchedulesInvalidDays})
			Expect(err).ToNot(BeNil())
			Expect(actual).To(BeNil())
		})
	})
})
