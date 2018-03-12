package scheduled_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	testDataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	testDataTypesBasal "github.com/tidepool-org/platform/data/types/basal/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "scheduled",
	}
}

func NewScheduled() *scheduled.Scheduled {
	datum := scheduled.New()
	datum.Basal = *testDataTypesBasal.NewBasal()
	datum.DeliveryType = "scheduled"
	datum.Duration = pointer.Int(test.RandomIntFromRange(scheduled.DurationMinimum, scheduled.DurationMaximum))
	datum.DurationExpected = pointer.Int(test.RandomIntFromRange(*datum.Duration, scheduled.DurationMaximum))
	datum.Rate = pointer.Float64(test.RandomFloat64FromRange(scheduled.RateMinimum, scheduled.RateMaximum))
	datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
	return datum
}

func CloneScheduled(datum *scheduled.Scheduled) *scheduled.Scheduled {
	if datum == nil {
		return nil
	}
	clone := scheduled.New()
	clone.Basal = *testDataTypesBasal.CloneBasal(&datum.Basal)
	clone.Duration = test.CloneInt(datum.Duration)
	clone.DurationExpected = test.CloneInt(datum.DurationExpected)
	clone.Rate = test.CloneFloat64(datum.Rate)
	clone.ScheduleName = test.CloneString(datum.ScheduleName)
	return clone
}

func NewTestScheduled(sourceTime interface{}, sourceDuration interface{}, sourceDurationExpected interface{}, sourceRate interface{}, scheduleName interface{}) *scheduled.Scheduled {
	datum := scheduled.Init()
	datum.DeviceID = pointer.String(id.New())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceDuration.(int); ok {
		datum.Duration = &val
	}
	if val, ok := sourceDurationExpected.(int); ok {
		datum.DurationExpected = &val
	}
	if val, ok := sourceRate.(float64); ok {
		datum.Rate = &val
	}
	if val, ok := scheduleName.(string); ok {
		datum.ScheduleName = &val
	}
	return datum
}

var _ = Describe("Scheduled", func() {
	It("DeliveryType is expected", func() {
		Expect(scheduled.DeliveryType).To(Equal("scheduled"))
	})

	It("DurationMaximum is expected", func() {
		Expect(scheduled.DurationMaximum).To(Equal(604800000))
	})

	It("DurationMinimum is expected", func() {
		Expect(scheduled.DurationMinimum).To(Equal(0))
	})

	It("RateMaximum is expected", func() {
		Expect(scheduled.RateMaximum).To(Equal(100.0))
	})

	It("RateMinimum is expected", func() {
		Expect(scheduled.RateMinimum).To(Equal(0.0))
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(scheduled.NewDatum()).To(Equal(&scheduled.Scheduled{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(scheduled.New()).To(Equal(&scheduled.Scheduled{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := scheduled.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("basal"))
			Expect(datum.DeliveryType).To(Equal("scheduled"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.Rate).To(BeNil())
			Expect(datum.ScheduleName).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *scheduled.Scheduled

		BeforeEach(func() {
			datum = NewScheduled()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("basal"))
				Expect(datum.DeliveryType).To(Equal("scheduled"))
				Expect(datum.Duration).To(BeNil())
				Expect(datum.DurationExpected).To(BeNil())
				Expect(datum.Rate).To(BeNil())
				Expect(datum.ScheduleName).To(BeNil())
			})
		})
	})

	Context("Scheduled", func() {
		Context("Parse", func() {
			var datum *scheduled.Scheduled

			BeforeEach(func() {
				datum = scheduled.Init()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *scheduled.Scheduled, expectedErrors []*service.Error) {
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
					Expect(datum.Rate).To(Equal(expectedDatum.Rate))
					Expect(datum.ScheduleName).To(Equal(expectedDatum.ScheduleName))
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
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid duration",
					&map[string]interface{}{"duration": 3600000},
					NewTestScheduled(nil, 3600000, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration",
					&map[string]interface{}{"duration": "invalid"},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
					}),
				Entry("parses object that has valid duration expected",
					&map[string]interface{}{"expectedDuration": 7200000},
					NewTestScheduled(nil, nil, 7200000, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration expected",
					&map[string]interface{}{"expectedDuration": "invalid"},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
					}),
				Entry("parses object that has valid rate",
					&map[string]interface{}{"rate": 1.0},
					NewTestScheduled(nil, nil, nil, 1.0, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid rate",
					&map[string]interface{}{"rate": "invalid"},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", NewMeta()),
					}),
				Entry("parses object that has valid schedule name",
					&map[string]interface{}{"scheduleName": "Weekday"},
					NewTestScheduled(nil, nil, nil, nil, "Weekday"),
					[]*service.Error{}),
				Entry("parses object that has invalid schedule name",
					&map[string]interface{}{"scheduleName": 0},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "duration": 3600000, "expectedDuration": 7200000, "rate": 1.0, "scheduleName": "Weekday"},
					NewTestScheduled("2016-09-06T13:45:58-07:00", 3600000, 7200000, 1.0, "Weekday"),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "duration": "invalid", "expectedDuration": "invalid", "rate": "invalid", "scheduleName": 0, "suppressed": "invalid"},
					NewTestScheduled(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *scheduled.Scheduled), expectedErrors ...error) {
					datum := NewScheduled()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *scheduled.Scheduled) {},
				),
				Entry("type missing",
					func(datum *scheduled.Scheduled) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &basal.Meta{DeliveryType: "scheduled"}),
				),
				Entry("type invalid",
					func(datum *scheduled.Scheduled) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "scheduled"}),
				),
				Entry("type basal",
					func(datum *scheduled.Scheduled) { datum.Type = "basal" },
				),
				Entry("delivery type missing",
					func(datum *scheduled.Scheduled) { datum.DeliveryType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deliveryType", &basal.Meta{Type: "basal"}),
				),
				Entry("delivery type invalid",
					func(datum *scheduled.Scheduled) { datum.DeliveryType = "invalidDeliveryType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "scheduled"), "/deliveryType", &basal.Meta{Type: "basal", DeliveryType: "invalidDeliveryType"}),
				),
				Entry("delivery type scheduled",
					func(datum *scheduled.Scheduled) { datum.DeliveryType = "scheduled" },
				),
				Entry("duration missing; duration expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604799999)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("rate missing",
					func(datum *scheduled.Scheduled) { datum.Rate = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rate", NewMeta()),
				),
				Entry("rate out of range (lower)",
					func(datum *scheduled.Scheduled) { datum.Rate = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("rate in range (lower)",
					func(datum *scheduled.Scheduled) { datum.Rate = pointer.Float64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *scheduled.Scheduled) { datum.Rate = pointer.Float64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *scheduled.Scheduled) { datum.Rate = pointer.Float64(100.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("schedule name empty",
					func(datum *scheduled.Scheduled) { datum.ScheduleName = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/scheduleName", NewMeta()),
				),
				Entry("schedule name valid",
					func(datum *scheduled.Scheduled) {
						datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
					},
				),
				Entry("multiple errors",
					func(datum *scheduled.Scheduled) {
						datum.Type = "invalidType"
						datum.DeliveryType = "invalidDeliveryType"
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800001)
						datum.Rate = pointer.Float64(100.1)
						datum.ScheduleName = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "scheduled"), "/deliveryType", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/scheduleName", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *scheduled.Scheduled)) {
					for _, origin := range structure.Origins() {
						datum := NewScheduled()
						mutator(datum)
						expectedDatum := CloneScheduled(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *scheduled.Scheduled) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *scheduled.Scheduled) { datum.Type = "" },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *scheduled.Scheduled) { datum.DeliveryType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *scheduled.Scheduled) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *scheduled.Scheduled) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; rate missing",
					func(datum *scheduled.Scheduled) { datum.Rate = nil },
				),
				Entry("does not modify the datum; schedule name missing",
					func(datum *scheduled.Scheduled) { datum.ScheduleName = nil },
				),
			)
		})
	})

	Context("ParseSuppressedScheduled", func() {
		// TODO
	})

	Context("NewSuppressedScheduled", func() {
		It("returns the expected datum", func() {
			Expect(scheduled.NewSuppressedScheduled()).To(Equal(&scheduled.SuppressedScheduled{
				Type:         pointer.String("basal"),
				DeliveryType: pointer.String("scheduled"),
			}))
		})
	})

	Context("SuppressedScheduled", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *scheduled.SuppressedScheduled), expectedErrors ...error) {
					datum := testDataTypesBasalScheduled.NewSuppressedScheduled()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *scheduled.SuppressedScheduled) {},
				),
				Entry("type missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Type = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *scheduled.SuppressedScheduled) { datum.Type = pointer.String("invalidType") },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type"),
				),
				Entry("type basal",
					func(datum *scheduled.SuppressedScheduled) { datum.Type = pointer.String("basal") },
				),
				Entry("delivery type missing",
					func(datum *scheduled.SuppressedScheduled) { datum.DeliveryType = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deliveryType"),
				),
				Entry("delivery type invalid",
					func(datum *scheduled.SuppressedScheduled) { datum.DeliveryType = pointer.String("invalidDeliveryType") },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "scheduled"), "/deliveryType"),
				),
				Entry("delivery type scheduled",
					func(datum *scheduled.SuppressedScheduled) { datum.DeliveryType = pointer.String("scheduled") },
				),
				Entry("annotations missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Annotations = nil },
				),
				Entry("annotations valid",
					func(datum *scheduled.SuppressedScheduled) { datum.Annotations = testData.NewBlobArray() },
				),
				Entry("rate missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate"),
				),
				Entry("rate in range (lower)",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = pointer.Float64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = pointer.Float64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = pointer.Float64(100.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
				),
				Entry("schedule name empty",
					func(datum *scheduled.SuppressedScheduled) { datum.ScheduleName = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scheduleName"),
				),
				Entry("schedule name valid",
					func(datum *scheduled.SuppressedScheduled) {
						datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
					},
				),
				Entry("multiple errors",
					func(datum *scheduled.SuppressedScheduled) {
						datum.Type = pointer.String("invalidType")
						datum.DeliveryType = pointer.String("invalidDeliveryType")
						datum.Rate = pointer.Float64(100.1)
						datum.ScheduleName = pointer.String("")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "scheduled"), "/deliveryType"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scheduleName"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *scheduled.SuppressedScheduled)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesBasalScheduled.NewSuppressedScheduled()
						mutator(datum)
						expectedDatum := testDataTypesBasalScheduled.CloneSuppressedScheduled(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *scheduled.SuppressedScheduled) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Type = nil },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *scheduled.SuppressedScheduled) { datum.DeliveryType = nil },
				),
				Entry("does not modify the datum; annotations missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Annotations = nil },
				),
				Entry("does not modify the datum; rate missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = nil },
				),
				Entry("does not modify the datum; schedule name missing",
					func(datum *scheduled.SuppressedScheduled) { datum.ScheduleName = nil },
				),
			)
		})
	})
})
