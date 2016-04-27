package cgm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/cgm"
)

var _ = Describe("Settings", func() {
	var settingsObj = fixtures.TestingDatumBase()

	var lowAlertsObj = make(map[string]interface{}, 0)
	var highAlertsObj = make(map[string]interface{}, 0)
	var outOfRangeAlertsObj = make(map[string]interface{}, 0)
	var rateOfChangeAlertsObj = make(map[string]map[string]interface{}, 0)

	var helper *types.TestingHelper

	Context("setting record from obj", func() {

		BeforeEach(func() {
			helper = types.NewTestingHelper()

			settingsObj["type"] = "cgmSettings"
			settingsObj["units"] = "mmol/L"
			settingsObj["transmitterId"] = "test"

			lowAlertsObj = map[string]interface{}{"enabled": true, "snooze": 0, "level": 3.6079861941795968}
			settingsObj["lowAlerts"] = lowAlertsObj

			highAlertsObj = map[string]interface{}{"enabled": true, "snooze": 0, "level": 8.3261219865683}
			settingsObj["highAlerts"] = highAlertsObj

			outOfRangeAlertsObj = map[string]interface{}{"enabled": false, "snooze": 1800000}

			settingsObj["outOfRangeAlerts"] = outOfRangeAlertsObj

			rateOfChangeAlertsObj = map[string]map[string]interface{}{
				"fallRate": {"enabled": false, "rate": -0.16652243973136602},
				"riseRate": {"enabled": false, "rate": 0.16652243973136602},
			}

			settingsObj["rateOfChangeAlerts"] = rateOfChangeAlertsObj

		})

		It("when valid", func() {
			Expect(helper.ValidDataType(cgm.Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("units", func() {

				It("is required", func() {
					delete(settingsObj, "units")
					Expect(
						helper.ErrorIsExpected(
							cgm.Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/units",
								Detail: "Must be one of mmol/L, mg/dL given '<nil>'",
							}),
					).To(BeNil())
				})

				It("can be mmol/l", func() {
					settingsObj["units"] = "mmol/l"
					Expect(helper.ValidDataType(cgm.Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("can be mg/dl", func() {
					settingsObj["units"] = "mg/dl"
					Expect(helper.ValidDataType(cgm.Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("cannot be anything else", func() {
					settingsObj["units"] = "grams"

					Expect(
						helper.ErrorIsExpected(
							cgm.Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/units",
								Detail: "Must be one of mmol/L, mg/dL given 'grams'",
							}),
					).To(BeNil())

				})
			})

			Context("transmitterId", func() {

				It("is required", func() {
					delete(settingsObj, "transmitterId")
					Expect(
						helper.ErrorIsExpected(
							cgm.Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/transmitterId",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

				It("is free text", func() {
					settingsObj["transmitterId"] = "my transmitter"
					Expect(helper.ValidDataType(cgm.Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("lowAlerts", func() {

				It("is required", func() {
					delete(settingsObj, "lowAlerts")

					expected := make(map[string]types.ExpectedErrorDetails, 0)
					expected["0/level"] = types.ExpectedErrorDetails{Detail: "Must be >= 3.0 and <= 15.0 given '<nil>'"}
					expected["0/enabled"] = types.ExpectedErrorDetails{Detail: "This is a required field given '<nil>'"}
					expected["0/snooze"] = types.ExpectedErrorDetails{Detail: "Must be >= 0 and <= 432000000 given '<nil>'"}

					Expect(
						helper.HasExpectedErrors(
							cgm.Build(settingsObj, helper.ErrorProcessing),
							expected,
						),
					).To(BeNil())
				})
			})
			Context("highAlerts", func() {

				It("is required", func() {
					delete(settingsObj, "highAlerts")

					expected := make(map[string]types.ExpectedErrorDetails, 0)
					expected["0/level"] = types.ExpectedErrorDetails{Detail: "Must be >= 3.0 and <= 15.0 given '<nil>'"}
					expected["0/enabled"] = types.ExpectedErrorDetails{Detail: "This is a required field given '<nil>'"}
					expected["0/snooze"] = types.ExpectedErrorDetails{Detail: "Must be >= 0 and <= 432000000 given '<nil>'"}

					Expect(
						helper.HasExpectedErrors(
							cgm.Build(settingsObj, helper.ErrorProcessing),
							expected,
						),
					).To(BeNil())
				})
			})
			Context("outOfRangeAlerts", func() {
				It("is not required", func() {
					delete(settingsObj, "outOfRangeAlerts")
					Expect(helper.ValidDataType(cgm.Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
			/*Context("rateOfChangeAlerts", func() {
				It("is required", func() {
					delete(settingsObj, "rateOfChangeAlerts")
					Expect(
						helper.ErrorIsExpected(
							cgm.Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/rateOfChangeAlerts",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

			})*/

		})
	})
})
