package insulin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/insulin"
	testDataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Compound", func() {
	It("CompoundAmountMinimum is expected", func() {
		Expect(insulin.CompoundAmountMinimum).To(Equal(0.0))
	})

	Context("ParseCompound", func() {
		// TODO
	})

	Context("NewCompound", func() {
		It("is successful", func() {
			Expect(insulin.NewCompound()).To(Equal(&insulin.Compound{}))
		})
	})

	Context("Compound", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.Compound), expectedErrors ...error) {
					datum := testDataTypesInsulin.NewCompound(3)
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.Compound) {},
				),
				Entry("amount missing",
					func(datum *insulin.Compound) { datum.Amount = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount out of range (lower)",
					func(datum *insulin.Compound) { datum.Amount = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-0.1, 0.0), "/amount"),
				),
				Entry("amount in range (lower)",
					func(datum *insulin.Compound) { datum.Amount = pointer.Float64(0.0) },
				),
				Entry("formulation missing",
					func(datum *insulin.Compound) { datum.Formulation = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/formulation"),
				),
				Entry("formulation invalid",
					func(datum *insulin.Compound) { datum.Formulation.Name = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/formulation/name"),
				),
				Entry("formulation valid",
					func(datum *insulin.Compound) { datum.Formulation = testDataTypesInsulin.NewFormulation(3) },
				),
				Entry("multiple errors",
					func(datum *insulin.Compound) {
						datum.Amount = nil
						datum.Formulation = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/formulation"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *insulin.Compound)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesInsulin.NewCompound(3)
						mutator(datum)
						expectedDatum := testDataTypesInsulin.CloneCompound(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.Compound) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *insulin.Compound) { datum.Amount = nil },
				),
				Entry("does not modify the datum; formulation missing",
					func(datum *insulin.Compound) { datum.Formulation = nil },
				),
			)
		})
	})

	Context("ParseCompoundArray", func() {
		// TODO
	})

	Context("NewCompoundArray", func() {
		It("is successful", func() {
			Expect(insulin.NewCompoundArray()).To(Equal(&insulin.CompoundArray{}))
		})
	})

	Context("CompoundArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.CompoundArray), expectedErrors ...error) {
					datum := insulin.NewCompoundArray()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.CompoundArray) {},
					structureValidator.ErrorValueEmpty(),
				),
				Entry("empty",
					func(datum *insulin.CompoundArray) { *datum = *insulin.NewCompoundArray() },
					structureValidator.ErrorValueEmpty(),
				),
				Entry("nil",
					func(datum *insulin.CompoundArray) { *datum = append(*datum, nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *insulin.CompoundArray) {
						invalid := testDataTypesInsulin.NewCompound(3)
						invalid.Amount = nil
						*datum = append(*datum, invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/amount"),
				),
				Entry("single valid",
					func(datum *insulin.CompoundArray) {
						*datum = append(*datum, testDataTypesInsulin.NewCompound(3))
					},
				),
				Entry("multiple invalid",
					func(datum *insulin.CompoundArray) {
						invalid := testDataTypesInsulin.NewCompound(3)
						invalid.Amount = nil
						*datum = append(*datum, testDataTypesInsulin.NewCompound(3), invalid, testDataTypesInsulin.NewCompound(3))
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
				Entry("multiple valid",
					func(datum *insulin.CompoundArray) {
						*datum = append(*datum, testDataTypesInsulin.NewCompound(3), testDataTypesInsulin.NewCompound(3), testDataTypesInsulin.NewCompound(3))
					},
				),
				Entry("multiple; length in range (upper)",
					func(datum *insulin.CompoundArray) {
						for len(*datum) < 100 {
							*datum = append(*datum, testDataTypesInsulin.NewCompound(1))
						}
					},
				),
				Entry("multiple; length out of range (upper)",
					func(datum *insulin.CompoundArray) {
						for len(*datum) < 101 {
							*datum = append(*datum, testDataTypesInsulin.NewCompound(1))
						}
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100),
				),
				Entry("multiple errors",
					func(datum *insulin.CompoundArray) {
						invalid := testDataTypesInsulin.NewCompound(3)
						invalid.Amount = nil
						*datum = append(*datum, nil, invalid, testDataTypesInsulin.NewCompound(3))
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *insulin.CompoundArray)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesInsulin.NewCompoundArray(3)
						mutator(datum)
						expectedDatum := testDataTypesInsulin.CloneCompoundArray(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.CompoundArray) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *insulin.CompoundArray) { (*datum)[0].Amount = nil },
				),
				Entry("does not modify the datum; formulation missing",
					func(datum *insulin.CompoundArray) { (*datum)[0].Formulation = nil },
				),
			)
		})
	})
})
