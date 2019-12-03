package transmitter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	"github.com/tidepool-org/platform/data/types/device/transmitter"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	test "github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "deviceTransmitter",
	}
}

func NewTransmitter() *transmitter.Transmitter {
	datum := transmitter.New()
	datum.Device = *dataTypesDeviceTest.NewDevice()
	datum.SubType = "deviceTransmitter"
	datum.EventType = pointer.FromString(test.RandomStringFromArray(transmitter.EventTypes()))
	return datum
}

func CloneTransmitter(datum *transmitter.Transmitter) *transmitter.Transmitter {
	if datum == nil {
		return nil
	}
	clone := transmitter.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.EventType = pointer.CloneString(datum.EventType)
	return clone
}

var _ = Describe("Change", func() {
	It("SubType is expected", func() {
		Expect(transmitter.SubType).To(Equal("deviceTransmitter"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := transmitter.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("deviceTransmitter"))
			Expect(datum.EventType).To(BeNil())
		})
	})

	Context("Transmitter", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *transmitter.Transmitter), expectedErrors ...error) {
					datum := NewTransmitter()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *transmitter.Transmitter) {},
				),
				Entry("type missing",
					func(datum *transmitter.Transmitter) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "deviceTransmitter"}),
				),
				Entry("type invalid",
					func(datum *transmitter.Transmitter) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "deviceTransmitter"}),
				),
				Entry("type device",
					func(datum *transmitter.Transmitter) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *transmitter.Transmitter) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *transmitter.Transmitter) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "deviceTransmitter"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type change",
					func(datum *transmitter.Transmitter) { datum.SubType = "deviceTransmitter" },
				),
				Entry("multiple errors",
					func(datum *transmitter.Transmitter) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "deviceTransmitter"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)

			DescribeTable("validates the datum with origin external",
				func(mutator func(datum *transmitter.Transmitter), expectedErrors ...error) {
					datum := NewTransmitter()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginExternal, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *transmitter.Transmitter) {},
				),
				Entry("event type valid",
					func(datum *transmitter.Transmitter) {
						datum.EventType = pointer.FromString(test.RandomStringFromArray(transmitter.EventTypes()))
					},
				),
				Entry("multiple errors",
					func(datum *transmitter.Transmitter) {
						datum.EventType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"start", "stop", "expired"}), "/eventType", NewMeta()),
				),
			)

			DescribeTable("validates the datum with origin internal/store",
				func(mutator func(datum *transmitter.Transmitter), expectedErrors ...error) {
					datum := NewTransmitter()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginInternal, expectedErrors...)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginStore, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *transmitter.Transmitter) {},
				),
				Entry("event type invalid",
					func(datum *transmitter.Transmitter) { datum.EventType = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"start", "stop", "expired"}), "/eventType", NewMeta()),
				),
				Entry("event type valid",
					func(datum *transmitter.Transmitter) {
						datum.EventType = pointer.FromString(test.RandomStringFromArray(transmitter.EventTypes()))
					},
				),
				Entry("multiple errors",
					func(datum *transmitter.Transmitter) {
						datum.EventType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"start", "stop", "expired"}), "/eventType", NewMeta()),
				),
			)
		})

		Context("Normalize", func() {
			It("does not modify datum if status is missing", func() {
				datum := NewTransmitter()
				expectedDatum := CloneTransmitter(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				Expect(datum).To(Equal(expectedDatum))
			})

			It("normalizes the datum and replaces status with status id", func() {
				datum := NewTransmitter()
				expectedDatum := CloneTransmitter(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(datum).To(Equal(expectedDatum))
			})
		})
	})
})
