package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Standard", func() {
	It("New returns an error if context is nil", func() {
		standard, err := validator.NewStandard(nil)
		Expect(standard).To(BeNil())
		Expect(err).To(HaveOccurred())
	})

	Context("new validator", func() {
		var standardContext *context.Standard
		var standard *validator.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(test.NewLogger())
			Expect(standardContext).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
			standard, err = validator.NewStandard(standardContext)
			Expect(err).ToNot(HaveOccurred())
		})

		It("exists", func() {
			Expect(standard).ToNot(BeNil())
		})

		It("Logger returns a logger", func() {
			Expect(standard.Logger()).ToNot(BeNil())
		})

		Context("SetMeta", func() {
			It("sets the meta on the context", func() {
				meta := "metametameta"
				standard.SetMeta(meta)
				Expect(standardContext.Meta()).To(BeIdenticalTo(meta))
			})
		})

		Context("AppendError", func() {
			It("appends an error on the context", func() {
				standard.AppendError("append-error", &service.Error{})
				Expect(standardContext.Errors()).To(HaveLen(1))
			})
		})

		Context("ValidateBoolean", func() {
			value := true

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateBoolean(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateBoolean("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateBoolean("reference", &value)).ToNot(BeNil())
			})
		})

		Context("ValidateInteger", func() {
			value := 1

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateInteger(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateInteger("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateInteger("reference", &value)).ToNot(BeNil())
			})
		})

		Context("ValidateFloat", func() {
			value := 1.0

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateFloat(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateFloat("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateFloat("reference", &value)).ToNot(BeNil())
			})
		})

		Context("ValidateString", func() {
			value := "string"

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateString(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateString("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateString("reference", &value)).ToNot(BeNil())
			})
		})

		Context("ValidateStringArray", func() {
			value := []string{"one", "two"}

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateStringArray(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateStringArray("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateStringArray("reference", &value)).ToNot(BeNil())
			})
		})

		Context("ValidateObject", func() {
			value := map[string]interface{}{"one": 1, "two": 2}

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateObject(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateObject("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateObject("reference", &value)).ToNot(BeNil())
			})
		})

		Context("ValidateObjectArray", func() {
			value := []map[string]interface{}{{"one": 1, "two": 2}, {"three": 3, "four": 4}}

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateObjectArray(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateObjectArray("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateObjectArray("reference", &value)).ToNot(BeNil())
			})
		})

		Context("ValidateInterface", func() {
			var value interface{} = "zero"

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateInterface(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateInterface("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateInterface("reference", &value)).ToNot(BeNil())
			})
		})

		Context("ValidateInterfaceArray", func() {
			value := []interface{}{"zero", "one"}

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateInterfaceArray(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateInterfaceArray("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateInterfaceArray("reference", &value)).ToNot(BeNil())
			})
		})

		Context("ValidateStringAsTime", func() {
			value := "time"

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateStringAsTime(nil, &value, "2006-01-02T15:04:05Z07:00")).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateStringAsTime("reference", nil, "2006-01-02T15:04:05Z07:00")).ToNot(BeNil())
			})

			It("returns nil when called with empty time layout", func() {
				Expect(standard.ValidateStringAsTime("reference", &value, "")).To(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateStringAsTime("reference", &value, "2006-01-02T15:04:05Z07:00")).ToNot(BeNil())
			})
		})

		Context("ValidateStringAsBloodGlucoseUnits", func() {
			value := "mmol/L"

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateStringAsBloodGlucoseUnits(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateStringAsBloodGlucoseUnits("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateStringAsBloodGlucoseUnits("reference", &value)).ToNot(BeNil())
			})
		})

		Context("ValidateFloatAsBloodGlucoseValue", func() {
			value := 12.345

			It("returns a validator when called with nil reference", func() {
				Expect(standard.ValidateFloatAsBloodGlucoseValue(nil, &value)).ToNot(BeNil())
			})

			It("returns a validator when called with nil value", func() {
				Expect(standard.ValidateFloatAsBloodGlucoseValue("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference and value", func() {
				Expect(standard.ValidateFloatAsBloodGlucoseValue("reference", &value)).ToNot(BeNil())
			})
		})

		Context("NewChildValidator", func() {
			It("returns a validator when called with nil reference", func() {
				Expect(standard.NewChildValidator(nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil reference", func() {
				Expect(standard.NewChildValidator("reference")).ToNot(BeNil())
			})

			It("Logger returns a logger", func() {
				Expect(standard.NewChildValidator("reference").Logger()).ToNot(BeNil())
			})
		})
	})
})
