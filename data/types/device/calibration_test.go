package device_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/device"
)

var _ = Describe("DeviceEvent", func() {

	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("calibration", func() {

		var deviceEventObj = fixtures.TestingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "calibration"
		deviceEventObj["value"] = 3.0
		deviceEventObj["units"] = "mg/dL"

		It("returns a Calibration if the obj is valid", func() {
			Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("units", func() {

				It("is required", func() {

					delete(deviceEventObj, "units")

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/units",
								Detail: "Must be one of mmol/L, mg/dL given '<nil>'",
							}),
					).To(BeNil())
				})

				It("fails if not mmol/L or mg/dL", func() {
					deviceEventObj["units"] = "other"
					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/units",
								Detail: "Must be one of mmol/L, mg/dL given 'other'",
							}),
					).To(BeNil())
				})

				It("valid if mmol/L", func() {
					deviceEventObj["units"] = "mmol/L"
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("valid if mg/dL", func() {
					deviceEventObj["units"] = "mg/dL"
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})
			})

			Context("value", func() {

				It("is required mg/dL", func() {

					delete(deviceEventObj, "value")
					deviceEventObj["units"] = "mg/dL"

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/value",
								Detail: "Must be between 0.0 and 1000.0 given '<nil>'",
							}),
					).To(BeNil())
				})

				It("is required mmol/L", func() {

					delete(deviceEventObj, "value")
					deviceEventObj["units"] = "mmol/L"

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/value",
								Detail: "Must be between 0.0 and 55.0 given '<nil>'",
							}),
					).To(BeNil())
				})
			})
		})
	})
})
