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

			if apply[0] != nil && apply[0]["$set"] != nil {
				applySet = apply[0]["$set"]
			}
			if len(apply) == 2 {
				if apply[1]["$unset"] != nil {
					applyUnset = apply[1]["$unset"]
				}
			}
			if revert[0] != nil && revert[0]["$unset"] != nil {
				revertUnset = revert[0]["$unset"]
			}
			if len(revert) == 2 {
				if revert[1]["$set"] != nil {
					revertSet = revert[1]["$set"]
				}
			}
			return applySet, applyUnset, revertUnset, revertSet
		}

		var _ = Describe("ProcessDatum", func() {

			It("smbg only sets or reverts _deduplicator and value", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.SMBGValueDatum))

				Expect(applySet).Should(HaveLen(2))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(HaveLen(1))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "Q3DdX+M2N0kmtylZBiObYDt7JoFzWNkLWJaYcXXd9Zw="}))
				Expect(applySet).Should(HaveKeyWithValue("value", 22.20299))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKeyWithValue("value", 22.202991964182132))

			})

			It("cbg only sets or reverts _deduplicator and value", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.CBGValueDatum))

				Expect(applySet).Should(HaveLen(2))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(HaveLen(1))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "kDdzWxsC4qNdfnnuWDYDX+fkZtFF7ZI/ZvvBL5PDa+s="}))
				Expect(applySet).Should(HaveKeyWithValue("value", 3.88552))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKeyWithValue("value", 3.8855235937318735))

			})

			It("bloodKetone only sets or reverts _deduplicator and value", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.BloodKetoneValueDatum))

				Expect(applySet).Should(HaveLen(2))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(HaveLen(1))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "nkLnx6jBepJGYnBs3xOKCT8wFP5jYTqzi5Dq2NXXy+A="}))
				Expect(applySet).Should(HaveKeyWithValue("value", 7.21597))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKeyWithValue("value", 7.2159723883591935))

			})

			It("basal only sets or reverts _deduplicator", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.AutomatedBasalTandem))

				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "YOItOWBgIIoEkqVsBq9yrOZ5utmsKTIezszpGBj5Vpc="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))

			})

			It("cgmSettings only sets or reverts _deduplicator", func() {
				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.CGMSetting))

				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "gyyB8OqbErdW2aOOo8POTXk1SNJmu5gDEIaCugTVn3M="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
			})

			It("pumpSettings only sets or reverts _deduplicator", func() {
				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.PumpSettingsTandem))

				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "l5e6HoVqMu3ZOUjqaky/m6ZNw+D0UFxbYw/fM9P4PXc="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
			})

			It("pumpSettings with _deduplicator and sleepSchedules updates", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.PumpSettingsWithSleepScheduleTandem))

				Expect(applySet).Should(HaveLen(2))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(HaveLen(1))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "l5e6HoVqMu3ZOUjqaky/m6ZNw+D0UFxbYw/fM9P4PXc="}))
				Expect(applySet).Should(HaveKey("sleepSchedules"))

				applyObj := applySet.(primitive.M)

				actualUpdatedSleepSchedules := applyObj["sleepSchedules"]

				expectedUpdatedSleepSchedules := map[string]interface{}{
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

				Expect(fmt.Sprintf("%v", actualUpdatedSleepSchedules)).To(Equal(fmt.Sprintf("%v", expectedUpdatedSleepSchedules)))

				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKey("sleepSchedules"))
				revertSetObj := revertSet.(primitive.M)
				actualRevertSetSleepSchedules := revertSetObj["sleepSchedules"]

				expectedRevertSleepSchedules := []map[string]interface{}{
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

				Expect(fmt.Sprintf("%v", actualRevertSetSleepSchedules)).To(Equal(fmt.Sprintf("%v", expectedRevertSleepSchedules)))

			})

			It("wizard only sets or reverts _deduplicator", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.WizardTandem))

				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "orP5cbifS8h0f3HWZcTOIf4B431HO1OReg9o1nmFnU4="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
			})

			It("pumpSettings only sets or reverts _deduplicator and ignores with bgTraget updates", func() {
				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.PumpSettingsCarelink))

				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "LgRaGs4QkIBV9sHUjurpMt/ALU+7F7ZlU8xNxhkTQwQ="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
			})

			It("will only sets or reverts _deduplicator and ignores empty payload", func() {
				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.EmptyPayloadDatum))
				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "dcXIRasQiatLHLG8oUjiG2yNSKetWpkC7GDMQ8ZpM/c="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
			})

			It("pumpSettings with _deduplicator and boluses updates", func() {
				bsonObj := getBSONData(test.PumpSettingsWithBolusDatum)
				applySet, applyUnset, revertUnset, revertSet := setup(bsonObj)

				Expect(applySet).Should(HaveLen(2))
				Expect(revertUnset).Should(HaveLen(2))

				Expect(applyUnset).Should(HaveLen(1))
				Expect(revertSet).Should(HaveLen(1))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "l5e6HoVqMu3ZOUjqaky/m6ZNw+D0UFxbYw/fM9P4PXc="}))
				Expect(applySet).Should(HaveKey("boluses"))
				Expect(revertSet).Should(HaveKey("bolus"))

				revertSetObj := revertSet.(primitive.M)
				Expect(applySet).Should(HaveKeyWithValue("boluses", revertSetObj["bolus"]))

				Expect(applyUnset).Should(HaveKeyWithValue("bolus", ""))
				Expect(revertUnset).Should(HaveKeyWithValue("boluses", ""))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
			})

			It("wizard only sets or reverts _deduplicator and ignores the bolus and bolusId link", func() {

				bsonObj := getBSONData(test.WizardTandem)
				Expect(bsonObj).Should(HaveKeyWithValue("bolus", "g2h6nohp5sdndpvl2l8kdete00lle4gt"))

				applySet, applyUnset, revertUnset, revertSet := setup(bsonObj)

				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "orP5cbifS8h0f3HWZcTOIf4B431HO1OReg9o1nmFnU4="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
			})

			It("deviceEvent only sets or reverts _deduplicator and ignores the status and statusId link", func() {
				bsonObj := getBSONData(test.ReservoirChangeWithStatus)
				Expect(bsonObj).Should(HaveKeyWithValue("status", "cvv61jde62b6i28bgot57f18bor5au1n"))
				applySet, applyUnset, revertUnset, revertSet := setup(bsonObj)

				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "yahFM0LCaLowGnmbqHijnOpfwkR3Ot/YVK7K5n5yIHg="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
			})

			It("deviceEvent only sets or reverts _deduplicator and ignores status suspended and will not update it", func() {
				bsonObj := getBSONData(test.AlarmDeviceEventDatum)
				Expect(bsonObj).Should(HaveKeyWithValue("status", "suspended"))

				applySet, applyUnset, revertUnset, revertSet := setup(bsonObj)

				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "co0AMaEqrFrInC2Ek+HqbvmZRr9WTT0rEnZ8JXpm2Hg="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
			})

			It("will only sets or reverts _deduplicator and ignore the converted payload", func() {
				bsonObj := getBSONData(test.CBGDexcomG5StringPayloadDatum)
				applySet, applyUnset, revertUnset, revertSet := setup(bsonObj)

				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "Kix7EaZBCVwTaOR/LQPj6iJ08mFJOR/IR2nsvyDGtGA="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
			})

			It("will only sets or reverts _deduplicator and ignore the converted annotations", func() {
				bsonObj := getBSONData(test.CBGDexcomG5StringAnnotationsDatum)
				applySet, applyUnset, revertUnset, revertSet := setup(bsonObj)

				Expect(applySet).Should(HaveLen(1))
				Expect(revertUnset).Should(HaveLen(1))
				Expect(applyUnset).Should(BeNil())
				Expect(revertSet).Should(BeNil())

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "Kix7EaZBCVwTaOR/LQPj6iJ08mFJOR/IR2nsvyDGtGA="}))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))

			})
		})
	})
})
