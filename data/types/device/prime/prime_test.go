package prime_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesDevice "github.com/tidepool-org/platform/data/types/device"
	dataTypesDevicePrime "github.com/tidepool-org/platform/data/types/device/prime"
	dataTypesDevicePrimeTest "github.com/tidepool-org/platform/data/types/device/prime/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() any {
	return &dataTypesDevice.Meta{
		Type:    dataTypesDevice.Type,
		SubType: dataTypesDevicePrime.SubType,
	}
}

var _ = Describe("Status", func() {
	It("SubType is expected", func() {
		Expect(dataTypesDevicePrime.SubType).To(Equal("prime"))
	})

	It("TargetCannula is expected", func() {
		Expect(dataTypesDevicePrime.TargetCannula).To(Equal("cannula"))
	})

	It("TargetTubing is expected", func() {
		Expect(dataTypesDevicePrime.TargetTubing).To(Equal("tubing"))
	})

	It("VolumeTargetCannulaMaximum is expected", func() {
		Expect(dataTypesDevicePrime.VolumeTargetCannulaMaximum).To(Equal(1000.0))
	})

	It("VolumeTargetCannulaMinimum is expected", func() {
		Expect(dataTypesDevicePrime.VolumeTargetCannulaMinimum).To(Equal(0.0))
	})

	It("VolumeTargetTubingMaximum is expected", func() {
		Expect(dataTypesDevicePrime.VolumeTargetTubingMaximum).To(Equal(1000.0))
	})

	It("VolumeTargetTubingMinimum is expected", func() {
		Expect(dataTypesDevicePrime.VolumeTargetTubingMinimum).To(Equal(0.0))
	})

	It("Targets returns expected", func() {
		Expect(dataTypesDevicePrime.Targets()).To(Equal([]string{"cannula", "tubing"}))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := dataTypesDevicePrime.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal(dataTypesDevice.Type))
			Expect(datum.SubType).To(Equal(dataTypesDevicePrime.SubType))
			Expect(datum.Target).To(BeNil())
			Expect(datum.Volume).To(BeNil())
		})
	})

	Context("Prime", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesDevicePrime.Prime), expectedErrors ...error) {
					datum := dataTypesDevicePrimeTest.RandomPrime()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDevicePrime.Prime) {},
				),
				Entry("type missing",
					func(datum *dataTypesDevicePrime.Prime) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypesDevice.Meta{SubType: dataTypesDevicePrime.SubType}),
				),
				Entry("type invalid",
					func(datum *dataTypesDevicePrime.Prime) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", dataTypesDevice.Type), "/type", &dataTypesDevice.Meta{Type: "invalidType", SubType: dataTypesDevicePrime.SubType}),
				),
				Entry("type device",
					func(datum *dataTypesDevicePrime.Prime) { datum.Type = dataTypesDevice.Type },
				),
				Entry("sub type missing",
					func(datum *dataTypesDevicePrime.Prime) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &dataTypesDevice.Meta{Type: dataTypesDevice.Type}),
				),
				Entry("sub type invalid",
					func(datum *dataTypesDevicePrime.Prime) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", dataTypesDevicePrime.SubType), "/subType", &dataTypesDevice.Meta{Type: dataTypesDevice.Type, SubType: "invalidSubType"}),
				),
				Entry("sub type prime",
					func(datum *dataTypesDevicePrime.Prime) { datum.SubType = dataTypesDevicePrime.SubType },
				),
				Entry("target missing",
					func(datum *dataTypesDevicePrime.Prime) { datum.Target = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/primeTarget", NewMeta()),
				),
				Entry("target invalid",
					func(datum *dataTypesDevicePrime.Prime) { datum.Target = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", dataTypesDevicePrime.Targets()), "/primeTarget", NewMeta()),
				),
				Entry("target cannula; volume missing",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Target = pointer.FromString(dataTypesDevicePrime.TargetCannula)
						datum.Volume = nil
					},
				),
				Entry("target cannula; volume out of range (lower)",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Target = pointer.FromString(dataTypesDevicePrime.TargetCannula)
						datum.Volume = pointer.FromFloat64(dataTypesDevicePrime.VolumeTargetCannulaMinimum - 0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesDevicePrime.VolumeTargetCannulaMinimum-0.1, dataTypesDevicePrime.VolumeTargetCannulaMinimum, dataTypesDevicePrime.VolumeTargetCannulaMaximum), "/volume", NewMeta()),
				),
				Entry("target cannula; volume in range (lower)",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Target = pointer.FromString(dataTypesDevicePrime.TargetCannula)
						datum.Volume = pointer.FromFloat64(dataTypesDevicePrime.VolumeTargetCannulaMinimum)
					},
				),
				Entry("target cannula; volume in range (upper)",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Target = pointer.FromString(dataTypesDevicePrime.TargetCannula)
						datum.Volume = pointer.FromFloat64(dataTypesDevicePrime.VolumeTargetCannulaMaximum)
					},
				),
				Entry("target cannula; volume out of range (upper)",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Target = pointer.FromString(dataTypesDevicePrime.TargetCannula)
						datum.Volume = pointer.FromFloat64(dataTypesDevicePrime.VolumeTargetCannulaMaximum + 0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesDevicePrime.VolumeTargetCannulaMaximum+0.1, dataTypesDevicePrime.VolumeTargetCannulaMinimum, dataTypesDevicePrime.VolumeTargetCannulaMaximum), "/volume", NewMeta()),
				),
				Entry("target tubing; volume missing",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Target = pointer.FromString(dataTypesDevicePrime.TargetTubing)
						datum.Volume = nil
					},
				),
				Entry("target tubing; volume out of range (lower)",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Target = pointer.FromString(dataTypesDevicePrime.TargetTubing)
						datum.Volume = pointer.FromFloat64(dataTypesDevicePrime.VolumeTargetTubingMinimum - 0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesDevicePrime.VolumeTargetTubingMinimum-0.1, dataTypesDevicePrime.VolumeTargetTubingMinimum, dataTypesDevicePrime.VolumeTargetTubingMaximum), "/volume", NewMeta()),
				),
				Entry("target tubing; volume in range (lower)",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Target = pointer.FromString(dataTypesDevicePrime.TargetTubing)
						datum.Volume = pointer.FromFloat64(dataTypesDevicePrime.VolumeTargetTubingMinimum)
					},
				),
				Entry("target tubing; volume in range (upper)",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Target = pointer.FromString(dataTypesDevicePrime.TargetTubing)
						datum.Volume = pointer.FromFloat64(dataTypesDevicePrime.VolumeTargetTubingMaximum)
					},
				),
				Entry("target tubing; volume out of range (upper)",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Target = pointer.FromString(dataTypesDevicePrime.TargetTubing)
						datum.Volume = pointer.FromFloat64(dataTypesDevicePrime.VolumeTargetTubingMaximum + 0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesDevicePrime.VolumeTargetTubingMaximum+0.1, dataTypesDevicePrime.VolumeTargetTubingMinimum, dataTypesDevicePrime.VolumeTargetTubingMaximum), "/volume", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *dataTypesDevicePrime.Prime) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Target = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", dataTypesDevice.Type), "/type", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", dataTypesDevicePrime.SubType), "/subType", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", dataTypesDevicePrime.Targets()), "/primeTarget", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesDevicePrime.Prime)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesDevicePrimeTest.RandomPrime()
						mutator(datum)
						expectedDatum := dataTypesDevicePrimeTest.ClonePrime(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesDevicePrime.Prime) {},
				),
				Entry("does not modify the datum; target missing",
					func(datum *dataTypesDevicePrime.Prime) { datum.Target = nil },
				),
				Entry("does not modify the datum; volume missing",
					func(datum *dataTypesDevicePrime.Prime) { datum.Volume = nil },
				),
			)
		})
	})
})
