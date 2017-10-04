package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"regexp"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("String", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New()
	})

	Context("NewString", func() {
		It("returns successfully", func() {
			value := "whatever"
			Expect(structureValidator.NewString(base, &value)).ToNot(BeNil())
		})
	})

	Context("with new validator with nil value", func() {
		var validator *structureValidator.String
		var result structure.String

		BeforeEach(func() {
			validator = structureValidator.NewString(base, nil)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotExists())))
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EqualTo", func() {
			BeforeEach(func() {
				result = validator.EqualTo("1")
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEqualTo", func() {
			BeforeEach(func() {
				result = validator.NotEqualTo("four")
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("OneOf", func() {
			BeforeEach(func() {
				result = validator.OneOf("1", "seven")
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotOneOf", func() {
			BeforeEach(func() {
				result = validator.NotOneOf("seven", "four")
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Matches", func() {
			BeforeEach(func() {
				result = validator.Matches(regexp.MustCompile(".*"))
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotMatches", func() {
			BeforeEach(func() {
				result = validator.NotMatches(regexp.MustCompile(".*"))
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with empty string value", func() {
		var validator *structureValidator.String
		var result structure.String
		var value string

		BeforeEach(func() {
			value = ""
			validator = structureValidator.NewString(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueExists())))
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueEmpty())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with value of 1", func() {
		var validator *structureValidator.String
		var result structure.String
		var value string

		BeforeEach(func() {
			value = "1"
			validator = structureValidator.NewString(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueExists())))
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotEmpty())))
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
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EqualTo", func() {
			BeforeEach(func() {
				result = validator.EqualTo("1")
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEqualTo", func() {
			BeforeEach(func() {
				result = validator.NotEqualTo("four")
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotGreaterThan(1, 3))))
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(1, 4))))
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
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("OneOf", func() {
			BeforeEach(func() {
				result = validator.OneOf("1", "seven")
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotOneOf", func() {
			BeforeEach(func() {
				result = validator.NotOneOf("seven", "four")
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Matches", func() {
			BeforeEach(func() {
				result = validator.Matches(regexp.MustCompile("^[0-9]$"))
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotMatches", func() {
			BeforeEach(func() {
				result = validator.NotMatches(regexp.MustCompile("^[a-z]$"))
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with value with length of four", func() {
		var validator *structureValidator.String
		var result structure.String
		var value string

		BeforeEach(func() {
			value = "four"
			validator = structureValidator.NewString(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueExists())))
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotEmpty())))
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
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EqualTo", func() {
			BeforeEach(func() {
				result = validator.EqualTo("1")
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotEqualTo("four", "1"))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEqualTo", func() {
			BeforeEach(func() {
				result = validator.NotEqualTo("four")
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueEqualTo("four", "four"))))
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotEqualTo(4, 1))))
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthEqualTo(4, 4))))
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotLessThan(4, 3))))
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotLessThanOrEqualTo(4, 1))))
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).To(BeNil())
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
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotInRange(4, 0, 3))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("OneOf", func() {
			BeforeEach(func() {
				result = validator.OneOf("1", "seven")
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueStringNotOneOf("four", []string{"1", "seven"}))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotOneOf", func() {
			BeforeEach(func() {
				result = validator.NotOneOf("seven", "four")
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueStringOneOf("four", []string{"seven", "four"}))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("OneOf with no allowed values", func() {
			BeforeEach(func() {
				result = validator.OneOf()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueStringNotOneOf("four", []string{}))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotOneOf with no disallowed values", func() {
			BeforeEach(func() {
				result = validator.NotOneOf()
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(BeNil())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Matches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^.no.$")
				result = validator.Matches(expression)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueStringNotMatches("four", expression))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^.ou.$")
				result = validator.NotMatches(expression)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueStringMatches("four", expression))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Matches with no expression", func() {
			BeforeEach(func() {
				result = validator.Matches(nil)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueStringNotMatches("four", nil))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotMatches with no expression", func() {
			BeforeEach(func() {
				result = validator.NotMatches(nil)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(BeNil())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueStringMatches("four", nil))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})
})
