package devicesettingschange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/devicesettingschange"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewDeviceSettingsChange() *devicesettingschange.DeviceSettingsChange {
	datum := devicesettingschange.New()
	datum.Device = *dataTypesDeviceTest.NewDevice()
	datum.SubType = "deviceSettingsChange"
	return datum
}

func CloneDeviceSettingsChange(datum *devicesettingschange.DeviceSettingsChange) *devicesettingschange.DeviceSettingsChange {
	if datum == nil {
		return nil
	}
	clone := devicesettingschange.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	return clone
}

var _ = Describe("Change", func() {
	It("SubType is expected", func() {
		Expect(devicesettingschange.SubType).To(Equal("deviceSettingsChange"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := devicesettingschange.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("deviceSettingsChange"))
		})
	})

	Context("DeviceSettingsChange", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *devicesettingschange.DeviceSettingsChange), expectedErrors ...error) {
					datum := NewDeviceSettingsChange()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *devicesettingschange.DeviceSettingsChange) {},
				),
				Entry("type missing",
					func(datum *devicesettingschange.DeviceSettingsChange) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "deviceSettingsChange"}),
				),
				Entry("type invalid",
					func(datum *devicesettingschange.DeviceSettingsChange) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "deviceSettingsChange"}),
				),
				Entry("type device",
					func(datum *devicesettingschange.DeviceSettingsChange) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *devicesettingschange.DeviceSettingsChange) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *devicesettingschange.DeviceSettingsChange) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "deviceSettingsChange"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type reservoir change",
					func(datum *devicesettingschange.DeviceSettingsChange) { datum.SubType = "deviceSettingsChange" },
				),
				Entry("multiple errors",
					func(datum *devicesettingschange.DeviceSettingsChange) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "deviceSettingsChange"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)

			DescribeTable("validates the datum with origin external",
				func(mutator func(datum *devicesettingschange.DeviceSettingsChange), expectedErrors ...error) {
					datum := NewDeviceSettingsChange()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginExternal, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *devicesettingschange.DeviceSettingsChange) {},
				),
			)

			DescribeTable("validates the datum with origin internal/store",
				func(mutator func(datum *devicesettingschange.DeviceSettingsChange), expectedErrors ...error) {
					datum := NewDeviceSettingsChange()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginInternal, expectedErrors...)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginStore, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *devicesettingschange.DeviceSettingsChange) {},
				),
			)
		})

		Context("Normalize", func() {
			It("does not modify datum if status is missing", func() {
				datum := NewDeviceSettingsChange()
				expectedDatum := CloneDeviceSettingsChange(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				Expect(datum).To(Equal(expectedDatum))
			})

			It("normalizes the datum and replaces status with status id", func() {
				datum := NewDeviceSettingsChange()
				expectedDatum := CloneDeviceSettingsChange(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(datum).To(Equal(expectedDatum))
			})
		})
	})
})
