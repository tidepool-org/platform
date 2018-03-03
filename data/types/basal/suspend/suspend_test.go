package suspend_test

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
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	testDataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	"github.com/tidepool-org/platform/data/types/basal/suspend"
	dataTypesBasalTemporary "github.com/tidepool-org/platform/data/types/basal/temporary"
	testDataTypesBasalTemporary "github.com/tidepool-org/platform/data/types/basal/temporary/test"
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
		DeliveryType: "suspend",
	}
}

func NewSuspend() *suspend.Suspend {
	datum := suspend.New()
	datum.Basal = *testDataTypesBasal.NewBasal()
	datum.DeliveryType = "suspend"
	datum.Duration = pointer.Int(test.RandomIntFromRange(suspend.DurationMinimum, suspend.DurationMaximum))
	datum.DurationExpected = pointer.Int(test.RandomIntFromRange(*datum.Duration, suspend.DurationMaximum))
	datum.Suppressed = testDataTypesBasalTemporary.NewSuppressedTemporary(testDataTypesBasalScheduled.NewSuppressedScheduled())
	return datum
}

func CloneSuspend(datum *suspend.Suspend) *suspend.Suspend {
	if datum == nil {
		return nil
	}
	clone := suspend.New()
	clone.Basal = *testDataTypesBasal.CloneBasal(&datum.Basal)
	clone.Duration = test.CloneInt(datum.Duration)
	clone.DurationExpected = test.CloneInt(datum.DurationExpected)
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalScheduled.SuppressedScheduled:
			clone.Suppressed = testDataTypesBasalScheduled.CloneSuppressedScheduled(suppressed)
		case *dataTypesBasalTemporary.SuppressedTemporary:
			clone.Suppressed = testDataTypesBasalTemporary.CloneSuppressedTemporary(suppressed)
		}
	}
	return clone
}

func NewTestSuspend(sourceTime interface{}, sourceDuration interface{}, sourceDurationExpected interface{}, sourceSuppressed suspend.Suppressed) *suspend.Suspend {
	datum := suspend.Init()
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
	datum.Suppressed = sourceSuppressed
	return datum
}

var _ = Describe("Suspend", func() {
	It("DurationMaximum is expected", func() {
		Expect(suspend.DurationMaximum).To(Equal(604800000))
	})

	It("DurationMinimum is expected", func() {
		Expect(suspend.DurationMinimum).To(Equal(0))
	})

	Context("DeliveryType", func() {
		It("returns the expected delivery type", func() {
			Expect(suspend.DeliveryType()).To(Equal("suspend"))
		})
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(suspend.NewDatum()).To(Equal(&suspend.Suspend{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(suspend.New()).To(Equal(&suspend.Suspend{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := suspend.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("basal"))
			Expect(datum.DeliveryType).To(Equal("suspend"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.Suppressed).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *suspend.Suspend

		BeforeEach(func() {
			datum = NewSuspend()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("basal"))
				Expect(datum.DeliveryType).To(Equal("suspend"))
				Expect(datum.Duration).To(BeNil())
				Expect(datum.DurationExpected).To(BeNil())
				Expect(datum.Suppressed).To(BeNil())
			})
		})
	})

	Context("Suspend", func() {
		Context("Parse", func() {
			var datum *suspend.Suspend

			BeforeEach(func() {
				datum = suspend.Init()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *suspend.Suspend, expectedErrors []*service.Error) {
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
					if expectedDatum.Suppressed != nil {
						Expect(datum.Suppressed).To(Equal(expectedDatum.Suppressed))
					} else {
						Expect(datum.Suppressed).To(BeNil())
					}
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestSuspend("2016-09-06T13:45:58-07:00", nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid duration",
					&map[string]interface{}{"duration": 3600000},
					NewTestSuspend(nil, 3600000, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration",
					&map[string]interface{}{"duration": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
					}),
				Entry("parses object that has valid duration expected",
					&map[string]interface{}{"expectedDuration": 7200000},
					NewTestSuspend(nil, nil, 7200000, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration expected",
					&map[string]interface{}{"expectedDuration": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
					}),
				Entry("parses object that has valid suppressed",
					&map[string]interface{}{"suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday"}},
					NewTestSuspend(nil, nil, nil, &dataTypesBasalScheduled.SuppressedScheduled{Type: pointer.String("basal"), DeliveryType: pointer.String("scheduled"), Rate: pointer.Float64(1.0), ScheduleName: pointer.String("Weekday")}),
					[]*service.Error{}),
				Entry("parses object that has invalid suppressed",
					&map[string]interface{}{"suppressed": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotObject("invalid"), "/suppressed", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "duration": 3600000, "expectedDuration": 7200000},
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 7200000, nil),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "duration": "invalid", "expectedDuration": "invalid", "suppressed": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotObject("invalid"), "/suppressed", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *suspend.Suspend), expectedErrors ...error) {
					datum := NewSuspend()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *suspend.Suspend) {},
				),
				Entry("type missing",
					func(datum *suspend.Suspend) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &basal.Meta{DeliveryType: "suspend"}),
				),
				Entry("type invalid",
					func(datum *suspend.Suspend) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "suspend"}),
				),
				Entry("type basal",
					func(datum *suspend.Suspend) { datum.Type = "basal" },
				),
				Entry("delivery type missing",
					func(datum *suspend.Suspend) { datum.DeliveryType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deliveryType", &basal.Meta{Type: "basal"}),
				),
				Entry("delivery type invalid",
					func(datum *suspend.Suspend) { datum.DeliveryType = "invalidDeliveryType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "suspend"), "/deliveryType", &basal.Meta{Type: "basal", DeliveryType: "invalidDeliveryType"}),
				),
				Entry("delivery type suspend",
					func(datum *suspend.Suspend) { datum.DeliveryType = "suspend" },
				),
				Entry("duration missing; duration expected missing",
					func(datum *suspend.Suspend) {
						datum.Duration = nil
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604799999)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("suppressed missing",
					func(datum *suspend.Suspend) { datum.Suppressed = nil },
				),
				Entry("suppressed scheduled",
					func(datum *suspend.Suspend) {
						datum.Suppressed = testDataTypesBasalScheduled.NewSuppressedScheduled()
					},
				),
				Entry("suppressed temporary with suppressed missing",
					func(datum *suspend.Suspend) {
						datum.Suppressed = testDataTypesBasalTemporary.NewSuppressedTemporary(nil)
					},
				),
				Entry("suppressed temporary with suppressed scheduled",
					func(datum *suspend.Suspend) {
						datum.Suppressed = testDataTypesBasalTemporary.NewSuppressedTemporary(testDataTypesBasalScheduled.NewSuppressedScheduled())
					},
				),
				Entry("suppressed temporary with suppressed temporary with suppressed missing",
					func(datum *suspend.Suspend) {
						datum.Suppressed = testDataTypesBasalTemporary.NewSuppressedTemporary(testDataTypesBasalTemporary.NewSuppressedTemporary(nil))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/suppressed/suppressed", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *suspend.Suspend) {
						datum.Type = "invalidType"
						datum.DeliveryType = "invalidDeliveryType"
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800001)
						datum.Suppressed = testDataTypesBasalTemporary.NewSuppressedTemporary(testDataTypesBasalTemporary.NewSuppressedTemporary(nil))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "suspend"), "/deliveryType", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/suppressed/suppressed", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *suspend.Suspend)) {
					for _, origin := range structure.Origins() {
						datum := NewSuspend()
						mutator(datum)
						expectedDatum := CloneSuspend(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *suspend.Suspend) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *suspend.Suspend) { datum.Type = "" },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *suspend.Suspend) { datum.DeliveryType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *suspend.Suspend) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *suspend.Suspend) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; suppressed missing",
					func(datum *suspend.Suspend) { datum.Suppressed = nil },
				),
			)
		})
	})
})
