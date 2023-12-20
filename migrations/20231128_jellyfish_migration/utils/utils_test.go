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
				existingBolusDatum["_id"] = expectedID
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
					}, bson.M{"$set": bson.M{"_deduplicator": bson.M{"hash": "FVjexdlY6mWkmoh5gdmtdhzVH4R03+iGE81ro08/KcE="}}}, false),
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

				Context("with mis-named jellyfish bolus", func() {
					var bolusData = &pump.BolusMap{
						"bolus-1": pumpTest.NewRandomBolus(),
						"bolus-2": pumpTest.NewRandomBolus(),
					}
					var settingsBolusDatum bson.M

					BeforeEach(func() {
						settingsBolusDatum = getBSONData(pumpSettingsDatum)
						settingsBolusDatum["bolus"] = bolusData
						settingsBolusDatum["_id"] = expectedID
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
							bson.M{"$set": bson.M{"_deduplicator": bson.M{"hash": "FVjexdlY6mWkmoh5gdmtdhzVH4R03+iGE81ro08/KcE="}}},
							false,
						),
						Entry("do nothing when has no bolus",
							func() bson.M {
								settingsBolusDatum["bolus"] = nil
								return settingsBolusDatum
							},
							bson.M{"$set": bson.M{"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="}}},
							false,
						),
						Entry("add boluses when bolus found",
							func() bson.M {
								settingsBolusDatum["bolus"] = bolusData
								return settingsBolusDatum
							},
							bson.M{
								"$set":    bson.M{"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="}},
								"$rename": bson.M{"bolus": "boluses"},
							},
							false,
						),
					)
				})
				Context("unordered sleepSchedules", func() {
					expectedSleepSchedulesMap := &pump.SleepScheduleMap{}
					var invalidDays *pump.SleepSchedule
					var s1Days *pump.SleepSchedule
					var s2Days *pump.SleepSchedule
					var sleepSchedulesDatum bson.M
					BeforeEach(func() {
						s1 := pumpTest.RandomSleepSchedule()
						s2 := pumpTest.RandomSleepSchedule()
						(*expectedSleepSchedulesMap)["1"] = s1
						(*expectedSleepSchedulesMap)["2"] = s2

						s1Days = pumpTest.CloneSleepSchedule(s1)
						for key, day := range *s1Days.Days {
							(*s1Days.Days)[key] = strings.ToUpper(day)
						}
						s2Days = pumpTest.CloneSleepSchedule(s2)
						for key, day := range *s2Days.Days {
							(*s2Days.Days)[key] = strings.ToUpper(day)
						}
						invalidDays = pumpTest.CloneSleepSchedule(s2)
						invalidDays.Days = &[]string{"not-a-day", common.DayFriday}

						//to ensure correct sorting
						expectedSleepSchedulesMap.Normalize(normalizer.New())

						Expect(expectedSleepSchedulesMap).ToNot(BeNil())
						pumpSettingsDatum.SleepSchedules = nil
						sleepSchedulesDatum = getBSONData(pumpSettingsDatum)
						sleepSchedulesDatum["_id"] = expectedID
						sleepSchedulesDatum["bolus"] = nil //remove as not testing here
					})

					It("does nothing when wrong type", func() {
						_, actual, err := utils.GetDatumUpdates(existingBolusDatum)
						Expect(err).To(BeNil())
						Expect(actual).To(Equal(bson.M{"$set": bson.M{"_deduplicator": bson.M{"hash": "FVjexdlY6mWkmoh5gdmtdhzVH4R03+iGE81ro08/KcE="}}}))
					})
					It("does nothing when no sleepSchedules", func() {
						sleepSchedulesDatum["sleepSchedules"] = nil
						_, actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
						Expect(err).To(BeNil())
						Expect(actual).To(Equal(bson.M{"$set": bson.M{"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="}}}))
					})
					It("returns updated sleepSchedules when valid", func() {
						sleepSchedulesDatum["sleepSchedules"] = []*pump.SleepSchedule{s1Days, s2Days}
						_, actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
						Expect(err).To(BeNil())
						Expect(actual).ToNot(BeNil())
						setData := actual["$set"].(bson.M)
						Expect(setData["sleepSchedules"]).ToNot(BeNil())
						Expect(setData["sleepSchedules"]).To(Equal(expectedSleepSchedulesMap))
					})
					It("returns error when sleepSchedules have invalid days", func() {
						sleepSchedulesDatum["sleepSchedules"] = []*pump.SleepSchedule{invalidDays}
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
