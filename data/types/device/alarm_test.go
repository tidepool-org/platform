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

	Context("alarm", func() {

		var deviceEventObj = fixtures.TestingDatumBase()
		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "alarm"
		deviceEventObj["alarmType"] = "low_insulin"
		deviceEventObj["status"] = "stuff"

		It("returns a Alarm if the obj is valid", func() {
			Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {
			Context("alarmType", func() {

				It("is required", func() {
					delete(deviceEventObj, "alarmType")

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/alarmType",
								Detail: "Must be one of low_insulin, no_insulin, low_power, no_power, occlusion, no_delivery, auto_off, over_limit, other given '<nil>'",
							}),
					).To(BeNil())
				})

				It("invalid if type not in allowed list", func() {
					deviceEventObj["alarmType"] = "nope"

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/alarmType",
								Detail: "Must be one of low_insulin, no_insulin, low_power, no_power, occlusion, no_delivery, auto_off, over_limit, other given 'nope'",
							}),
					).To(BeNil())
				})

				It("is case sensitive", func() {
					deviceEventObj["alarmType"] = "Low_Insulin"

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/alarmType",
								Detail: "Must be one of low_insulin, no_insulin, low_power, no_power, occlusion, no_delivery, auto_off, over_limit, other given 'Low_Insulin'",
							}),
					).To(BeNil())
				})

				It("valid if in the list", func() {
					deviceEventObj["alarmType"] = "no_power"
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
			Context("status", func() {
				It("is not required", func() {
					delete(deviceEventObj, "status")
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})
				It("is free text", func() {
					deviceEventObj["status"] = "moarstuff"
					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
		})
	})
})
