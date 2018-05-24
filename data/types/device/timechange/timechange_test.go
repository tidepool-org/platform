package timechange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	testDataTypesDevice "github.com/tidepool-org/platform/data/types/device/test"
	"github.com/tidepool-org/platform/data/types/device/timechange"
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
		SubType: "timeChange",
	}
}

func NewTimeChange() *timechange.TimeChange {
	datum := timechange.New()
	datum.Device = *testDataTypesDevice.NewDevice()
	datum.SubType = "timeChange"
	datum.Change = NewChange()
	return datum
}

func CloneTimeChange(datum *timechange.TimeChange) *timechange.TimeChange {
	if datum == nil {
		return nil
	}
	clone := timechange.New()
	clone.Device = *testDataTypesDevice.CloneDevice(&datum.Device)
	clone.Change = CloneChange(datum.Change)
	return clone
}

func NewTestTimeChange(sourceTime interface{}, sourceChange *timechange.Change) *timechange.TimeChange {
	datum := timechange.New()
	datum.DeviceID = pointer.FromString(id.New())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	datum.Change = sourceChange
	return datum
}

var _ = Describe("Change", func() {
	It("SubType is expected", func() {
		Expect(timechange.SubType).To(Equal("timeChange"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := timechange.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("timeChange"))
			Expect(datum.Change).To(BeNil())
		})
	})

	Context("TimeChange", func() {
		Context("Parse", func() {
			var datum *timechange.TimeChange

			BeforeEach(func() {
				datum = timechange.New()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *timechange.TimeChange, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.Change).To(Equal(expectedDatum.Change))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestTimeChange(nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestTimeChange(nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestTimeChange("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestTimeChange(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid change",
					&map[string]interface{}{"change": map[string]interface{}{"agent": "manual"}},
					NewTestTimeChange(nil, NewTestChange("manual", nil, nil)),
					[]*service.Error{}),
				Entry("parses object that has invalid change",
					&map[string]interface{}{"change": map[string]interface{}{"agent": 123}},
					NewTestTimeChange(nil, NewTestChange(nil, nil, nil)),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(123), "/change/agent", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"change": map[string]interface{}{"agent": "manual", "from": "2016-03-21T09:45:37", "to": "2016-03-21T10:42:00"}},
					NewTestTimeChange(nil, NewTestChange("manual", "2016-03-21T09:45:37", "2016-03-21T10:42:00")),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "change": map[string]interface{}{"agent": 123, "from": 456, "to": 789}},
					NewTestTimeChange(nil, NewTestChange(nil, nil, nil)),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(123), "/change/agent", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(456), "/change/from", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(789), "/change/to", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *timechange.TimeChange), expectedErrors ...error) {
					datum := NewTimeChange()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *timechange.TimeChange) {},
				),
				Entry("type missing",
					func(datum *timechange.TimeChange) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "timeChange"}),
				),
				Entry("type invalid",
					func(datum *timechange.TimeChange) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "timeChange"}),
				),
				Entry("type device",
					func(datum *timechange.TimeChange) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *timechange.TimeChange) { datum.SubType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *timechange.TimeChange) { datum.SubType = "invalidSubType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "timeChange"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type timeChange",
					func(datum *timechange.TimeChange) { datum.SubType = "timeChange" },
				),
				Entry("change missing",
					func(datum *timechange.TimeChange) { datum.Change = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/change", NewMeta()),
				),
				Entry("change invalid",
					func(datum *timechange.TimeChange) { datum.Change.Agent = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/change/agent", NewMeta()),
				),
				Entry("change valid",
					func(datum *timechange.TimeChange) { datum.Change = NewChange() },
				),
				Entry("multiple errors",
					func(datum *timechange.TimeChange) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Change.Agent = nil
						datum.Change.From = nil
						datum.Change.To = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "timeChange"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/change/agent", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/change/from", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/change/to", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *timechange.TimeChange)) {
					for _, origin := range structure.Origins() {
						datum := NewTimeChange()
						mutator(datum)
						expectedDatum := CloneTimeChange(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *timechange.TimeChange) {},
				),
				Entry("does not modify the datum; change missing",
					func(datum *timechange.TimeChange) { datum.Change = nil },
				),
			)
		})
	})
})
