package data

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Basal", func() {

	const (
		userid   = "b676436f60"
		groupid  = "43099shgs55"
		uploadid = "upid_b856b0e6e519"
	)

	var (
		basalObj = map[string]interface{}{
			"userId":           userid, //userid would have been injected by now via the builder
			"groupId":          groupid,
			"uploadId":         uploadid,
			"time":             "2016-02-25T23:02:00.000Z",
			"timezoneOffset":   -480,
			"clockDriftOffset": 0,
			"conversionOffset": 0,
			"deviceTime":       "2016-02-25T15:02:00.000Z",
			"deviceId":         "IR1285-79-36047-15",
			"type":             "basal",
			"deliveryType":     "scheduled",
			"scheduleName":     "DEFAULT",
			"rate":             1.75,
			"duration":         28800000,
		}

		processing validate.ErrorProcessing
	)

	Context("datum from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
		})

		It("should return a basal if the obj is valid", func() {
			basal := BuildBasal(basalObj, processing)
			var basalType *Basal
			Expect(basal).To(BeAssignableToTypeOf(basalType))
		})

		It("should produce no error when valid", func() {
			BuildBasal(basalObj, processing)
			Expect(processing.HasErrors()).To(BeFalse())
		})

	})

	Context("validation", func() {

		Context("rate", func() {

			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
			})

			Context("is invalid when", func() {

				It("zero", func() {
					basalObj["rate"] = -1.0
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("zero and gives me context in error", func() {
					basalObj["rate"] = -0.1
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Rate' failed with 'Must be greater than 0.0' when given '-0.1'"))
				})

			})
			Context("is valid when", func() {

				It("greater than zero", func() {
					basalObj["rate"] = 0.7
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
		Context("duration", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				It("zero", func() {
					basalObj["duration"] = -1
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("gives detailed error", func() {
					basalObj["duration"] = -1
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Duration' failed with 'Must be greater than 0' when given '-1'"))
				})

			})
			Context("is valid when", func() {

				It("greater than zero", func() {
					basalObj["duration"] = 4000
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
		Context("deliveryType", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				It("there is no matching type", func() {
					basalObj["deliveryType"] = "superfly"
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})
				It("gives detailed error", func() {
					basalObj["deliveryType"] = "superfly"
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'DeliveryType' failed with 'Must be one of injected,scheduled,suspend,temp' when given 'superfly'"))
				})
			})
			Context("is valid when", func() {
				It("injected type", func() {
					basalObj["deliveryType"] = "injected"
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
				It("scheduled type", func() {
					basalObj["deliveryType"] = "scheduled"
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
				It("suspend type", func() {
					basalObj["deliveryType"] = "suspend"
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
				It("temp type", func() {
					basalObj["deliveryType"] = "temp"
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
			})
		})
		Context("insulin", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				It("there is no matching type", func() {
					basalObj["insulin"] = "good"
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("gives detailed error", func() {
					basalObj["insulin"] = "good"
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Insulin' failed with 'Must be one of levemir,lantus' when given 'good'"))
				})

			})
			Context("is valid when", func() {

				It("levemir type", func() {
					basalObj["insulin"] = "levemir"
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

				It("lantus type", func() {
					basalObj["insulin"] = "lantus"
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})
			})
		})
		Context("value", func() {
			BeforeEach(func() {
				processing = validate.ErrorProcessing{BasePath: "0/basal", ErrorsArray: validate.NewErrorsArray()}
			})
			Context("is invalid when", func() {

				/*It("zero", func() {
					basalObj["value"] = -1
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
				})

				It("gives detailed error", func() {
					basalObj["value"] = -1
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(processing.Errors[0].Detail).To(ContainSubstring("'Insulin' failed with 'Must be one of levemir,lantus' when given 'good'"))
				})*/

			})
			Context("is valid when", func() {

				It("greater than zero", func() {
					basalObj["value"] = 1
					basal := BuildBasal(basalObj, processing)
					getPlatformValidator().Struct(basal, processing)
					Expect(processing.HasErrors()).To(BeFalse())
				})

			})
		})
	})
})
