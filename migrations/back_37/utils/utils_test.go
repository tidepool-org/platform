package utils_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	"github.com/tidepool-org/platform/migrations/back_37/utils"
)

var _ = Describe("back-37", func() {

	var _ = Describe("utils", func() {

		Context("GetBGValuePlatformPrecision", func() {
			DescribeTable("return the expected mmol/L value",
				func(jellyfishVal float64, expectedVal float64) {
					actual := utils.GetBGValuePlatformPrecision(jellyfishVal)
					Expect(actual).To(Equal(expectedVal))
				},
				Entry("original mmol/L value", 10.1, 10.1),
				Entry("converted mgd/L of 100", 5.550747991045533, 5.55075),
				Entry("converted mgd/L of 150", 8.3261219865683, 8.32612),
				Entry("converted mgd/L of 65", 3.6079861941795968, 3.60799),
			)
		})

		var _ = Describe("GetDatumUpdates", func() {
			var existingDatum bson.M

			BeforeEach(func() {
				existingDatum = bson.M{}
				existingDatum["_userId"] = "some-user-id"
				existingDatum["deviceId"] = "some-device-id"
				theTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
				existingDatum["time"] = theTime
				existingDatum["type"] = "bolous"
				Expect(existingDatum).ToNot(BeNil())
			})
			Context("_deduplicator hash", func() {
				DescribeTable("should",
					func(getInput func() bson.M, expected bson.M, expectError bool) {
						input := getInput()
						actual, err := utils.GetDatumUpdates(input)
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
					Entry("error when missing _userId", func() bson.M {
						existingDatum["_userId"] = nil
						return existingDatum
					}, nil, true),
					Entry("error when missing deviceId", func() bson.M {
						existingDatum["deviceId"] = nil
						return existingDatum
					}, nil, true),
					Entry("error when missing time", func() bson.M {
						existingDatum["time"] = nil
						return existingDatum
					}, nil, true),
					Entry("error when missing type", func() bson.M {
						existingDatum["type"] = nil
						return existingDatum
					}, nil, true),
					Entry("adds hash when vaild", func() bson.M {
						return existingDatum
					}, bson.M{"_deduplicator": bson.M{"hash": "fDOlBxqBW5/iv5zV1Rcsawt4wpiSoHdd/yf5WAXW4/c="}}, false),
				)
			})

			Context("bolus", func() {
				var bolusData = &pump.BolusMap{
					"bolus-1": pumpTest.NewRandomBolus(),
					"bolus-2": pumpTest.NewRandomBolus(),
				}
				var bolusDatum bson.M

				BeforeEach(func() {
					bolusDatum = existingDatum
					bolusDatum["type"] = "pumpSettings"
					bolusDatum["bolus"] = bolusData
				})

				DescribeTable("should",
					func(getInput func() bson.M, expected bson.M, expectError bool) {
						input := getInput()
						actual, err := utils.GetDatumUpdates(input)
						if expectError {
							Expect(err).ToNot(BeNil())
							Expect(actual).To(BeNil())
							return
						}
						Expect(err).To(BeNil())
						if expected != nil {
							Expect(actual).To(BeEquivalentTo(expected))
						} else {
							Expect(actual).To(BeNil())
						}
					},

					Entry("do nothing when wrong type",
						func() bson.M {
							bolusDatum["type"] = "other"
							return bolusDatum
						},
						bson.M{"_deduplicator": bson.M{"hash": "eZburze+ZpJwbBfrNLtugyZybLY1WJH22dWAoVBpyvg="}},
						false,
					),
					Entry("do nothing when has no bolus",
						func() bson.M {
							bolusDatum["bolus"] = nil
							return bolusDatum
						},
						bson.M{"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="}},
						false,
					),
					Entry("error bolus when invalid",
						func() bson.M {
							bolusDatum["bolus"] = "wrong"
							return bolusDatum
						},
						nil,
						true,
					),
					Entry("add boluses when bolus found",
						func() bson.M {
							bolusDatum["bolus"] = bolusData
							return bolusDatum
						},
						bson.M{
							"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="},
							"boluses":       bolusData,
						},
						false,
					),
				)
			})
			Context("sleepSchedules", func() {
				var sleepSchedulesExpected *pump.SleepScheduleMap
				var sleepSchedulesStored *pump.SleepScheduleMap
				var sleepSchedulesInvalidDays *pump.SleepScheduleMap
				var sleepSchedulesDatum bson.M

				BeforeEach(func() {

					sleepSchedulesDatum = existingDatum
					sleepSchedulesDatum["type"] = "pumpSettings"
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
					Expect(sleepSchedulesExpected).ToNot(Equal(sleepSchedulesStored))
					Expect(sleepSchedulesInvalidDays).ToNot(BeNil())
				})

				It("does nothing when wrong type", func() {
					sleepSchedulesDatum["type"] = "other"
					actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
					Expect(err).To(BeNil())
					Expect(actual).To(Equal(bson.M{"_deduplicator": bson.M{"hash": "eZburze+ZpJwbBfrNLtugyZybLY1WJH22dWAoVBpyvg="}}))
				})
				It("does nothing when no sleepSchedules", func() {
					sleepSchedulesDatum["sleepSchedules"] = nil
					actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
					Expect(err).To(BeNil())
					Expect(actual).To(Equal(bson.M{"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="}}))
				})
				It("returns error when sleepSchedules is invalid", func() {
					sleepSchedulesDatum["sleepSchedules"] = "wrong"
					actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("pumpSettings.sleepSchedules is not the expected type wrong"))
					Expect(actual).To(BeNil())
				})
				It("returns updated sleepSchedules when valid", func() {
					Expect(sleepSchedulesExpected).ToNot(Equal(sleepSchedulesStored))
					sleepSchedulesDatum["sleepSchedules"] = sleepSchedulesStored
					actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
					Expect(err).To(BeNil())
					Expect(actual).ToNot(BeNil())
					Expect(actual["sleepSchedules"]).ToNot(BeNil())
					Expect(actual["sleepSchedules"]).To(Equal(sleepSchedulesExpected))
				})
				It("returns error when sleepSchedules have invalid days", func() {
					sleepSchedulesDatum["sleepSchedules"] = sleepSchedulesInvalidDays
					actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("pumpSettings.sleepSchedules has an invalid day of week not-a-day"))
					Expect(actual).To(BeNil())
				})
			})
		})
	})
})
