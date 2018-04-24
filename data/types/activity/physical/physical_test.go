package physical_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/activity/physical"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "physicalActivity",
	}
}

func NewPhysical() *physical.Physical {
	datum := physical.New()
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "physicalActivity"
	datum.Duration = NewDuration()
	datum.ReportedIntensity = pointer.String(test.RandomStringFromStringArray(physical.ReportedIntensities()))
	return datum
}

func ClonePhysical(datum *physical.Physical) *physical.Physical {
	if datum == nil {
		return nil
	}
	clone := physical.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.Duration = CloneDuration(datum.Duration)
	clone.ReportedIntensity = test.CloneString(datum.ReportedIntensity)
	return clone
}

var _ = Describe("Physical", func() {
	It("ReportedIntensityHigh is expected", func() {
		Expect(physical.ReportedIntensityHigh).To(Equal("high"))
	})

	It("ReportedIntensityLow is expected", func() {
		Expect(physical.ReportedIntensityLow).To(Equal("low"))
	})

	It("ReportedIntensityMedium is expected", func() {
		Expect(physical.ReportedIntensityMedium).To(Equal("medium"))
	})

	It("ReportedIntensities returns expected", func() {
		Expect(physical.ReportedIntensities()).To(Equal([]string{"high", "low", "medium"}))
	})

	Context("Type", func() {
		It("returns the expected type", func() {
			Expect(physical.Type()).To(Equal("physicalActivity"))
		})
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(physical.NewDatum()).To(Equal(&physical.Physical{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(physical.New()).To(Equal(&physical.Physical{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := physical.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("physicalActivity"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.ReportedIntensity).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *physical.Physical

		BeforeEach(func() {
			datum = NewPhysical()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("physicalActivity"))
				Expect(datum.Duration).To(BeNil())
				Expect(datum.ReportedIntensity).To(BeNil())
			})
		})
	})

	Context("Physical", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *physical.Physical), expectedErrors ...error) {
					datum := NewPhysical()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *physical.Physical) {},
				),
				Entry("type missing",
					func(datum *physical.Physical) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *physical.Physical) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "physicalActivity"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type physicalActivity",
					func(datum *physical.Physical) { datum.Type = "physicalActivity" },
				),
				Entry("duration missing",
					func(datum *physical.Physical) { datum.Duration = nil },
				),
				Entry("duration invalid",
					func(datum *physical.Physical) {
						datum.Duration.Units = nil
						datum.Duration.Value = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration/units", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration/value", NewMeta()),
				),
				Entry("duration valid",
					func(datum *physical.Physical) { datum.Duration = NewDuration() },
				),
				Entry("reported intensity missing",
					func(datum *physical.Physical) { datum.ReportedIntensity = nil },
				),
				Entry("reported intensity invalid",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.String("invalid") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"high", "low", "medium"}), "/reportedIntensity", NewMeta()),
				),
				Entry("reported intensity high",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.String("high") },
				),
				Entry("reported intensity low",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.String("low") },
				),
				Entry("reported intensity medium",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.String("medium") },
				),
				Entry("multiple errors",
					func(datum *physical.Physical) {
						datum.Type = "invalidType"
						datum.Duration.Units = nil
						datum.Duration.Value = nil
						datum.ReportedIntensity = pointer.String("invalid")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "physicalActivity"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration/units", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration/value", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"high", "low", "medium"}), "/reportedIntensity", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *physical.Physical)) {
					for _, origin := range structure.Origins() {
						datum := NewPhysical()
						mutator(datum)
						expectedDatum := ClonePhysical(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *physical.Physical) {},
				),
				Entry("does not modify the datum; duration missing",
					func(datum *physical.Physical) { datum.Duration = nil },
				),
				Entry("does not modify the datum; reported intensity missing",
					func(datum *physical.Physical) { datum.ReportedIntensity = nil },
				),
			)
		})
	})
})
