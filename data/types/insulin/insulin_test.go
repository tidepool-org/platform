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
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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
	datum.Dose = NewDose()
	return datum
}

func CloneInsulin(datum *insulin.Insulin) *insulin.Insulin {
	if datum == nil {
		return nil
	}
	clone := insulin.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.Dose = CloneDose(datum.Dose)
	return clone
}

var _ = Describe("Insulin", func() {
	It("Type is expected", func() {
		Expect(insulin.Type).To(Equal("insulin"))
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(insulin.NewDatum()).To(Equal(&insulin.Insulin{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(insulin.New()).To(Equal(&insulin.Insulin{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := insulin.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("insulin"))
			Expect(datum.Dose).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *insulin.Insulin

		BeforeEach(func() {
			datum = insulin.New()
			Expect(datum).ToNot(BeNil())
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("insulin"))
				Expect(datum.Dose).To(BeNil())
			})
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
				),
				Entry("dose invalid",
					func(datum *insulin.Insulin) { datum.Dose.Total = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose/total", NewMeta()),
				),
				Entry("dose valid",
					func(datum *insulin.Insulin) { datum.Dose = NewDose() },
				),
				Entry("multiple errors",
					func(datum *insulin.Insulin) {
						datum.Type = "invalidType"
						datum.Dose.Total = nil
						datum.Dose.Units = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "insulin"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose/total", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose/units", &types.Meta{Type: "invalidType"}),
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
			)
		})
	})
})
