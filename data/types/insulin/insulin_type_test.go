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

var _ = Describe("InsulinType", func() {
	Context("ParseInsulinType", func() {
		// TODO
	})

	Context("NewInsulinType", func() {
		It("is successful", func() {
			Expect(insulin.NewInsulinType()).To(Equal(&insulin.InsulinType{}))
		})
	})

	Context("InsulinType", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.InsulinType), expectedErrors ...error) {
					datum := testDataTypesInsulin.NewInsulinType()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.InsulinType) {},
				),
				Entry("formulation missing; mix missing",
					func(datum *insulin.InsulinType) {
						datum.Formulation = nil
						datum.Mix = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/formulation"),
				),
				Entry("formulation missing; mix invalid",
					func(datum *insulin.InsulinType) {
						datum.Formulation = nil
						datum.Mix = testDataTypesInsulin.NewMix()
						(*datum.Mix)[0] = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mix/0"),
				),
				Entry("formulation missing; mix valid",
					func(datum *insulin.InsulinType) {
						datum.Formulation = nil
						datum.Mix = testDataTypesInsulin.NewMix()
					},
				),
				Entry("formulation invalid; mix missing",
					func(datum *insulin.InsulinType) {
						datum.Formulation = testDataTypesInsulin.NewFormulation()
						datum.Formulation.Name = pointer.String("")
						datum.Mix = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/formulation/name"),
				),
				Entry("formulation invalid; mix invalid",
					func(datum *insulin.InsulinType) {
						datum.Formulation = testDataTypesInsulin.NewFormulation()
						datum.Formulation.Name = pointer.String("")
						datum.Mix = testDataTypesInsulin.NewMix()
						(*datum.Mix)[0] = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/formulation/name"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/mix"),
				),
				Entry("formulation invalid; mix valid",
					func(datum *insulin.InsulinType) {
						datum.Formulation = testDataTypesInsulin.NewFormulation()
						datum.Formulation.Name = pointer.String("")
						datum.Mix = testDataTypesInsulin.NewMix()
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/formulation/name"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/mix"),
				),
				Entry("formulation valid; mix missing",
					func(datum *insulin.InsulinType) {
						datum.Formulation = testDataTypesInsulin.NewFormulation()
						datum.Mix = nil
					},
				),
				Entry("formulation valid; mix invalid",
					func(datum *insulin.InsulinType) {
						datum.Formulation = testDataTypesInsulin.NewFormulation()
						datum.Mix = testDataTypesInsulin.NewMix()
						(*datum.Mix)[0] = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/mix"),
				),
				Entry("formulation valid; mix valid",
					func(datum *insulin.InsulinType) {
						datum.Formulation = testDataTypesInsulin.NewFormulation()
						datum.Mix = testDataTypesInsulin.NewMix()
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/mix"),
				),
				Entry("multiple errors",
					func(datum *insulin.InsulinType) {
						datum.Formulation = testDataTypesInsulin.NewFormulation()
						datum.Formulation.Name = pointer.String("")
						datum.Mix = testDataTypesInsulin.NewMix()
						(*datum.Mix)[0] = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/formulation/name"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/mix"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *insulin.InsulinType)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesInsulin.NewInsulinType()
						mutator(datum)
						expectedDatum := testDataTypesInsulin.CloneInsulinType(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.InsulinType) {},
				),
				Entry("does not modify the datum; formulation missing",
					func(datum *insulin.InsulinType) { datum.Formulation = nil },
				),
				Entry("does not modify the datum; mix missing",
					func(datum *insulin.InsulinType) { datum.Mix = nil },
				),
			)
		})
	})
})
