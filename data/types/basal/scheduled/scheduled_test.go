package scheduled_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	"github.com/tidepool-org/platform/data/types/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "scheduled",
	}
}

func NewTestScheduled(sourceTime interface{}, sourceDuration interface{}, sourceExpectedDuration interface{}, sourceRate interface{}, scheduleName interface{}) *scheduled.Scheduled {
	testScheduled := scheduled.Init()
	testScheduled.DeviceID = app.StringAsPointer(app.NewID())
	if value, ok := sourceTime.(string); ok {
		testScheduled.Time = app.StringAsPointer(value)
	}
	if value, ok := sourceDuration.(int); ok {
		testScheduled.Duration = app.IntegerAsPointer(value)
	}
	if value, ok := sourceExpectedDuration.(int); ok {
		testScheduled.ExpectedDuration = app.IntegerAsPointer(value)
	}
	if value, ok := sourceRate.(float64); ok {
		testScheduled.Rate = app.FloatAsPointer(value)
	}
	if value, ok := scheduleName.(string); ok {
		testScheduled.ScheduleName = app.StringAsPointer(value)
	}
	return testScheduled
}

var _ = Describe("Temporary", func() {
	Context("DeliveryType", func() {
		It("returns the expected type", func() {
			Expect(scheduled.DeliveryType()).To(Equal("scheduled"))
		})
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(scheduled.NewDatum()).To(Equal(&scheduled.Scheduled{}))
		})
	})

	Context("New", func() {
		It("returns the expected scheduled", func() {
			Expect(scheduled.New()).To(Equal(&scheduled.Scheduled{}))
		})
	})

	Context("Init", func() {
		It("returns the expected scheduled with all values initialized", func() {
			testScheduled := scheduled.Init()
			Expect(testScheduled).ToNot(BeNil())
			Expect(testScheduled.ID).ToNot(BeEmpty())
			Expect(testScheduled.Type).To(Equal("basal"))
			Expect(testScheduled.DeliveryType).To(Equal("scheduled"))
			Expect(testScheduled.Duration).To(BeNil())
			Expect(testScheduled.ExpectedDuration).To(BeNil())
			Expect(testScheduled.Rate).To(BeNil())
			Expect(testScheduled.ScheduleName).To(BeNil())
		})
	})

	Context("with new scheduled", func() {
		var testScheduled *scheduled.Scheduled

		BeforeEach(func() {
			testScheduled = scheduled.New()
			Expect(testScheduled).ToNot(BeNil())
		})

		Context("Init", func() {
			It("initializes the scheduled", func() {
				testScheduled.Init()
				Expect(testScheduled.ID).ToNot(BeEmpty())
				Expect(testScheduled.Type).To(Equal("basal"))
				Expect(testScheduled.DeliveryType).To(Equal("scheduled"))
				Expect(testScheduled.Duration).To(BeNil())
				Expect(testScheduled.ExpectedDuration).To(BeNil())
				Expect(testScheduled.Rate).To(BeNil())
				Expect(testScheduled.ScheduleName).To(BeNil())
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				testScheduled.Init()
			})

			DescribeTable("Parse",
				func(sourceObject *map[string]interface{}, expectedScheduled *scheduled.Scheduled, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(testScheduled.Parse(testParser)).To(Succeed())
					Expect(testScheduled.Time).To(Equal(expectedScheduled.Time))
					Expect(testScheduled.Duration).To(Equal(expectedScheduled.Duration))
					Expect(testScheduled.ExpectedDuration).To(Equal(expectedScheduled.ExpectedDuration))
					Expect(testScheduled.Rate).To(Equal(expectedScheduled.Rate))
					Expect(testScheduled.ScheduleName).To(Equal(expectedScheduled.ScheduleName))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestScheduled("2016-09-06T13:45:58-07:00", nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid duration",
					&map[string]interface{}{"duration": 3600000},
					NewTestScheduled(nil, 3600000, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration",
					&map[string]interface{}{"duration": "invalid"},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
					}),
				Entry("parses object that has valid expected duration",
					&map[string]interface{}{"expectedDuration": 7200000},
					NewTestScheduled(nil, nil, 7200000, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid expected duration",
					&map[string]interface{}{"expectedDuration": "invalid"},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
					}),
				Entry("parses object that has valid rate",
					&map[string]interface{}{"rate": 1.0},
					NewTestScheduled(nil, nil, nil, 1.0, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid rate",
					&map[string]interface{}{"rate": "invalid"},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", NewMeta()),
					}),
				Entry("parses object that has valid schedule name",
					&map[string]interface{}{"scheduleName": "Weekday"},
					NewTestScheduled(nil, nil, nil, nil, "Weekday"),
					[]*service.Error{}),
				Entry("parses object that has invalid schedule name",
					&map[string]interface{}{"scheduleName": 0},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "duration": 3600000, "expectedDuration": 7200000, "rate": 1.0, "scheduleName": "Weekday"},
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 7200000, 1.0, "Weekday"),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "duration": "invalid", "expectedDuration": "invalid", "rate": "invalid", "scheduleName": 0, "suppressed": "invalid"},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testing.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
						testing.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
						testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", NewMeta()),
						testing.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", NewMeta()),
					}),
			)

			DescribeTable("Validate",
				func(sourceScheduled *scheduled.Scheduled, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceScheduled.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 7200000, 1.0, "Weekday"),
					[]*service.Error{}),
				Entry("missing time",
					NewTestScheduled(nil, 3600000, 7200000, 1.0, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
					}),
				Entry("missing duration",
					NewTestScheduled("2016-09-06T13:45:58-07:00", nil, 7200000, 1.0, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
					}),
				Entry("duration out of range (lower)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", -1, 7200000, 1.0, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					}),
				Entry("duration in range (lower)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 0, 7200000, 1.0, "Weekday"),
					[]*service.Error{}),
				Entry("duration in range (upper)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 604800000, nil, 1.0, "Weekday"),
					[]*service.Error{}),
				Entry("duration out of range (upper)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 604800001, nil, 1.0, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					}),
				Entry("expected duration out of range (lower)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 3599999, 1.0, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotInRange(3599999, 3600000, 604800000), "/expectedDuration", NewMeta()),
					}),
				Entry("expected duration in range (lower)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 3600000, 1.0, "Weekday"),
					[]*service.Error{}),
				Entry("expected duration in range (upper)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 604800000, 1.0, "Weekday"),
					[]*service.Error{}),
				Entry("expected duration out of range (upper)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 604800001, 1.0, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotInRange(604800001, 3600000, 604800000), "/expectedDuration", NewMeta()),
					}),
				Entry("missing duration; expected duration out of range (lower)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", nil, -1, 1.0, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
						testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
					}),
				Entry("missing duration; expected duration in range (lower)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", nil, 0, 1.0, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
					}),
				Entry("missing duration; expected duration in range (upper)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", nil, 604800000, 1.0, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
					}),
				Entry("missing duration; expected duration out of range (upper)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", nil, 604800001, 1.0, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
						testing.ComposeError(service.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
					}),
				Entry("missing rate",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 7200000, nil, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/rate", NewMeta()),
					}),
				Entry("rate out of range (lower)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 7200000, -0.1, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", NewMeta()),
					}),
				Entry("rate in range (lower)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 7200000, 0.0, "Weekday"),
					[]*service.Error{}),
				Entry("rate in range (upper)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 7200000, 100.0, "Weekday"),
					[]*service.Error{}),
				Entry("rate out of range (upper)",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 7200000, 100.1, "Weekday"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", NewMeta()),
					}),
				Entry("schedule name empty",
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 7200000, 1.0, ""),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueEmpty(), "/scheduleName", NewMeta()),
					}),
				Entry("multiple",
					NewTestScheduled(nil, nil, 604800001, 100.1, ""),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
						testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
						testing.ComposeError(service.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
						testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", NewMeta()),
						testing.ComposeError(service.ErrorValueEmpty(), "/scheduleName", NewMeta()),
					}),
			)

			Context("Normalize", func() {
				It("succeeds", func() {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testNormalizer, err := normalizer.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testNormalizer).ToNot(BeNil())
					Expect(testScheduled.Normalize(testNormalizer)).To(Succeed())
				})
			})
		})
	})
})
