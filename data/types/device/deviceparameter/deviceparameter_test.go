package deviceparameter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/deviceparameter"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

const (
	updateTimeFormat = "2006-01-02T15:04:05.000Z"
)

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "deviceParameter",
	}
}

func NewDeviceParameter() *deviceparameter.DeviceParameter {
	updateTime := test.RandomTime()

	datum := deviceparameter.New()
	datum.Device = *dataTypesDeviceTest.NewDevice()
	datum.SubType = "deviceParameter"
	datum.Name = pointer.FromString(test.RandomString())
	datum.Value = pointer.FromString(test.RandomString())
	datum.LastUpdateDate = pointer.FromString(updateTime.Format(updateTimeFormat))
	datum.PreviousValue = pointer.FromString(test.RandomString())
	datum.Level = pointer.FromString(test.RandomStringFromArray(deviceparameter.LevelValues()))
	datum.MinValue = pointer.FromString(test.RandomString())
	datum.MaxValue = pointer.FromString(test.RandomString())
	datum.Processed = pointer.FromString(test.RandomStringFromArray(deviceparameter.ProcessedValues()))
	datum.LinkedSubType = pointer.FromStringArray([]string{test.RandomString()})
	return datum
}

func CloneDeviceParameter(datum *deviceparameter.DeviceParameter) *deviceparameter.DeviceParameter {
	if datum == nil {
		return nil
	}
	clone := deviceparameter.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Value = pointer.CloneString(datum.Value)
	clone.LastUpdateDate = pointer.CloneString(datum.LastUpdateDate)
	clone.PreviousValue = pointer.CloneString(datum.PreviousValue)
	clone.Level = pointer.CloneString(datum.Level)
	clone.MinValue = pointer.CloneString(datum.MinValue)
	clone.MaxValue = pointer.CloneString(datum.MaxValue)
	clone.Processed = pointer.CloneString(datum.Processed)
	clone.LinkedSubType = pointer.CloneStringArray(datum.LinkedSubType)

	return clone
}

var _ = Describe("Status", func() {
	It("SubType is expected", func() {
		Expect(deviceparameter.SubType).To(Equal("deviceParameter"))
	})

	It("ProcessedYes is expected", func() {
		Expect(deviceparameter.ProcessedYes).To(Equal("yes"))
	})

	It("ProcessedNo is expected", func() {
		Expect(deviceparameter.ProcessedNo).To(Equal("no"))
	})

	It("LastUpdateDateTimeFormat is expected", func() {
		Expect(deviceparameter.LastUpdateDateTimeFormat).To(Equal(updateTimeFormat))
	})

	It("ProcessedValues returns expected", func() {
		Expect(deviceparameter.ProcessedValues()).To(Equal([]string{"yes", "no"}))
	})

	It("LevelValues returns expected", func() {
		Expect(deviceparameter.LevelValues()).To(Equal([]string{"1", "2", "3"}))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := deviceparameter.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("deviceParameter"))
			Expect(datum.Name).To(BeNil())
			Expect(datum.Value).To(BeNil())
			Expect(datum.Units).To(BeNil())
			Expect(datum.LastUpdateDate).To(BeNil())
			Expect(datum.PreviousValue).To(BeNil())
			Expect(datum.Level).To(BeNil())
			Expect(datum.MinValue).To(BeNil())
			Expect(datum.MaxValue).To(BeNil())
			Expect(datum.Processed).To(BeNil())
			Expect(datum.LinkedSubType).To(BeNil())
		})
	})

	Context("DeviceParameter", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *deviceparameter.DeviceParameter), expectedErrors ...error) {
					datum := NewDeviceParameter()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *deviceparameter.DeviceParameter) {},
				),
				Entry("type missing",
					func(datum *deviceparameter.DeviceParameter) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "deviceParameter"}),
				),
				Entry("type invalid",
					func(datum *deviceparameter.DeviceParameter) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "deviceParameter"}),
				),
				Entry("type device",
					func(datum *deviceparameter.DeviceParameter) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *deviceparameter.DeviceParameter) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *deviceparameter.DeviceParameter) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "deviceParameter"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type DeviceParameter",
					func(datum *deviceparameter.DeviceParameter) { datum.SubType = "deviceParameter" },
				),
				Entry("name missing",
					func(datum *deviceparameter.DeviceParameter) { datum.Name = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/name", NewMeta()),
				),
				Entry("value missing",
					func(datum *deviceparameter.DeviceParameter) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("lastUpdateDate missing",
					func(datum *deviceparameter.DeviceParameter) { datum.LastUpdateDate = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/lastUpdateDate", NewMeta()),
				),
				Entry("level missing",
					func(datum *deviceparameter.DeviceParameter) { datum.Level = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/level", NewMeta()),
				),
				Entry("level with wrong value",
					func(datum *deviceparameter.DeviceParameter) { *datum.Level = "nil" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("nil", []string{"1", "2", "3"}), "/level", NewMeta()),
				),
				Entry("processed with wrong value",
					func(datum *deviceparameter.DeviceParameter) { *datum.Processed = "nil" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("nil", []string{"yes", "no"}), "/processed", NewMeta()),
				),
				Entry("linkedSubType missing when processed is yes",
					func(datum *deviceparameter.DeviceParameter) {
						*datum.Processed = "yes"
						datum.LinkedSubType = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/linkedSubType", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *deviceparameter.DeviceParameter) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						// *datum.Processed = "nil"
						// datum.Target = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "deviceParameter"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *deviceparameter.DeviceParameter)) {
					for _, origin := range structure.Origins() {
						datum := NewDeviceParameter()
						mutator(datum)
						expectedDatum := CloneDeviceParameter(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *deviceparameter.DeviceParameter) {},
				),
				Entry("does not modify the datum; name missing",
					func(datum *deviceparameter.DeviceParameter) { datum.Name = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *deviceparameter.DeviceParameter) { datum.Value = nil },
				),
				Entry("does not modify the datum; lastUpdateDate missing",
					func(datum *deviceparameter.DeviceParameter) { datum.LastUpdateDate = nil },
				),
				Entry("does not modify the datum; level missing",
					func(datum *deviceparameter.DeviceParameter) { datum.Level = nil },
				),
				Entry("does not modify the datum; processed with wrong value",
					func(datum *deviceparameter.DeviceParameter) { *datum.Processed = "nil" },
				),
				Entry("does not modify the datum; linkedSubType missing when processed is yes",
					func(datum *deviceparameter.DeviceParameter) {
						*datum.Processed = "yes"
						datum.LinkedSubType = nil
					},
				),
			)
		})
	})
})
