package utils_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils/test"
)

var _ = Describe("back-37", func() {
	var _ = Describe("utils", func() {
		var getBSONData = func(datum interface{}) bson.M {
			var bsonData bson.M
			bsonAsByte, _ := bson.Marshal(&datum)
			bson.Unmarshal(bsonAsByte, &bsonData)
			return bsonData
		}

		// var datumSetup = func(testObj map[string]interface{}) (map[string]interface{}, error) {
		// 	bsonData := getBSONData(testObj)
		// 	objType := fmt.Sprintf("%v", bsonData["type"])
		// 	utils.ApplyBaseChanges(bsonData, objType)
		// 	incomingJSONData, err := json.Marshal(bsonData)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	cleanedObject := map[string]interface{}{}
		// 	if err := json.Unmarshal(incomingJSONData, &cleanedObject); err != nil {
		// 		return nil, err
		// 	}
		// 	return cleanedObject, nil
		// }

		// var _ = Describe("BuildPlatformDatum", func() {

		// 	It("should successfully build basal datum", func() {
		// 		Skip("todo")
		// 		basalData, err := datumSetup(test.AutomatedBasalTandem)
		// 		Expect(err).To(BeNil())
		// 		datum, err := utils.BuildPlatformDatum(fmt.Sprintf("%v", basalData["_id"]), basal.Type, basalData)
		// 		Expect(err).To(BeNil())
		// 		Expect(datum).ToNot(BeNil())
		// 		Expect((*datum).GetType()).To(Equal(basal.Type))
		// 	})
		// 	It("should successfully build dexcom g5 datum", func() {
		// 		Skip("todo")
		// 		cbgData, err := datumSetup(test.CBGDexcomG5MobDatum)
		// 		Expect(err).To(BeNil())
		// 		datum, err := utils.BuildPlatformDatum(fmt.Sprintf("%v", cbgData["_id"]), continuous.Type, cbgData)
		// 		Expect(err).To(BeNil())
		// 		Expect(datum).ToNot(BeNil())
		// 		Expect((*datum).GetType()).To(Equal(continuous.Type))
		// 	})
		// 	It("should successfully build carelink pump settings", func() {
		// 		Skip("todo")
		// 		pSettingsData, err := datumSetup(test.PumpSettingsCarelink)
		// 		Expect(err).To(BeNil())
		// 		datum, err := utils.BuildPlatformDatum(fmt.Sprintf("%v", pSettingsData["_id"]), pump.Type, pSettingsData)
		// 		Expect(err).To(BeNil())
		// 		Expect(datum).ToNot(BeNil())
		// 		Expect((*datum).GetType()).To(Equal(pump.Type))
		// 	})
		// 	It("should successfully build omnipod pump settings", func() {
		// 		Skip("todo")
		// 		pSettingsData, err := datumSetup(test.PumpSettingsOmnipod)
		// 		Expect(err).To(BeNil())
		// 		datum, err := utils.BuildPlatformDatum(fmt.Sprintf("%v", pSettingsData["_id"]), pump.Type, pSettingsData)
		// 		Expect(err).To(BeNil())
		// 		Expect(datum).ToNot(BeNil())
		// 		Expect((*datum).GetType()).To(Equal(pump.Type))
		// 	})
		// 	It("should successfully build tandem pump settings", func() {
		// 		Skip("todo")
		// 		pSettingsData, err := datumSetup(test.PumpSettingsTandem)
		// 		Expect(err).To(BeNil())
		// 		datum, err := utils.BuildPlatformDatum(fmt.Sprintf("%v", pSettingsData["_id"]), pump.Type, pSettingsData)
		// 		Expect(err).To(BeNil())
		// 		Expect(datum).ToNot(BeNil())
		// 		Expect((*datum).GetType()).To(Equal(pump.Type))
		// 	})
		// 	It("should successfully build tandem wizard", func() {
		// 		Skip("todo")
		// 		calcData, err := datumSetup(test.WizardTandem)
		// 		Expect(err).To(BeNil())
		// 		datum, err := utils.BuildPlatformDatum(fmt.Sprintf("%v", calcData["_id"]), calculator.Type, calcData)
		// 		Expect(err).To(BeNil())
		// 		Expect(datum).ToNot(BeNil())
		// 		Expect((*datum).GetType()).To(Equal(calculator.Type))
		// 	})
		// 	It("should successfully build device event", func() {
		// 		Skip("todo")
		// 		deviceEventData, err := datumSetup(test.ReservoirChange)
		// 		Expect(err).To(BeNil())
		// 		datum, err := utils.BuildPlatformDatum(fmt.Sprintf("%v", deviceEventData["_id"]), device.Type, deviceEventData)
		// 		Expect(err).To(BeNil())
		// 		Expect(datum).ToNot(BeNil())
		// 		Expect((*datum).GetType()).To(Equal(device.Type))
		// 	})
		// 	It("should successfully build cgm settings", func() {
		// 		Skip("todo")
		// 		deviceEventData, err := datumSetup(test.CGMSetting)
		// 		Expect(err).To(BeNil())
		// 		datum, err := utils.BuildPlatformDatum(fmt.Sprintf("%v", deviceEventData["_id"]), cgm.Type, deviceEventData)
		// 		Expect(err).To(BeNil())
		// 		Expect(datum).ToNot(BeNil())
		// 		Expect((*datum).GetType()).To(Equal(cgm.Type))
		// 	})
		// })

		// var _ = Describe("ApplyBaseChanges", func() {
		// 	const expectedID = "some-id"
		// 	var pumpSettingsDatum *pump.Pump

		// 	BeforeEach(func() {
		// 		mmolL := pump.DisplayBloodGlucoseUnitsMmolPerL
		// 		pumpSettingsDatum = pumpTest.NewPump(&mmolL)
		// 		*pumpSettingsDatum.ID = expectedID
		// 		*pumpSettingsDatum.UserID = "some-user-id"
		// 		*pumpSettingsDatum.DeviceID = "some-device-id"
		// 		theTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
		// 		*pumpSettingsDatum.Time = theTime
		// 	})
		// 	Context("pumpSettings datum with mis-named jellyfish bolus", func() {
		// 		var bolusData = &pump.BolusMap{
		// 			"bolus-1": pumpTest.NewRandomBolus(),
		// 			"bolus-2": pumpTest.NewRandomBolus(),
		// 		}
		// 		var settingsBolusDatum bson.M
		// 		var datumType string

		// 		BeforeEach(func() {
		// 			settingsBolusDatum = getBSONData(pumpSettingsDatum)
		// 			settingsBolusDatum["bolus"] = bolusData
		// 			settingsBolusDatum["_id"] = expectedID
		// 			datumType = fmt.Sprintf("%v", settingsBolusDatum["type"])
		// 		})

		// 		It("should do nothing when has no bolus", func() {
		// 			Skip("todo")
		// 			settingsBolusDatum["bolus"] = nil
		// 			Expect(settingsBolusDatum["bolus"]).To(BeNil())
		// 			err := utils.ApplyBaseChanges(settingsBolusDatum, datumType)
		// 			Expect(err).To(BeNil())
		// 			Expect(settingsBolusDatum["bolus"]).To(BeNil())
		// 			Expect(settingsBolusDatum["boluses"]).To(BeNil())
		// 		})

		// 		It("should rename as boluses when bolus found", func() {
		// 			Skip("todo")
		// 			Expect(settingsBolusDatum["bolus"]).ToNot(BeNil())
		// 			err := utils.ApplyBaseChanges(settingsBolusDatum, datumType)
		// 			Expect(err).To(BeNil())
		// 			Expect(settingsBolusDatum["bolus"]).To(BeNil())
		// 			Expect(settingsBolusDatum["boluses"]).ToNot(BeNil())
		// 			Expect(settingsBolusDatum["boluses"]).To(Equal(bolusData))
		// 		})
		// 	})
		// 	Context("pumpSettings datum with unordered sleepSchedules", func() {
		// 		expectedSleepSchedulesMap := &pump.SleepScheduleMap{}
		// 		var invalidDays *pump.SleepSchedule
		// 		var s1Days *pump.SleepSchedule
		// 		var s2Days *pump.SleepSchedule
		// 		var sleepSchedulesDatum bson.M
		// 		var datumType string

		// 		BeforeEach(func() {
		// 			sleepSchedulesDatum = getBSONData(pumpSettingsDatum)
		// 			datumType = fmt.Sprintf("%v", sleepSchedulesDatum["type"])
		// 			s1 := pumpTest.RandomSleepSchedule()
		// 			s2 := pumpTest.RandomSleepSchedule()
		// 			(*expectedSleepSchedulesMap)["1"] = s1
		// 			(*expectedSleepSchedulesMap)["2"] = s2

		// 			s1Days = pumpTest.CloneSleepSchedule(s1)
		// 			for key, day := range *s1Days.Days {
		// 				(*s1Days.Days)[key] = strings.ToUpper(day)
		// 			}
		// 			s2Days = pumpTest.CloneSleepSchedule(s2)
		// 			for key, day := range *s2Days.Days {
		// 				(*s2Days.Days)[key] = strings.ToUpper(day)
		// 			}
		// 			invalidDays = pumpTest.CloneSleepSchedule(s2)
		// 			invalidDays.Days = &[]string{"not-a-day", common.DayFriday}
		// 			Expect(expectedSleepSchedulesMap).ToNot(BeNil())
		// 			pumpSettingsDatum.SleepSchedules = nil
		// 			sleepSchedulesDatum = getBSONData(pumpSettingsDatum)
		// 			sleepSchedulesDatum["_id"] = expectedID
		// 			sleepSchedulesDatum["bolus"] = nil //remove as not testing here
		// 		})

		// 		It("does nothing when no sleepSchedules", func() {
		// 			Skip("todo")
		// 			sleepSchedulesDatum["sleepSchedules"] = nil
		// 			err := utils.ApplyBaseChanges(sleepSchedulesDatum, datumType)
		// 			Expect(err).To(BeNil())
		// 			Expect(sleepSchedulesDatum["sleepSchedules"]).To(BeNil())
		// 		})
		// 		It("returns updated sleepSchedules when valid", func() {
		// 			Skip("todo")
		// 			sleepSchedulesDatum["sleepSchedules"] = []*pump.SleepSchedule{s1Days, s2Days}
		// 			err := utils.ApplyBaseChanges(sleepSchedulesDatum, datumType)
		// 			Expect(err).To(BeNil())
		// 			Expect(sleepSchedulesDatum["sleepSchedules"]).ToNot(BeNil())
		// 			Expect(sleepSchedulesDatum["sleepSchedules"]).To(Equal(expectedSleepSchedulesMap))
		// 		})
		// 		It("returns error when sleepSchedules have invalid days", func() {
		// 			Skip("todo")
		// 			sleepSchedulesDatum["sleepSchedules"] = []*pump.SleepSchedule{invalidDays}
		// 			err := utils.ApplyBaseChanges(sleepSchedulesDatum, datumType)
		// 			Expect(err).ToNot(BeNil())
		// 			Expect(err.Error()).To(Equal("pumpSettings.sleepSchedules has an invalid day of week not-a-day"))
		// 		})
		// 	})
		// 	Context("datum with glucose", func() {
		// 		var newContinuous = func(units *string) *continuous.Continuous {
		// 			datum := continuous.New()
		// 			datum.Glucose = *glucoseTest.NewGlucose(units)
		// 			datum.Type = "cbg"
		// 			*datum.ID = expectedID
		// 			*datum.UserID = "some-user-id"
		// 			*datum.DeviceID = "some-device-id"
		// 			theTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
		// 			*datum.Time = theTime
		// 			return datum
		// 		}
		// 		datumType := "cbg"
		// 		It("should do nothing when value is already correct", func() {
		// 			Skip("todo")
		// 			mmoll := glucose.MmolL
		// 			cbg := newContinuous(&mmoll)
		// 			cbgData := getBSONData(cbg)
		// 			cbgData["_id"] = expectedID
		// 			cbgData["value"] = 4.88466

		// 			Expect(cbgData["value"]).To(Equal(4.88466))
		// 			err := utils.ApplyBaseChanges(cbgData, datumType)
		// 			Expect(err).To(BeNil())
		// 			Expect(cbgData["value"]).To(Equal(4.88466))
		// 		})
		// 		It("should update the value when the precesion is too accurate correct", func() {
		// 			Skip("todo")
		// 			mmoll := glucose.MmolL
		// 			cbg := newContinuous(&mmoll)
		// 			cbgData := getBSONData(cbg)
		// 			cbgData["_id"] = expectedID
		// 			cbgData["value"] = 4.88465823212007

		// 			Expect(cbgData["value"]).To(Equal(4.88465823212007))
		// 			err := utils.ApplyBaseChanges(cbgData, datumType)
		// 			Expect(err).To(BeNil())
		// 			floatVal := 4.88466
		// 			Expect(cbgData["value"]).To(Equal(&floatVal))
		// 		})
		// 	})
		// 	Context("reservoirChange deviceEvent datum with status string", func() {
		// 		var newReservoirChange = func() *reservoirchange.ReservoirChange {
		// 			datum := reservoirchange.New()
		// 			datum.Device = *dataTypesDeviceTest.RandomDevice()
		// 			datum.SubType = "reservoirChange"
		// 			return datum
		// 		}
		// 		It("should convert to statusId", func() {
		// 			Skip("todo")
		// 			deviceEvent := newReservoirChange()
		// 			deviceEventData := getBSONData(deviceEvent)
		// 			deviceEventData["status"] = "some-status-id"
		// 			err := utils.ApplyBaseChanges(deviceEventData, deviceEvent.Type)
		// 			Expect(err).To(BeNil())
		// 			Expect(deviceEventData["status"]).To(BeNil())
		// 			Expect(deviceEventData["statusId"]).To(Equal("some-status-id"))
		// 		})
		// 	})
		// 	Context("wizard datum with bolus string", func() {
		// 		It("should convert to bolusId for datum validation", func() {
		// 			Skip("todo")
		// 			wizardBSON := getBSONData(test.WizardTandem)
		// 			Expect(wizardBSON["bolus"]).ToNot(BeNil())
		// 			err := utils.ApplyBaseChanges(wizardBSON, calculator.Type)
		// 			Expect(err).To(BeNil())
		// 			Expect(wizardBSON["bolus"]).To(BeNil())
		// 		})
		// 	})
		// 	Context("datum with string payload", func() {
		// 		var datumWithPayload primitive.M
		// 		var datumType string
		// 		var payload *metadata.Metadata
		// 		BeforeEach(func() {
		// 			datumWithPayload = getBSONData(pumpSettingsDatum)
		// 			payload = metadataTest.RandomMetadata()
		// 			datumWithPayload["payload"] = *payload
		// 			datumType = fmt.Sprintf("%v", datumWithPayload["type"])
		// 		})

		// 		It("should do nothing when value is already correct", func() {
		// 			Skip("todo")
		// 			Expect(datumWithPayload["payload"]).To(Equal(*payload))
		// 			err := utils.ApplyBaseChanges(datumWithPayload, datumType)
		// 			Expect(err).To(BeNil())
		// 			Expect(datumWithPayload["payload"]).To(Equal(*payload))
		// 		})
		// 		It("should update the payload when it is a string", func() {
		// 			Skip("todo")
		// 			datumWithPayload["payload"] = `{"transmitterId":"410X6M","transmitterTicks":5796922,"trend":"flat"}`
		// 			err := utils.ApplyBaseChanges(datumWithPayload, datumType)
		// 			Expect(err).To(BeNil())
		// 			Expect(datumWithPayload["payload"]).To(Equal(&metadata.Metadata{
		// 				"transmitterId":    "410X6M",
		// 				"transmitterTicks": float64(5796922),
		// 				"trend":            "flat",
		// 			}))
		// 		})
		// 		It("should remove the payload when it is empty", func() {
		// 			Skip("todo")
		// 			datumWithPayload["payload"] = bson.M{}
		// 			err := utils.ApplyBaseChanges(datumWithPayload, datumType)
		// 			Expect(err).To(BeNil())
		// 			Expect(datumWithPayload["payload"]).To(BeNil())
		// 		})
		// 	})
		// 	Context("datum with string annotations", func() {
		// 		var datumWithAnnotation primitive.M
		// 		var annotations *metadata.MetadataArray
		// 		var datumType string
		// 		BeforeEach(func() {
		// 			datumWithAnnotation = getBSONData(pumpSettingsDatum)
		// 			annotations = metadataTest.RandomMetadataArray()
		// 			datumWithAnnotation["annotations"] = *annotations
		// 			datumType = fmt.Sprintf("%v", datumWithAnnotation["type"])
		// 		})

		// 		It("should do nothing when value is already correct", func() {
		// 			Skip("todo")
		// 			Expect(datumWithAnnotation["annotations"]).To(Equal(*annotations))
		// 			err := utils.ApplyBaseChanges(datumWithAnnotation, datumType)
		// 			Expect(err).To(BeNil())
		// 			Expect(datumWithAnnotation["annotations"]).To(Equal(*annotations))
		// 		})
		// 		It("should update the annotations when it is a string", func() {
		// 			Skip("todo")
		// 			datumWithAnnotation["annotations"] = `[{"code":"bg/out-of-range","threshold":40,"value":"low"}]`
		// 			err := utils.ApplyBaseChanges(datumWithAnnotation, datumType)
		// 			Expect(err).To(BeNil())
		// 			Expect(datumWithAnnotation["annotations"]).To(Equal(&metadata.MetadataArray{
		// 				&metadata.Metadata{
		// 					"code":      "bg/out-of-range",
		// 					"threshold": float64(40),
		// 					"value":     "low",
		// 				},
		// 			}))
		// 		})
		// 	})
		// })

		// var _ = Describe("GetDatumChanges", func() {

		// 	const expectedID = "difference-id"

		// 	var getRawData = func(datum interface{}) map[string]interface{} {
		// 		var rawObject map[string]interface{}
		// 		asByte, _ := json.Marshal(&datum)
		// 		json.Unmarshal(asByte, &rawObject)
		// 		return rawObject
		// 	}

		// 	It("has no difference", func() {
		// 		Skip("Todo")
		// 		datumObject := getBSONData(test.AutomatedBasalTandem)
		// 		incomingObject := getRawData(test.AutomatedBasalTandem)
		// 		apply, revert, err := utils.GetDatumChanges(expectedID, datumObject, incomingObject)
		// 		Expect(err).To(BeNil())
		// 		Expect(apply).ToNot(BeNil())
		// 		Expect(apply).To(Equal([]bson.M{}))
		// 		Expect(revert).ToNot(BeNil())
		// 		Expect(revert).To(Equal([]bson.M{}))
		// 	})
		// 	It("set for missing properties", func() {
		// 		Skip("Todo")
		// 		datumObject := getBSONData(test.AutomatedBasalTandem)
		// 		incomingObject := getRawData(test.AutomatedBasalTandem)
		// 		delete(incomingObject, "deliveryType")
		// 		apply, revert, err := utils.GetDatumChanges(expectedID, datumObject, incomingObject)
		// 		Expect(err).To(BeNil())
		// 		Expect(apply).To(Equal([]bson.M{{"$set": bson.M{"deliveryType": "automated"}}}))
		// 		Expect(revert).To(Equal([]bson.M{{"$unset": bson.M{"deliveryType": ""}}}))
		// 	})
		// 	It("set _deduplicator correctly", func() {
		// 		Skip("Todo")
		// 		calcData, _ := datumSetup(test.WizardTandem)
		// 		id := fmt.Sprintf("%v", calcData["_id"])
		// 		datum, _ := utils.BuildPlatformDatum(id, calculator.Type, calcData)
		// 		apply, revert, err := utils.GetDatumChanges(id, datum, calcData)
		// 		Expect(err).To(BeNil())
		// 		Expect(len(apply)).To(Equal(2))
		// 		Expect(apply[0]["$set"]).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "o6ybZQtDZ95FvuV0zYGphri2SIGesbLCbkHxc1wbbEE="}))
		// 		Expect(len(revert)).To(Equal(2))
		// 		Expect(revert[0]).Should(HaveKeyWithValue("$unset", bson.M{"_deduplicator": ""}))
		// 	})
		// 	It("unset for unwanted properties", func() {
		// 		Skip("Todo")
		// 		datumObject := getBSONData(test.AutomatedBasalTandem)
		// 		incomingObject := getRawData(test.AutomatedBasalTandem)
		// 		incomingObject["random"] = map[string]interface{}{"extra": true}
		// 		apply, revert, err := utils.GetDatumChanges(expectedID, datumObject, incomingObject)
		// 		Expect(err).To(BeNil())
		// 		Expect(len(apply)).To(Equal(1))
		// 		Expect(apply[0]).Should(HaveKeyWithValue("$unset", bson.M{"random": ""}))
		// 		Expect(len(revert)).To(Equal(1))
		// 		Expect(revert[0]).Should(HaveKeyWithValue("$set", bson.M{"random": map[string]interface{}{"extra": true}}))
		// 	})
		// 	It("allow for removing deeply nested properties", func() {
		// 		Skip("Todo")
		// 		datumObject := getBSONData(test.AutomatedBasalTandem)
		// 		incomingObject := getRawData(test.AutomatedBasalTandem)

		// 		ofThings := map[string]interface{}{
		// 			"of": map[string]interface{}{
		// 				"things": map[string]interface{}{
		// 					"go": map[string]interface{}{
		// 						"here": true}}}}
		// 		incomingObject["lots"] = ofThings
		// 		apply, revert, err := utils.GetDatumChanges(expectedID, datumObject, incomingObject)
		// 		Expect(err).To(BeNil())
		// 		Expect(len(apply)).To(Equal(1))
		// 		Expect(apply[0]).Should(HaveKeyWithValue("$unset", bson.M{"lots": ""}))
		// 		Expect(len(revert)).To(Equal(1))
		// 		Expect(revert[0]).Should(HaveKeyWithValue("$set", bson.M{"lots": ofThings}))
		// 	})
		// 	It("allow for updating nested array properties", func() {
		// 		Skip("Todo")
		// 		datumObject := getBSONData(test.PumpSettingsTandem)
		// 		incomingObject := getRawData(test.PumpSettingsTandem)

		// 		datumObject["insulinSensitivities"] = map[string]interface{}{
		// 			"Simple": []map[string]interface{}{
		// 				{"amount": 1.2, "start": 0},
		// 				{"amount": 2.6, "start": 46800000},
		// 			},
		// 			"Standard": []map[string]interface{}{
		// 				{"amount": 2.7753739955227665, "start": 1000},
		// 				{"amount": 2.7753739955227665, "start": 46800000},
		// 			},
		// 		}

		// 		apply, revert, err := utils.GetDatumChanges(expectedID, datumObject, incomingObject)
		// 		Expect(err).To(BeNil())
		// 		Expect(len(apply)).To(Equal(1))
		// 		Expect(apply[0]).Should(HaveKeyWithValue("$set", bson.M{
		// 			"insulinSensitivities.Simple.0.amount":  1.2,
		// 			"insulinSensitivities.Simple.1.amount":  2.6,
		// 			"insulinSensitivities.Standard.0.start": float64(1000),
		// 		}))
		// 		Expect(len(revert)).To(Equal(1))
		// 		Expect(revert[0]).Should(HaveKeyWithValue("$set", bson.M{
		// 			"insulinSensitivities.Simple.0.amount":  2.7753739955227665,
		// 			"insulinSensitivities.Simple.1.amount":  2.7753739955227665,
		// 			"insulinSensitivities.Standard.0.start": float64(0),
		// 		}))
		// 	})
		// 	It("no difference when inner payload changes", func() {
		// 		Skip("Todo")
		// 		datumObject := getBSONData(test.AutomatedBasalTandem)
		// 		incomingObject := getRawData(test.AutomatedBasalTandem)
		// 		datumObject["payload"] = map[string]interface{}{"stuff": true}
		// 		apply, revert, err := utils.GetDatumChanges(expectedID, datumObject, incomingObject)
		// 		Expect(err).To(BeNil())
		// 		Expect(apply).To(Equal([]bson.M{}))
		// 		Expect(revert).To(Equal([]bson.M{}))
		// 	})

		// 	It("should convert to bolusId for datum validation", func() {
		// 		Skip("Todo")
		// 		bsonData := getBSONData(test.WizardTandem)
		// 		datumID := fmt.Sprintf("%v", bsonData["_id"])
		// 		datumType := fmt.Sprintf("%v", bsonData["type"])
		// 		apply, _, err := utils.ProcessDatum(datumID, datumType, bsonData)
		// 		Expect(err).To(BeNil())
		// 		Expect(apply[1]["$unset"]).ShouldNot(HaveKey("bolusId"))
		// 	})

		// 	It("should update all bgTraget values", func() {

		// 		bsonObject := getBSONData(test.PumpSettingsCarelink)
		// 		datumID := fmt.Sprintf("%v", bsonObject["_id"])
		// 		datumType := fmt.Sprintf("%v", bsonObject["type"])
		// 		apply, revert, err := utils.ProcessDatum(datumID, datumType, bsonObject)

		// 		Expect(err).To(BeNil())
		// 		Expect(apply[0]["$set"]).ToNot(BeNil())
		// 		Expect(len(apply)).To(Equal(2))
		// 		Expect(apply[0]["$set"]).Should(HaveKeyWithValue("units.bg", "mmol/L"))
		// 		Expect(apply[0]["$set"]).Should(HaveKeyWithValue("bgTarget.0.target", float64(5.55074)))
		// 		Expect(apply[0]["$set"]).Should(HaveKeyWithValue("bgTarget.1.target", float64(5.55074)))

		// 		Expect(revert[0]["$set"]).ToNot(BeNil())
		// 		Expect(len(revert)).To(Equal(2))
		// 		Expect(revert[0]["$set"]).Should(HaveKeyWithValue("units.bg", "mmol/L"))
		// 		Expect(revert[0]["$set"]).Should(HaveKeyWithValue("bgTarget.0.target", float64(5.550747991045533)))
		// 		Expect(revert[0]["$set"]).Should(HaveKeyWithValue("bgTarget.1.target", float64(5.550747991045533)))
		// 	})

		// 	It("pump settings omnipod", func() {
		// 		Skip("Todo")
		// 		incomingObject := getRawData(test.PumpSettingsOmnipod)
		// 		datumID := fmt.Sprintf("%v", incomingObject["_id"])
		// 		datum, _ := utils.BuildPlatformDatum(datumID, pump.Type, incomingObject)
		// 		apply, revert, err := utils.GetDatumChanges(datumID, datum, incomingObject)
		// 		Expect(err).To(BeNil())
		// 		Expect(len(apply)).To(Equal(2))
		// 		Expect(apply[0]["$set"]).Should(HaveKeyWithValue("units.bg", "mmol/L"))
		// 		Expect(apply[0]["$set"]).Should(HaveKeyWithValue("bgTarget.0.high", float64(0.40054)))
		// 		Expect(apply[0]["$set"]).Should(HaveKeyWithValue("bgTarget.1.high", float64(0.40054)))
		// 		Expect(apply[0]["$set"]).Should(HaveKeyWithValue("bgTarget.0.target", float64(0.30811)))
		// 		Expect(apply[0]["$set"]).Should(HaveKeyWithValue("bgTarget.1.target", float64(0.30811)))
		// 		Expect(len(revert)).To(Equal(2))
		// 		Expect(revert[0]["$set"]).Should(HaveKeyWithValue("units.bg", "mmol/L"))
		// 		Expect(revert[0]["$set"]).Should(HaveKeyWithValue("bgTarget.0.high", float64(0.40054)))
		// 		Expect(revert[0]["$set"]).Should(HaveKeyWithValue("bgTarget.1.high", float64(0.40054)))
		// 		Expect(revert[0]["$set"]).Should(HaveKeyWithValue("bgTarget.0.target", float64(0.30811)))
		// 		Expect(revert[0]["$set"]).Should(HaveKeyWithValue("bgTarget.1.target", float64(0.30811)))
		// 	})

		// })

		var _ = Describe("ProcessDatum", func() {

			It("basal with unwanted percent feild", func() {

				bsonObject := getBSONData(test.AutomatedBasalTandem)
				datumID := fmt.Sprintf("%v", bsonObject["_id"])
				datumType := fmt.Sprintf("%v", bsonObject["type"])

				apply, revert, err := utils.ProcessDatum(datumID, datumType, bsonObject)
				Expect(err).To(BeNil())
				Expect(apply).ToNot(BeNil())
				Expect(revert).ToNot(BeNil())

				Expect(apply[0]["$set"]).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "CFDp66+LJvYW7rxf+4ndFd8hoTMq+ymzwLnuEUEqhVs="}))
				Expect(apply[1]["$unset"]).Should(HaveKeyWithValue("percent", ""))
				Expect(revert[0]["$unset"]).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revert[1]["$set"]).Should(HaveKeyWithValue("percent", float64(0.47857142857142865)))
			})

			It("pump settings with blood glucose precsion updates", func() {

				bsonObject := getBSONData(test.PumpSettingsTandem)
				datumID := fmt.Sprintf("%v", bsonObject["_id"])
				datumType := fmt.Sprintf("%v", bsonObject["type"])

				apply, revert, err := utils.ProcessDatum(datumID, datumType, bsonObject)
				Expect(err).To(BeNil())
				Expect(apply).ToNot(BeNil())
				Expect(revert).ToNot(BeNil())

				Expect(apply[0]["$set"]).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "bpKLJbi5JfqD7N0WJ1vj0ck03c9EZ3U0H09TCLhdd3k="}))
				//Expect(apply[0]["$set"]).Should(HaveKeyWithValue("bgTargets.Simple.1.target", 0))
				Expect(apply[1]["$unset"]).Should(HaveKeyWithValue("localTime", ""))
				Expect(revert[0]["$unset"]).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revert[1]["$set"]).Should(HaveKeyWithValue("localTime", "2017-11-05T12:56:51.000Z"))
			})

			It("wizard with bgInput and bgTarget glucose updates", func() {

				bsonObject := getBSONData(test.WizardTandem)
				datumID := fmt.Sprintf("%v", bsonObject["_id"])
				datumType := fmt.Sprintf("%v", bsonObject["type"])

				apply, revert, err := utils.ProcessDatum(datumID, datumType, bsonObject)
				Expect(err).To(BeNil())
				Expect(apply).ToNot(BeNil())
				Expect(revert).ToNot(BeNil())

				applySet := apply[0]["$set"]
				applyUnset := apply[1]["$unset"]

				revertSet := revert[1]["$set"]
				revertUnset := revert[0]["$unset"]

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "o6ybZQtDZ95FvuV0zYGphri2SIGesbLCbkHxc1wbbEE="}))
				Expect(applySet).Should(HaveKeyWithValue("bgInput", 4.4406))
				Expect(applySet).Should(HaveKeyWithValue("bgTarget.target", 4.4406))
				Expect(applyUnset).Should(HaveKeyWithValue("localTime", ""))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKeyWithValue("localTime", "2017-11-05T12:56:51.000Z"))
				Expect(revertSet).Should(HaveKeyWithValue("bgInput", 4.440598392836427))
				Expect(revertSet).Should(HaveKeyWithValue("bgTarget.target", 4.440598392836427))
			})

		})
	})
})
