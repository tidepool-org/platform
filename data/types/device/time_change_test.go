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
	var change = make(map[string]interface{}, 0)

	BeforeEach(func() {
		helper = types.NewTestingHelper()

		deviceEventObj["type"] = "deviceEvent"
		deviceEventObj["subType"] = "timeChange"
		change = map[string]interface{}{
			"from":     "2015-03-08T12:02:00",
			"to":       "2015-03-08T13:00:00",
			"agent":    "manual",
			"reasons":  []string{"to_daylight_savings", "correction"},
			"timezone": "US/Pacific",
		}

		deviceEventObj["change"] = change
	})

	Context("timeChange", func() {

		It("returns a TimeChange if the obj is valid", func() {
			Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("reasons", func() {
				It("not required", func() {
					delete(change, "reasons")
					deviceEventObj["change"] = change

					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})
				/*It("can be empty", func() {

					change["reasons"] = []string{}
					deviceEventObj["change"] = change

					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("can be any of the approved types", func() {

					change["reasons"] = []string{"from_daylight_savings", "to_daylight_savings", "travel", "correction", "other"}
					deviceEventObj["change"] = change

					Expect(helper.ValidDataType(device.Build(deviceEventObj, helper.ErrorProcessing))).To(BeNil())
				})
				It("cannot be an un-approved type ", func() {

					change["reasons"] = []string{"from_daylight_savings", "nope", "travel", "correction", "other"}
					deviceEventObj["change"] = change

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "change/reasons/1",
								Detail: "Must be one of from_daylight_savings, to_daylight_savings, travel, correction, other given '[from_daylight_savings nope travel correction other]'",
							}),
					).To(BeNil())
				})*/
			})
			Context("agent", func() {
				It("is required", func() {
					delete(change, "agent")
					deviceEventObj["change"] = change

					Expect(
						helper.ErrorIsExpected(
							device.Build(deviceEventObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/agent",
								Detail: "Must be one of manual, automatic given '<nil>'",
							}),
					).To(BeNil())
				})
			})
		})
	})
})
