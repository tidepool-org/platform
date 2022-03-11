package flush_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/flush"
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
		SubType: "flush",
	}
}

func NewFlush() *flush.Flush {
	datum := flush.New()
	datum.Device = *dataTypesDeviceTest.NewDevice()
	datum.SubType = "flush"
	datum.Status = pointer.FromString(test.RandomStringFromArray(flush.Statuses()))
	datum.StatusCode = pointer.FromInt(test.RandomIntFromArray(flush.StatusCodes()))
	datum.Volume = pointer.FromFloat64(test.RandomFloat64FromRange(flush.VolumeTargetMinimum, flush.VolumeTargetMaximum))
	return datum
}

func CloneFlush(datum *flush.Flush) *flush.Flush {
	if datum == nil {
		return nil
	}
	clone := flush.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.Status = pointer.CloneString(datum.Status)
	clone.StatusCode = pointer.CloneInt(datum.StatusCode)
	clone.Volume = pointer.CloneFloat64(datum.Volume)
	return clone
}

var _ = Describe("Status", func() {
	It("SubType is expected", func() {
		Expect(flush.SubType).To(Equal("flush"))
	})

	It("Succeeded is expected", func() {
		Expect(flush.Succeeded).To(Equal("success"))
	})
	It("Failed is expected", func() {
		Expect(flush.Failed).To(Equal("failure"))
	})
	It("Statuses returns expected expected", func() {
		Expect(flush.Statuses()).To(Equal([]string{"success", "failure"}))
	})

	It("VolumeTargetMaximum is expected", func() {
		Expect(flush.VolumeTargetMaximum).To(Equal(10.0))
	})

	It("VolumeTargetMinimum is expected", func() {
		Expect(flush.VolumeTargetMinimum).To(Equal(0.0))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := flush.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("flush"))
			Expect(datum.Status).To(BeNil())
			Expect(datum.StatusCode).To(BeNil())
			Expect(datum.Volume).To(BeNil())
		})
	})

	Context("Flush", func() {
		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *flush.Flush), expectedErrors ...error) {
					datum := NewFlush()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *flush.Flush) {},
				),
				Entry("type missing",
					func(datum *flush.Flush) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "flush"}),
				),
				Entry("type invalid",
					func(datum *flush.Flush) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "flush"}),
				),
				Entry("type device",
					func(datum *flush.Flush) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *flush.Flush) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *flush.Flush) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "flush"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type flush",
					func(datum *flush.Flush) { datum.SubType = "flush" },
				),
				Entry("status missing",
					func(datum *flush.Flush) { datum.Status = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/status", NewMeta()),
				),
				Entry("Status invalid",
					func(datum *flush.Flush) {
						datum.Status = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"success", "failure"}), "/status", NewMeta()),
				),
				Entry("Status valid, StatusCode missing",
					func(datum *flush.Flush) {
						datum.SubType = "flush"
						datum.Status = pointer.FromString("success")
						datum.StatusCode = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/statusCode", NewMeta()),
				),
				Entry("Status success, volume missing",
					func(datum *flush.Flush) {
						datum.Status = pointer.FromString("success")
						datum.Volume = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/volume", NewMeta()),
				),
				Entry("Status success; volume out of range (lower)",
					func(datum *flush.Flush) {
						datum.Status = pointer.FromString("success")
						datum.Volume = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0, 10), "/volume", NewMeta()),
				),
				Entry("Status success; volume in range (lower)",
					func(datum *flush.Flush) {
						datum.Status = pointer.FromString("success")
						datum.Volume = pointer.FromFloat64(0.0)
					},
				),
				Entry("Status success; volume in range (upper)",
					func(datum *flush.Flush) {
						datum.Status = pointer.FromString("success")
						datum.Volume = pointer.FromFloat64(10.0)
					},
				),
				Entry("Status success; volume out of range (upper)",
					func(datum *flush.Flush) {
						datum.Status = pointer.FromString("success")
						datum.Volume = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0, 10), "/volume", NewMeta()),
				),
				Entry("Volume missing",
					func(datum *flush.Flush) { datum.Volume = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/volume", NewMeta()),
				),
				Entry("statusCode missing",
					func(datum *flush.Flush) { datum.StatusCode = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/statusCode", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *flush.Flush) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Status = pointer.FromString("invalid")
						datum.StatusCode = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "flush"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"success", "failure"}), "/status", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/statusCode", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *flush.Flush)) {
					for _, origin := range structure.Origins() {
						datum := NewFlush()
						mutator(datum)
						expectedDatum := CloneFlush(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *flush.Flush) {},
				),
				Entry("does not modify the datum; Status missing",
					func(datum *flush.Flush) { datum.Status = nil },
				),
				Entry("does not modify the datum; StatusCode missing",
					func(datum *flush.Flush) { datum.StatusCode = nil },
				),
				Entry("does not modify the datum; volume missing",
					func(datum *flush.Flush) { datum.Volume = nil },
				),
			)
		})
	})
})
