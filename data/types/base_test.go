package types

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
)

var _ = Describe("Base", func() {

	var baseObj Datum
	var helper *TestingHelper

	BeforeEach(func() {
		helper = NewTestingHelper()
		baseObj = fixtures.TestingDatumBase()
		baseObj["type"] = "testing"
	})

	Context("can be built with all fields", func() {

		It("and be of type base if the obj is valid", func() {
			Expect(helper.ValidDataType(BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("time", func() {

				It("there is no date", func() {
					baseObj["time"] = ""

					Expect(
						helper.ErrorIsExpected(
							BuildBase(baseObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "Times need to be ISO 8601 format and not in the future given ''",
							}),
					).To(BeNil())

				})

				It("the date is not the right spec", func() {
					baseObj["time"] = "Monday, 02 Jan 2016"

					Expect(
						helper.ErrorIsExpected(
							BuildBase(baseObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "Times need to be ISO 8601 format and not in the future given 'Monday, 02 Jan 2016'",
							}),
					).To(BeNil())

				})

				It("the date does not include hours and mins", func() {
					baseObj["time"] = "2016-02-05"

					Expect(
						helper.ErrorIsExpected(
							BuildBase(baseObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "Times need to be ISO 8601 format and not in the future given '2016-02-05'",
							}),
					).To(BeNil())
				})

				It("the date does not include mins", func() {
					baseObj["time"] = "2016-02-05T20"

					Expect(
						helper.ErrorIsExpected(
							BuildBase(baseObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "Times need to be ISO 8601 format and not in the future given '2016-02-05T20'",
							}),
					).To(BeNil())
				})

				It("the date is RFC3339 formated - e.g. 1", func() {
					baseObj["time"] = "2016-03-14T20:22:21+13:00"
					Expect(helper.ValidDataType(BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("the date is RFC3339 formated - e.g. 2", func() {
					baseObj["time"] = "2016-02-05T15:53:00"
					Expect(helper.ValidDataType(BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("the date is RFC3339 formated - e.g. 3", func() {
					baseObj["time"] = "2016-02-05T15:53:00.000Z"
					Expect(helper.ValidDataType(BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("deviceId", func() {

				It("is required", func() {
					delete(baseObj, "deviceId")

					Expect(
						helper.ErrorIsExpected(
							BuildBase(baseObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/deviceId",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

				It("cannot be empty", func() {
					baseObj["deviceId"] = ""

					Expect(
						helper.ErrorIsExpected(
							BuildBase(baseObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/deviceId",
								Detail: "This is a required field given ''",
							}),
					).To(BeNil())
				})

			})
			Context("timezoneOffset", func() {

				It("less then -840", func() {
					baseObj["timezoneOffset"] = -841
					Expect(
						helper.ErrorIsExpected(
							BuildBase(baseObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/timezoneOffset",
								Detail: "needs to be in minutes and >= -840 and <= 720 given '-841'",
							}),
					).To(BeNil())
				})

				It("greater than 720", func() {
					baseObj["timezoneOffset"] = 721
					Expect(
						helper.ErrorIsExpected(
							BuildBase(baseObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/timezoneOffset",
								Detail: "needs to be in minutes and >= -840 and <= 720 given '721'",
							}),
					).To(BeNil())

				})

				It("-840", func() {
					baseObj["timezoneOffset"] = -840
					Expect(helper.ValidDataType(BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("720", func() {
					baseObj["timezoneOffset"] = 720
					Expect(helper.ValidDataType(BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("0", func() {
					baseObj["timezoneOffset"] = 0
					Expect(helper.ValidDataType(BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("payload", func() {

				It("an interface", func() {
					baseObj["payload"] = map[string]string{"some": "stuff", "in": "here"}
					Expect(helper.ValidDataType(BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("annotations", func() {

				It("many annotations", func() {
					baseObj["annotations"] = []interface{}{"some", "stuff", "in", "here"}
					Expect(helper.ValidDataType(BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("one annotation", func() {
					baseObj["annotations"] = []interface{}{"one"}
					Expect(helper.ValidDataType(BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})
		Context("convertions", func() {

			It("int when zero", func() {
				var intVal = Datum{"myint": 0}
				zero := 0

				converted := intVal.ToInt("myint", helper.ErrorProcessing)
				Expect(converted).To(Equal(&zero))
				Expect(helper.HasErrors()).To(BeFalse())

			})

		})

	})
})
