package water_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/data/types/water"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "water",
	}
}

func NewWater() *water.Water {
	datum := water.New()
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "water"
	datum.Amount = NewAmount()
	return datum
}

func CloneWater(datum *water.Water) *water.Water {
	if datum == nil {
		return nil
	}
	clone := water.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.Amount = CloneAmount(datum.Amount)
	return clone
}

var _ = Describe("Water", func() {
	It("Type is expected", func() {
		Expect(water.Type).To(Equal("water"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := water.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("water"))
			Expect(datum.Amount).To(BeNil())
		})
	})

	Context("Water", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *water.Water), expectedErrors ...error) {
					datum := NewWater()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *water.Water) {},
				),
				Entry("type missing",
					func(datum *water.Water) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *water.Water) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "water"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type water",
					func(datum *water.Water) { datum.Type = "water" },
				),
				Entry("amount missing",
					func(datum *water.Water) { datum.Amount = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/amount", NewMeta()),
				),
				Entry("amount invalid",
					func(datum *water.Water) { datum.Amount.Units = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/amount/units", NewMeta()),
				),
				Entry("amount valid",
					func(datum *water.Water) { datum.Amount = NewAmount() },
				),
				Entry("multiple errors",
					func(datum *water.Water) {
						datum.Type = "invalidType"
						datum.Amount = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "water"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/amount", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *water.Water)) {
					for _, origin := range structure.Origins() {
						datum := NewWater()
						mutator(datum)
						expectedDatum := CloneWater(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *water.Water) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *water.Water) { datum.Amount = nil },
				),
			)
		})
	})
})
