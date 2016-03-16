package data_test

import (
	"time"

	. "github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/validate"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Base", func() {

	const (
		userid   = "b676436f60"
		groupid  = "43099shgs55"
		uploadid = "upid_b856b0e6e519"
	)

	Context("can be built with all fields", func() {
		var (
			basalObj = map[string]interface{}{
				"userId":           userid,  //userid would have been injected by now via the builder
				"groupId":          groupid, //groupId would have been injected by now via the builder
				"uploadId":         uploadid,
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
				"userId":     userid, //userid would have been injected by now via the builder
				"groupId":    groupid,
				"uploadId":   uploadid,
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
				validator.RegisterValidation("timestr", TimeStringValidator)
			})
			type testStruct struct {
				GivenDate string `json:"givenDate"  valid:"timestr"`
			}
			Context("is invalid when", func() {
				It("there is no date", func() {
					nodate := testStruct{GivenDate: ""}
					Expect(validator.ValidateStruct(nodate)).ToNot(BeNil())
				})
				It("the date is not the right spec", func() {
					wrongspec := testStruct{GivenDate: "Monday, 02 Jan 2016"}
					Expect(validator.ValidateStruct(wrongspec)).ToNot(BeNil())
				})
				It("the date does not include hours and mins", func() {
					notime := testStruct{GivenDate: "2016-02-05"}
					Expect(validator.ValidateStruct(notime)).ToNot(BeNil())
				})
				It("the date does not include mins", func() {
					notime := testStruct{GivenDate: "2016-02-05T20"}
					Expect(validator.ValidateStruct(notime)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("the date is RFC3339 formated - e.g. 1", func() {
					validdate := testStruct{GivenDate: "2016-03-14T20:22:21+13:00"}
					Expect(validator.ValidateStruct(validdate)).To(BeNil())
				})
				It("the date is RFC3339 formated - e.g. 2", func() {
					validdate := testStruct{GivenDate: "2016-02-05T15:53:00"}
					Expect(validator.ValidateStruct(validdate)).To(BeNil())
				})
				It("the date is RFC3339 formated - e.g. 3", func() {
					validdate := testStruct{GivenDate: "2016-02-05T15:53:00.000Z"}
					Expect(validator.ValidateStruct(validdate)).To(BeNil())
				})
			})
		})
		Context("TimeObjectValidator", func() {
			type testStruct struct {
				GivenDate time.Time `json:"givenDate"  valid:"timeobj"`
			}
			BeforeEach(func() {
				validator.RegisterValidation("timeobj", TimeObjectValidator)
			})
			Context("is invalid when", func() {

				It("in the future", func() {
					furturedate := testStruct{GivenDate: time.Now().Add(time.Hour * 36)}
					Expect(validator.ValidateStruct(furturedate)).ToNot(BeNil())
				})
				It("zero", func() {
					zerodate := testStruct{}
					Expect(validator.ValidateStruct(zerodate)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {

				It("set now", func() {
					nowdate := testStruct{GivenDate: time.Now()}
					Expect(validator.ValidateStruct(nowdate)).To(BeNil())
				})
				It("set in the past", func() {
					pastdate := testStruct{GivenDate: time.Now().AddDate(0, -2, 0)}
					Expect(validator.ValidateStruct(pastdate)).To(BeNil())
				})
			})
		})
		Context("TimezoneOffsetValidator", func() {
			type testStruct struct {
				Offset int `json:"offset"  valid:"tzoffset"`
			}
			BeforeEach(func() {
				validator.RegisterValidation("tzoffset", TimezoneOffsetValidator)
			})
			Context("is invalid when", func() {
				It("less then -840", func() {
					under := testStruct{Offset: -841}
					Expect(validator.ValidateStruct(under)).ToNot(BeNil())
				})
				It("greater than 720", func() {
					over := testStruct{Offset: 721}
					Expect(validator.ValidateStruct(over)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("-840", func() {
					under := testStruct{Offset: -840}
					Expect(validator.ValidateStruct(under)).To(BeNil())
				})
				It("720", func() {
					over := testStruct{Offset: 720}
					Expect(validator.ValidateStruct(over)).To(BeNil())
				})
				It("0", func() {
					over := testStruct{Offset: 0}
					Expect(validator.ValidateStruct(over)).To(BeNil())
				})
			})
		})
		Context("Payload", func() {
			type testStruct struct {
				Payload interface{} `json:"payload"  valid:"payload"`
			}
			BeforeEach(func() {
				validator.RegisterValidation("payload", PayloadValidator)
			})
			Context("is valid when", func() {
				It("an interface", func() {
					payload := testStruct{Payload: map[string]string{"some": "stuff", "in": "here"}}
					Expect(validator.ValidateStruct(payload)).To(BeNil())
				})
				It("not an interface", func() {
					type testStruct struct {
						Payload int `json:"payload"  valid:"payload"`
					}
					intPayload := testStruct{Payload: 100}
					Expect(validator.ValidateStruct(intPayload)).To(BeNil())
				})
			})
		})
		Context("Annotations", func() {
			type testStruct struct {
				Annotations []interface{} `json:"annotations"  valid:"annotations"`
			}
			BeforeEach(func() {
				validator.RegisterValidation("annotations", AnnotationsValidator)
			})
			Context("is valid when", func() {
				It("many annotations", func() {
					annotations := testStruct{Annotations: []interface{}{"some", "stuff", "in", "here"}}
					Expect(validator.ValidateStruct(annotations)).To(BeNil())
				})
				It("one annotation", func() {
					annotation := testStruct{Annotations: []interface{}{"some"}}
					Expect(validator.ValidateStruct(annotation)).To(BeNil())
				})
			})
			Context("is invalid when", func() {
				It("not an array", func() {
					type testStruct struct {
						Annotations interface{} `json:"annotations"  valid:"annotations"`
					}
					badAnnotation := testStruct{Annotations: "some"}
					Expect(validator.ValidateStruct(badAnnotation)).ToNot(BeNil())
				})
			})
		})
	})
})
