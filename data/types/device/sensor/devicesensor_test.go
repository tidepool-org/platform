package sensor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/sensor"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
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
		SubType: "deviceSensor",
	}
}

func NewSensor() *sensor.Sensor {
	datum := sensor.New()
	datum.Device = *dataTypesDeviceTest.NewDevice()
	datum.SubType = "deviceSensor"
	datum.EventType = pointer.FromString(test.RandomStringFromArray(sensor.EventTypes()))
	return datum
}

func CloneSensor(datum *sensor.Sensor) *sensor.Sensor {
	if datum == nil {
		return nil
	}
	clone := sensor.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.EventType = pointer.CloneString(datum.EventType)
	return clone
}

var _ = Describe("Change", func() {
	It("SubType is expected", func() {
		Expect(sensor.SubType).To(Equal("deviceSensor"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := sensor.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("deviceSensor"))
			Expect(datum.EventType).To(BeNil())
		})
	})

	Context("Sensor", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *sensor.Sensor), expectedErrors ...error) {
					datum := NewSensor()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *sensor.Sensor) {},
				),
				Entry("type missing",
					func(datum *sensor.Sensor) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "deviceSensor"}),
				),
				Entry("type invalid",
					func(datum *sensor.Sensor) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "deviceSensor"}),
				),
				Entry("type device",
					func(datum *sensor.Sensor) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *sensor.Sensor) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *sensor.Sensor) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "deviceSensor"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type change",
					func(datum *sensor.Sensor) { datum.SubType = "deviceSensor" },
				),
				Entry("multiple errors",
					func(datum *sensor.Sensor) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "deviceSensor"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)

			DescribeTable("validates the datum with origin external",
				func(mutator func(datum *sensor.Sensor), expectedErrors ...error) {
					datum := NewSensor()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginExternal, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *sensor.Sensor) {},
				),
				Entry("event type valid",
					func(datum *sensor.Sensor) {
						datum.EventType = pointer.FromString(test.RandomStringFromArray(sensor.EventTypes()))
					},
				),
				Entry("multiple errors",
					func(datum *sensor.Sensor) {
						datum.EventType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"start", "stop", "expired"}), "/eventType", NewMeta()),
				),
			)

			DescribeTable("validates the datum with origin internal/store",
				func(mutator func(datum *sensor.Sensor), expectedErrors ...error) {
					datum := NewSensor()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginInternal, expectedErrors...)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginStore, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *sensor.Sensor) {},
				),
				Entry("event type invalid",
					func(datum *sensor.Sensor) { datum.EventType = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"start", "stop", "expired"}), "/eventType", NewMeta()),
				),
				Entry("event type valid",
					func(datum *sensor.Sensor) {
						datum.EventType = pointer.FromString(test.RandomStringFromArray(sensor.EventTypes()))
					},
				),
				Entry("multiple errors",
					func(datum *sensor.Sensor) {
						datum.EventType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"start", "stop", "expired"}), "/eventType", NewMeta()),
				),
			)
		})

		Context("Normalize", func() {
			It("does not modify datum if status is missing", func() {
				datum := NewSensor()
				expectedDatum := CloneSensor(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				Expect(datum).To(Equal(expectedDatum))
			})

			It("normalizes the datum and replaces status with status id", func() {
				datum := NewSensor()
				expectedDatum := CloneSensor(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(datum).To(Equal(expectedDatum))
			})
		})
	})
})
