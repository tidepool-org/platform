package insulin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/insulin"
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
	datum.ActingType = pointer.String(test.RandomStringFromArray(insulin.ActingTypes()))
	datum.Brand = pointer.String(test.NewText(1, 100))
	datum.Dose = NewDose()
	datum.Name = pointer.String(test.NewText(1, 100))
	datum.Site = pointer.String(test.NewText(1, 100))
	return datum
}

func CloneInsulin(datum *insulin.Insulin) *insulin.Insulin {
	if datum == nil {
		return nil
	}
	clone := insulin.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.ActingType = test.CloneString(datum.ActingType)
	clone.Brand = test.CloneString(datum.Brand)
	clone.Dose = CloneDose(datum.Dose)
	clone.Name = test.CloneString(datum.Name)
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
			Expect(datum.ActingType).To(BeNil())
			Expect(datum.Brand).To(BeNil())
			Expect(datum.Dose).To(BeNil())
			Expect(datum.Name).To(BeNil())
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
				Entry("acting type missing",
					func(datum *insulin.Insulin) { datum.ActingType = nil },
				),
				Entry("acting type invalid",
					func(datum *insulin.Insulin) { datum.ActingType = pointer.String("invalid") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"intermediate", "long", "rapid", "short"}), "/actingType", NewMeta()),
				),
				Entry("acting type intermediate",
					func(datum *insulin.Insulin) { datum.ActingType = pointer.String("intermediate") },
				),
				Entry("acting type long",
					func(datum *insulin.Insulin) { datum.ActingType = pointer.String("long") },
				),
				Entry("acting type rapid",
					func(datum *insulin.Insulin) { datum.ActingType = pointer.String("rapid") },
				),
				Entry("acting type short",
					func(datum *insulin.Insulin) { datum.ActingType = pointer.String("short") },
				),
				Entry("brand missing",
					func(datum *insulin.Insulin) { datum.Brand = nil },
				),
				Entry("brand empty",
					func(datum *insulin.Insulin) { datum.Brand = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/brand", NewMeta()),
				),
				Entry("brand invalid",
					func(datum *insulin.Insulin) { datum.Brand = pointer.String(test.NewText(101, 101)) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/brand", NewMeta()),
				),
				Entry("brand valid",
					func(datum *insulin.Insulin) { datum.Brand = pointer.String(test.NewText(1, 100)) },
				),
				Entry("dose missing",
					func(datum *insulin.Insulin) { datum.Dose = nil },
				),
				Entry("dose invalid",
					func(datum *insulin.Insulin) { datum.Dose.Total = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose/total", NewMeta()),
				),
				Entry("dose valid",
					func(datum *insulin.Insulin) { datum.Dose = NewDose() },
				),
				Entry("name missing",
					func(datum *insulin.Insulin) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *insulin.Insulin) { datum.Name = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", NewMeta()),
				),
				Entry("name invalid",
					func(datum *insulin.Insulin) { datum.Name = pointer.String(test.NewText(101, 101)) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name", NewMeta()),
				),
				Entry("name valid",
					func(datum *insulin.Insulin) { datum.Name = pointer.String(test.NewText(1, 100)) },
				),
				Entry("site missing",
					func(datum *insulin.Insulin) { datum.Site = nil },
				),
				Entry("site empty",
					func(datum *insulin.Insulin) { datum.Site = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/site", NewMeta()),
				),
				Entry("site invalid",
					func(datum *insulin.Insulin) { datum.Site = pointer.String(test.NewText(101, 101)) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/site", NewMeta()),
				),
				Entry("site valid",
					func(datum *insulin.Insulin) { datum.Site = pointer.String(test.NewText(1, 100)) },
				),
				Entry("multiple errors",
					func(datum *insulin.Insulin) {
						datum.Type = "invalidType"
						datum.ActingType = pointer.String("invalid")
						datum.Brand = pointer.String("")
						datum.Dose.Total = nil
						datum.Dose.Units = nil
						datum.Name = pointer.String("")
						datum.Site = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "insulin"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"intermediate", "long", "rapid", "short"}), "/actingType", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/brand", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose/total", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose/units", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", &types.Meta{Type: "invalidType"}),
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
				Entry("does not modify the datum; acting type nil",
					func(datum *insulin.Insulin) { datum.ActingType = nil },
				),
				Entry("does not modify the datum; brand nil",
					func(datum *insulin.Insulin) { datum.Brand = nil },
				),
				Entry("does not modify the datum; dose nil",
					func(datum *insulin.Insulin) { datum.Dose = nil },
				),
				Entry("does not modify the datum; name nil",
					func(datum *insulin.Insulin) { datum.Name = nil },
				),
				Entry("does not modify the datum; site nil",
					func(datum *insulin.Insulin) { datum.Site = nil },
				),
			)
		})
	})
})
