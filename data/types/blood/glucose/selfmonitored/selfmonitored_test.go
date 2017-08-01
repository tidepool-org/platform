package selfmonitored_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "smbg",
	}
}

func NewTestSelfMonitored(sourceTime interface{}, sourceUnits interface{}, sourceValue interface{}, sourceSubType interface{}) *selfmonitored.SelfMonitored {
	testSelfMonitored := selfmonitored.Init()
	testSelfMonitored.DeviceID = pointer.String(id.New())
	if value, ok := sourceTime.(string); ok {
		testSelfMonitored.Time = pointer.String(value)
	}
	if value, ok := sourceUnits.(string); ok {
		testSelfMonitored.Units = pointer.String(value)
	}
	if value, ok := sourceValue.(float64); ok {
		testSelfMonitored.Value = pointer.Float(value)
	}
	if value, ok := sourceSubType.(string); ok {
		testSelfMonitored.SubType = pointer.String(value)
	}
	return testSelfMonitored
}

var _ = Describe("SelfMonitored", func() {
	Context("Type", func() {
		It("returns the expected type", func() {
			Expect(selfmonitored.Type()).To(Equal("smbg"))
		})
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(selfmonitored.NewDatum()).To(Equal(&selfmonitored.SelfMonitored{}))
		})
	})

	Context("New", func() {
		It("returns the expected self monitored", func() {
			Expect(selfmonitored.New()).To(Equal(&selfmonitored.SelfMonitored{}))
		})
	})

	Context("Init", func() {
		It("returns the expected self monitored", func() {
			testSelfMonitored := selfmonitored.Init()
			Expect(testSelfMonitored).ToNot(BeNil())
			Expect(testSelfMonitored.ID).ToNot(BeEmpty())
			Expect(testSelfMonitored.Type).To(Equal("smbg"))
		})
	})

	Context("with new self monitored", func() {
		var testSelfMonitored *selfmonitored.SelfMonitored

		BeforeEach(func() {
			testSelfMonitored = selfmonitored.New()
			Expect(testSelfMonitored).ToNot(BeNil())
		})

		Context("Init", func() {
			It("initializes the self monitored", func() {
				testSelfMonitored.Init()
				Expect(testSelfMonitored.ID).ToNot(BeEmpty())
				Expect(testSelfMonitored.Type).To(Equal("smbg"))
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				testSelfMonitored.Init()
			})

			DescribeTable("Parse",
				func(sourceObject *map[string]interface{}, expectedSelfMonitored *selfmonitored.SelfMonitored, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(testSelfMonitored.Parse(testParser)).To(Succeed())
					Expect(testSelfMonitored.Time).To(Equal(expectedSelfMonitored.Time))
					Expect(testSelfMonitored.Units).To(Equal(expectedSelfMonitored.Units))
					Expect(testSelfMonitored.Value).To(Equal(expectedSelfMonitored.Value))
					Expect(testSelfMonitored.SubType).To(Equal(expectedSelfMonitored.SubType))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid units",
					&map[string]interface{}{"units": "mmol/L"},
					NewTestSelfMonitored(nil, "mmol/L", nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid units",
					&map[string]interface{}{"units": 0},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/units", NewMeta()),
					}),
				Entry("parses object that has valid value",
					&map[string]interface{}{"value": 10.0},
					NewTestSelfMonitored(nil, nil, 10.0, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid value",
					&map[string]interface{}{"value": "invalid"},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/value", NewMeta()),
					}),
				Entry("parses object that has valid sub type",
					&map[string]interface{}{"subType": "linked"},
					NewTestSelfMonitored(nil, nil, nil, "linked"),
					[]*service.Error{}),
				Entry("parses object that has invalid sub type",
					&map[string]interface{}{"subType": 0},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/subType", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "units": "mmol/L", "value": 10.0, "subType": "linked"},
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", 10.0, "linked"),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "units": 0, "value": "invalid", "subType": 0},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(0), "/units", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/value", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(0), "/subType", NewMeta()),
					}),
			)

			DescribeTable("Validate",
				func(sourceSelfMonitored *selfmonitored.SelfMonitored, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceSelfMonitored.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", 10.0, "linked"),
					[]*service.Error{}),
				Entry("missing time",
					NewTestSelfMonitored(nil, "mmol/L", 10.0, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
					}),
				Entry("missing units",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", nil, 10.0, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/units", NewMeta()),
					}),
				Entry("unknown units",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "unknown", 10.0, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
					}),
				Entry("mmol/L units",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", 10.0, "linked"),
					[]*service.Error{}),
				Entry("mmol/l units",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/l", 10.0, "linked"),
					[]*service.Error{}),
				Entry("mg/dL units",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mg/dL", 180.0, "linked"),
					[]*service.Error{}),
				Entry("mg/dl units",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mg/dl", 180.0, "linked"),
					[]*service.Error{}),
				Entry("missing value",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", nil, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/value", NewMeta()),
					}),
				Entry("unknown units; value in range (lower)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "unknown", -math.MaxFloat64, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
					}),
				Entry("unknown units; value in range (upper)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "unknown", math.MaxFloat64, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
					}),
				Entry("mmol/L units; value out of range (lower)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", -0.1, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
					}),
				Entry("mmol/L units; value in range (lower)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", 0.0, "linked"),
					[]*service.Error{}),
				Entry("mmol/L units; value in range (upper)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", 55.0, "linked"),
					[]*service.Error{}),
				Entry("mmol/L units; value out of range (upper)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", 55.1, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
					}),
				Entry("mmol/l units; value out of range (lower)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/l", -0.1, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
					}),
				Entry("mmol/l units; value in range (lower)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/l", 0.0, "linked"),
					[]*service.Error{}),
				Entry("mmol/l units; value in range (upper)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/l", 55.0, "linked"),
					[]*service.Error{}),
				Entry("mmol/l units; value out of range (upper)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/l", 55.1, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
					}),
				Entry("mg/dL units; value out of range (lower)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mg/dL", -0.1, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
					}),
				Entry("mg/dL units; value in range (lower)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mg/dL", 0.0, "linked"),
					[]*service.Error{}),
				Entry("mg/dL units; value in range (upper)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mg/dL", 1000.0, "linked"),
					[]*service.Error{}),
				Entry("mg/dL units; value out of range (upper)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mg/dL", 1000.1, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
					}),
				Entry("mg/dl units; value out of range (lower)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mg/dl", -0.1, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
					}),
				Entry("mg/dl units; value in range (lower)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mg/dl", 0.0, "linked"),
					[]*service.Error{}),
				Entry("mg/dl units; value in range (upper)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mg/dl", 1000.0, "linked"),
					[]*service.Error{}),
				Entry("mg/dl units; value out of range (upper)",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mg/dl", 1000.1, "linked"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
					}),
				Entry("unknown sub type",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", 10.0, "unknown"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"linked", "manual"}), "/subType", NewMeta()),
					}),
				Entry("linked sub type",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", 10.0, "linked"),
					[]*service.Error{}),
				Entry("manual sub type",
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", 10.0, "manual"),
					[]*service.Error{}),
				Entry("multiple",
					NewTestSelfMonitored(nil, "unknown", nil, "unknown"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
						testData.ComposeError(service.ErrorValueNotExists(), "/value", NewMeta()),
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"linked", "manual"}), "/subType", NewMeta()),
					}),
			)
		})
	})
})
