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

var _ = Describe("Simple", func() {
	It("SimpleActingTypeIntermediate is expected", func() {
		Expect(insulin.SimpleActingTypeIntermediate).To(Equal("intermediate"))
	})

	It("SimpleActingTypeLong is expected", func() {
		Expect(insulin.SimpleActingTypeLong).To(Equal("long"))
	})

	It("SimpleActingTypeRapid is expected", func() {
		Expect(insulin.SimpleActingTypeRapid).To(Equal("rapid"))
	})

	It("SimpleActingTypeShort is expected", func() {
		Expect(insulin.SimpleActingTypeShort).To(Equal("short"))
	})

	It("SimpleBrandLengthMaximum is expected", func() {
		Expect(insulin.SimpleBrandLengthMaximum).To(Equal(100))
	})

	It("SimpleActingTypes returns expected", func() {
		Expect(insulin.SimpleActingTypes()).To(Equal([]string{"intermediate", "long", "rapid", "short"}))
	})

	Context("ParseSimple", func() {
		// TODO
	})

	Context("NewSimple", func() {
		It("is successful", func() {
			Expect(insulin.NewSimple()).To(Equal(&insulin.Simple{}))
		})
	})

	Context("Simple", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.Simple), expectedErrors ...error) {
					datum := testDataTypesInsulin.NewSimple()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.Simple) {},
				),
				Entry("acting type missing",
					func(datum *insulin.Simple) { datum.ActingType = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/actingType"),
				),
				Entry("acting type invalid",
					func(datum *insulin.Simple) { datum.ActingType = pointer.FromString("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"intermediate", "long", "rapid", "short"}), "/actingType"),
				),
				Entry("acting type intermediate",
					func(datum *insulin.Simple) { datum.ActingType = pointer.FromString("intermediate") },
				),
				Entry("acting type long",
					func(datum *insulin.Simple) { datum.ActingType = pointer.FromString("long") },
				),
				Entry("acting type rapid",
					func(datum *insulin.Simple) { datum.ActingType = pointer.FromString("rapid") },
				),
				Entry("acting type short",
					func(datum *insulin.Simple) { datum.ActingType = pointer.FromString("short") },
				),
				Entry("brand missing",
					func(datum *insulin.Simple) { datum.Brand = nil },
				),
				Entry("brand empty",
					func(datum *insulin.Simple) { datum.Brand = pointer.FromString("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/brand"),
				),
				Entry("brand invalid",
					func(datum *insulin.Simple) { datum.Brand = pointer.FromString(test.NewText(101, 101)) },
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/brand"),
				),
				Entry("brand valid",
					func(datum *insulin.Simple) { datum.Brand = pointer.FromString(test.NewText(1, 100)) },
				),
				Entry("concentration missing",
					func(datum *insulin.Simple) { datum.Concentration = nil },
				),
				Entry("concentration invalid",
					func(datum *insulin.Simple) { datum.Concentration.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/concentration/units"),
				),
				Entry("concentration valid",
					func(datum *insulin.Simple) { datum.Concentration = testDataTypesInsulin.NewConcentration() },
				),
				Entry("multiple errors",
					func(datum *insulin.Simple) {
						datum.ActingType = nil
						datum.Brand = pointer.FromString("")
						datum.Concentration.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/actingType"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/brand"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/concentration/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *insulin.Simple)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesInsulin.NewSimple()
						mutator(datum)
						expectedDatum := testDataTypesInsulin.CloneSimple(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.Simple) {},
				),
				Entry("does not modify the datum; acting type nil",
					func(datum *insulin.Simple) { datum.ActingType = nil },
				),
				Entry("does not modify the datum; brand nil",
					func(datum *insulin.Simple) { datum.Brand = nil },
				),
				Entry("does not modify the datum; concentration nil",
					func(datum *insulin.Simple) { datum.Concentration = nil },
				),
			)
		})
	})
})
