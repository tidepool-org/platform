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

			It("basal with unwanted percent feild", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.AutomatedBasalTandem))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "CFDp66+LJvYW7rxf+4ndFd8hoTMq+ymzwLnuEUEqhVs="}))
				Expect(applyUnset).Should(HaveKeyWithValue("percent", ""))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKeyWithValue("percent", float64(0.47857142857142865)))
			})

			It("pump settings with blood glucose precsion updates", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.PumpSettingsTandem))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "bpKLJbi5JfqD7N0WJ1vj0ck03c9EZ3U0H09TCLhdd3k="}))
				Expect(applySet).Should(HaveKeyWithValue("bgTargets.Simple.0.target", 5.55075))
				Expect(applySet).Should(HaveKeyWithValue("bgTargets.Simple.1.target", 5.55075))
				Expect(applySet).Should(HaveKeyWithValue("bgTargets.Standard.0.target", 5.55075))
				Expect(applySet).Should(HaveKeyWithValue("bgTargets.Standard.1.target", 5.55075))
				Expect(applySet).Should(HaveKeyWithValue("units.bg", "mmol/L"))
				Expect(applyUnset).Should(HaveKeyWithValue("localTime", ""))
				Expect(revertUnset).Should(HaveKeyWithValue("_deduplicator", ""))
				Expect(revertSet).Should(HaveKeyWithValue("localTime", "2017-11-05T12:56:51.000Z"))
				Expect(revertSet).Should(HaveKeyWithValue("bgTargets.Simple.0.target", 5.550747991045533))
				Expect(revertSet).Should(HaveKeyWithValue("bgTargets.Simple.1.target", 5.550747991045533))
				Expect(revertSet).Should(HaveKeyWithValue("bgTargets.Standard.0.target", 5.550747991045533))
				Expect(revertSet).Should(HaveKeyWithValue("bgTargets.Standard.1.target", 5.550747991045533))
				Expect(revertSet).Should(HaveKeyWithValue("units.bg", "mg/dL"))
			})

			It("wizard with bgInput and bgTarget glucose updates", func() {

				applySet, applyUnset, revertUnset, revertSet := setup(getBSONData(test.WizardTandem))

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "o6ybZQtDZ95FvuV0zYGphri2SIGesbLCbkHxc1wbbEE="}))
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

				Expect(applySet).Should(HaveKeyWithValue("_deduplicator", map[string]interface{}{"hash": "NC17pw1UAaab50iChhQXJ+N9dTi6GduTy9UjsMHolow="}))
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

		})
	})
})
