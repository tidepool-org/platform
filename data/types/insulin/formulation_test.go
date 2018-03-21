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
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Formulation", func() {
	It("FormulationActingTypeIntermediate is expected", func() {
		Expect(insulin.FormulationActingTypeIntermediate).To(Equal("intermediate"))
	})

	It("FormulationActingTypeLong is expected", func() {
		Expect(insulin.FormulationActingTypeLong).To(Equal("long"))
	})

	It("FormulationActingTypeRapid is expected", func() {
		Expect(insulin.FormulationActingTypeRapid).To(Equal("rapid"))
	})

	It("FormulationActingTypeShort is expected", func() {
		Expect(insulin.FormulationActingTypeShort).To(Equal("short"))
	})

	It("FormulationBrandLengthMaximum is expected", func() {
		Expect(insulin.FormulationBrandLengthMaximum).To(Equal(100))
	})

	It("FormulationNameLengthMaximum is expected", func() {
		Expect(insulin.FormulationNameLengthMaximum).To(Equal(100))
	})

	It("FormulationActingTypes returns expected", func() {
		Expect(insulin.FormulationActingTypes()).To(Equal([]string{"intermediate", "long", "rapid", "short"}))
	})

	Context("ParseFormulation", func() {
		// TODO
	})

	Context("NewFormulation", func() {
		It("is successful", func() {
			Expect(insulin.NewFormulation()).To(Equal(&insulin.Formulation{}))
		})
	})

	Context("Formulation", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.Formulation), expectedErrors ...error) {
					datum := testDataTypesInsulin.NewFormulation()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.Formulation) {},
				),
				Entry("acting type missing",
					func(datum *insulin.Formulation) { datum.ActingType = nil },
				),
				Entry("acting type invalid",
					func(datum *insulin.Formulation) { datum.ActingType = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"intermediate", "long", "rapid", "short"}), "/actingType"),
				),
				Entry("acting type intermediate",
					func(datum *insulin.Formulation) { datum.ActingType = pointer.String("intermediate") },
				),
				Entry("acting type long",
					func(datum *insulin.Formulation) { datum.ActingType = pointer.String("long") },
				),
				Entry("acting type rapid",
					func(datum *insulin.Formulation) { datum.ActingType = pointer.String("rapid") },
				),
				Entry("acting type short",
					func(datum *insulin.Formulation) { datum.ActingType = pointer.String("short") },
				),
				Entry("brand missing",
					func(datum *insulin.Formulation) { datum.Brand = nil },
				),
				Entry("brand empty",
					func(datum *insulin.Formulation) { datum.Brand = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/brand"),
				),
				Entry("brand invalid",
					func(datum *insulin.Formulation) { datum.Brand = pointer.String(test.NewText(101, 101)) },
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/brand"),
				),
				Entry("brand valid",
					func(datum *insulin.Formulation) { datum.Brand = pointer.String(test.NewText(1, 100)) },
				),
				Entry("concentration missing",
					func(datum *insulin.Formulation) { datum.Concentration = nil },
				),
				Entry("concentration invalid",
					func(datum *insulin.Formulation) { datum.Concentration.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/concentration/units"),
				),
				Entry("concentration valid",
					func(datum *insulin.Formulation) { datum.Concentration = testDataTypesInsulin.NewConcentration() },
				),
				Entry("name missing",
					func(datum *insulin.Formulation) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *insulin.Formulation) { datum.Name = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name invalid",
					func(datum *insulin.Formulation) { datum.Name = pointer.String(test.NewText(101, 101)) },
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("name valid",
					func(datum *insulin.Formulation) { datum.Name = pointer.String(test.NewText(1, 100)) },
				),
				Entry("multiple errors",
					func(datum *insulin.Formulation) {
						datum.ActingType = pointer.String("invalid")
						datum.Brand = pointer.String("")
						datum.Concentration.Units = nil
						datum.Name = pointer.String("")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"intermediate", "long", "rapid", "short"}), "/actingType"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/brand"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/concentration/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *insulin.Formulation)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesInsulin.NewFormulation()
						mutator(datum)
						expectedDatum := testDataTypesInsulin.CloneFormulation(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.Formulation) {},
				),
				Entry("does not modify the datum; acting type nil",
					func(datum *insulin.Formulation) { datum.ActingType = nil },
				),
				Entry("does not modify the datum; brand nil",
					func(datum *insulin.Formulation) { datum.Brand = nil },
				),
				Entry("does not modify the datum; concentration nil",
					func(datum *insulin.Formulation) { datum.Concentration = nil },
				),
				Entry("does not modify the datum; name nil",
					func(datum *insulin.Formulation) { datum.Name = nil },
				),
			)
		})
	})
})
