package validator_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"time"

// 	"github.com/tidepool-org/platform/structure"
// 	testStructure "github.com/tidepool-org/platform/structure/test"
// 	structureValidator "github.com/tidepool-org/platform/structure/validator"
// 	"github.com/tidepool-org/platform/test"
// )

// var _ = Describe("Validator", func() {
// 	var base *testStructure.Base

// 	BeforeEach(func() {
// 		base = testStructure.NewBase()
// 	})

// 	AfterEach(func() {
// 		base.Expectations()
// 	})

// 	Context("New", func() {
// 		It("returns successfully", func() {
// 			Expect(structureValidator.New()).ToNot(BeNil())
// 		})
// 	})

// 	Context("NewValidator", func() {
// 		It("returns successfully", func() {
// 			Expect(structureValidator.NewValidator(base)).ToNot(BeNil())
// 		})
// 	})

// 	Context("with new validator", func() {
// 		var validator *structureValidator.Validator

// 		BeforeEach(func() {
// 			validator = structureValidator.NewValidator(base)
// 			Expect(validator).ToNot(BeNil())
// 		})

// 		Context("with references", func() {
// 			BeforeEach(func() {
// 				base.WithReferenceOutputs = []structure.Base{testStructure.NewBase()}
// 			})

// 			Context("Validating", func() {
// 				It("returns a validator when called with nil value", func() {
// 					Expect(validator.Validating("reference", nil)).ToNot(BeNil())
// 				})

// 				It("returns a validator when called with non-nil value", func() {
// 					value := testStructure.NewValidatable()
// 					Expect(validator.Validating("reference", value)).ToNot(BeNil())
// 				})
// 			})

// 			Context("Bool", func() {
// 				It("returns a validator when called with nil value", func() {
// 					Expect(validator.Bool("reference", nil)).ToNot(BeNil())
// 				})

// 				It("returns a validator when called with non-nil value", func() {
// 					value := true
// 					Expect(validator.Bool("reference", &value)).ToNot(BeNil())
// 				})
// 			})

// 			Context("Float64", func() {
// 				It("returns a validator when called with nil value", func() {
// 					Expect(validator.Float64("reference", nil)).ToNot(BeNil())
// 				})

// 				It("returns a validator when called with non-nil value", func() {
// 					value := 3.45
// 					Expect(validator.Float64("reference", &value)).ToNot(BeNil())
// 				})
// 			})

// 			Context("Int", func() {
// 				It("returns a validator when called with nil value", func() {
// 					Expect(validator.Int("reference", nil)).ToNot(BeNil())
// 				})

// 				It("returns a validator when called with non-nil value", func() {
// 					value := 2
// 					Expect(validator.Int("reference", &value)).ToNot(BeNil())
// 				})
// 			})

// 			Context("String", func() {
// 				It("returns a validator when called with nil value", func() {
// 					Expect(validator.String("reference", nil)).ToNot(BeNil())
// 				})

// 				It("returns a validator when called with non-nil value", func() {
// 					value := "six"
// 					Expect(validator.String("reference", &value)).ToNot(BeNil())
// 				})
// 			})

// 			Context("StringArray", func() {
// 				It("returns a validator when called with nil value", func() {
// 					Expect(validator.StringArray("reference", nil)).ToNot(BeNil())
// 				})

// 				It("returns a validator when called with non-nil value", func() {
// 					value := []string{"seven", "eight"}
// 					Expect(validator.StringArray("reference", &value)).ToNot(BeNil())
// 				})
// 			})

// 			Context("Time", func() {
// 				It("returns a validator when called with nil value", func() {
// 					Expect(validator.Time("reference", nil)).ToNot(BeNil())
// 				})

// 				It("returns a validator when called with non-nil value", func() {
// 					value := time.Now()
// 					Expect(validator.Time("reference", &value)).ToNot(BeNil())
// 				})
// 			})
// 		})

// 		Context("WithSource", func() {
// 			It("returns new validator", func() {
// 				base.WithSourceOutputs = []structure.Base{testStructure.NewBase()}
// 				withSource := testStructure.NewSource()
// 				result := validator.WithSource(withSource)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(validator))
// 				Expect(base.WithSourceInputs).To(Equal([]structure.Source{withSource}))
// 			})
// 		})

// 		Context("WithMeta", func() {
// 			It("returns new validator", func() {
// 				base.WithMetaOutputs = []structure.Base{testStructure.NewBase()}
// 				withMeta := test.NewText(1, 128)
// 				result := validator.WithMeta(withMeta)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(validator))
// 				Expect(base.WithMetaInputs).To(Equal([]interface{}{withMeta}))
// 			})
// 		})

// 		Context("WithReference", func() {
// 			It("returns new validator", func() {
// 				base.WithReferenceOutputs = []structure.Base{testStructure.NewBase()}
// 				withReference := testStructure.NewReference()
// 				result := validator.WithReference(withReference)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(validator))
// 				Expect(base.WithReferenceInputs).To(Equal([]string{withReference}))
// 			})
// 		})
// 	})
// })
