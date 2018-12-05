package validator_test

import (
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("StringArray", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New().WithSource(structure.NewPointerSource())
	})

	Context("NewStringArray", func() {
		It("returns successfully", func() {
			value := []string{"one", "two"}
			Expect(structureValidator.NewStringArray(base, &value)).ToNot(BeNil())
		})
	})

	Context("with new validator with nil value", func() {
		var validator *structureValidator.StringArray
		var result structure.StringArray

		BeforeEach(func() {
			validator = structureValidator.NewStringArray(base, nil)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueNotExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthEqualTo(1)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthNotEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthNotEqualTo(4)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThan", func() {
			BeforeEach(func() {
				result = validator.LengthLessThan(3)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthLessThan(1)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThan", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThan(3)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThanOrEqualTo(4)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthInRange", func() {
			BeforeEach(func() {
				result = validator.LengthInRange(0, 1)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Each", func() {
			var invocations int

			BeforeEach(func() {
				invocations = 0
				result = validator.Each(func(stringValidator structure.String) {
					Expect(stringValidator.NotEmpty()).To(BeIdenticalTo(stringValidator))
					invocations++
				})
			})

			It("has the expected invocations", func() {
				Expect(invocations).To(Equal(0))
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotEmpty", func() {
			BeforeEach(func() {
				result = validator.EachNotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachOneOf", func() {
			BeforeEach(func() {
				result = validator.EachOneOf("1", "seven")
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotOneOf", func() {
			BeforeEach(func() {
				result = validator.EachNotOneOf("seven", "four")
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^[0-9]*$")
				result = validator.EachMatches(expression)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^.ou.$")
				result = validator.EachNotMatches(expression)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachUsing", func() {
			var values []string

			BeforeEach(func() {
				values = []string{}
				result = validator.EachUsing(func(v string, errorReporter structure.ErrorReporter) {
					values = append(values, v)
					errorReporter.ReportError(errors.New(v))
				})
			})

			It("has the expected values", func() {
				Expect(values).To(BeEmpty())
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachUnique", func() {
			BeforeEach(func() {
				result = validator.EachUnique()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using", func() {
			BeforeEach(func() {
				result = validator.Using(func(value []string, errorReporter structure.ErrorReporter) {
					errorReporter.ReportError(structureValidator.ErrorValueExists())
				})
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with empty string array value", func() {
		var validator *structureValidator.StringArray
		var result structure.StringArray
		var value []string

		BeforeEach(func() {
			value = []string{}
			validator = structureValidator.NewStringArray(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using", func() {
			BeforeEach(func() {
				result = validator.Using(func(value []string, errorReporter structure.ErrorReporter) {
					Expect(value).To(Equal(value))
					errorReporter.ReportError(structureValidator.ErrorValueExists())
				})
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using (without func)", func() {
			BeforeEach(func() {
				result = validator.Using(nil)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with value with length of 1", func() {
		var validator *structureValidator.StringArray
		var result structure.StringArray
		var value []string

		BeforeEach(func() {
			value = []string{"1"}
			validator = structureValidator.NewStringArray(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueNotEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthEqualTo(1)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthNotEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthNotEqualTo(4)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThan", func() {
			BeforeEach(func() {
				result = validator.LengthLessThan(3)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthLessThanOrEqualTo(1)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThan", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThan(3)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorLengthNotGreaterThan(1, 3))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThanOrEqualTo(4)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorLengthNotGreaterThanOrEqualTo(1, 4))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthInRange", func() {
			BeforeEach(func() {
				result = validator.LengthInRange(0, 3)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Each", func() {
			var invocations int

			BeforeEach(func() {
				invocations = 0
				result = validator.Each(func(stringValidator structure.String) {
					Expect(stringValidator.NotEmpty()).To(BeIdenticalTo(stringValidator))
					invocations++
				})
			})

			It("has the expected invocations", func() {
				Expect(invocations).To(Equal(1))
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotEmpty", func() {
			BeforeEach(func() {
				result = validator.EachNotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachOneOf", func() {
			BeforeEach(func() {
				result = validator.EachOneOf("1", "seven")
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotOneOf", func() {
			BeforeEach(func() {
				result = validator.EachNotOneOf("seven", "four")
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^[0-9]*$")
				result = validator.EachMatches(expression)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^.ou.$")
				result = validator.EachNotMatches(expression)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachUsing", func() {
			var values []string

			BeforeEach(func() {
				values = []string{}
				result = validator.EachUsing(func(v string, errorReporter structure.ErrorReporter) {
					values = append(values, v)
					errorReporter.ReportError(errors.New(v))
				})
			})

			It("has the expected values", func() {
				Expect(values).To(Equal(value))
			})

			It("has the expected errors", func() {
				testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(errors.New("1"), "/0"))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachUnique", func() {
			BeforeEach(func() {
				result = validator.EachUnique()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using", func() {
			BeforeEach(func() {
				result = validator.Using(func(value []string, errorReporter structure.ErrorReporter) {
					Expect(value).To(Equal(value))
					errorReporter.ReportError(structureValidator.ErrorValueExists())
				})
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using (without func)", func() {
			BeforeEach(func() {
				result = validator.Using(nil)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with value with length of 4", func() {
		var validator *structureValidator.StringArray
		var result structure.StringArray
		var value []string

		BeforeEach(func() {
			value = []string{"1", "two", "", "four"}
			validator = structureValidator.NewStringArray(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueNotEmpty())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthEqualTo(1)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorLengthNotEqualTo(4, 1))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthNotEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthNotEqualTo(4)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorLengthEqualTo(4, 4))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThan", func() {
			BeforeEach(func() {
				result = validator.LengthLessThan(3)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorLengthNotLessThan(4, 3))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthLessThanOrEqualTo(1)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorLengthNotLessThanOrEqualTo(4, 1))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThan", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThan(3)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThanOrEqualTo(4)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthInRange", func() {
			BeforeEach(func() {
				result = validator.LengthInRange(0, 3)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorLengthNotInRange(4, 0, 3))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Each", func() {
			var invocations int

			BeforeEach(func() {
				invocations = 0
				result = validator.Each(func(stringValidator structure.String) {
					Expect(stringValidator.NotEmpty()).To(BeIdenticalTo(stringValidator))
					invocations++
				})
			})

			It("has the expected invocations", func() {
				Expect(invocations).To(Equal(4))
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/2"))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotEmpty", func() {
			BeforeEach(func() {
				result = validator.EachNotEmpty()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/2"))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachOneOf", func() {
			BeforeEach(func() {
				result = validator.EachOneOf("1", "seven")
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), errors.Append(
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("two", []string{"1", "seven"}), "/1"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", []string{"1", "seven"}), "/2"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("four", []string{"1", "seven"}), "/3"),
				))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotOneOf", func() {
			BeforeEach(func() {
				result = validator.EachNotOneOf("seven", "four")
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureValidator.ErrorValueStringOneOf("four", []string{"seven", "four"}), "/3"))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachOneOf with no allowed values", func() {
			BeforeEach(func() {
				result = validator.EachOneOf()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), errors.Append(
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("1", []string{}), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("two", []string{}), "/1"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", []string{}), "/2"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("four", []string{}), "/3"),
				))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotOneOf with no disallowed values", func() {
			BeforeEach(func() {
				result = validator.EachNotOneOf()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^[0-9]+$")
				result = validator.EachMatches(expression)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), errors.Append(
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotMatches("two", expression), "/1"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotMatches("", expression), "/2"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotMatches("four", expression), "/3"),
				))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^.ou.$")
				result = validator.EachNotMatches(expression)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureValidator.ErrorValueStringMatches("four", expression), "/3"))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachMatches with no expression", func() {
			BeforeEach(func() {
				result = validator.EachMatches(nil)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), errors.Append(
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotMatches("1", nil), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotMatches("two", nil), "/1"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotMatches("", nil), "/2"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotMatches("four", nil), "/3"),
				))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotMatches with no expression", func() {
			BeforeEach(func() {
				result = validator.EachNotMatches(nil)
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), errors.Append(
					testErrors.WithPointerSource(structureValidator.ErrorValueStringMatches("1", nil), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringMatches("two", nil), "/1"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringMatches("", nil), "/2"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringMatches("four", nil), "/3"),
				))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachUsing", func() {
			var values []string

			BeforeEach(func() {
				values = []string{}
				result = validator.EachUsing(func(v string, errorReporter structure.ErrorReporter) {
					values = append(values, v)
					errorReporter.ReportError(errors.New(v))
				})
			})

			It("has the expected values", func() {
				Expect(values).To(Equal(value))
			})

			It("has the expected errors", func() {
				testErrors.ExpectEqual(base.Error(),
					testErrors.WithPointerSource(errors.New("1"), "/0"),
					testErrors.WithPointerSource(errors.New("two"), "/1"),
					testErrors.WithPointerSource(errors.New(""), "/2"),
					testErrors.WithPointerSource(errors.New("four"), "/3"),
				)
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachUnique", func() {
			BeforeEach(func() {
				result = validator.EachUnique()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using", func() {
			BeforeEach(func() {
				result = validator.Using(func(value []string, errorReporter structure.ErrorReporter) {
					Expect(value).To(Equal(value))
					errorReporter.ReportError(structureValidator.ErrorValueExists())
				})
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(), structureValidator.ErrorValueExists())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Using (without func)", func() {
			BeforeEach(func() {
				result = validator.Using(nil)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with duplicates", func() {
		var validator *structureValidator.StringArray
		var result structure.StringArray
		var value []string

		BeforeEach(func() {
			value = []string{"one", "two", "four", "two", "four", "five", "one"}
			validator = structureValidator.NewStringArray(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("EachUnique", func() {
			BeforeEach(func() {
				result = validator.EachUnique()
			})

			It("reports multiple errors", func() {
				Expect(base.Error()).To(HaveOccurred())
				testErrors.ExpectEqual(base.Error(),
					testErrors.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/3"),
					testErrors.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/4"),
					testErrors.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/6"),
				)
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})
})
