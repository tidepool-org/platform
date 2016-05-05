package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("Base", func() {

	var baseObj types.Datum
	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
		baseObj = fixtures.TestingDatumBase()
		baseObj["type"] = "testing"
	})

	Context("CurrentSchemaVersion is set", func() {
		It("as 10", func() {
			Expect(types.CurrentSchemaVersion).To(Equal(10))
		})
	})

	Context("can be built with all fields", func() {

		It("and be of type base if the obj is valid", func() {
			Expect(helper.ValidDataType(types.BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
		})

		It("will have schema version set as CurrentSchemaVersion", func() {
			base := types.BuildBase(baseObj, helper.ErrorProcessing)
			Expect(base.SchemaVersion).To(Equal(types.CurrentSchemaVersion))
		})

		Context("validation", func() {

			Context("time", func() {

				It("there is no date", func() {
					baseObj["time"] = ""

					Expect(
						helper.ErrorIsExpected(
							types.BuildBase(baseObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "An ISO 8601-formatted timestamp including either a timezone offset from UTC OR converted to UTC with a final Z for 'Zulu' time. e.g.2013-05-04T03:58:44.584Z OR 2013-05-04T03:58:44-08:00 given ''",
							}),
					).To(BeNil())

				})

				It("the date is not the right spec", func() {
					baseObj["time"] = "Monday, 02 Jan 2016"

					Expect(
						helper.ErrorIsExpected(
							types.BuildBase(baseObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "An ISO 8601-formatted timestamp including either a timezone offset from UTC OR converted to UTC with a final Z for 'Zulu' time. e.g.2013-05-04T03:58:44.584Z OR 2013-05-04T03:58:44-08:00 given 'Monday, 02 Jan 2016'",
							}),
					).To(BeNil())

				})

				It("the date does not include hours and mins", func() {
					baseObj["time"] = "2016-02-05"

					Expect(
						helper.ErrorIsExpected(
							types.BuildBase(baseObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "An ISO 8601-formatted timestamp including either a timezone offset from UTC OR converted to UTC with a final Z for 'Zulu' time. e.g.2013-05-04T03:58:44.584Z OR 2013-05-04T03:58:44-08:00 given '2016-02-05'",
							}),
					).To(BeNil())
				})

				It("the date does not include mins", func() {
					baseObj["time"] = "2016-02-05T20"

					Expect(
						helper.ErrorIsExpected(
							types.BuildBase(baseObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/time",
								Detail: "An ISO 8601-formatted timestamp including either a timezone offset from UTC OR converted to UTC with a final Z for 'Zulu' time. e.g.2013-05-04T03:58:44.584Z OR 2013-05-04T03:58:44-08:00 given '2016-02-05T20'",
							}),
					).To(BeNil())
				})

			})
			Context("deviceId", func() {

				It("is required", func() {
					delete(baseObj, "deviceId")

					Expect(
						helper.ErrorIsExpected(
							types.BuildBase(baseObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceId",
								Detail: "This is a required field given '<nil>'",
							}),
					).To(BeNil())
				})

				It("cannot be empty", func() {
					baseObj["deviceId"] = ""

					Expect(
						helper.ErrorIsExpected(
							types.BuildBase(baseObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
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
							types.BuildBase(baseObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/timezoneOffset",
								Detail: "needs to be in minutes and >= -840 and <= 720 given '-841'",
							}),
					).To(BeNil())
				})

				It("greater than 720", func() {
					baseObj["timezoneOffset"] = 721
					Expect(
						helper.ErrorIsExpected(
							types.BuildBase(baseObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/timezoneOffset",
								Detail: "needs to be in minutes and >= -840 and <= 720 given '721'",
							}),
					).To(BeNil())

				})

				It("-840", func() {
					baseObj["timezoneOffset"] = -840
					Expect(helper.ValidDataType(types.BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("720", func() {
					baseObj["timezoneOffset"] = 720
					Expect(helper.ValidDataType(types.BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("0", func() {
					baseObj["timezoneOffset"] = 0
					Expect(helper.ValidDataType(types.BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("payload", func() {

				It("an interface", func() {
					baseObj["payload"] = map[string]string{"some": "stuff", "in": "here"}
					Expect(helper.ValidDataType(types.BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("annotations", func() {

				It("many annotations", func() {
					baseObj["annotations"] = []interface{}{"some", "stuff", "in", "here"}
					Expect(helper.ValidDataType(types.BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("one annotation", func() {
					baseObj["annotations"] = []interface{}{"one"}
					Expect(helper.ValidDataType(types.BuildBase(baseObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})
		Context("convertions", func() {

			It("int when zero", func() {
				var intVal = types.Datum{"myint": 0}
				zero := 0

				converted := intVal.ToInt("myint", helper.ErrorProcessing)
				Expect(converted).To(Equal(&zero))
				Expect(helper.ErrorProcessing.HasErrors()).To(BeFalse())

			})

		})

	})
})
