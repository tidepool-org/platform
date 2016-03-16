package data_test

import (
	. "github.com/tidepool-org/platform/data"
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
				validator.RegisterValidation("basalrate", BasalRateValidator)
			})
			type testStruct struct {
				Rate float64 `json:"rate"  valid:"basalrate"`
			}
			Context("is invalid when", func() {
				It("zero", func() {
					zeroRate := testStruct{Rate: 0}
					Expect(validator.ValidateStruct(zeroRate)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("greater than zero", func() {
					greaterThanZeroRate := testStruct{Rate: 10.6}
					Expect(validator.ValidateStruct(greaterThanZeroRate)).To(BeNil())
				})
			})
		})
		Context("BasalDurationValidator", func() {
			BeforeEach(func() {
				validator.RegisterValidation("basalduration", BasalDurationValidator)
			})
			type testStruct struct {
				Duration int `json:"duration"  valid:"basalduration"`
			}
			Context("is invalid when", func() {
				It("zero", func() {
					zeroDuration := testStruct{Duration: 0}
					Expect(validator.ValidateStruct(zeroDuration)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("greater than zero", func() {
					greaterThanZeroDuration := testStruct{Duration: 4000}
					Expect(validator.ValidateStruct(greaterThanZeroDuration)).To(BeNil())
				})
			})
		})
		Context("BasalDeliveryTypeValidator", func() {
			BeforeEach(func() {
				validator.RegisterValidation("basaldeliverytype", BasalDeliveryTypeValidator)
			})
			type testStruct struct {
				DeliveryType string `json:"deliverytype"  valid:"basaldeliverytype"`
			}
			Context("is invalid when", func() {
				It("there is no matching type", func() {
					naType := testStruct{DeliveryType: "superfly"}
					Expect(validator.ValidateStruct(naType)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("injected type", func() {
					injectedType := testStruct{DeliveryType: "injected"}
					Expect(validator.ValidateStruct(injectedType)).To(BeNil())
				})
				It("scheduled type", func() {
					scheduledType := testStruct{DeliveryType: "scheduled"}
					Expect(validator.ValidateStruct(scheduledType)).To(BeNil())
				})
				It("suspend type", func() {
					suspendType := testStruct{DeliveryType: "suspend"}
					Expect(validator.ValidateStruct(suspendType)).To(BeNil())
				})
				It("temp type", func() {
					tempType := testStruct{DeliveryType: "temp"}
					Expect(validator.ValidateStruct(tempType)).To(BeNil())
				})
			})
		})
		Context("BasalInjectionValidator", func() {
			BeforeEach(func() {
				validator.RegisterValidation("injection", BasalInjectionValidator)
			})
			type testStruct struct {
				Injection string `json:"injection"  valid:"injection"`
			}
			Context("is invalid when", func() {
				It("there is no matching type", func() {
					naType := testStruct{Injection: "good"}
					Expect(validator.ValidateStruct(naType)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("levemir type", func() {
					levemirType := testStruct{Injection: "levemir"}
					Expect(validator.ValidateStruct(levemirType)).To(BeNil())
				})
				It("lantus type", func() {
					lantusType := testStruct{Injection: "lantus"}
					Expect(validator.ValidateStruct(lantusType)).To(BeNil())
				})
			})
		})
		Context("BasalInjectionValueValidator", func() {
			BeforeEach(func() {
				validator.RegisterValidation("injectionvalue", BasalInjectionValueValidator)
			})
			type testStruct struct {
				Value int `json:"value"  valid:"injectionvalue"`
			}
			Context("is invalid when", func() {
				It("zero", func() {
					zeroValue := testStruct{Value: 0}
					Expect(validator.ValidateStruct(zeroValue)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("greater than zero", func() {
					greaterThanZeroValue := testStruct{Value: 1}
					Expect(validator.ValidateStruct(greaterThanZeroValue)).To(BeNil())
				})
			})
		})
	})
})
