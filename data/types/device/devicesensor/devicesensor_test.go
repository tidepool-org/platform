package devicesensor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/devicesensor"
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

func NewDeviceSensor() *devicesensor.DeviceSensor {
	datum := devicesensor.New()
	datum.Device = *dataTypesDeviceTest.NewDevice()
	datum.SubType = "deviceSensor"
	datum.EventType = pointer.FromString(test.RandomStringFromArray(devicesensor.EventTypes()))
	return datum
}

func CloneDeviceSensor(datum *devicesensor.DeviceSensor) *devicesensor.DeviceSensor {
	if datum == nil {
		return nil
	}
	clone := devicesensor.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.EventType = pointer.CloneString(datum.EventType)
	return clone
}

var _ = Describe("Change", func() {
	It("SubType is expected", func() {
		Expect(devicesensor.SubType).To(Equal("deviceSensor"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := devicesensor.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("deviceSensor"))
			Expect(datum.EventType).To(BeNil())
		})
	})

	Context("DeviceSensor", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *devicesensor.DeviceSensor), expectedErrors ...error) {
					datum := NewDeviceSensor()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *devicesensor.DeviceSensor) {},
				),
				Entry("type missing",
					func(datum *devicesensor.DeviceSensor) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "deviceSensor"}),
				),
				Entry("type invalid",
					func(datum *devicesensor.DeviceSensor) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "deviceSensor"}),
				),
				Entry("type device",
					func(datum *devicesensor.DeviceSensor) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *devicesensor.DeviceSensor) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *devicesensor.DeviceSensor) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "deviceSensor"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type change",
					func(datum *devicesensor.DeviceSensor) { datum.SubType = "deviceSensor" },
				),
				Entry("multiple errors",
					func(datum *devicesensor.DeviceSensor) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "deviceSensor"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)

			DescribeTable("validates the datum with origin external",
				func(mutator func(datum *devicesensor.DeviceSensor), expectedErrors ...error) {
					datum := NewDeviceSensor()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginExternal, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *devicesensor.DeviceSensor) {},
				),
				Entry("event type valid",
					func(datum *devicesensor.DeviceSensor) {
						datum.EventType = pointer.FromString(test.RandomStringFromArray(devicesensor.EventTypes()))
					},
				),
				Entry("multiple errors",
					func(datum *devicesensor.DeviceSensor) {
						datum.EventType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"start", "stop", "expired"}), "/eventType", NewMeta()),
				),
			)

			DescribeTable("validates the datum with origin internal/store",
				func(mutator func(datum *devicesensor.DeviceSensor), expectedErrors ...error) {
					datum := NewDeviceSensor()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginInternal, expectedErrors...)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginStore, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *devicesensor.DeviceSensor) {},
				),
				Entry("event type invalid",
					func(datum *devicesensor.DeviceSensor) { datum.EventType = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"start", "stop", "expired"}), "/eventType", NewMeta()),
				),
				Entry("event type valid",
					func(datum *devicesensor.DeviceSensor) {
						datum.EventType = pointer.FromString(test.RandomStringFromArray(devicesensor.EventTypes()))
					},
				),
				Entry("multiple errors",
					func(datum *devicesensor.DeviceSensor) {
						datum.EventType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"start", "stop", "expired"}), "/eventType", NewMeta()),
				),
			)
		})

		Context("Normalize", func() {
			It("does not modify datum if status is missing", func() {
				datum := NewDeviceSensor()
				expectedDatum := CloneDeviceSensor(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				Expect(datum).To(Equal(expectedDatum))
			})

			It("normalizes the datum and replaces status with status id", func() {
				datum := NewDeviceSensor()
				expectedDatum := CloneDeviceSensor(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(datum).To(Equal(expectedDatum))
			})
		})
	})
})
