package prime_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/prime"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "prime",
	}
}

func NewPrime() *prime.Prime {
	datum := prime.New()
	datum.Device = *dataTypesDeviceTest.RandomDevice()
	datum.SubType = "prime"
	datum.Target = pointer.FromString(test.RandomStringFromArray(prime.Targets()))
	switch *datum.Target {
	case "cannula":
		datum.Volume = pointer.FromFloat64(test.RandomFloat64FromRange(prime.VolumeTargetCannulaMinimum, prime.VolumeTargetCannulaMaximum))
	case "tubing":
		datum.Volume = pointer.FromFloat64(test.RandomFloat64FromRange(prime.VolumeTargetTubingMinimum, prime.VolumeTargetTubingMaximum))
	}
	return datum
}

func ClonePrime(datum *prime.Prime) *prime.Prime {
	if datum == nil {
		return nil
	}
	clone := prime.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.Target = pointer.CloneString(datum.Target)
	clone.Volume = pointer.CloneFloat64(datum.Volume)
	return clone
}

var _ = Describe("Status", func() {
	It("SubType is expected", func() {
		Expect(prime.SubType).To(Equal("prime"))
	})

	It("TargetCannula is expected", func() {
		Expect(prime.TargetCannula).To(Equal("cannula"))
	})

	It("TargetTubing is expected", func() {
		Expect(prime.TargetTubing).To(Equal("tubing"))
	})

	It("VolumeTargetCannulaMaximum is expected", func() {
		Expect(prime.VolumeTargetCannulaMaximum).To(Equal(10.0))
	})

	It("VolumeTargetCannulaMinimum is expected", func() {
		Expect(prime.VolumeTargetCannulaMinimum).To(Equal(0.0))
	})

	It("VolumeTargetTubingMaximum is expected", func() {
		Expect(prime.VolumeTargetTubingMaximum).To(Equal(100.0))
	})

	It("VolumeTargetTubingMinimum is expected", func() {
		Expect(prime.VolumeTargetTubingMinimum).To(Equal(0.0))
	})

	It("Targets returns expected", func() {
		Expect(prime.Targets()).To(Equal([]string{"cannula", "tubing"}))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := prime.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("prime"))
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
				func(mutator func(datum *prime.Prime), expectedErrors ...error) {
					datum := NewPrime()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *prime.Prime) {},
				),
				Entry("type missing",
					func(datum *prime.Prime) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "prime"}),
				),
				Entry("type invalid",
					func(datum *prime.Prime) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "prime"}),
				),
				Entry("type device",
					func(datum *prime.Prime) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *prime.Prime) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *prime.Prime) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "prime"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type prime",
					func(datum *prime.Prime) { datum.SubType = "prime" },
				),
				Entry("target missing",
					func(datum *prime.Prime) { datum.Target = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/primeTarget", NewMeta()),
				),
				Entry("target invalid",
					func(datum *prime.Prime) { datum.Target = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"cannula", "tubing"}), "/primeTarget", NewMeta()),
				),
				Entry("target cannula; volume missing",
					func(datum *prime.Prime) {
						datum.Target = pointer.FromString("cannula")
						datum.Volume = nil
					},
				),
				Entry("target cannula; volume out of range (lower)",
					func(datum *prime.Prime) {
						datum.Target = pointer.FromString("cannula")
						datum.Volume = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0, 10), "/volume", NewMeta()),
				),
				Entry("target cannula; volume in range (lower)",
					func(datum *prime.Prime) {
						datum.Target = pointer.FromString("cannula")
						datum.Volume = pointer.FromFloat64(0.0)
					},
				),
				Entry("target cannula; volume in range (upper)",
					func(datum *prime.Prime) {
						datum.Target = pointer.FromString("cannula")
						datum.Volume = pointer.FromFloat64(10.0)
					},
				),
				Entry("target cannula; volume out of range (upper)",
					func(datum *prime.Prime) {
						datum.Target = pointer.FromString("cannula")
						datum.Volume = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0, 10), "/volume", NewMeta()),
				),
				Entry("target tubing; volume missing",
					func(datum *prime.Prime) {
						datum.Target = pointer.FromString("tubing")
						datum.Volume = nil
					},
				),
				Entry("target tubing; volume out of range (lower)",
					func(datum *prime.Prime) {
						datum.Target = pointer.FromString("tubing")
						datum.Volume = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/volume", NewMeta()),
				),
				Entry("target tubing; volume in range (lower)",
					func(datum *prime.Prime) {
						datum.Target = pointer.FromString("tubing")
						datum.Volume = pointer.FromFloat64(0.0)
					},
				),
				Entry("target tubing; volume in range (upper)",
					func(datum *prime.Prime) {
						datum.Target = pointer.FromString("tubing")
						datum.Volume = pointer.FromFloat64(100.0)
					},
				),
				Entry("target tubing; volume out of range (upper)",
					func(datum *prime.Prime) {
						datum.Target = pointer.FromString("tubing")
						datum.Volume = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/volume", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *prime.Prime) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Target = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "prime"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"cannula", "tubing"}), "/primeTarget", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *prime.Prime)) {
					for _, origin := range structure.Origins() {
						datum := NewPrime()
						mutator(datum)
						expectedDatum := ClonePrime(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *prime.Prime) {},
				),
				Entry("does not modify the datum; target missing",
					func(datum *prime.Prime) { datum.Target = nil },
				),
				Entry("does not modify the datum; volume missing",
					func(datum *prime.Prime) { datum.Volume = nil },
				),
			)
		})
	})
})
