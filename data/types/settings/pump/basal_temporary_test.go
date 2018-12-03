package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewBasalTemporary() *pump.BasalTemporary {
	datum := pump.NewBasalTemporary()
	datum.Type = pointer.FromString(test.RandomStringFromArray(pump.BasalTemporaryTypes()))
	return datum
}

func CloneBasalTemporary(datum *pump.BasalTemporary) *pump.BasalTemporary {
	if datum == nil {
		return nil
	}
	clone := pump.NewBasalTemporary()
	clone.Type = test.CloneString(datum.Type)
	return clone
}

var _ = Describe("BasalTemporary", func() {
	It("BasalTemporaryTypeOff is expected", func() {
		Expect(pump.BasalTemporaryTypeOff).To(Equal("off"))
	})

	It("BasalTemporaryTypePercent is expected", func() {
		Expect(pump.BasalTemporaryTypePercent).To(Equal("percent"))
	})

	It("BasalTemporaryTypeUnitsPerHour is expected", func() {
		Expect(pump.BasalTemporaryTypeUnitsPerHour).To(Equal("Units/hour"))
	})

	It("BasalTemporaryTypes returns expected", func() {
		Expect(pump.BasalTemporaryTypes()).To(Equal([]string{"off", "percent", "Units/hour"}))
	})

	Context("ParseBasalTemporary", func() {
		// TODO
	})

	Context("NewBasalTemporary", func() {
		It("is successful", func() {
			Expect(pump.NewBasalTemporary()).To(Equal(&pump.BasalTemporary{}))
		})
	})

	Context("BasalTemporary", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BasalTemporary), expectedErrors ...error) {
					datum := NewBasalTemporary()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalTemporary) {},
				),
				Entry("type missing",
					func(datum *pump.BasalTemporary) { datum.Type = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *pump.BasalTemporary) { datum.Type = pointer.FromString("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"off", "percent", "Units/hour"}), "/type"),
				),
				Entry("type off",
					func(datum *pump.BasalTemporary) { datum.Type = pointer.FromString("off") },
				),
				Entry("type percent",
					func(datum *pump.BasalTemporary) { datum.Type = pointer.FromString("percent") },
				),
				Entry("type Units/hour",
					func(datum *pump.BasalTemporary) { datum.Type = pointer.FromString("Units/hour") },
				),
				Entry("multiple errors",
					func(datum *pump.BasalTemporary) {
						datum.Type = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalTemporary)) {
					for _, origin := range structure.Origins() {
						datum := NewBasalTemporary()
						mutator(datum)
						expectedDatum := CloneBasalTemporary(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BasalTemporary) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *pump.BasalTemporary) { datum.Type = nil },
				),
			)
		})
	})
})
