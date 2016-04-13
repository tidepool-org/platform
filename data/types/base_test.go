package types

import (
	"github.com/tidepool-org/platform/data/_fixtures"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Base", func() {

	var basalObj = fixtures.TestingDatumBase()
	var helper *TestingHelper

	BeforeEach(func() {
		helper = NewTestingHelper()
		basalObj["type"] = "basal"
		basalObj["deliveryType"] = "scheduled"
		basalObj["scheduleName"] = "Standard"
		basalObj["rate"] = 2.2
		basalObj["duration"] = 21600000
	})

	Context("can be built with all fields", func() {

		It("and be of type base if the obj is valid", func() {
			Expect(helper.ValidDataType(BuildBase(basalObj, helper.ErrorProcessing))).To(BeNil())
		})

	})

	Context("can be built with only core fields", func() {

		core := fixtures.TestingDatumBase()
		core["type"] = "tbd"

		It("and be of type base if the obj is valid", func() {
			Expect(helper.ValidDataType(BuildBase(core, helper.ErrorProcessing))).To(BeNil())
		})
	})

	Context("validation", func() {

		Context("time", func() {

			Context("is invalid when", func() {
				It("there is no date", func() {
					basalObj["time"] = ""

					Expect(
						helper.ErrorIsExpected(
							BuildBase(basalObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "Times need to be ISO 8601 format and not in the future given ''",
							}),
					).To(BeNil())

				})

				It("the date is not the right spec", func() {
					basalObj["time"] = "Monday, 02 Jan 2016"

					Expect(
						helper.ErrorIsExpected(
							BuildBase(basalObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "Times need to be ISO 8601 format and not in the future given 'Monday, 02 Jan 2016'",
							}),
					).To(BeNil())

				})

				It("the date does not include hours and mins", func() {
					basalObj["time"] = "2016-02-05"

					Expect(
						helper.ErrorIsExpected(
							BuildBase(basalObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "Times need to be ISO 8601 format and not in the future given '2016-02-05'",
							}),
					).To(BeNil())
				})

				It("the date does not include mins", func() {
					basalObj["time"] = "2016-02-05T20"

					Expect(
						helper.ErrorIsExpected(
							BuildBase(basalObj, helper.ErrorProcessing),
							ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "Times need to be ISO 8601 format and not in the future given '2016-02-05T20'",
							}),
					).To(BeNil())

				})
			})
			Context("is valid when", func() {
				It("the date is RFC3339 formated - e.g. 1", func() {
					basalObj["time"] = "2016-03-14T20:22:21+13:00"
					Expect(helper.ValidDataType(BuildBase(basalObj, helper.ErrorProcessing))).To(BeNil())
				})
				It("the date is RFC3339 formated - e.g. 2", func() {
					basalObj["time"] = "2016-02-05T15:53:00"
					Expect(helper.ValidDataType(BuildBase(basalObj, helper.ErrorProcessing))).To(BeNil())
				})
				It("the date is RFC3339 formated - e.g. 3", func() {
					basalObj["time"] = "2016-02-05T15:53:00.000Z"
					Expect(helper.ValidDataType(BuildBase(basalObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
			Context("timezoneOffset", func() {

				Context("is invalid when", func() {

					It("less then -840", func() {
						basalObj["timezoneOffset"] = -841
						Expect(
							helper.ErrorIsExpected(
								BuildBase(basalObj, helper.ErrorProcessing),
								ExpectedErrorDetails{
									Path:   "0/timezoneOffset",
									Detail: "needs to be in minutes and >= -840 and <= 720 given '-841'",
								}),
						).To(BeNil())
					})

					It("greater than 720", func() {
						basalObj["timezoneOffset"] = 721
						Expect(
							helper.ErrorIsExpected(
								BuildBase(basalObj, helper.ErrorProcessing),
								ExpectedErrorDetails{
									Path:   "0/timezoneOffset",
									Detail: "needs to be in minutes and >= -840 and <= 720 given '721'",
								}),
						).To(BeNil())

					})
				})
				Context("is valid when", func() {
					It("-840", func() {
						basalObj["timezoneOffset"] = -840
						Expect(helper.ValidDataType(BuildBase(basalObj, helper.ErrorProcessing))).To(BeNil())
					})

					It("720", func() {
						basalObj["timezoneOffset"] = 720
						Expect(helper.ValidDataType(BuildBase(basalObj, helper.ErrorProcessing))).To(BeNil())
					})

					It("0", func() {
						basalObj["timezoneOffset"] = 0
						Expect(helper.ValidDataType(BuildBase(basalObj, helper.ErrorProcessing))).To(BeNil())
					})
				})
			})
			Context("payload", func() {

				Context("is valid when", func() {

					It("an interface", func() {
						basalObj["payload"] = map[string]string{"some": "stuff", "in": "here"}
						Expect(helper.ValidDataType(BuildBase(basalObj, helper.ErrorProcessing))).To(BeNil())
					})

				})
			})
			Context("annotations", func() {

				Context("valid when", func() {

					It("many annotations", func() {
						basalObj["annotations"] = []interface{}{"some", "stuff", "in", "here"}
						Expect(helper.ValidDataType(BuildBase(basalObj, helper.ErrorProcessing))).To(BeNil())
					})

					It("one annotation", func() {
						basalObj["annotations"] = []interface{}{"one"}
						Expect(helper.ValidDataType(BuildBase(basalObj, helper.ErrorProcessing))).To(BeNil())
					})
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
