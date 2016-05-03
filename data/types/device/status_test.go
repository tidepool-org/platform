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

	Context("status", func() {

		var deviceEventObj = fixtures.TestingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "status"
		deviceEventObj["status"] = "suspended"
		deviceEventObj["duration"] = 1000
		deviceEventObj["reason"] = map[string]interface{}{"suspended": "automatic"}

		It("returns a Status if the obj is valid", func() {
			Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {
			Context("status", func() {

				It("is required", func() {

					delete(deviceEventObj, "status")

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/status",
								Detail: "Must be one of suspended, resumed given '<nil>'",
							}),
					).To(BeNil())
				})

				It("fails if not suspended or resumed", func() {
					deviceEventObj["status"] = "other"
					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/status",
								Detail: "Must be one of suspended, resumed given 'other'",
							}),
					).To(BeNil())
				})

				It("valid if suspended", func() {
					deviceEventObj["status"] = "suspended"
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("valid if resumed", func() {
					deviceEventObj["status"] = "resumed"
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
			Context("reason", func() {

				It("is required", func() {

					delete(deviceEventObj, "reason")

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/reason",
								Detail: "Must be one of manual, automatic given '<nil>'",
							}),
					).To(BeNil())
				})

				It("fails if not manual or automatic", func() {
					deviceEventObj["reason"] = map[string]interface{}{"suspended": "other"}
					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path: "0/reason",
								//TODO: no one should need to know what a map is outside of the platform.
								Detail: "Must be one of manual, automatic given 'map[suspended:other]'",
							}),
					).To(BeNil())
				})

				It("valid if manual", func() {
					deviceEventObj["reason"] = map[string]interface{}{"suspended": "manual"}
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("valid if automatic", func() {
					deviceEventObj["reason"] = map[string]interface{}{"suspended": "automatic"}
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
			Context("duration", func() {

				It("is not required", func() {

					delete(deviceEventObj, "duration")
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())

				})

				It("can't be < 0", func() {
					deviceEventObj["duration"] = -1
					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/duration",
								Detail: "Must be one of manual, automatic given '-1'",
							}),
					).To(BeNil())
				})

				It("valid if >= 0 ", func() {
					deviceEventObj["duration"] = 60000
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
		})
	})
})
