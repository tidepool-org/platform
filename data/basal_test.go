package data

import (
	"github.com/tidepool-org/platform/validate"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
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
	)

	Context("datum from obj", func() {
		It("should return a basal if the obj is valid", func() {
			basal, _ := BuildBasal(basalObj)
			var basalType *Basal
			Expect(basal).To(BeAssignableToTypeOf(basalType))
		})
		It("should produce no error when valid", func() {
			_, err := BuildBasal(basalObj)
			Expect(err).To(BeNil())
		})
	})

	Context("dataset from builder", func() {
		It("should return a basal if the obj is valid", func() {
			basal, _ := BuildBasal(basalObj)
			var basalType *Basal
			Expect(basal).To(BeAssignableToTypeOf(basalType))
		})
		It("should produce no error when valid", func() {
			_, err := BuildBasal(basalObj)
			Expect(err).To(BeNil())
		})
	})
	Context("validation", func() {
		var (
			validator = validate.NewPlatformValidator()
		)

		Context("BasalRateValidator", func() {
			BeforeEach(func() {
				validator.RegisterValidation(rateTag, BasalRateValidator)
			})

			Context("is invalid when", func() {
				It("zero", func() {
					zeroRate := Basal{Rate: 0}
					Expect(validator.Field(zeroRate.Rate, rateTag)).ToNot(BeNil())
				})
				It("zero and gives me context in error", func() {
					zeroRate := Basal{Rate: 0}
					errs := validator.Field(zeroRate.Rate, rateTag)

					if errs != nil {
						Expect(errs.GetError(basalFailureReasons).Error()).To(Equal("Error:Field validation failed with 'Must be greater than 0.000000' when given '0' for type 'float64'"))
					}

				})
			})
			Context("is valid when", func() {
				It("greater than zero", func() {
					greaterThanZeroRate := Basal{Rate: 10.6}
					Expect(validator.Field(greaterThanZeroRate.Rate, rateTag)).To(BeNil())
				})
			})
		})
		Context("BasalDurationValidator", func() {
			BeforeEach(func() {
				validator.RegisterValidation(durationTag, BasalDurationValidator)
			})
			Context("is invalid when", func() {
				It("zero", func() {
					zeroDuration := Basal{Duration: 0}
					Expect(validator.Field(zeroDuration.Duration, durationTag)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("greater than zero", func() {
					greaterThanZeroDuration := Basal{Duration: 4000}
					Expect(validator.Field(greaterThanZeroDuration.Duration, durationTag)).To(BeNil())
				})
			})
		})
		Context("BasalDeliveryTypeValidator", func() {
			BeforeEach(func() {
				validator.RegisterValidation(deliveryTypeTag, BasalDeliveryTypeValidator)
			})
			Context("is invalid when", func() {
				It("there is no matching type", func() {
					naType := Basal{DeliveryType: "superfly"}
					Expect(validator.Field(naType.DeliveryType, deliveryTypeTag)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("injected type", func() {
					injectedType := Basal{DeliveryType: "injected"}
					Expect(validator.Field(injectedType.DeliveryType, deliveryTypeTag)).To(BeNil())
				})
				It("scheduled type", func() {
					scheduledType := Basal{DeliveryType: "scheduled"}
					Expect(validator.Field(scheduledType.DeliveryType, deliveryTypeTag)).To(BeNil())
				})
				It("suspend type", func() {
					suspendType := Basal{DeliveryType: "suspend"}
					Expect(validator.Field(suspendType.DeliveryType, deliveryTypeTag)).To(BeNil())
				})
				It("temp type", func() {
					tempType := Basal{DeliveryType: "temp"}
					Expect(validator.Field(tempType.DeliveryType, deliveryTypeTag)).To(BeNil())
				})
			})
		})
		Context("BasalInsulinValidator", func() {
			BeforeEach(func() {
				validator.RegisterValidation(insulinTag, BasalInsulinValidator)
			})
			Context("is invalid when", func() {
				It("there is no matching type", func() {
					naType := Basal{Insulin: "good"}
					Expect(validator.Field(naType.Insulin, insulinTag)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("levemir type", func() {
					levemirType := Basal{Insulin: "levemir"}
					Expect(validator.Field(levemirType.Insulin, insulinTag)).To(BeNil())
				})
				It("lantus type", func() {
					lantusType := Basal{Insulin: "lantus"}
					Expect(validator.Field(lantusType.Insulin, insulinTag)).To(BeNil())
				})
			})
		})
		Context("BasalValueValidator", func() {
			BeforeEach(func() {
				validator.RegisterValidation(valueTag, BasalValueValidator)
			})
			Context("is invalid when", func() {
				It("zero", func() {
					zeroValue := Basal{Value: 0}
					Expect(validator.Field(zeroValue.Value, valueTag)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("greater than zero", func() {
					greaterThanZeroValue := Basal{Value: 1}
					Expect(validator.Field(greaterThanZeroValue.Value, valueTag)).To(BeNil())
				})
			})
		})
	})
})
