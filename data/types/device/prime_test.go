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
	var deviceEventObj = fixtures.TestingDatumBase()

	BeforeEach(func() {
		helper = types.NewTestingHelper()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "prime"
		deviceEventObj["primeTarget"] = "cannula"
		deviceEventObj["volume"] = 1.0
	})

	Context("prime", func() {

		It("returns a PrimeDeviceEvent if the obj is valid", func() {
			Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {
			Context("primeTarget", func() {

				It("is required", func() {

					delete(deviceEventObj, "primeTarget")

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/primeTarget",
								Detail: "Must be one of cannula, tubing given '<nil>'",
							}),
					).To(BeNil())
				})

				It("can be tubing", func() {

					deviceEventObj["primeTarget"] = "tubing"

					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("can be cannula", func() {

					deviceEventObj["primeTarget"] = "cannula"

					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("can't be anything else required", func() {

					deviceEventObj["primeTarget"] = "other"

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/primeTarget",
								Detail: "Must be one of cannula, tubing given 'other'",
							}),
					).To(BeNil())
				})
			})

			Context("volume", func() {

				It("is not required", func() {

					delete(deviceEventObj, "volume")

					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("if cannula in range of 0.0 to 3.0", func() {
					deviceEventObj["primeTarget"] = "cannula"
					deviceEventObj["volume"] = 1.0

					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("if cannula > 3.0 fails", func() {

					deviceEventObj["primeTarget"] = "cannula"
					deviceEventObj["volume"] = 3.1

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/volume",
								Detail: "Must be >= 0.0 and <= 3.0 given '3.1'",
							}),
					).To(BeNil())
				})

				It("if cannula < 0.0 fails", func() {

					deviceEventObj["primeTarget"] = "cannula"
					deviceEventObj["volume"] = -0.1

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/volume",
								Detail: "Must be >= 0.0 and <= 3.0 given '-0.1'",
							}),
					).To(BeNil())
				})

				It("if tubing in range of 0.0 to 100.0", func() {

					deviceEventObj["primeTarget"] = "tubing"
					deviceEventObj["volume"] = 55.0

					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("if tubing > 100.0 fails", func() {

					deviceEventObj["primeTarget"] = "tubing"
					deviceEventObj["volume"] = 100.1

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/volume",
								Detail: "Must be >= 0.0 and <= 100.0 given '100.1'",
							}),
					).To(BeNil())
				})

				It("if tubing < 0.0 fails", func() {

					deviceEventObj["primeTarget"] = "tubing"
					deviceEventObj["volume"] = -0.1

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/volume",
								Detail: "Must be >= 0.0 and <= 100.0 given '-0.1'",
							}),
					).To(BeNil())
				})

			})

		})
	})

})
