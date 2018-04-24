package status_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/status"
	testDataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "status",
	}
}

func NewTestStatus(sourceTime interface{}, sourceDuration interface{}, sourceName interface{}, sourceReason *data.Blob) *status.Status {
	datum := status.New()
	datum.DeviceID = pointer.String(id.New())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceDuration.(int); ok {
		datum.Duration = &val
	}
	if val, ok := sourceName.(string); ok {
		datum.Name = &val
	}
	datum.Reason = sourceReason
	return datum
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
			var datum *status.Status

			BeforeEach(func() {
				datum = status.New()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *status.Status, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.Duration).To(Equal(expectedDatum.Duration))
					Expect(datum.DurationExpected).To(Equal(expectedDatum.DurationExpected))
					Expect(datum.Name).To(Equal(expectedDatum.Name))
					Expect(datum.Reason).To(Equal(expectedDatum.Reason))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestStatus(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestStatus(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestStatus("2016-09-06T13:45:58-07:00", nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestStatus(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid duration",
					&map[string]interface{}{"duration": 1000000},
					NewTestStatus(nil, 1000000, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration",
					&map[string]interface{}{"duration": "invalid"},
					NewTestStatus(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
					}),
				Entry("parses object that has valid name",
					&map[string]interface{}{"status": "suspended"},
					NewTestStatus(nil, nil, "suspended", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid name",
					&map[string]interface{}{"status": 123},
					NewTestStatus(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(123), "/status", NewMeta()),
					}),
				Entry("parses object that has valid reason",
					&map[string]interface{}{"reason": map[string]interface{}{"a": "one", "b": 2}},
					NewTestStatus(nil, nil, nil, &data.Blob{"a": "one", "b": 2}),
					[]*service.Error{}),
				Entry("parses object that has invalid reason",
					&map[string]interface{}{"reason": "invalid"},
					NewTestStatus(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotObject("invalid"), "/reason", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "duration": 1000000, "status": "suspended", "reason": map[string]interface{}{"a": "one", "b": 2}},
					NewTestStatus("2016-09-06T13:45:58-07:00", 1000000, "suspended", &data.Blob{"a": "one", "b": 2}),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "duration": "invalid", "status": 123, "reason": "invalid"},
					NewTestStatus(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(123), "/status", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotObject("invalid"), "/reason", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *status.Status), expectedErrors ...error) {
					datum := testDataTypesDeviceStatus.NewStatus()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *status.Status) {},
				),
				Entry("type missing",
					func(datum *status.Status) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "status"}),
				),
				Entry("type invalid",
					func(datum *status.Status) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "status"}),
				),
				Entry("type device",
					func(datum *status.Status) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *status.Status) { datum.SubType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *status.Status) { datum.SubType = "invalidSubType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "status"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
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
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *status.Status) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(0)
					},
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *status.Status) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *status.Status) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *status.Status) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/duration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *status.Status) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *status.Status) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *status.Status) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(0)
					},
				),
				Entry("duration in range; duration expected missing",
					func(datum *status.Status) {
						datum.Duration = pointer.Int(1)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range; duration expected out of range",
					func(datum *status.Status) {
						datum.Duration = pointer.Int(1)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(0, 1), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range; duration expected in range",
					func(datum *status.Status) {
						datum.Duration = pointer.Int(1)
						datum.DurationExpected = pointer.Int(1)
					},
				),
				Entry("name missing",
					func(datum *status.Status) { datum.Name = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/status", NewMeta()),
				),
				Entry("name invalid",
					func(datum *status.Status) { datum.Name = pointer.String("invalid") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"resumed", "suspended"}), "/status", NewMeta()),
				),
				Entry("name resumed",
					func(datum *status.Status) { datum.Name = pointer.String("resumed") },
				),
				Entry("name suspended",
					func(datum *status.Status) { datum.Name = pointer.String("suspended") },
				),
				Entry("reason missing",
					func(datum *status.Status) { datum.Reason = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/reason", NewMeta()),
				),
				Entry("reason exists",
					func(datum *status.Status) { datum.Reason = testData.NewBlob() },
				),
				Entry("multiple errors",
					func(datum *status.Status) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(-1)
						datum.Name = pointer.String("invalid")
						datum.Reason = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "status"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/duration", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/expectedDuration", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"resumed", "suspended"}), "/status", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/reason", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *status.Status)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesDeviceStatus.NewStatus()
						mutator(datum)
						expectedDatum := testDataTypesDeviceStatus.CloneStatus(datum)
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
