package data

import (
	"github.com/tidepool-org/platform/validate"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Base", func() {

	const (
		userID   = "b676436f60"
		groupID  = "43099shgs55"
		uploadID = "upid_b856b0e6e519"
	)

	var basalObj = map[string]interface{}{
		"userId":           userID,  //userID would have been injected by now via the builder
		"groupId":          groupID, //groupId would have been injected by now via the builder
		"uploadId":         uploadID,
		"deviceTime":       "2014-06-11T06:00:00.000Z",
		"time":             "2014-06-11T06:00:00.000Z",
		"timezoneOffset":   0,
		"conversionOffset": 0,
		"clockDriftOffset": 0,
		"type":             "basal",
		"deliveryType":     "scheduled",
		"scheduleName":     "Standard",
		"rate":             2.2,
		"duration":         21600000,
		"deviceId":         "InsOmn-111111111",
	}

	Context("can be built with all fields", func() {
		var (
			processing = validate.ErrorProcessing{BasePath: "0/base", ErrorsArray: validate.NewErrorsArray()}
		)
		It("should return a the base types if the obj is valid", func() {
			base := BuildBase(basalObj, processing)
			var baseType Base
			Expect(base).To(BeAssignableToTypeOf(baseType))
		})
		It("should return and error object that is empty but not nil", func() {
			BuildBase(basalObj, processing)
			Expect(processing.HasErrors()).To(BeFalse())
		})
	})

	Context("can be built with only core fields", func() {

		var (
			basalObj = map[string]interface{}{
				"userId":     userID, //userID would have been injected by now via the builder
				"groupId":    groupID,
				"uploadId":   uploadID,
				"deviceTime": "2014-06-11T06:00:00.000Z",
				"time":       "2014-06-11T06:00:00.000Z",
				"type":       "basal",
				"deviceId":   "InsOmn-111111111",
			}
			processing = validate.ErrorProcessing{BasePath: "0/base", ErrorsArray: validate.NewErrorsArray()}
		)
		It("should return a the base types if the obj is valid", func() {
			base := BuildBase(basalObj, processing)
			var baseType Base
			Expect(base).To(BeAssignableToTypeOf(baseType))
		})
		It("should return and error object that is empty but not nil", func() {
			BuildBase(basalObj, processing)
			Expect(processing.HasErrors()).To(BeFalse())
		})
	})

	Context("validation", func() {

		var processing validate.ErrorProcessing

		Context("time", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/base", ErrorsArray: validate.NewErrorsArray()}
			})

			Context("is invalid when", func() {
				It("there is no date", func() {
					basalObj["time"] = ""
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Time' failed with 'Times need to be ISO 8601 format and not in the future' when given ''"))
				})
				It("the date is not the right spec", func() {
					basalObj["time"] = "Monday, 02 Jan 2016"
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Time' failed with 'Times need to be ISO 8601 format and not in the future' when given 'Monday, 02 Jan 2016'"))
				})
				It("the date does not include hours and mins", func() {
					basalObj["time"] = "2016-02-05"
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Time' failed with 'Times need to be ISO 8601 format and not in the future' when given '2016-02-05'"))
				})
				It("the date does not include mins", func() {
					basalObj["time"] = "2016-02-05T20"
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Time' failed with 'Times need to be ISO 8601 format and not in the future' when given '2016-02-05T20'"))
				})
			})
			Context("is valid when", func() {
				It("the date is RFC3339 formated - e.g. 1", func() {
					basalObj["time"] = "2016-03-14T20:22:21+13:00"
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
				It("the date is RFC3339 formated - e.g. 2", func() {
					basalObj["time"] = "2016-02-05T15:53:00"
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
				It("the date is RFC3339 formated - e.g. 3", func() {
					basalObj["time"] = "2016-02-05T15:53:00.000Z"
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
			})
		})
		Context("timezoneOffset", func() {

			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/base", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				It("less then -840", func() {
					basalObj["timezoneOffset"] = -841
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'TimezoneOffset' failed with 'TimezoneOffset needs to be in minutes and greater than -840 and less than 720' when given '-841'"))
				})

				It("greater than 720", func() {
					basalObj["timezoneOffset"] = 721
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'TimezoneOffset' failed with 'TimezoneOffset needs to be in minutes and greater than -840 and less than 720' when given '721'"))
				})
			})
			Context("is valid when", func() {
				It("-840", func() {
					basalObj["timezoneOffset"] = -840
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("720", func() {
					basalObj["timezoneOffset"] = 720
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("0", func() {
					basalObj["timezoneOffset"] = 0
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
			})
		})
		Context("payload", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/base", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is valid when", func() {

				It("an interface", func() {
					basalObj["payload"] = map[string]string{"some": "stuff", "in": "here"}
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
		Context("annotations", func() {

			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/base", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("valid when", func() {

				It("many annotations", func() {
					basalObj["annotations"] = []interface{}{"some", "stuff", "in", "here"}
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("one annotation", func() {
					basalObj["annotations"] = []interface{}{"one"}
					base := BuildBase(basalObj, processing)
					getPlatformValidator().Struct(base, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
			})
		})
	})
	Context("convertions", func() {
		It("int when zero", func() {
			var intVal = map[string]interface{}{"myint": 0}
			var processing = validate.ErrorProcessing{BasePath: "0/test", ErrorsArray: validate.NewErrorsArray()}
			zero := 0

			converted := ToInt("myint", intVal["myint"], processing)
			Expect(converted).To(Equal(&zero))
			Expect(processing.HasErrors()).To(BeFalse())

		})
	})
})
