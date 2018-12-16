package prime_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/prime"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
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
	datum.Device = *dataTypesDeviceTest.NewDevice()
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
	clone.Target = test.CloneString(datum.Target)
	clone.Volume = test.CloneFloat64(datum.Volume)
	return clone
}

func NewTestPrime(sourceTime interface{}, sourceTarget interface{}, sourceVolume interface{}) *prime.Prime {
	datum := prime.New()
	datum.DeviceID = pointer.FromString(dataTest.NewDeviceID())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceTarget.(string); ok {
		datum.Target = &val
	}
	if val, ok := sourceVolume.(float64); ok {
		datum.Volume = &val
	}
	return datum
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
			var datum *prime.Prime

			BeforeEach(func() {
				datum = prime.New()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *prime.Prime, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.Target).To(Equal(expectedDatum.Target))
					Expect(datum.Volume).To(Equal(expectedDatum.Volume))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestPrime(nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestPrime(nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestPrime("2016-09-06T13:45:58-07:00", nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestPrime(nil, nil, nil),
					[]*service.Error{
						dataTest.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid target",
					&map[string]interface{}{"primeTarget": "cannula"},
					NewTestPrime(nil, "cannula", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid target",
					&map[string]interface{}{"primeTarget": 123},
					NewTestPrime(nil, nil, nil),
					[]*service.Error{
						dataTest.ComposeError(service.ErrorTypeNotString(123), "/primeTarget", NewMeta()),
					}),
				Entry("parses object that has valid volume",
					&map[string]interface{}{"volume": 0.3},
					NewTestPrime(nil, nil, 0.3),
					[]*service.Error{}),
				Entry("parses object that has invalid volume",
					&map[string]interface{}{"volume": "invalid"},
					NewTestPrime(nil, nil, nil),
					[]*service.Error{
						dataTest.ComposeError(service.ErrorTypeNotFloat("invalid"), "/volume", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "primeTarget": "cannula", "volume": 0.3},
					NewTestPrime("2016-09-06T13:45:58-07:00", "cannula", 0.3),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "primeTarget": 123, "volume": "invalid"},
					NewTestPrime(nil, nil, nil),
					[]*service.Error{
						dataTest.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						dataTest.ComposeError(service.ErrorTypeNotString(123), "/primeTarget", NewMeta()),
						dataTest.ComposeError(service.ErrorTypeNotFloat("invalid"), "/volume", NewMeta()),
					}),
			)
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
