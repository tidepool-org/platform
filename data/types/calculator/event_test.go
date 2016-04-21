package calculator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("Event", func() {
	var calculatorObj = fixtures.TestingDatumBase()
	var recommendedObj = make(map[string]interface{})
	var bolusObj = make(map[string]interface{})
	var bgTargetObj = make(map[string]interface{})
	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
		calculatorObj["type"] = "wizard"
		calculatorObj["carbInput"] = 45
		calculatorObj["bgInput"] = 99.0
		calculatorObj["insulinOnBoard"] = 1.3
		calculatorObj["insulinSensitivity"] = 75
		calculatorObj["units"] = "mg/dL"

		recommendedObj["carb"] = 4.0
		recommendedObj["correction"] = 1.0
		recommendedObj["net"] = 4.0

		calculatorObj["recommended"] = recommendedObj

		bolusObj["type"] = "bolus"
		bolusObj["subType"] = "normal"
		bolusObj["deviceId"] = "test"
		bolusObj["time"] = "2014-01-01T01:00:00.000Z"

		calculatorObj["bolus"] = bolusObj

		bgTargetObj["high"] = 120.0
		bgTargetObj["low"] = 80.0

		calculatorObj["bgTarget"] = bgTargetObj

	})

	Context("calculator record from obj", func() {

		It("when valid", func() {
			Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("carbInput", func() {

				It("is not required", func() {
					delete(calculatorObj, "carbInput")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("bgInput", func() {

				It("is not required", func() {
					delete(calculatorObj, "bgInput")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("insulinOnBoard", func() {

				It("is not required", func() {
					delete(calculatorObj, "insulinOnBoard")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

			})

			Context("units", func() {

				It("is required", func() {
					delete(calculatorObj, "units")

					Expect(
						helper.ErrorIsExpected(
							Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/units",
								Detail: "Must be one of mmol/L, mg/dL given '<nil>'",
							}),
					).To(BeNil())
				})

			})

			Context("recommended", func() {

				It("is not required", func() {
					delete(calculatorObj, "recommended")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("if present requires net", func() {

					delete(recommendedObj, "net")
					calculatorObj["recommended"] = recommendedObj

					Expect(
						helper.ErrorIsExpected(
							Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/net",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

				It("if present requires correction", func() {

					delete(recommendedObj, "correction")
					calculatorObj["recommended"] = recommendedObj

					Expect(
						helper.ErrorIsExpected(
							Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/correction",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

				It("if present requires carb", func() {

					delete(recommendedObj, "carb")
					calculatorObj["recommended"] = recommendedObj

					Expect(
						helper.ErrorIsExpected(
							Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/carb",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

			})

			Context("bolus", func() {

				It("is not required", func() {
					delete(calculatorObj, "bolus")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("if present requires subType", func() {

					delete(bolusObj, "subType")
					calculatorObj["bolus"] = bolusObj

					Expect(
						helper.ErrorIsExpected(
							Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/subType",
								Detail: "Must be one of normal, square, dual/square given '<nil>'",
							}),
					).To(BeNil())
				})

				It("if present requires time", func() {

					delete(bolusObj, "time")
					calculatorObj["bolus"] = bolusObj

					Expect(
						helper.ErrorIsExpected(
							Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "Times need to be ISO 8601 format and not in the future given '<nil>'",
							}),
					).To(BeNil())
				})

				It("if present requires deviceId", func() {

					delete(bolusObj, "deviceId")
					calculatorObj["bolus"] = bolusObj

					Expect(
						helper.ErrorIsExpected(
							Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceId",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

			})

			Context("bgTarget", func() {

				It("is not required", func() {
					delete(calculatorObj, "bgTarget")
					Expect(helper.ValidDataType(Build(calculatorObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("if present requires high", func() {

					delete(bgTargetObj, "high")
					calculatorObj["bgTarget"] = bgTargetObj

					Expect(
						helper.ErrorIsExpected(
							Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/high",
								Detail: "Must be greater than 0.0 given '<nil>'",
							}),
					).To(BeNil())
				})

				It("if present requires low", func() {

					delete(bgTargetObj, "low")
					calculatorObj["bgTarget"] = bgTargetObj

					Expect(
						helper.ErrorIsExpected(
							Build(calculatorObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/low",
								Detail: "Must be greater than 0.0 given '<nil>'",
							}),
					).To(BeNil())
				})

			})
		})

	})
})
