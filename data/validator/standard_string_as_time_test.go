package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/test"
)

var _ = Describe("StandardStringAsTime", func() {
	It("New returns nil if context is nil", func() {
		value := "2015-12-31T13:14:16-08:00"
		Expect(validator.NewStandardStringAsTime(nil, "ghost", &value, "2006-01-02T15:04:05Z07:00")).To(BeNil())
	})

	It("New returns nil if time layout is empty string", func() {
		value := "2015-12-31T13:14:16-08:00"
		standardContext, err := context.NewStandard(test.NewLogger())
		Expect(standardContext).ToNot(BeNil())
		Expect(err).To(Succeed())
		Expect(validator.NewStandardStringAsTime(standardContext, "ghost", &value, "")).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(test.NewLogger())
			Expect(standardContext).ToNot(BeNil())
			Expect(err).To(Succeed())
		})

		Context("new validator with nil reference and nil value", func() {
			var standard *validator.StandardStringAsTime
			var result data.Time

			BeforeEach(func() {
				standard = validator.NewStandardStringAsTime(standardContext, nil, nil, "2006-01-02T15:04:05Z07:00")
			})

			It("exists", func() {
				Expect(standard).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standard.Exists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-does-not-exist"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value does not exist"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value does not exist"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/<nil>"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("After", func() {
				BeforeEach(func() {
					result = standard.After(time.Unix(1451567655, 0).UTC())
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("AfterNow", func() {
				BeforeEach(func() {
					result = standard.AfterNow()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("Before", func() {
				BeforeEach(func() {
					result = standard.Before(time.Unix(1451567655, 0).UTC())
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("BeforeNow", func() {
				BeforeEach(func() {
					result = standard.BeforeNow()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})
		})

		Context("new validator with valid reference and an invalid value", func() {
			var standard *validator.StandardStringAsTime
			var result data.Time

			BeforeEach(func() {
				value := "invalid"
				standard = validator.NewStandardStringAsTime(standardContext, "ghost", &value, "2006-01-02T15:04:05Z07:00")
			})

			It("exists", func() {
				Expect(standard).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standard.Exists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("time-not-valid"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not a valid time"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"invalid\" is not a valid time of format \"2006-01-02T15:04:05Z07:00\""))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghost"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("After", func() {
				BeforeEach(func() {
					result = standard.After(time.Unix(1451567655, 0).UTC())
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("time-not-valid"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not a valid time"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"invalid\" is not a valid time of format \"2006-01-02T15:04:05Z07:00\""))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghost"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("AfterNow", func() {
				BeforeEach(func() {
					result = standard.AfterNow()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("time-not-valid"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not a valid time"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"invalid\" is not a valid time of format \"2006-01-02T15:04:05Z07:00\""))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghost"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("Before", func() {
				BeforeEach(func() {
					result = standard.Before(time.Unix(1451567655, 0).UTC())
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("time-not-valid"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not a valid time"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"invalid\" is not a valid time of format \"2006-01-02T15:04:05Z07:00\""))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghost"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("BeforeNow", func() {
				BeforeEach(func() {
					result = standard.BeforeNow()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("time-not-valid"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not a valid time"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"invalid\" is not a valid time of format \"2006-01-02T15:04:05Z07:00\""))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghost"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})
		})

		Context("new validator with valid reference and value well into the past", func() {
			var standard *validator.StandardStringAsTime
			var result data.Time

			BeforeEach(func() {
				value := "1990-01-01T14:15:16Z"
				standard = validator.NewStandardStringAsTime(standardContext, "ghost", &value, "2006-01-02T15:04:05Z07:00")
			})

			It("exists", func() {
				Expect(standard).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standard.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("After", func() {
				BeforeEach(func() {
					result = standard.After(time.Unix(1451567655, 0).UTC())
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("time-not-after"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not after the specified time"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"1990-01-01T14:15:16Z\" is not after \"2015-12-31T13:14:15Z\""))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghost"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("AfterNow", func() {
				BeforeEach(func() {
					result = standard.AfterNow()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("time-not-after"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not after the specified time"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"1990-01-01T14:15:16Z\" is not after now"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghost"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("Before", func() {
				BeforeEach(func() {
					result = standard.Before(time.Unix(1451567655, 0).UTC())
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("BeforeNow", func() {
				BeforeEach(func() {
					result = standard.BeforeNow()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})
		})

		Context("new validator with valid reference and value well into the future", func() {
			var standard *validator.StandardStringAsTime
			var result data.Time

			BeforeEach(func() {
				value := "2090-01-01T14:15:16Z"
				standard = validator.NewStandardStringAsTime(standardContext, "ghost", &value, "2006-01-02T15:04:05Z07:00")
			})

			It("exists", func() {
				Expect(standard).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standard.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("After", func() {
				BeforeEach(func() {
					result = standard.After(time.Unix(1451567655, 0).UTC())
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("AfterNow", func() {
				BeforeEach(func() {
					result = standard.AfterNow()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("Before", func() {
				BeforeEach(func() {
					result = standard.Before(time.Unix(1451567655, 0).UTC())
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("time-not-before"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not before the specified time"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"2090-01-01T14:15:16Z\" is not before \"2015-12-31T13:14:15Z\""))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghost"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})

			Context("BeforeNow", func() {
				BeforeEach(func() {
					result = standard.BeforeNow()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("time-not-before"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not before the specified time"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"2090-01-01T14:15:16Z\" is not before now"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghost"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})
		})

		Context("new validator with valid reference and a variety of values", func() {
			It("exists using YYYY-MM-DD layout", func() {
				value := "1990-01-01"
				validator.NewStandardStringAsTime(standardContext, "ghost", &value, "2006-01-02").Exists()
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("exists using time.RFC3339Nano layout", func() {
				value := "1990-12-31T11:12:13.1234567-08:00"
				validator.NewStandardStringAsTime(standardContext, "ghost", &value, time.RFC3339Nano).Exists()
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("exists using time.Kitchen layout", func() {
				value := "3:51PM"
				validator.NewStandardStringAsTime(standardContext, "ghost", &value, time.Kitchen).Exists()
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})
	})
})
