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

var _ = Describe("Mix", func() {
	It("MixElementAmountMinimum is expected", func() {
		Expect(insulin.MixElementAmountMinimum).To(Equal(0.0))
	})

	Context("ParseMixElement", func() {
		// TODO
	})

	Context("NewMixElement", func() {
		It("is successful", func() {
			Expect(insulin.NewMixElement()).To(Equal(&insulin.MixElement{}))
		})
	})

	Context("MixElement", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.MixElement), expectedErrors ...error) {
					datum := testDataTypesInsulin.NewMixElement()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.MixElement) {},
				),
				Entry("amount missing",
					func(datum *insulin.MixElement) { datum.Amount = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount out of range (lower)",
					func(datum *insulin.MixElement) { datum.Amount = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-0.1, 0.0), "/amount"),
				),
				Entry("amount in range (lower)",
					func(datum *insulin.MixElement) { datum.Amount = pointer.Float64(0.0) },
				),
				Entry("formulation missing",
					func(datum *insulin.MixElement) { datum.Formulation = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/formulation"),
				),
				Entry("formulation invalid",
					func(datum *insulin.MixElement) { datum.Formulation.Name = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/formulation/name"),
				),
				Entry("formulation valid",
					func(datum *insulin.MixElement) { datum.Formulation = testDataTypesInsulin.NewFormulation() },
				),
				Entry("multiple errors",
					func(datum *insulin.MixElement) {
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
				func(mutator func(datum *insulin.MixElement)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesInsulin.NewMixElement()
						mutator(datum)
						expectedDatum := testDataTypesInsulin.CloneMixElement(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.MixElement) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *insulin.MixElement) { datum.Amount = nil },
				),
				Entry("does not modify the datum; formulation missing",
					func(datum *insulin.MixElement) { datum.Formulation = nil },
				),
			)
		})
	})

	Context("ParseMix", func() {
		// TODO
	})

	Context("NewMix", func() {
		It("is successful", func() {
			Expect(insulin.NewMix()).To(Equal(&insulin.Mix{}))
		})
	})

	Context("Mix", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.Mix), expectedErrors ...error) {
					datum := insulin.NewMix()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.Mix) {},
					structureValidator.ErrorValueEmpty(),
				),
				Entry("empty",
					func(datum *insulin.Mix) { *datum = *insulin.NewMix() },
					structureValidator.ErrorValueEmpty(),
				),
				Entry("nil",
					func(datum *insulin.Mix) { *datum = append(*datum, nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *insulin.Mix) {
						invalid := testDataTypesInsulin.NewMixElement()
						invalid.Amount = nil
						*datum = append(*datum, invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/amount"),
				),
				Entry("single valid",
					func(datum *insulin.Mix) {
						*datum = append(*datum, testDataTypesInsulin.NewMixElement())
					},
				),
				Entry("multiple invalid",
					func(datum *insulin.Mix) {
						invalid := testDataTypesInsulin.NewMixElement()
						invalid.Amount = nil
						*datum = append(*datum, testDataTypesInsulin.NewMixElement(), invalid, testDataTypesInsulin.NewMixElement())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
				Entry("multiple valid",
					func(datum *insulin.Mix) {
						*datum = append(*datum, testDataTypesInsulin.NewMixElement(), testDataTypesInsulin.NewMixElement(), testDataTypesInsulin.NewMixElement())
					},
				),
				Entry("multiple errors",
					func(datum *insulin.Mix) {
						invalid := testDataTypesInsulin.NewMixElement()
						invalid.Amount = nil
						*datum = append(*datum, nil, invalid, testDataTypesInsulin.NewMixElement())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *insulin.Mix)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesInsulin.NewMix()
						mutator(datum)
						expectedDatum := testDataTypesInsulin.CloneMix(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.Mix) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *insulin.Mix) { (*datum)[0].Amount = nil },
				),
				Entry("does not modify the datum; formulation missing",
					func(datum *insulin.Mix) { (*datum)[0].Formulation = nil },
				),
			)
		})
	})
})
