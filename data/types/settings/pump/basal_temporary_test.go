package pump_test

import (
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
					datum := pumpTest.NewBasalTemporary()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalTemporary) {},
				),
				Entry("type missing",
					func(datum *pump.BasalTemporary) { datum.Type = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *pump.BasalTemporary) { datum.Type = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"off", "percent", "Units/hour"}), "/type"),
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
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalTemporary)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewBasalTemporary()
						mutator(datum)
						expectedDatum := pumpTest.CloneBasalTemporary(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
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
