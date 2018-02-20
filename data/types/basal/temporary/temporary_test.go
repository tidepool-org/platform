package temporary_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/temporary"
	basalTest "github.com/tidepool-org/platform/data/types/basal/test"
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
		DeliveryType: "temp",
	}
}

func NewTemporary() *temporary.Temporary {
	datum := temporary.New()
	datum.Basal = *testDataTypesBasal.NewBasal()
	datum.DeliveryType = "temp"
	datum.Duration = pointer.Int(test.RandomIntFromRange(temporary.DurationMinimum, temporary.DurationMaximum))
	datum.DurationExpected = pointer.Int(test.RandomIntFromRange(*datum.Duration, temporary.DurationMaximum))
	datum.Percent = pointer.Float64(test.RandomFloat64FromRange(temporary.PercentMinimum, temporary.PercentMaximum))
	datum.Rate = pointer.Float64(test.RandomFloat64FromRange(temporary.RateMinimum, temporary.RateMaximum))
	datum.Suppressed = testDataTypesBasal.NewSuppressedScheduled()
	return datum
}

func CloneTemporary(datum *temporary.Temporary) *temporary.Temporary {
	if datum == nil {
		return nil
	}
	clone := temporary.New()
	clone.Basal = *testDataTypesBasal.CloneBasal(&datum.Basal)
	clone.Duration = test.CloneInt(datum.Duration)
	clone.DurationExpected = test.CloneInt(datum.DurationExpected)
	clone.Percent = test.CloneFloat64(datum.Percent)
	clone.Rate = test.CloneFloat64(datum.Rate)
	clone.Suppressed = testDataTypesBasal.CloneSuppressed(datum.Suppressed)
	return clone
}

func NewTestTemporary(sourceTime interface{}, sourceDuration interface{}, sourceDurationExpected interface{}, sourceRate interface{}, sourcePercent interface{}, sourceSuppressed *basal.Suppressed) *temporary.Temporary {
	datum := temporary.Init()
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
	if val, ok := sourcePercent.(float64); ok {
		datum.Percent = &val
	}
	datum.Suppressed = sourceSuppressed
	return datum
}

var _ = Describe("Temporary", func() {
	It("DurationMaximum is expected", func() {
		Expect(temporary.DurationMaximum).To(Equal(604800000))
	})

	It("DurationMinimum is expected", func() {
		Expect(temporary.DurationMinimum).To(Equal(0))
	})

	It("PercentMaximum is expected", func() {
		Expect(temporary.PercentMaximum).To(Equal(10.0))
	})

	It("PercentMinimum is expected", func() {
		Expect(temporary.PercentMinimum).To(Equal(0.0))
	})

	It("RateMaximum is expected", func() {
		Expect(temporary.RateMaximum).To(Equal(100.0))
	})

	It("RateMinimum is expected", func() {
		Expect(temporary.RateMinimum).To(Equal(0.0))
	})

	It("SuppressedDeliveryTypes returns expected", func() {
		Expect(temporary.SuppressedDeliveryTypes()).To(Equal([]string{"scheduled"}))
	})

	Context("DeliveryType", func() {
		It("returns the expected delivery type", func() {
			Expect(temporary.DeliveryType()).To(Equal("temp"))
		})
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(temporary.NewDatum()).To(Equal(&temporary.Temporary{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(temporary.New()).To(Equal(&temporary.Temporary{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := temporary.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("basal"))
			Expect(datum.DeliveryType).To(Equal("temp"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.Percent).To(BeNil())
			Expect(datum.Rate).To(BeNil())
			Expect(datum.Suppressed).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *temporary.Temporary

		BeforeEach(func() {
			datum = NewTemporary()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("basal"))
				Expect(datum.DeliveryType).To(Equal("temp"))
				Expect(datum.Duration).To(BeNil())
				Expect(datum.DurationExpected).To(BeNil())
				Expect(datum.Percent).To(BeNil())
				Expect(datum.Rate).To(BeNil())
				Expect(datum.Suppressed).To(BeNil())
			})
		})
	})

	Context("Temporary", func() {
		Context("Parse", func() {
			var datum *temporary.Temporary

			BeforeEach(func() {
				datum = temporary.Init()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *temporary.Temporary, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.Duration).To(Equal(expectedDatum.Duration))
					Expect(datum.DurationExpected).To(Equal(expectedDatum.DurationExpected))
					Expect(datum.Rate).To(Equal(expectedDatum.Rate))
					Expect(datum.Percent).To(Equal(expectedDatum.Percent))
					Expect(datum.Suppressed).To(Equal(expectedDatum.Suppressed))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestTemporary(nil, nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestTemporary(nil, nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestTemporary("2016-09-06T13:45:58-07:00", nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestTemporary(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid duration",
					&map[string]interface{}{"duration": 3600000},
					NewTestTemporary(nil, 3600000, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration",
					&map[string]interface{}{"duration": "invalid"},
					NewTestTemporary(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
					}),
				Entry("parses object that has valid duration expected",
					&map[string]interface{}{"expectedDuration": 7200000},
					NewTestTemporary(nil, nil, 7200000, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration expected",
					&map[string]interface{}{"expectedDuration": "invalid"},
					NewTestTemporary(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
					}),
				Entry("parses object that has valid rate",
					&map[string]interface{}{"rate": 2.0},
					NewTestTemporary(nil, nil, nil, 2.0, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid rate",
					&map[string]interface{}{"rate": "invalid"},
					NewTestTemporary(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", NewMeta()),
					}),
				Entry("parses object that has valid percent",
					&map[string]interface{}{"percent": 0.5},
					NewTestTemporary(nil, nil, nil, nil, 0.5, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid percent",
					&map[string]interface{}{"percent": "invalid"},
					NewTestTemporary(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/percent", NewMeta()),
					}),
				Entry("parses object that has valid suppressed",
					&map[string]interface{}{"suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 2.0, "scheduleName": "Weekday"}},
					NewTestTemporary(nil, nil, nil, nil, nil, basalTest.NewTestSuppressed("basal", "scheduled", nil, 2.0, "Weekday", nil)),
					[]*service.Error{}),
				Entry("parses object that has invalid suppressed",
					&map[string]interface{}{"suppressed": "invalid"},
					NewTestTemporary(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotObject("invalid"), "/suppressed", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "duration": 3600000, "expectedDuration": 7200000, "rate": 2.0, "percent": 0.5},
					NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, 2.0, 0.5, nil),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "duration": "invalid", "expectedDuration": "invalid", "rate": "invalid", "percent": "invalid", "suppressed": "invalid"},
					NewTestTemporary(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/percent", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotObject("invalid"), "/suppressed", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *temporary.Temporary), expectedErrors ...error) {
					datum := NewTemporary()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *temporary.Temporary) {},
				),
				Entry("type missing",
					func(datum *temporary.Temporary) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &basal.Meta{DeliveryType: "temp"}),
				),
				Entry("type invalid",
					func(datum *temporary.Temporary) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "temp"}),
				),
				Entry("type basal",
					func(datum *temporary.Temporary) { datum.Type = "basal" },
				),
				Entry("delivery type missing",
					func(datum *temporary.Temporary) { datum.DeliveryType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deliveryType", &basal.Meta{Type: "basal"}),
				),
				Entry("delivery type invalid",
					func(datum *temporary.Temporary) { datum.DeliveryType = "invalidDeliveryType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "temp"), "/deliveryType", &basal.Meta{Type: "basal", DeliveryType: "invalidDeliveryType"}),
				),
				Entry("delivery type temp",
					func(datum *temporary.Temporary) { datum.DeliveryType = "temp" },
				),
				Entry("duration missing; duration expected missing",
					func(datum *temporary.Temporary) {
						datum.Duration = nil
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604799999)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("percent missing",
					func(datum *temporary.Temporary) { datum.Percent = nil },
				),
				Entry("percent out of range (lower)",
					func(datum *temporary.Temporary) { datum.Percent = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/percent", NewMeta()),
				),
				Entry("percent in range (lower)",
					func(datum *temporary.Temporary) { datum.Percent = pointer.Float64(0.0) },
				),
				Entry("percent in range (upper)",
					func(datum *temporary.Temporary) { datum.Percent = pointer.Float64(10.0) },
				),
				Entry("percent out of range (upper)",
					func(datum *temporary.Temporary) { datum.Percent = pointer.Float64(10.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/percent", NewMeta()),
				),
				Entry("rate missing",
					func(datum *temporary.Temporary) { datum.Rate = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rate", NewMeta()),
				),
				Entry("rate out of range (lower)",
					func(datum *temporary.Temporary) { datum.Rate = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("rate in range (lower)",
					func(datum *temporary.Temporary) { datum.Rate = pointer.Float64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *temporary.Temporary) { datum.Rate = pointer.Float64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *temporary.Temporary) { datum.Rate = pointer.Float64(100.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("suppressed missing",
					func(datum *temporary.Temporary) { datum.Suppressed = nil },
				),
				Entry("suppressed scheduled with suppressed missing",
					func(datum *temporary.Temporary) {
						datum.Suppressed = testDataTypesBasal.NewSuppressedScheduled()
						datum.Suppressed.Suppressed = nil
					},
				),
				Entry("suppressed scheduled with suppressed exists",
					func(datum *temporary.Temporary) {
						datum.Suppressed = testDataTypesBasal.NewSuppressedScheduled()
						datum.Suppressed.Suppressed = testDataTypesBasal.NewSuppressedScheduled()
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/suppressed/suppressed", NewMeta()),
				),
				Entry("suppressed suspend",
					func(datum *temporary.Temporary) {
						datum.Suppressed = testDataTypesBasal.NewSuppressedSuspend()
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("suspend", []string{"scheduled"}), "/suppressed/deliveryType", NewMeta()),
				),
				Entry("suppressed temporary",
					func(datum *temporary.Temporary) { datum.Suppressed = testDataTypesBasal.NewSuppressedTemporary() },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("temp", []string{"scheduled"}), "/suppressed/deliveryType", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *temporary.Temporary) {
						datum.Type = "invalidType"
						datum.DeliveryType = "invalidDeliveryType"
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800001)
						datum.Percent = pointer.Float64(10.1)
						datum.Rate = pointer.Float64(100.1)
						datum.Suppressed = testDataTypesBasal.NewSuppressedSuspend()
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "temp"), "/deliveryType", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/percent", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("suspend", []string{"scheduled"}), "/suppressed/deliveryType", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *temporary.Temporary)) {
					for _, origin := range structure.Origins() {
						datum := NewTemporary()
						mutator(datum)
						expectedDatum := CloneTemporary(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *temporary.Temporary) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *temporary.Temporary) { datum.Type = "" },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *temporary.Temporary) { datum.DeliveryType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *temporary.Temporary) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *temporary.Temporary) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; percent missing",
					func(datum *temporary.Temporary) { datum.Percent = nil },
				),
				Entry("does not modify the datum; reate missing",
					func(datum *temporary.Temporary) { datum.Rate = nil },
				),
				Entry("does not modify the datum; suppressed missing",
					func(datum *temporary.Temporary) { datum.Suppressed = nil },
				),
			)
		})
	})
})
