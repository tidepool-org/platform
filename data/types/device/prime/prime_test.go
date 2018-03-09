package prime_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/prime"
	testDataTypesDevice "github.com/tidepool-org/platform/data/types/device/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
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
	datum.Device = *testDataTypesDevice.NewDevice()
	datum.SubType = "prime"
	datum.Target = pointer.String(test.RandomStringFromArray(prime.Targets()))
	switch *datum.Target {
	case "cannula":
		datum.Volume = pointer.Float64(test.RandomFloat64FromRange(prime.VolumeTargetCannulaMinimum, prime.VolumeTargetCannulaMaximum))
	case "tubing":
		datum.Volume = pointer.Float64(test.RandomFloat64FromRange(prime.VolumeTargetTubingMinimum, prime.VolumeTargetTubingMaximum))
	}
	return datum
}

func ClonePrime(datum *prime.Prime) *prime.Prime {
	if datum == nil {
		return nil
	}
	clone := prime.New()
	clone.Device = *testDataTypesDevice.CloneDevice(&datum.Device)
	clone.Target = test.CloneString(datum.Target)
	clone.Volume = test.CloneFloat64(datum.Volume)
	return clone
}

func NewTestPrime(sourceTime interface{}, sourceTarget interface{}, sourceVolume interface{}) *prime.Prime {
	datum := prime.Init()
	datum.DeviceID = pointer.String(id.New())
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

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(prime.NewDatum()).To(Equal(&prime.Prime{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(prime.New()).To(Equal(&prime.Prime{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := prime.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("prime"))
			Expect(datum.Target).To(BeNil())
			Expect(datum.Volume).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *prime.Prime

		BeforeEach(func() {
			datum = NewPrime()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("deviceEvent"))
				Expect(datum.SubType).To(Equal("prime"))
				Expect(datum.Target).To(BeNil())
				Expect(datum.Volume).To(BeNil())
			})
		})
	})

	Context("Prime", func() {
		Context("Parse", func() {
			var datum *prime.Prime

			BeforeEach(func() {
				datum = prime.Init()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *prime.Prime, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
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
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid target",
					&map[string]interface{}{"primeTarget": "cannula"},
					NewTestPrime(nil, "cannula", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid target",
					&map[string]interface{}{"primeTarget": 123},
					NewTestPrime(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(123), "/primeTarget", NewMeta()),
					}),
				Entry("parses object that has valid volume",
					&map[string]interface{}{"volume": 0.3},
					NewTestPrime(nil, nil, 0.3),
					[]*service.Error{}),
				Entry("parses object that has invalid volume",
					&map[string]interface{}{"volume": "invalid"},
					NewTestPrime(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/volume", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "primeTarget": "cannula", "volume": 0.3},
					NewTestPrime("2016-09-06T13:45:58-07:00", "cannula", 0.3),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "primeTarget": 123, "volume": "invalid"},
					NewTestPrime(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(123), "/primeTarget", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/volume", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *prime.Prime), expectedErrors ...error) {
					datum := NewPrime()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *prime.Prime) {},
				),
				Entry("type missing",
					func(datum *prime.Prime) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "prime"}),
				),
				Entry("type invalid",
					func(datum *prime.Prime) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "prime"}),
				),
				Entry("type device",
					func(datum *prime.Prime) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *prime.Prime) { datum.SubType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *prime.Prime) { datum.SubType = "invalidSubType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "prime"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type prime",
					func(datum *prime.Prime) { datum.SubType = "prime" },
				),
				Entry("target missing",
					func(datum *prime.Prime) { datum.Target = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/primeTarget", NewMeta()),
				),
				Entry("target invalid",
					func(datum *prime.Prime) { datum.Target = pointer.String("invalid") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"cannula", "tubing"}), "/primeTarget", NewMeta()),
				),
				Entry("target cannula; volume missing",
					func(datum *prime.Prime) {
						datum.Target = pointer.String("cannula")
						datum.Volume = nil
					},
				),
				Entry("target cannula; volume out of range (lower)",
					func(datum *prime.Prime) {
						datum.Target = pointer.String("cannula")
						datum.Volume = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0, 10), "/volume", NewMeta()),
				),
				Entry("target cannula; volume in range (lower)",
					func(datum *prime.Prime) {
						datum.Target = pointer.String("cannula")
						datum.Volume = pointer.Float64(0.0)
					},
				),
				Entry("target cannula; volume in range (upper)",
					func(datum *prime.Prime) {
						datum.Target = pointer.String("cannula")
						datum.Volume = pointer.Float64(10.0)
					},
				),
				Entry("target cannula; volume out of range (upper)",
					func(datum *prime.Prime) {
						datum.Target = pointer.String("cannula")
						datum.Volume = pointer.Float64(10.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0, 10), "/volume", NewMeta()),
				),
				Entry("target tubing; volume missing",
					func(datum *prime.Prime) {
						datum.Target = pointer.String("tubing")
						datum.Volume = nil
					},
				),
				Entry("target tubing; volume out of range (lower)",
					func(datum *prime.Prime) {
						datum.Target = pointer.String("tubing")
						datum.Volume = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/volume", NewMeta()),
				),
				Entry("target tubing; volume in range (lower)",
					func(datum *prime.Prime) {
						datum.Target = pointer.String("tubing")
						datum.Volume = pointer.Float64(0.0)
					},
				),
				Entry("target tubing; volume in range (upper)",
					func(datum *prime.Prime) {
						datum.Target = pointer.String("tubing")
						datum.Volume = pointer.Float64(100.0)
					},
				),
				Entry("target tubing; volume out of range (upper)",
					func(datum *prime.Prime) {
						datum.Target = pointer.String("tubing")
						datum.Volume = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/volume", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *prime.Prime) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Target = pointer.String("invalid")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "prime"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"cannula", "tubing"}), "/primeTarget", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
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
