package insulin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/insulin"
	testDataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "insulin",
	}
}

func NewInsulin() *insulin.Insulin {
	datum := insulin.New()
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "insulin"
	datum.Dose = testDataTypesInsulin.NewDose()
	datum.Formulation = testDataTypesInsulin.NewFormulation(3)
	datum.Site = pointer.FromString(test.NewText(1, 100))
	return datum
}

func CloneInsulin(datum *insulin.Insulin) *insulin.Insulin {
	if datum == nil {
		return nil
	}
	clone := insulin.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.Dose = testDataTypesInsulin.CloneDose(datum.Dose)
	clone.Formulation = testDataTypesInsulin.CloneFormulation(datum.Formulation)
	clone.Site = test.CloneString(datum.Site)
	return clone
}

var _ = Describe("Insulin", func() {
	It("Type is expected", func() {
		Expect(insulin.Type).To(Equal("insulin"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := insulin.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("insulin"))
			Expect(datum.Dose).To(BeNil())
			Expect(datum.Formulation).To(BeNil())
			Expect(datum.Site).To(BeNil())
		})
	})

	Context("Insulin", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.Insulin), expectedErrors ...error) {
					datum := NewInsulin()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.Insulin) {},
				),
				Entry("type missing",
					func(datum *insulin.Insulin) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *insulin.Insulin) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "insulin"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type insulin",
					func(datum *insulin.Insulin) { datum.Type = "insulin" },
				),
				Entry("dose missing",
					func(datum *insulin.Insulin) { datum.Dose = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose", NewMeta()),
				),
				Entry("dose invalid",
					func(datum *insulin.Insulin) { datum.Dose.Total = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose/total", NewMeta()),
				),
				Entry("dose valid",
					func(datum *insulin.Insulin) { datum.Dose = testDataTypesInsulin.NewDose() },
				),
				Entry("formulation missing",
					func(datum *insulin.Insulin) { datum.Formulation = nil },
				),
				Entry("formulation invalid",
					func(datum *insulin.Insulin) {
						datum.Formulation.Name = pointer.FromString("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/formulation/name", NewMeta()),
				),
				Entry("formulation valid",
					func(datum *insulin.Insulin) { datum.Formulation = testDataTypesInsulin.NewFormulation(3) },
				),
				Entry("site missing",
					func(datum *insulin.Insulin) { datum.Site = nil },
				),
				Entry("site empty",
					func(datum *insulin.Insulin) { datum.Site = pointer.FromString("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/site", NewMeta()),
				),
				Entry("site invalid",
					func(datum *insulin.Insulin) { datum.Site = pointer.FromString(test.NewText(101, 101)) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/site", NewMeta()),
				),
				Entry("site valid",
					func(datum *insulin.Insulin) { datum.Site = pointer.FromString(test.NewText(1, 100)) },
				),
				Entry("multiple errors",
					func(datum *insulin.Insulin) {
						datum.Type = "invalidType"
						datum.Dose = nil
						datum.Formulation.Name = pointer.FromString("")
						datum.Site = pointer.FromString("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "insulin"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/formulation/name", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/site", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *insulin.Insulin)) {
					for _, origin := range structure.Origins() {
						datum := NewInsulin()
						mutator(datum)
						expectedDatum := CloneInsulin(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.Insulin) {},
				),
				Entry("does not modify the datum; dose nil",
					func(datum *insulin.Insulin) { datum.Dose = nil },
				),
				Entry("does not modify the datum; formulation nil",
					func(datum *insulin.Insulin) { datum.Formulation = nil },
				),
				Entry("does not modify the datum; site nil",
					func(datum *insulin.Insulin) { datum.Site = nil },
				),
			)
		})
	})
})
