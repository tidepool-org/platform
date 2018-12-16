package insulin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/insulin"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Formulation", func() {
	It("FormulationNameLengthMaximum is expected", func() {
		Expect(insulin.FormulationNameLengthMaximum).To(Equal(100))
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
					datum := dataTypesInsulinTest.NewFormulation(3)
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.Formulation) {},
				),
				Entry("compounds, name, and simple missing",
					func(datum *insulin.Formulation) {
						datum.Compounds = nil
						datum.Name = nil
						datum.Simple = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/simple"),
				),
				Entry("compounds missing; simple missing",
					func(datum *insulin.Formulation) {
						datum.Compounds = nil
						datum.Simple = nil
					},
				),
				Entry("compounds missing; simple invalid",
					func(datum *insulin.Formulation) {
						datum.Compounds = nil
						datum.Simple = dataTypesInsulinTest.NewSimple()
						datum.Simple.ActingType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/simple/actingType"),
				),
				Entry("compounds missing; simple valid",
					func(datum *insulin.Formulation) {
						datum.Compounds = nil
						datum.Simple = dataTypesInsulinTest.NewSimple()
					},
				),
				Entry("compounds invalid; simple missing",
					func(datum *insulin.Formulation) {
						datum.Compounds = insulin.NewCompoundArray()
						*datum.Compounds = append(*datum.Compounds, nil)
						datum.Simple = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/compounds/0"),
				),
				Entry("compounds invalid; simple invalid",
					func(datum *insulin.Formulation) {
						datum.Compounds = insulin.NewCompoundArray()
						*datum.Compounds = append(*datum.Compounds, nil)
						datum.Simple = dataTypesInsulinTest.NewSimple()
						datum.Simple.ActingType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/compounds"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/simple/actingType"),
				),
				Entry("compounds invalid; simple valid",
					func(datum *insulin.Formulation) {
						datum.Compounds = insulin.NewCompoundArray()
						*datum.Compounds = append(*datum.Compounds, nil)
						datum.Simple = dataTypesInsulinTest.NewSimple()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/compounds"),
				),
				Entry("compounds valid; simple missing",
					func(datum *insulin.Formulation) {
						datum.Compounds = dataTypesInsulinTest.NewCompoundArray(3)
						datum.Simple = nil
					},
				),
				Entry("compounds valid; simple invalid",
					func(datum *insulin.Formulation) {
						datum.Compounds = dataTypesInsulinTest.NewCompoundArray(3)
						datum.Simple = dataTypesInsulinTest.NewSimple()
						datum.Simple.ActingType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/compounds"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/simple/actingType"),
				),
				Entry("compounds valid; simple valid",
					func(datum *insulin.Formulation) {
						datum.Compounds = dataTypesInsulinTest.NewCompoundArray(3)
						datum.Simple = dataTypesInsulinTest.NewSimple()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/compounds"),
				),
				Entry("name missing",
					func(datum *insulin.Formulation) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *insulin.Formulation) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name invalid",
					func(datum *insulin.Formulation) { datum.Name = pointer.FromString(test.NewText(101, 101)) },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("name valid",
					func(datum *insulin.Formulation) { datum.Name = pointer.FromString(test.NewText(1, 100)) },
				),
				Entry("multiple errors",
					func(datum *insulin.Formulation) {
						datum.Compounds = dataTypesInsulinTest.NewCompoundArray(3)
						datum.Name = pointer.FromString("")
						datum.Simple = dataTypesInsulinTest.NewSimple()
						datum.Simple.ActingType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/compounds"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/simple/actingType"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *insulin.Formulation)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesInsulinTest.NewFormulation(3)
						mutator(datum)
						expectedDatum := dataTypesInsulinTest.CloneFormulation(datum)
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
				Entry("does not modify the datum; compounds missing",
					func(datum *insulin.Formulation) { datum.Compounds = nil },
				),
				Entry("does not modify the datum; name missing",
					func(datum *insulin.Formulation) { datum.Name = nil },
				),
				Entry("does not modify the datum; simple missing",
					func(datum *insulin.Formulation) { datum.Simple = nil },
				),
			)
		})
	})
})
