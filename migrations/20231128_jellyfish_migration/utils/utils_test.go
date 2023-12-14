package utils_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data/normalizer"
	bolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
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
			var existingBolusDatum bson.M
			const expectedID = "some-id"

			var getBSONData = func(datum interface{}) bson.M {
				var bsonData bson.M
				bsonAsByte, _ := bson.Marshal(&datum)
				bson.Unmarshal(bsonAsByte, &bsonData)
				return bsonData
			}

			BeforeEach(func() {
				datum := bolusTest.RandomBolus()
				*datum.ID = expectedID
				*datum.UserID = "some-user-id"
				*datum.DeviceID = "some-device-id"
				datum.SubType = "some-subtype"
				theTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
				*datum.Time = theTime
				existingBolusDatum = getBSONData(datum)
				Expect(existingBolusDatum).ToNot(BeNil())
			})

			Context("_deduplicator hash", func() {
				DescribeTable("should",
					func(getInput func() bson.M, expectedUpdates bson.M, expectError bool) {
						input := getInput()
						actualID, actualUpdates, err := utils.GetDatumUpdates(input)
						if expectError {
							Expect(err).ToNot(BeNil())
							Expect(actualUpdates).To(BeNil())
							return
						}
						Expect(err).To(BeNil())
						if expectedUpdates != nil {
							Expect(actualUpdates).To(Equal(expectedUpdates))
							Expect(actualID).To(Equal(expectedID))
						} else {
							Expect(actualUpdates).To(BeNil())
						}
					},
					Entry("error when missing _userId", func() bson.M {
						existingBolusDatum["_userId"] = nil
						return existingBolusDatum
					}, nil, true),
					Entry("error when missing deviceId", func() bson.M {
						existingBolusDatum["deviceId"] = nil
						return existingBolusDatum
					}, nil, true),
					Entry("error when missing time", func() bson.M {
						existingBolusDatum["time"] = nil
						return existingBolusDatum
					}, nil, true),
					Entry("error when missing type", func() bson.M {
						existingBolusDatum["type"] = nil
						return existingBolusDatum
					}, nil, true),
					Entry("adds hash when vaild", func() bson.M {
						return existingBolusDatum
					}, bson.M{"_deduplicator": bson.M{"hash": "FVjexdlY6mWkmoh5gdmtdhzVH4R03+iGE81ro08/KcE="}}, false),
				)
			})

			Context("pumpSettings", func() {

				var pumpSettingsDatum *pump.Pump

				BeforeEach(func() {
					mmolL := pump.DisplayBloodGlucoseUnitsMmolPerL
					pumpSettingsDatum = pumpTest.NewPump(&mmolL)
					*pumpSettingsDatum.ID = expectedID
					*pumpSettingsDatum.UserID = "some-user-id"
					*pumpSettingsDatum.DeviceID = "some-device-id"
					theTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
					*pumpSettingsDatum.Time = theTime
				})

				Context("with incorrect jellyfish bolus", func() {
					var bolusData = &pump.BolusMap{
						"bolus-1": pumpTest.NewRandomBolus(),
						"bolus-2": pumpTest.NewRandomBolus(),
					}
					var settingsBolusDatum bson.M

					BeforeEach(func() {
						settingsBolusDatum = getBSONData(pumpSettingsDatum)
						//as currently set in jellyfish
						settingsBolusDatum["bolus"] = bolusData
					})

					DescribeTable("should",
						func(getInput func() bson.M, expected bson.M, expectError bool) {
							input := getInput()
							_, actual, err := utils.GetDatumUpdates(input)
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
								return existingBolusDatum
							},
							bson.M{"_deduplicator": bson.M{"hash": "FVjexdlY6mWkmoh5gdmtdhzVH4R03+iGE81ro08/KcE="}},
							false,
						),
						Entry("do nothing when has no bolus",
							func() bson.M {
								settingsBolusDatum["bolus"] = nil
								return settingsBolusDatum
							},
							bson.M{"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="}},
							false,
						),
						Entry("error bolus when invalid",
							func() bson.M {
								settingsBolusDatum["bolus"] = "wrong"
								return settingsBolusDatum
							},
							nil,
							true,
						),
						Entry("add boluses when bolus found",
							func() bson.M {
								settingsBolusDatum["bolus"] = bolusData
								return settingsBolusDatum
							},
							bson.M{
								"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="},
								"boluses":       bolusData,
							},
							false,
						),
					)
				})
				Context("unordered sleepSchedules", func() {
					var sleepSchedulesExpected *pump.SleepScheduleMap
					var sleepSchedulesStored *pump.SleepScheduleMap
					var sleepSchedulesInvalidDays *pump.SleepScheduleMap
					var sleepSchedulesDatum bson.M

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
						Expect(sleepSchedulesExpected).ToNot(Equal(sleepSchedulesStored))
						Expect(sleepSchedulesInvalidDays).ToNot(BeNil())
						pumpSettingsDatum.SleepSchedules = sleepSchedulesStored
						sleepSchedulesDatum = getBSONData(pumpSettingsDatum)
						sleepSchedulesDatum["bolus"] = nil //remove as not testng here
					})

					It("does nothing when wrong type", func() {
						_, actual, err := utils.GetDatumUpdates(existingBolusDatum)
						Expect(err).To(BeNil())
						Expect(actual).To(Equal(bson.M{"_deduplicator": bson.M{"hash": "FVjexdlY6mWkmoh5gdmtdhzVH4R03+iGE81ro08/KcE="}}))
					})
					It("does nothing when no sleepSchedules", func() {
						sleepSchedulesDatum["sleepSchedules"] = nil
						_, actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
						Expect(err).To(BeNil())
						Expect(actual).To(Equal(bson.M{"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="}}))
					})
					It("returns updated sleepSchedules when valid", func() {
						Expect(sleepSchedulesExpected).ToNot(Equal(sleepSchedulesStored))
						sleepSchedulesDatum["sleepSchedules"] = sleepSchedulesStored
						_, actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
						Expect(err).To(BeNil())
						Expect(actual).ToNot(BeNil())
						Expect(actual["sleepSchedules"]).ToNot(BeNil())
						Expect(actual["sleepSchedules"]).To(Equal(sleepSchedulesExpected))
					})
					It("returns error when sleepSchedules have invalid days", func() {
						sleepSchedulesDatum["sleepSchedules"] = sleepSchedulesInvalidDays
						_, actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
						Expect(err).ToNot(BeNil())
						Expect(err.Error()).To(Equal("pumpSettings.sleepSchedules has an invalid day of week not-a-day"))
						Expect(actual).To(BeNil())
					})
				})
			})
		})
	})
})
