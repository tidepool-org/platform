package mode_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypesCommonTest "github.com/tidepool-org/platform/data/types/common/test"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/mode"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "zen",
	}
}

func NewMode() *mode.Mode {
	datum := mode.New(mode.ZenMode)
	datum.Device = *dataTypesDeviceTest.NewDevice()
	datum.SubType = mode.ZenMode
	datum.EventID = pointer.FromString("123456789")
	datum.Duration = dataTypesCommonTest.NewDuration()
	return datum
}

func CloneMode(datum *mode.Mode) *mode.Mode {
	if datum == nil {
		return nil
	}
	clone := mode.New(datum.SubType)
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.EventID = pointer.FromString("123456789")
	clone.Duration = dataTypesCommonTest.CloneDuration(datum.Duration)
	return clone
}

var _ = Describe("Change", func() {
	It("SubType is expected", func() {
		Expect(mode.ConfidentialMode).To(Equal("confidential"))
		Expect(mode.ZenMode).To(Equal("zen"))
	})

	Context("New", func() {
		It("returns the expected datum with all Zen values initialized", func() {
			datum := mode.New(mode.ZenMode)
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("zen"))
		})
		It("returns the expected datum with all confidential values initialized", func() {
			datum := mode.New(mode.ConfidentialMode)
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("confidential"))
		})
	})

	Context("Mode", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *mode.Mode), expectedErrors ...error) {
					datum := NewMode()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *mode.Mode) {},
				),
				Entry("type missing",
					func(datum *mode.Mode) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "zen"}),
				),
				Entry("type invalid",
					func(datum *mode.Mode) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "zen"}),
				),
				Entry("type device",
					func(datum *mode.Mode) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *mode.Mode) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("", []string{"confidential", "zen"}), "/subType", &device.Meta{Type: "deviceEvent", SubType: ""}),
				),
				Entry("sub type invalid",
					func(datum *mode.Mode) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalidSubType", []string{"confidential", "zen"}), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type zen",
					func(datum *mode.Mode) { datum.SubType = "zen" },
				),
				Entry("multiple errors",
					func(datum *mode.Mode) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalidSubType", []string{"confidential", "zen"}), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
				Entry("EventId is missing",
					func(datum *mode.Mode) {
						datum.EventID = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/eventId", &device.Meta{Type: "deviceEvent", SubType: "zen"}),
				),
				Entry("EventId is missing",
					func(datum *mode.Mode) {
						datum.SubType = "confidential"
						datum.EventID = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/eventId", &device.Meta{Type: "deviceEvent", SubType: "confidential"}),
				),
			)

			DescribeTable("validates the datum with origin internal/store",
				func(mutator func(datum *mode.Mode), expectedErrors ...error) {
					datum := NewMode()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginInternal, expectedErrors...)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginStore, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *mode.Mode) {},
				),
			)
		})

		Context("Normalize", func() {
		})
	})
})
