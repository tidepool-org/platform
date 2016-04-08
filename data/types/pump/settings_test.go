package pump

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("Settings", func() {
	var settingsObj = fixtures.TestingDatumBase()
	var unitsObj = make(map[string]interface{})
	var carbRatiosObj = make([]map[string]interface{}, 0)
	var insulinSensitivitiesObj = make([]map[string]interface{}, 0)
	var basalSchedulesObj = make(map[string][]map[string]interface{}, 0)
	var bgTargetsObj = make([]map[string]interface{}, 0)
	var helper *types.TestingHelper

	Context("setting record from obj", func() {

		BeforeEach(func() {
			helper = types.NewTestingHelper()

			settingsObj["type"] = "pumpSettings"
			settingsObj["activeSchedule"] = "standard"

			unitsObj["carb"] = "grams"
			unitsObj["bg"] = "mmol/L"

			settingsObj["units"] = unitsObj

			carbRatiosObj = []map[string]interface{}{{"amount": 12.0, "start": 0}, {"amount": 10.0, "start": 21600000}}
			settingsObj["carbRatio"] = carbRatiosObj

			bgTargetsObj = []map[string]interface{}{{"low": 5.5, "high": 6.7, "start": 0}, {"low": 5.0, "high": 6.1, "start": 18000000}}
			settingsObj["bgTarget"] = bgTargetsObj

			insulinSensitivitiesObj = []map[string]interface{}{{"amount": 3.6, "start": 0}, {"amount": 2.5, "start": 18000000}}
			settingsObj["insulinSensitivity"] = insulinSensitivitiesObj

			basalSchedulesObj = map[string][]map[string]interface{}{
				"standard":  {{"rate": 0.8, "start": 0}, {"rate": 0.75, "start": 3600000}},
				"pattern a": {{"rate": 0.95, "start": 0}, {"rate": 0.9, "start": 3600000}},
			}

			settingsObj["basalSchedules"] = basalSchedulesObj

		})

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

				It("will have two", func() {
					settings := Build(settingsObj, helper.ErrorProcessing)
					Expect(helper.ValidDataType(settings)).To(BeNil())
					Expect(len(settings.CarbohydrateRatios)).To(Equal(2))
				})

				It("if present requires start", func() {

					carbRatiosObj = []map[string]interface{}{{"amount": 12.0}, {"amount": 10.0}}
					settingsObj["carbRatio"] = carbRatiosObj

					Expect(
						helper.ErrorsAreExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/start",
								Detail: "Needs to be in the range of >= 0 and < 86400000 given '<nil>'",
							}),
					).To(BeNil())
				})

				It("start needs to be greater than equal to 0", func() {

					carbRatiosObj = []map[string]interface{}{{"amount": 12.0, "start": -10}, {"amount": 10.0, "start": 0}}
					settingsObj["carbRatio"] = carbRatiosObj

					Expect(
						helper.ErrorIsExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/start",
								Detail: "Needs to be in the range of >= 0 and < 86400000 given '-10'",
							}),
					).To(BeNil())
				})

				It("start needs to be less than 86400000", func() {

					carbRatiosObj = []map[string]interface{}{{"amount": 12.0, "start": 0}, {"amount": 10.0, "start": 86400000}}
					settingsObj["carbRatio"] = carbRatiosObj

					Expect(
						helper.ErrorIsExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/start",
								Detail: "Needs to be in the range of >= 0 and < 86400000 given '86400000'",
							}),
					).To(BeNil())
				})

				It("if present requires amount", func() {

					carbRatiosObj = []map[string]interface{}{{"start": 0}, {"start": 6400000}}
					settingsObj["carbRatio"] = carbRatiosObj

					Expect(
						helper.ErrorsAreExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/amount",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

			})

			Context("bgTarget", func() {

				It("is not required", func() {
					delete(settingsObj, "bgTarget")
					Expect(helper.ValidDataType(Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("will have two", func() {
					settings := Build(settingsObj, helper.ErrorProcessing)
					Expect(helper.ValidDataType(settings)).To(BeNil())
					Expect(len(settings.BloodGlucoseTargets)).To(Equal(2))
				})

				It("if present requires start", func() {

					bgTargetsObj = []map[string]interface{}{{"low": 5.5, "high": 6.7}, {"low": 5.0, "high": 6.1, "start": 18000000}}
					settingsObj["bgTarget"] = bgTargetsObj

					Expect(
						helper.ErrorIsExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/start",
								Detail: "Needs to be in the range of >= 0 and < 86400000 given '<nil>'",
							}),
					).To(BeNil())
				})

				It("if present requires low", func() {

					bgTargetsObj = []map[string]interface{}{{"low": 5.5, "high": 6.7, "start": 0}, {"high": 6.1, "start": 18000000}}
					settingsObj["bgTarget"] = bgTargetsObj

					Expect(
						helper.ErrorIsExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/low",
								Detail: "Must be greater than 0.0 given '<nil>'",
							}),
					).To(BeNil())
				})

				It("if present requires high", func() {

					bgTargetsObj = []map[string]interface{}{{"low": 5.5, "start": 0}, {"low": 5.0, "high": 6.1, "start": 18000000}}
					settingsObj["bgTarget"] = bgTargetsObj

					Expect(
						helper.ErrorIsExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/high",
								Detail: "Must be greater than 0.0 given '<nil>'",
							}),
					).To(BeNil())
				})

			})

			Context("insulinSensitivity", func() {

				It("is not required", func() {
					delete(settingsObj, "insulinSensitivity")
					Expect(helper.ValidDataType(Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("will have two", func() {
					settings := Build(settingsObj, helper.ErrorProcessing)
					Expect(helper.ValidDataType(settings)).To(BeNil())
					Expect(len(settings.InsulinSensitivities)).To(Equal(2))
				})

				It("if present requires start", func() {

					insulinSensitivitiesObj = []map[string]interface{}{{"amount": 3.6, "start": 0}, {"amount": 2.5}}
					settingsObj["insulinSensitivity"] = insulinSensitivitiesObj

					Expect(
						helper.ErrorsAreExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/start",
								Detail: "Needs to be in the range of >= 0 and < 86400000 given '<nil>'",
							}),
					).To(BeNil())
				})

				It("if present requires amount", func() {

					insulinSensitivitiesObj = []map[string]interface{}{{"start": 0}, {"amount": 2.5, "start": 18000000}}
					settingsObj["insulinSensitivity"] = insulinSensitivitiesObj

					Expect(
						helper.ErrorsAreExpected(
							Build(settingsObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/amount",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

			})

			Context("basalSchedules", func() {

				It("is not required", func() {
					delete(settingsObj, "basalSchedules")
					Expect(helper.ValidDataType(Build(settingsObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("will have two", func() {
					settings := Build(settingsObj, helper.ErrorProcessing)
					Expect(helper.ValidDataType(settings)).To(BeNil())
					Expect(len(settings.BasalSchedules)).To(Equal(2))
				})

			})
		})
	})
})
