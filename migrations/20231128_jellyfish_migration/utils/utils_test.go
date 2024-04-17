package utils_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

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

		var setup = func(bsonObj bson.M) (applySet interface{}, applyUnset interface{}, revertUnset interface{}, revertSet interface{}) {
			datumType := fmt.Sprintf("%v", bsonObj["type"])
			datumID := fmt.Sprintf("%v", bsonObj["_id"])
			apply, revert, err := utils.ProcessDatum(datumID, datumType, bsonObj)
			Expect(err).To(BeNil())
			Expect(apply).ToNot(BeNil())
			Expect(revert).ToNot(BeNil())

			applySet = apply[0]["$set"]
			applyUnset = apply[1]["$unset"]
			revertUnset = revert[0]["$unset"]
			revertSet = revert[1]["$set"]

			return applySet, applyUnset, revertUnset, revertSet
		}

		var _ = Describe("ProcessDatum", func() {

			It("basal with unwanted percent field", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.AutomatedBasalTandem))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "YOItOWBgIIoEkqVsBq9yrOZ5utmsKTIezszpGBj5Vpc="}))
				Expect(applyUnset).Should(HaveKeyWithValue("percent", ""))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKeyWithValue("percent", float64(0.47857142857142865)))
			})

			It("bolus out of range expection", func() {
				bsonObj := getBSONData(test.AutomatedBolus)
				datumType := fmt.Sprintf("%v", bsonObj["type"])
				datumID := fmt.Sprintf("%v", bsonObj["_id"])
				_, _, err := utils.ProcessDatum(datumID, datumType, bsonObj)
				Expect(err).To(BeNil())
			})

			It("cgm settings with blood glucose precsion updates", func() {
				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.CGMSetting))

				Expect(applySet).Should(HaveLen(4))
				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "gyyB8OqbErdW2aOOo8POTXk1SNJmu5gDEIaCugTVn3M="}))
				Expect(applySet).Should(HaveKeyWithValue("lowAlerts.level", 3.88552))
				Expect(applySet).Should(HaveKeyWithValue("highAlerts.level", 22.20299))
				// NOTE `rateOfChangeAlert` does not truncate the fallRate.rate and riseRate.rate
				// see platform/data/types/settings/cgm/rate_alert_DEPRECATED.go RateDEPRECATEDMmolLThree and RateDEPRECATEDMmolLTwo
				Expect(applySet).Should(HaveKeyWithValue("rateOfChangeAlert", map[string]interface{}{
					"fallRate": map[string]interface{}{
						"rate":    -0.16652243973136602,
						"enabled": false,
					},
					"riseRate": map[string]interface{}{
						"rate":    0.16652243973136602,
						"enabled": false,
					},
				}))

				Expect(applyUnset).Should(HaveLen(2))
				Expect(applyUnset).Should(HaveKey("rateOfChangeAlerts"))
				Expect(applyUnset).Should(HaveKey("localTime"))

				Expect(revertSet).Should(HaveLen(4))
				Expect(revertSet).Should(HaveKeyWithValue("lowAlerts.level", 3.8855235937318735))
				Expect(revertSet).Should(HaveKeyWithValue("highAlerts.level", 22.202991964182132))
				Expect(revertSet).Should(HaveKeyWithValue("rateOfChangeAlerts", map[string]interface{}{
					"fallRate": map[string]interface{}{
						"rate":    -0.16652243973136602,
						"enabled": false,
					},
					"riseRate": map[string]interface{}{
						"rate":    0.16652243973136602,
						"enabled": false,
					},
				}))

				Expect(revertUnset).Should(HaveLen(2))
				Expect(revertUnset).Should(HaveKey("_deduplicator"))
				Expect(revertUnset).Should(HaveKey("rateOfChangeAlert"))
			})

			It("pump settings with blood glucose precsion updates", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.PumpSettingsTandem))

				Expect(applySet).Should(HaveLen(6))
				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "l5e6HoVqMu3ZOUjqaky/m6ZNw+D0UFxbYw/fM9P4PXc="}))
				Expect(applySet).Should(HaveKeyWithValue("bgTargets.Simple.0.target", 5.55075))
				Expect(applySet).Should(HaveKeyWithValue("bgTargets.Simple.1.target", 5.55075))
				Expect(applySet).Should(HaveKeyWithValue("bgTargets.Standard.0.target", 5.55075))
				Expect(applySet).Should(HaveKeyWithValue("bgTargets.Standard.1.target", 5.55075))
				Expect(applySet).Should(HaveKeyWithValue("units.bg", "mmol/L"))

				Expect(applyUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(HaveKeyWithValue("localTime", ""))

				Expect(revertUnset).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))

				Expect(revertSet).Should(HaveLen(6))
				Expect(revertSet).Should(HaveKeyWithValue("localTime", "2017-11-05T12:56:51.000Z"))
				Expect(revertSet).Should(HaveKeyWithValue("bgTargets.Simple.0.target", 5.550747991045533))
				Expect(revertSet).Should(HaveKeyWithValue("bgTargets.Simple.1.target", 5.550747991045533))
				Expect(revertSet).Should(HaveKeyWithValue("bgTargets.Standard.0.target", 5.550747991045533))
				Expect(revertSet).Should(HaveKeyWithValue("bgTargets.Standard.1.target", 5.550747991045533))
				Expect(revertSet).Should(HaveKeyWithValue("units.bg", "mg/dL"))
			})

			It("pump settings with sleep schedule updates", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.PumpSettingsWithSleepScheduleTandem))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "l5e6HoVqMu3ZOUjqaky/m6ZNw+D0UFxbYw/fM9P4PXc="}))
				Expect(applySet).Should(HaveKey("sleepSchedules"))

				applyObj := applySet.(primitive.M)

				actualSchedules := applyObj["sleepSchedules"]

				expectedSchedules := map[string]interface{}{
					"1": map[string]interface{}{
						"enabled": true,
						"days":    []interface{}{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"},
						"start":   82800,
						"end":     25200,
					},
					"2": map[string]interface{}{
						"enabled": false,
						"days":    []interface{}{"sunday"},
						"start":   3600,
						"end":     32400,
					},
				}

				Expect(fmt.Sprintf("%v", actualSchedules)).To(Equal(fmt.Sprintf("%v", expectedSchedules)))

				Expect(applyUnset).Should(HaveKeyWithValue("localTime", ""))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKey("sleepSchedules"))
				revertSetObj := revertSet.(primitive.M)
				actualRevrtSchedules := revertSetObj["sleepSchedules"]

				originalSchedules := []map[string]interface{}{
					{
						"enabled": true,
						"days":    []interface{}{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
						"start":   82800,
						"end":     25200,
					},
					{
						"enabled": false,
						"days":    []interface{}{"Sunday"},
						"start":   3600,
						"end":     32400,
					},
				}

				Expect(fmt.Sprintf("%v", actualRevrtSchedules)).To(Equal(fmt.Sprintf("%v", originalSchedules)))

			})

			It("wizard with bgInput and bgTarget glucose updates", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.WizardTandem))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "orP5cbifS8h0f3HWZcTOIf4B431HO1OReg9o1nmFnU4="}))
				Expect(applySet).Should(HaveKeyWithValue("bgInput", 4.4406))
				Expect(applySet).Should(HaveKeyWithValue("bgTarget.target", 4.4406))
				Expect(applyUnset).Should(HaveKeyWithValue("localTime", ""))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKeyWithValue("localTime", "2017-11-05T12:56:51.000Z"))
				Expect(revertSet).Should(HaveKeyWithValue("bgInput", 4.440598392836427))
				Expect(revertSet).Should(HaveKeyWithValue("bgTarget.target", 4.440598392836427))
			})

			It("pump settings with bgTraget glucose updates", func() {

				applySet, _, revertUnset, revertSet := setup(getBSONData(test.PumpSettingsCarelink))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "LgRaGs4QkIBV9sHUjurpMt/ALU+7F7ZlU8xNxhkTQwQ="}))
				Expect(applySet).Should(HaveKeyWithValue("bgTarget.0.target", 5.55075))
				Expect(applySet).Should(HaveKeyWithValue("bgTarget.1.target", 5.55075))
				Expect(applySet).Should(HaveKeyWithValue("units.bg", "mmol/L"))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKeyWithValue("bgTarget.0.target", 5.550747991045533))
				Expect(revertSet).Should(HaveKeyWithValue("bgTarget.1.target", 5.550747991045533))
				Expect(revertSet).Should(HaveKeyWithValue("units.bg", "mg/dL"))
			})

			It("will remove empty payload", func() {
				_, applyUnset, _, revertSet := setup(getBSONData(test.EmptyPayloadDatum))
				Expect(applyUnset).Should(HaveKeyWithValue("payload", ""))
				Expect(revertSet).Should(HaveKeyWithValue("payload", map[string]interface{}{}))
			})

			It("will move misnamed bolus to boluses for pump setting", func() {
				bsonObj := getBSONData(test.PumpSettingsWithBolusDatum)
				applySet, applyUnset, revertUnset, revertSet := setup(bsonObj)
				Expect(applyUnset).Should(HaveKeyWithValue("bolus", ""))
				Expect(applySet).Should(HaveKey("boluses"))
				Expect(revertSet).Should(HaveKey("bolus"))
				Expect(revertUnset).Should(HaveKeyWithValue("boluses", ""))
			})

			It("wizard datum will not have bolus link removed", func() {
				bsonObj := getBSONData(test.WizardTandem)
				Expect(bsonObj).Should(HaveKeyWithValue("bolus", "g2h6nohp5sdndpvl2l8kdete00lle4gt"))
				applySet, applyUnset, revertUnset, revertSet := setup(bsonObj)
				Expect(applyUnset).ShouldNot(HaveKeyWithValue("bolus", ""))
				Expect(applySet).ShouldNot(HaveKey("bolusId"))
				Expect(revertSet).ShouldNot(HaveKey("bolus"))
				Expect(revertUnset).ShouldNot(HaveKey("bolusId"))
			})

			It("device event datum will not have status link removed", func() {
				bsonObj := getBSONData(test.ReservoirChangeWithStatus)
				Expect(bsonObj).Should(HaveKeyWithValue("status", "cvv61jde62b6i28bgot57f18bor5au1n"))
				applySet, applyUnset, revertUnset, revertSet := setup(bsonObj)
				Expect(applyUnset).ShouldNot(HaveKeyWithValue("status", ""))
				Expect(applySet).ShouldNot(HaveKey("statusId"))
				Expect(revertSet).ShouldNot(HaveKey("status"))
				Expect(revertUnset).ShouldNot(HaveKey("statusId"))
			})

			It("status device event datum with suspended status as suspended will not update it", func() {
				bsonObj := getBSONData(test.AlarmDeviceEventDatum)
				Expect(bsonObj).Should(HaveKeyWithValue("status", "suspended"))
				applySet, _, _, revertSet := setup(bsonObj)
				Expect(applySet).ShouldNot(HaveKey("status"))
				Expect(revertSet).ShouldNot(HaveKey("status"))
			})

			It("will convert payload that is stored as a string", func() {
				bsonObj := getBSONData(test.CBGDexcomG5StringPayloadDatum)
				applySet, _, _, revertSet := setup(bsonObj)
				Expect(applySet).Should(HaveKeyWithValue("payload", map[string]interface{}{"systemTime": "2017-11-05T18:56:51Z", "transmitterId": "410X6M", "transmitterTicks": 5.796922e+06, "trend": "flat", "trendRate": 0.6, "trendRateUnits": "mg/dL/min"}))
				Expect(revertSet).Should(HaveKeyWithValue("payload", "{\"systemTime\":\"2017-11-05T18:56:51Z\",\"transmitterId\":\"410X6M\",\"transmitterTicks\":5796922,\"trend\":\"flat\",\"trendRate\":0.6,\"trendRateUnits\":\"mg/dL/min\"}"))
			})

			It("will convert annotations that are stored as a string", func() {
				bsonObj := getBSONData(test.CBGDexcomG5StringAnnotationsDatum)
				applySet, _, _, revertSet := setup(bsonObj)

				Expect(applySet).Should(HaveKey("annotations"))

				expectedAnnotations := []interface{}{
					map[string]interface{}{"code": "bg/out-of-range", "threshold": 40, "value": "low"},
				}
				applyObj := applySet.(primitive.M)
				actualAnnotations := applyObj["annotations"]

				Expect(fmt.Sprintf("%v", actualAnnotations)).To(Equal(fmt.Sprintf("%v", expectedAnnotations)))

				Expect(revertSet).Should(HaveKeyWithValue("annotations", "[{\"code\":\"bg/out-of-range\",\"threshold\":40,\"value\":\"low\"}]"))
			})
		})
	})
})
