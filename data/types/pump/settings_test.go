package pump

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("Settings", func() {
	var settingsObj = fixtures.TestingDatumBase()
	var unitsObj = make(map[string]interface{})
	var carbRatiosObj = make([]map[string]interface{}, 0)
	var insulinSensitivitiesObj = make([]map[string]interface{}, 0)
	var bgTargetsObj = make([]map[string]interface{}, 0)
	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()

		/*
			var goodObject = {
			  type: 'pumpSettings',
			  "activeSchedule": "standard",
			  "units": {
			    "carb": "grams",
			    "bg": "mmol/L"
			  },
			  "basalSchedules": {
			    "standard": [
			      { "rate": 0.8, "start": 0 },
			      { "rate": 0.75, "start": 3600000 }
			    ],
			    "pattern a": [
			      { "rate": 0.95, "start": 0 },
			      { "rate": 0.9, "start": 3600000 }
			    ]
			  },
			  "carbRatio": [
			    { "amount": 12, "start": 0 },
			    { "amount": 10, "start": 21600000 }
			  ],
			  "insulinSensitivity": [
			    { "amount": 3.6, "start": 0 },
			    { "amount": 2.5, "start": 18000000 }
			  ],
			  "bgTarget": [
			    { "low": 5.5, "high": 6.7, "start": 0 },
			    { "low": 5, "high": 6.1, "start": 18000000 }
			  ]
			};
		*/

		settingsObj["type"] = "pumpSettings"
		settingsObj["activeSchedule"] = "standard"

		unitsObj["carb"] = "grams"
		unitsObj["bg"] = "mmol/L"

		settingsObj["units"] = unitsObj

		carbRatiosObj = []map[string]interface{}{{"amount": 12.0, "start": 0}, {"amount": 10.0, "start": 21600000}}
		settingsObj["carbRatio"] = carbRatiosObj

		bgTargetsObj = []map[string]interface{}{{"low": 5.5, "high": 6.7, "start": 0}, {"low": 5, "high": 6.1, "start": 18000000}}
		settingsObj["bgTarget"] = bgTargetsObj

		insulinSensitivitiesObj = []map[string]interface{}{{"amount": 3.6, "start": 0}, {"amount": 2.5, "start": 18000000}}
		settingsObj["insulinSensitivity"] = insulinSensitivitiesObj

	})

	Context("setting record from obj", func() {

		It("when valid", func() {
			Expect(helper.ValidDataType(Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("activeSchedule", func() {

				It("is required", func() {
					delete(settingsObj, "activeSchedule")
					Expect(
						helper.ErrorIsExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/activeSchedule",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

			})

			Context("units", func() {

				It("is not required", func() {
					delete(settingsObj, "units")
					Expect(helper.ValidDataType(Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("if present requires carb", func() {

					delete(unitsObj, "carb")
					settingsObj["units"] = unitsObj

					Expect(
						helper.ErrorIsExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/carb",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

				It("if present requires bg", func() {

					delete(unitsObj, "bg")
					settingsObj["units"] = unitsObj

					Expect(
						helper.ErrorIsExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/bg",
								Detail: "Must be one of mmol/L, mg/dL given '<nil>'",
							}),
					).To(BeNil())
				})

			})

			Context("carbRatio", func() {

				It("is not required", func() {
					delete(settingsObj, "carbRatio")
					Expect(helper.ValidDataType(Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("bgTarget", func() {

				It("is not required", func() {
					delete(settingsObj, "bgTarget")
					Expect(helper.ValidDataType(Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("insulinSensitivity", func() {

				It("is not required", func() {
					delete(settingsObj, "insulinSensitivity")
					Expect(helper.ValidDataType(Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})
	})
})
