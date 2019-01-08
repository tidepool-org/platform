package status_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/status"
	dataTypesDeviceStatusTest "github.com/tidepool-org/platform/data/types/device/status/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "status",
	}
}

var _ = Describe("Status", func() {
	It("SubType is expected", func() {
		Expect(status.SubType).To(Equal("status"))
	})

	It("DurationMinimum is expected", func() {
		Expect(status.DurationMinimum).To(Equal(0))
	})

	It("NameResumed is expected", func() {
		Expect(status.NameResumed).To(Equal("resumed"))
	})

	It("NameSuspended is expected", func() {
		Expect(status.NameSuspended).To(Equal("suspended"))
	})

	It("Names returns expected", func() {
		Expect(status.Names()).To(Equal([]string{"resumed", "suspended"}))
	})

	Context("NewStatusDatum", func() {
		// TODO
	})

	Context("ParseStatusDatum", func() {
		// TODO
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := status.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("status"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.Name).To(BeNil())
			Expect(datum.Reason).To(BeNil())
		})
	})

	Context("Status", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *status.Status), expectedErrors ...error) {
					datum := dataTypesDeviceStatusTest.NewStatus()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *status.Status) {},
				),
				Entry("type missing",
					func(datum *status.Status) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "status"}),
				),
				Entry("type invalid",
					func(datum *status.Status) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "status"}),
				),
				Entry("type device",
					func(datum *status.Status) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *status.Status) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *status.Status) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "status"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type status",
					func(datum *status.Status) { datum.SubType = "status" },
				),
				Entry("duration missing; duration expected missing",
					func(datum *status.Status) {
						datum.Duration = nil
						datum.DurationExpected = nil
					},
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *status.Status) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *status.Status) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *status.Status) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *status.Status) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *status.Status) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/duration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *status.Status) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *status.Status) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *status.Status) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration in range; duration expected missing",
					func(datum *status.Status) {
						datum.Duration = pointer.FromInt(1)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range; duration expected out of range",
					func(datum *status.Status) {
						datum.Duration = pointer.FromInt(1)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(0, 1), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range; duration expected in range",
					func(datum *status.Status) {
						datum.Duration = pointer.FromInt(1)
						datum.DurationExpected = pointer.FromInt(1)
					},
				),
				Entry("name missing",
					func(datum *status.Status) { datum.Name = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/status", NewMeta()),
				),
				Entry("name invalid",
					func(datum *status.Status) { datum.Name = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"resumed", "suspended"}), "/status", NewMeta()),
				),
				Entry("name resumed",
					func(datum *status.Status) { datum.Name = pointer.FromString("resumed") },
				),
				Entry("name suspended",
					func(datum *status.Status) { datum.Name = pointer.FromString("suspended") },
				),
				Entry("reason missing",
					func(datum *status.Status) { datum.Reason = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/reason", NewMeta()),
				),
				Entry("reason exists",
					func(datum *status.Status) { datum.Reason = dataTest.NewBlob() },
				),
				Entry("multiple errors",
					func(datum *status.Status) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
						datum.Name = pointer.FromString("invalid")
						datum.Reason = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "status"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/duration", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/expectedDuration", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"resumed", "suspended"}), "/status", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/reason", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *status.Status)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesDeviceStatusTest.NewStatus()
						mutator(datum)
						expectedDatum := dataTypesDeviceStatusTest.CloneStatus(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *status.Status) {},
				),
				Entry("does not modify the datum; duration missing",
					func(datum *status.Status) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *status.Status) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; name missing",
					func(datum *status.Status) { datum.Name = nil },
				),
				Entry("does not modify the datum; reason missing",
					func(datum *status.Status) { datum.Reason = nil },
				),
			)
		})
	})
})
