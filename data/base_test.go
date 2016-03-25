package data

import (
	"time"

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

	Context("can be built with all fields", func() {
		var (
			basalObj = map[string]interface{}{
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
		)
		It("should return a the base types if the obj is valid", func() {
			base, _ := BuildBase(basalObj)
			var baseType Base
			Expect(base).To(BeAssignableToTypeOf(baseType))
		})
		It("should return and error object that is empty but not nil", func() {
			_, err := BuildBase(basalObj)
			Expect(err).To(Not(BeNil()))
			Expect(err.IsEmpty()).To(BeTrue())
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
		)
		It("should return a the base types if the obj is valid", func() {
			base, _ := BuildBase(basalObj)
			var baseType Base
			Expect(base).To(BeAssignableToTypeOf(baseType))
		})
		It("should return and error object that is empty but not nil", func() {
			_, err := BuildBase(basalObj)
			Expect(err).To(Not(BeNil()))
			Expect(err.IsEmpty()).To(BeTrue())
		})
	})

	Context("validation", func() {

		var (
			validator = validate.NewPlatformValidator()
		)

		Context("TimeStringValidator", func() {
			BeforeEach(func() {
				validator.RegisterValidation(timeStringTag, TimeStringValidator)
			})

			Context("is invalid when", func() {
				It("there is no date", func() {
					nodate := Base{DeviceTime: ""}
					Expect(validator.Field(nodate.DeviceTime, timeStringTag)).ToNot(BeNil())
				})
				It("the date is not the right spec", func() {
					wrongspec := Base{DeviceTime: "Monday, 02 Jan 2016"}
					Expect(validator.Field(wrongspec.DeviceTime, timeStringTag)).ToNot(BeNil())
				})
				It("the date does not include hours and mins", func() {
					notime := Base{DeviceTime: "2016-02-05"}
					Expect(validator.Field(notime.DeviceTime, timeStringTag)).ToNot(BeNil())
				})
				It("the date does not include mins", func() {
					notime := Base{DeviceTime: "2016-02-05T20"}
					Expect(validator.Field(notime.DeviceTime, timeStringTag)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("the date is RFC3339 formated - e.g. 1", func() {
					valid := Base{DeviceTime: "2016-03-14T20:22:21+13:00"}
					Expect(validator.Field(valid.DeviceTime, timeStringTag)).To(BeNil())
				})
				It("the date is RFC3339 formated - e.g. 2", func() {
					valid := Base{DeviceTime: "2016-02-05T15:53:00"}
					Expect(validator.Field(valid.DeviceTime, timeStringTag)).To(BeNil())
				})
				It("the date is RFC3339 formated - e.g. 3", func() {
					valid := Base{DeviceTime: "2016-02-05T15:53:00.000Z"}
					Expect(validator.Field(valid.DeviceTime, timeStringTag)).To(BeNil())
				})
			})
		})
		Context("TimeObjectValidator", func() {
			type testStruct struct {
				GivenDate time.Time `json:"givenDate" valid:"timeobj"`
			}
			BeforeEach(func() {
				validator.RegisterValidation(timeObjectTag, TimeObjectValidator)
			})
			Context("is invalid when", func() {

				It("in the future", func() {
					furturedate := testStruct{GivenDate: time.Now().Add(time.Hour * 36)}
					Expect(validator.Struct(furturedate)).ToNot(BeNil())
				})
				It("zero", func() {
					zerodate := testStruct{}
					Expect(validator.Struct(zerodate)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {

				It("set now", func() {
					nowdate := testStruct{GivenDate: time.Now()}
					Expect(validator.Struct(nowdate)).To(BeNil())
				})
				It("set in the past", func() {
					pastdate := testStruct{GivenDate: time.Now().AddDate(0, -2, 0)}
					Expect(validator.Struct(pastdate)).To(BeNil())
				})
			})
		})
		Context("TimezoneOffsetValidator", func() {

			BeforeEach(func() {
				validator.RegisterValidation(timeZoneOffsetTag, TimezoneOffsetValidator)
			})
			Context("is invalid when", func() {
				It("less then -840", func() {
					under := Base{TimezoneOffset: -841}
					Expect(validator.Field(under.TimezoneOffset, timeZoneOffsetTag)).ToNot(BeNil())
				})
				It("greater than 720", func() {
					over := Base{TimezoneOffset: 721}
					Expect(validator.Field(over.TimezoneOffset, timeZoneOffsetTag)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("-840", func() {
					onLower := Base{TimezoneOffset: -840}
					Expect(validator.Field(onLower.TimezoneOffset, timeZoneOffsetTag)).To(BeNil())
				})
				It("720", func() {
					onUpper := Base{TimezoneOffset: 720}
					Expect(validator.Field(onUpper.TimezoneOffset, timeZoneOffsetTag)).To(BeNil())
				})
				It("0", func() {
					zero := Base{TimezoneOffset: 0}
					Expect(validator.Field(zero.TimezoneOffset, timeZoneOffsetTag)).To(BeNil())
				})
			})
		})
		Context("Payload", func() {
			BeforeEach(func() {
				validator.RegisterValidation(payloadTag, PayloadValidator)
			})
			Context("is valid when", func() {
				It("an interface", func() {
					base := Base{Payload: map[string]string{"some": "stuff", "in": "here"}}
					Expect(validator.Field(base.Payload, payloadTag)).To(BeNil())
				})
			})
		})
		Context("Annotations", func() {

			BeforeEach(func() {
				validator.RegisterValidation(annotationsTag, AnnotationsValidator)
			})
			Context("is valid when", func() {
				It("many annotations", func() {
					base := Base{Annotations: []interface{}{"some", "stuff", "in", "here"}}
					Expect(validator.Field(base.Annotations, annotationsTag)).To(BeNil())
				})
				It("one annotation", func() {
					base := Base{Annotations: []interface{}{"some"}}
					Expect(validator.Field(base.Annotations, annotationsTag)).To(BeNil())
				})
			})
		})
	})
})
